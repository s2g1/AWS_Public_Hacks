package compliance

import (
	"strconv"
	"strings"
	"time"
	"unicode"

	"federal-payment-processing/internal/models"
)

// OFACMatchThreshold defines the minimum similarity score to consider a payee
// as matching an OFAC sanctions list entry.
const OFACMatchThreshold = 0.85

// SanctionsList is an interface for the OFAC sanctions data source,
// enabling mocking in tests.
type SanctionsList interface {
	// GetEntries returns all entries in the sanctions list.
	GetEntries() []string
}

// ComplianceHandler performs compliance checks on payment extraction results.
type ComplianceHandler struct {
	sanctionsList SanctionsList
	debarmentList DebarmentList
	spendingStore SpendingStore
	thresholds    SpendingThresholds
}

// NewComplianceHandler creates a new ComplianceHandler with the given sanctions
// and debarment lists.
func NewComplianceHandler(sanctionsList SanctionsList, debarmentList DebarmentList) *ComplianceHandler {
	return &ComplianceHandler{
		sanctionsList: sanctionsList,
		debarmentList: debarmentList,
	}
}

// NewComplianceHandlerWithSpending creates a ComplianceHandler with spending
// threshold checks enabled.
func NewComplianceHandlerWithSpending(sanctionsList SanctionsList, debarmentList DebarmentList, store SpendingStore, thresholds SpendingThresholds) *ComplianceHandler {
	return &ComplianceHandler{
		sanctionsList: sanctionsList,
		debarmentList: debarmentList,
		spendingStore: store,
		thresholds:    thresholds,
	}
}

// DebarmentMatchThreshold defines the minimum similarity score to consider a payee
// as matching a federal debarment list entry.
const DebarmentMatchThreshold = 0.85

// CheckCompliance evaluates the extraction result against OFAC sanctions and federal
// debarment list. On OFAC or debarment match: returns NON_COMPLIANT immediately with
// a BLOCKING flag.
func (h *ComplianceHandler) CheckCompliance(extraction *models.ExtractionResult) *models.ComplianceResult {
	result := &models.ComplianceResult{
		Status:    models.ComplianceStatusCompliant,
		Flags:     []models.ComplianceFlag{},
		Rules:     []string{"OFAC_SANCTIONS", "DEBARMENT"},
		CheckedAt: time.Now(),
	}

	// Extract payee name from the extraction result.
	payeeField, ok := extraction.Fields["payee"]
	if !ok {
		// No payee field — cannot screen, return compliant for this check.
		return result
	}

	payeeName := payeeField.Normalized
	if payeeName == "" {
		payeeName = payeeField.Value
	}

	normalizedPayee := normalizeName(payeeName)

	// Screen against OFAC sanctions list.
	entries := h.sanctionsList.GetEntries()
	for _, entry := range entries {
		similarity := JaroWinklerSimilarity(normalizedPayee, normalizeName(entry))
		if similarity >= OFACMatchThreshold {
			result.Status = models.ComplianceStatusNonCompliant
			result.Flags = append(result.Flags, models.ComplianceFlag{
				Rule:     "OFAC_SANCTIONS",
				Severity: models.FlagSeverityBlocking,
				Message:  "Payee matches OFAC sanctions list: " + entry,
			})
			// Fail-fast: return immediately on OFAC match.
			return result
		}
	}

	// Screen against federal debarment list.
	if h.debarmentList != nil {
		debarmentEntries := h.debarmentList.GetEntries()
		for _, entry := range debarmentEntries {
			similarity := JaroWinklerSimilarity(normalizedPayee, normalizeName(entry))
			if similarity >= DebarmentMatchThreshold {
				result.Status = models.ComplianceStatusNonCompliant
				result.Flags = append(result.Flags, models.ComplianceFlag{
					Rule:     "DEBARMENT",
					Severity: models.FlagSeverityBlocking,
					Message:  "Payee matches federal debarment list: " + entry,
				})
				// Fail-fast: return immediately on debarment match.
				return result
			}
		}
	}

	// Spending threshold checks (after OFAC and debarment pass).
	if h.spendingStore != nil {
		// Parse amount from the extraction result.
		amountField, hasAmount := extraction.Fields["amount"]
		if hasAmount {
			amountStr := amountField.Normalized
			if amountStr == "" {
				amountStr = amountField.Value
			}
			amount := parseAmount(amountStr)
			if amount > 0 {
				spendingFlags := checkSpendingThresholds(payeeName, amount, h.thresholds, h.spendingStore)
				result.Flags = append(result.Flags, spendingFlags...)
			}
		}
	}

	// Determine final status based on all accumulated flags.
	result.Status = DetermineComplianceStatus(result.Flags)

	return result
}

// normalizeName normalizes a name for comparison by lowercasing and removing
// non-alphanumeric characters (except spaces).
func normalizeName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	var b strings.Builder
	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == ' ' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// JaroWinklerSimilarity computes the Jaro-Winkler similarity between two strings.
// Returns a value between 0.0 (no similarity) and 1.0 (exact match).
func JaroWinklerSimilarity(s1, s2 string) float64 {
	if s1 == s2 {
		return 1.0
	}
	if len(s1) == 0 || len(s2) == 0 {
		return 0.0
	}

	jaroSim := jaroSimilarity(s1, s2)

	// Winkler modification: boost for common prefix (up to 4 chars).
	prefixLen := 0
	maxPrefix := 4
	if len(s1) < maxPrefix {
		maxPrefix = len(s1)
	}
	if len(s2) < maxPrefix {
		maxPrefix = len(s2)
	}
	for i := 0; i < maxPrefix; i++ {
		if s1[i] == s2[i] {
			prefixLen++
		} else {
			break
		}
	}

	// Standard Winkler scaling factor is 0.1.
	const p = 0.1
	return jaroSim + float64(prefixLen)*p*(1.0-jaroSim)
}

// jaroSimilarity computes the Jaro similarity between two strings.
func jaroSimilarity(s1, s2 string) float64 {
	if s1 == s2 {
		return 1.0
	}

	len1 := len(s1)
	len2 := len(s2)

	// Maximum distance for matching characters.
	maxDist := len1
	if len2 > maxDist {
		maxDist = len2
	}
	maxDist = maxDist/2 - 1
	if maxDist < 0 {
		maxDist = 0
	}

	s1Matches := make([]bool, len1)
	s2Matches := make([]bool, len2)

	matches := 0
	transpositions := 0

	// Find matching characters.
	for i := 0; i < len1; i++ {
		start := i - maxDist
		if start < 0 {
			start = 0
		}
		end := i + maxDist + 1
		if end > len2 {
			end = len2
		}
		for j := start; j < end; j++ {
			if s2Matches[j] || s1[i] != s2[j] {
				continue
			}
			s1Matches[i] = true
			s2Matches[j] = true
			matches++
			break
		}
	}

	if matches == 0 {
		return 0.0
	}

	// Count transpositions.
	k := 0
	for i := 0; i < len1; i++ {
		if !s1Matches[i] {
			continue
		}
		for !s2Matches[k] {
			k++
		}
		if s1[i] != s2[k] {
			transpositions++
		}
		k++
	}

	fMatches := float64(matches)
	return (fMatches/float64(len1) + fMatches/float64(len2) + (fMatches-float64(transpositions)/2.0)/fMatches) / 3.0
}

// parseAmount parses a currency string to a float64. Handles values with or
// without a dollar sign and commas.
func parseAmount(s string) float64 {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "$")
	s = strings.ReplaceAll(s, ",", "")
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return val
}
