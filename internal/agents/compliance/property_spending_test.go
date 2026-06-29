package compliance

import (
	"testing"

	"federal-payment-processing/internal/models"

	"pgregory.net/rapid"
)

// **Validates: Requirements 8.2, 8.4**

// mockSpendingStoreForProperty implements SpendingStore with a configurable cumulative spend.
type mockSpendingStoreForProperty struct {
	cumulativeSpend float64
}

func (m *mockSpendingStoreForProperty) GetCumulativeSpend(payee string, fiscalYear int) (float64, error) {
	return m.cumulativeSpend, nil
}

// genPositiveAmount generates a positive payment amount.
func genPositiveAmount() *rapid.Generator[float64] {
	return rapid.Float64Range(0.01, 10_000_000.0)
}

// genPositiveThreshold generates a positive threshold value.
func genPositiveThreshold() *rapid.Generator[float64] {
	return rapid.Float64Range(1.0, 5_000_000.0)
}

// genCumulativeSpend generates a non-negative cumulative spend value.
func genCumulativeSpend() *rapid.Generator[float64] {
	return rapid.Float64Range(0.0, 10_000_000.0)
}

// genPayeeName generates a random payee name string.
func genPayeeName() *rapid.Generator[string] {
	return rapid.StringMatching(`[A-Z][a-z]{2,10} [A-Z][a-z]{2,10}`)
}

// hasFlag checks if a flag with the given rule exists in the flags slice.
func hasFlag(flags []models.ComplianceFlag, rule string) bool {
	for _, f := range flags {
		if f.Rule == rule {
			return true
		}
	}
	return false
}

// getFlagSeverity returns the severity of a flag with the given rule, or empty string if not found.
func getFlagSeverity(flags []models.ComplianceFlag, rule string) models.FlagSeverity {
	for _, f := range flags {
		if f.Rule == rule {
			return f.Severity
		}
	}
	return ""
}

// TestProperty_SpendingThreshold_SingleTransactionExceeded verifies that when
// amount > SingleTransactionMax, a THRESHOLD_EXCEEDED flag with REQUIRES_REVIEW severity is always present.
func TestProperty_SpendingThreshold_SingleTransactionExceeded(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		singleMax := genPositiveThreshold().Draw(t, "singleTransactionMax")
		// Generate amount that exceeds single transaction max
		amount := rapid.Float64Range(singleMax+0.01, singleMax+5_000_000.0).Draw(t, "amount")
		annualMax := rapid.Float64Range(amount+1.0, amount+10_000_000.0).Draw(t, "annualMax")
		payee := genPayeeName().Draw(t, "payee")

		store := &mockSpendingStoreForProperty{cumulativeSpend: 0.0}
		thresholds := SpendingThresholds{
			SingleTransactionMax: singleMax,
			AnnualMax:            annualMax,
		}

		flags := checkSpendingThresholds(payee, amount, thresholds, store)

		if !hasFlag(flags, "THRESHOLD_EXCEEDED") {
			t.Fatalf("expected THRESHOLD_EXCEEDED flag when amount %.2f > singleMax %.2f, got flags: %+v",
				amount, singleMax, flags)
		}
		if getFlagSeverity(flags, "THRESHOLD_EXCEEDED") != models.FlagSeverityRequiresReview {
			t.Fatalf("expected REQUIRES_REVIEW severity for THRESHOLD_EXCEEDED, got %s",
				getFlagSeverity(flags, "THRESHOLD_EXCEEDED"))
		}
	})
}

