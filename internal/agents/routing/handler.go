package routing

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"federal-payment-processing/internal/models"
)

// DueDateUrgencyDays is the number of days within which a payment due date
// triggers URGENT priority regardless of amount.
const DueDateUrgencyDays = 3

// Amount thresholds for approval level determination.
const (
	ThresholdPurchaseCard           = 2500.00
	ThresholdSupervisor             = 25000.00
	ThresholdContractingOfficer     = 250000.00
	ThresholdSeniorContractingOfficer = 1000000.00
)

// RoutingHandler determines the appropriate approval authority and priority
// based on payment amount and compliance status.
type RoutingHandler struct{}

// NewRoutingHandler creates a new RoutingHandler.
func NewRoutingHandler() *RoutingHandler {
	return &RoutingHandler{}
}

// DetermineRoute evaluates the extraction result and compliance result to produce
// a routing decision with the appropriate approval level and priority.
func (h *RoutingHandler) DetermineRoute(extraction *models.ExtractionResult, compliance *models.ComplianceResult) (*models.RoutingDecision, error) {
	return h.DetermineRouteWithTime(extraction, compliance, time.Now())
}

// DetermineRouteWithTime is like DetermineRoute but accepts an explicit "now" time
// for testability of due-date urgency logic.
func (h *RoutingHandler) DetermineRouteWithTime(extraction *models.ExtractionResult, compliance *models.ComplianceResult, now time.Time) (*models.RoutingDecision, error) {
	amount, err := parseAmount(extraction)
	if err != nil {
		return nil, fmt.Errorf("routing: failed to parse amount: %w", err)
	}

	approvalLevel, priority := determineApprovalLevelAndPriority(amount)

	// Elevate approval level and priority if compliance has conditions.
	if compliance != nil && compliance.Status == models.ComplianceStatusCompliantWithConditions {
		approvalLevel = elevateApprovalLevel(approvalLevel)
		priority = elevatePriority(priority)
	}

	// Override priority to URGENT if due date is within 3 days.
	if dueDate, ok := parseDueDate(extraction); ok {
		daysUntilDue := int(dueDate.Sub(now).Hours() / 24)
		if daysUntilDue <= DueDateUrgencyDays {
			priority = models.PriorityUrgent
		}
	}

	rationale := buildRationale(amount, approvalLevel, priority)

	return &models.RoutingDecision{
		Status:        models.RoutingStatusRouted,
		ApprovalLevel: approvalLevel,
		Priority:      priority,
		Rationale:     rationale,
		RoutedAt:      now,
	}, nil
}

// determineApprovalLevelAndPriority maps a payment amount to the base approval
// level and priority per federal delegation of authority thresholds.
func determineApprovalLevelAndPriority(amount float64) (models.ApprovalLevel, models.Priority) {
	switch {
	case amount <= ThresholdPurchaseCard:
		return models.ApprovalLevelPurchaseCard, models.PriorityLow
	case amount <= ThresholdSupervisor:
		return models.ApprovalLevelSupervisor, models.PriorityNormal
	case amount <= ThresholdContractingOfficer:
		return models.ApprovalLevelContractingOfficer, models.PriorityNormal
	case amount <= ThresholdSeniorContractingOfficer:
		return models.ApprovalLevelSeniorContractingOfficer, models.PriorityHigh
	default:
		return models.ApprovalLevelAgencyHead, models.PriorityUrgent
	}
}

// parseAmount extracts and parses the payment amount from the extraction result.
// It looks for the "amount" or "totalAmount" field and handles currency formatting.
func parseAmount(extraction *models.ExtractionResult) (float64, error) {
	// Try "amount" field first, then "totalAmount".
	field, ok := extraction.Fields["amount"]
	if !ok {
		field, ok = extraction.Fields["totalAmount"]
		if !ok {
			return 0, fmt.Errorf("no amount or totalAmount field found in extraction result")
		}
	}

	// Use normalized value if available, otherwise raw value.
	raw := field.Normalized
	if raw == "" {
		raw = field.Value
	}

	return parseCurrencyString(raw)
}

// parseCurrencyString converts a currency string (e.g., "$1,234.56") to a float64.
func parseCurrencyString(s string) (float64, error) {
	s = strings.TrimSpace(s)
	// Remove currency symbol and commas.
	s = strings.ReplaceAll(s, "$", "")
	s = strings.ReplaceAll(s, ",", "")
	s = strings.TrimSpace(s)

	if s == "" {
		return 0, fmt.Errorf("empty amount string")
	}

	amount, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid amount format %q: %w", s, err)
	}

	return amount, nil
}

// buildRationale constructs a human-readable explanation of the routing decision.
func buildRationale(amount float64, level models.ApprovalLevel, priority models.Priority) string {
	return fmt.Sprintf(
		"Payment amount $%.2f routed to %s with %s priority based on federal delegation of authority thresholds",
		amount, level, priority,
	)
}

// elevateApprovalLevel moves the approval level up by one tier using ApprovalLevelOrder.
// If already at the maximum (AGENCY_HEAD), it stays at the maximum.
func elevateApprovalLevel(current models.ApprovalLevel) models.ApprovalLevel {
	order := models.ApprovalLevelOrder
	for i, level := range order {
		if level == current {
			if i+1 < len(order) {
				return order[i+1]
			}
			return current // already at max
		}
	}
	return current
}

// elevatePriority moves the priority up by one level using PriorityOrder.
// If already at the maximum (URGENT), it stays at the maximum.
func elevatePriority(current models.Priority) models.Priority {
	order := models.PriorityOrder
	for i, p := range order {
		if p == current {
			if i+1 < len(order) {
				return order[i+1]
			}
			return current // already at max
		}
	}
	return current
}

// parseDueDate attempts to extract and parse the "date" field from the extraction result.
// Returns the parsed time and true if successful, or zero time and false if not available.
func parseDueDate(extraction *models.ExtractionResult) (time.Time, bool) {
	field, ok := extraction.Fields["date"]
	if !ok {
		return time.Time{}, false
	}

	raw := field.Normalized
	if raw == "" {
		raw = field.Value
	}
	if raw == "" {
		return time.Time{}, false
	}

	// Try common date formats.
	formats := []string{
		"2006-01-02",
		"01/02/2006",
		"1/2/2006",
		"2006-01-02T15:04:05Z",
		"Jan 2, 2006",
		"January 2, 2006",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, raw); err == nil {
			return t, true
		}
	}

	return time.Time{}, false
}