// TestProperty_SpendingThreshold_AnnualLimitExceeded verifies that when
// cumulativeSpend + amount > AnnualMax, an ANNUAL_LIMIT flag with REQUIRES_REVIEW severity is always present.
func TestProperty_SpendingThreshold_AnnualLimitExceeded(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		annualMax := genPositiveThreshold().Draw(t, "annualMax")
		// Generate cumulative + amount that exceeds annual max
		cumulativeSpend := rapid.Float64Range(0.0, annualMax).Draw(t, "cumulativeSpend")
		// amount must cause cumulative + amount > annualMax
		remaining := annualMax - cumulativeSpend
		amount := rapid.Float64Range(remaining+0.01, remaining+5_000_000.0).Draw(t, "amount")
		// Set single transaction max high enough to not trigger that flag
		singleMax := amount + 1_000_000.0
		payee := genPayeeName().Draw(t, "payee")

		store := &mockSpendingStoreForProperty{cumulativeSpend: cumulativeSpend}
		thresholds := SpendingThresholds{
			SingleTransactionMax: singleMax,
			AnnualMax:            annualMax,
		}

		flags := checkSpendingThresholds(payee, amount, thresholds, store)

		if !hasFlag(flags, "ANNUAL_LIMIT") {
			t.Fatalf("expected ANNUAL_LIMIT flag when cumulative %.2f + amount %.2f > annualMax %.2f, got flags: %+v",
				cumulativeSpend, amount, annualMax, flags)
		}
		if getFlagSeverity(flags, "ANNUAL_LIMIT") != models.FlagSeverityRequiresReview {
			t.Fatalf("expected REQUIRES_REVIEW severity for ANNUAL_LIMIT, got %s",
				getFlagSeverity(flags, "ANNUAL_LIMIT"))
		}
	})
}

// TestProperty_SpendingThreshold_NoFlags_WhenUnderBothThresholds verifies that when
// amount <= SingleTransactionMax AND cumulativeSpend + amount <= AnnualMax, no flags are produced.
func TestProperty_SpendingThreshold_NoFlags_WhenUnderBothThresholds(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		singleMax := genPositiveThreshold().Draw(t, "singleTransactionMax")
		annualMax := rapid.Float64Range(singleMax, singleMax+10_000_000.0).Draw(t, "annualMax")
		// Amount must be <= singleMax
		amount := rapid.Float64Range(0.01, singleMax).Draw(t, "amount")
		// Cumulative spend must satisfy: cumulative + amount <= annualMax
		maxCumulative := annualMax - amount
		cumulativeSpend := rapid.Float64Range(0.0, maxCumulative).Draw(t, "cumulativeSpend")
		payee := genPayeeName().Draw(t, "payee")

		store := &mockSpendingStoreForProperty{cumulativeSpend: cumulativeSpend}
		thresholds := SpendingThresholds{
			SingleTransactionMax: singleMax,
			AnnualMax:            annualMax,
		}

		flags := checkSpendingThresholds(payee, amount, thresholds, store)

		if len(flags) != 0 {
			t.Fatalf("expected no flags when amount %.2f <= singleMax %.2f and cumulative %.2f + amount <= annualMax %.2f, got flags: %+v",
				amount, singleMax, cumulativeSpend, annualMax, flags)
		}
	})
}

// TestProperty_SpendingThreshold_NeverBlockingSeverity verifies that THRESHOLD_EXCEEDED
// and ANNUAL_LIMIT flags never have BLOCKING severity (they're always REQUIRES_REVIEW).
func TestProperty_SpendingThreshold_NeverBlockingSeverity(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		singleMax := genPositiveThreshold().Draw(t, "singleTransactionMax")
		annualMax := genPositiveThreshold().Draw(t, "annualMax")
		amount := genPositiveAmount().Draw(t, "amount")
		cumulativeSpend := genCumulativeSpend().Draw(t, "cumulativeSpend")
		payee := genPayeeName().Draw(t, "payee")

		store := &mockSpendingStoreForProperty{cumulativeSpend: cumulativeSpend}
		thresholds := SpendingThresholds{
			SingleTransactionMax: singleMax,
			AnnualMax:            annualMax,
		}

		flags := checkSpendingThresholds(payee, amount, thresholds, store)

		for _, flag := range flags {
			if flag.Rule == "THRESHOLD_EXCEEDED" && flag.Severity == models.FlagSeverityBlocking {
				t.Fatalf("THRESHOLD_EXCEEDED must never have BLOCKING severity, got BLOCKING for amount %.2f, singleMax %.2f",
					amount, singleMax)
			}
			if flag.Rule == "ANNUAL_LIMIT" && flag.Severity == models.FlagSeverityBlocking {
				t.Fatalf("ANNUAL_LIMIT must never have BLOCKING severity, got BLOCKING for cumulative %.2f + amount %.2f, annualMax %.2f",
					cumulativeSpend, amount, annualMax)
			}
		}
	})
}
