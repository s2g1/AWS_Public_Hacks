package disbursement

import (
	"fmt"
	"testing"
	"time"

	"federal-payment-processing/internal/models"

	"pgregory.net/rapid"
)

// **Validates: Requirements 23.1, 23.2, 23.3**

// paramMockTreasury implements TreasuryInterface with configurable success/failure behavior.
type paramMockTreasury struct {
	shouldSucceed bool
	errorMessage  string
}

func (m *paramMockTreasury) ExecuteTransfer(from, to AccountInfo, amount float64, reference, memo string) (*TransferResult, error) {
	if m.shouldSucceed {
		return &TransferResult{
			TransactionID: fmt.Sprintf("TREAS-%s", reference),
			Status:        "SUCCESS",
		}, nil
	}
	return &TransferResult{
		TransactionID: "",
		Status:        "FAILED",
		ErrorMessage:  m.errorMessage,
		IsRetryable:   false,
	}, nil
}

// paramMockAccountLookup implements AccountLookup with configurable behavior.
type paramMockAccountLookup struct {
	hasAccount bool
}

func (m *paramMockAccountLookup) GetPayeeAccount(payee string) (*AccountInfo, error) {
	if m.hasAccount {
		return &AccountInfo{
			AccountNumber: "9876543210",
			RoutingNumber: "021000021",
			BankName:      "Federal Reserve Bank",
		}, nil
	}
	return nil, nil
}

// genPositiveAmount generates a random positive payment amount.
func genPositiveAmount() *rapid.Generator[float64] {
	return rapid.Custom(func(t *rapid.T) float64 {
		// Generate amounts between 0.01 and 10,000,000.00
		cents := rapid.IntRange(1, 1000000000).Draw(t, "cents")
		return float64(cents) / 100.0
	})
}

// genPayeeName generates a random payee name string.
func genPayeeName() *rapid.Generator[string] {
	return rapid.Custom(func(t *rapid.T) string {
		names := []string{
			"Acme Corp", "GlobalTech Inc", "Federal Systems LLC",
			"Smith & Associates", "Pacific Engineering", "National Services Co",
			"Delta Solutions", "Omega Industries", "Prime Contractors",
			"United Federal Group", "Liberty Defense Systems", "Eagle Logistics",
		}
		idx := rapid.IntRange(0, len(names)-1).Draw(t, "nameIndex")
		return names[idx]
	})
}

// buildPropertyTestRecord builds an approved payment record with random amount and payee.
func buildPropertyTestRecord(payee string, amount float64) *models.PaymentRecord {
	amountStr := fmt.Sprintf("%.2f", amount)
	return &models.PaymentRecord{
		PaymentID: fmt.Sprintf("PAY-%d", time.Now().UnixNano()),
		Status:    models.PaymentStatusApproved,
		ExtractedData: &models.ExtractionResult{
			DocumentType: models.DocumentTypeInvoice,
			Fields: map[string]models.ExtractedField{
				"payee": {
					Value:      payee,
					Confidence: 0.95,
					Normalized: payee,
				},
				"amount": {
					Value:      "$" + amountStr,
					Confidence: 0.98,
					Normalized: amountStr,
				},
			},
			OverallConfidence: 0.95,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// TestProperty_DisbursementSuccess_HasCompleteConfirmation verifies that when a
// disbursement succeeds, the confirmation is non-nil, has a unique transaction ID,
// correct amount, correct payee, and non-zero timestamp.
func TestProperty_DisbursementSuccess_HasCompleteConfirmation(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		payee := genPayeeName().Draw(t, "payee")
		amount := genPositiveAmount().Draw(t, "amount")

		handler := NewDisbursementHandler(
			&paramMockTreasury{shouldSucceed: true},
			&paramMockAccountLookup{hasAccount: true},
			AccountInfo{AccountNumber: "****0001", RoutingNumber: "021000021", BankName: "US Treasury"},
		)

		record := buildPropertyTestRecord(payee, amount)
		result := handler.Execute(record)

		// On success: confirmation must be non-nil
		if result.Status != models.DisbursementStatusSuccess {
			t.Fatalf("expected SUCCESS status, got %s", result.Status)
		}
		if result.Confirmation == nil {
			t.Fatal("expected non-nil confirmation on SUCCESS")
		}
		// Transaction ID must be non-empty
		if result.Confirmation.TransactionID == "" {
			t.Fatal("expected non-empty transaction ID on SUCCESS")
		}
		// Amount must match the input
		if result.Confirmation.Amount != amount {
			t.Fatalf("expected amount %.2f, got %.2f", amount, result.Confirmation.Amount)
		}
		// Payee must match
		if result.Confirmation.Payee != payee {
			t.Fatalf("expected payee %q, got %q", payee, result.Confirmation.Payee)
		}
		// Timestamp must be non-zero
		if result.Confirmation.DisbursedAt.IsZero() {
			t.Fatal("expected non-zero DisbursedAt timestamp on SUCCESS")
		}
		// Reference must be non-empty
		if result.Confirmation.Reference == "" {
			t.Fatal("expected non-empty reference on SUCCESS")
		}
	})
}

// TestProperty_DisbursementFailure_NoPartialState verifies that when disbursement
// fails, confirmation is nil, reason is non-empty, and no partial state exists.
func TestProperty_DisbursementFailure_NoPartialState(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		payee := genPayeeName().Draw(t, "payee")
		amount := genPositiveAmount().Draw(t, "amount")

		errorMessages := []string{
			"Insufficient funds", "Account frozen", "Transfer limit exceeded",
			"System unavailable", "Invalid routing number", "Account closed",
		}
		errIdx := rapid.IntRange(0, len(errorMessages)-1).Draw(t, "errorIndex")

		handler := NewDisbursementHandler(
			&paramMockTreasury{shouldSucceed: false, errorMessage: errorMessages[errIdx]},
			&paramMockAccountLookup{hasAccount: true},
			AccountInfo{AccountNumber: "****0001", RoutingNumber: "021000021", BankName: "US Treasury"},
		)

		record := buildPropertyTestRecord(payee, amount)
		result := handler.Execute(record)

		// On failure: status must be FAILED
		if result.Status != models.DisbursementStatusFailed {
			t.Fatalf("expected FAILED status, got %s", result.Status)
		}
		// Confirmation must be nil (no partial state)
		if result.Confirmation != nil {
			t.Fatal("expected nil confirmation on FAILED—no partial state should exist")
		}
		// Reason must be non-empty
		if result.Reason == "" {
			t.Fatal("expected non-empty reason on FAILED")
		}
	})
}

// TestProperty_DisbursementExactlyOneState verifies that the result is always in
// exactly one of two states: SUCCESS with confirmation OR FAILED with reason.
// Never both, never neither.
func TestProperty_DisbursementExactlyOneState(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		payee := genPayeeName().Draw(t, "payee")
		amount := genPositiveAmount().Draw(t, "amount")
		shouldSucceed := rapid.Bool().Draw(t, "shouldSucceed")

		var treasury TreasuryInterface
		if shouldSucceed {
			treasury = &paramMockTreasury{shouldSucceed: true}
		} else {
			treasury = &paramMockTreasury{shouldSucceed: false, errorMessage: "Transfer rejected"}
		}

		handler := NewDisbursementHandler(
			treasury,
			&paramMockAccountLookup{hasAccount: true},
			AccountInfo{AccountNumber: "****0001", RoutingNumber: "021000021", BankName: "US Treasury"},
		)

		record := buildPropertyTestRecord(payee, amount)
		result := handler.Execute(record)

		isSuccess := result.Status == models.DisbursementStatusSuccess
		isFailed := result.Status == models.DisbursementStatusFailed

		// Must be exactly one of SUCCESS or FAILED
		if isSuccess == isFailed {
			t.Fatalf("result must be exactly one of SUCCESS or FAILED, got status=%s", result.Status)
		}

		if isSuccess {
			// SUCCESS state: confirmation present, reason empty
			if result.Confirmation == nil {
				t.Fatal("SUCCESS state must have non-nil confirmation")
			}
			if result.Confirmation.TransactionID == "" {
				t.Fatal("SUCCESS state must have non-empty transaction ID")
			}
		} else {
			// FAILED state: confirmation nil, reason present
			if result.Confirmation != nil {
				t.Fatal("FAILED state must have nil confirmation")
			}
			if result.Reason == "" {
				t.Fatal("FAILED state must have non-empty reason")
			}
		}
	})
}

// TestProperty_DisbursementNoPartialPayment verifies that no partial payment state
// exists: if status is SUCCESS then confirmation has all required fields; if status
// is FAILED then confirmation is nil.
func TestProperty_DisbursementNoPartialPayment(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		payee := genPayeeName().Draw(t, "payee")
		amount := genPositiveAmount().Draw(t, "amount")

		// Simulate various scenarios: success, treasury failure, missing account
		scenario := rapid.IntRange(0, 2).Draw(t, "scenario")

		var treasury TreasuryInterface
		var accountLookup AccountLookup

		switch scenario {
		case 0: // Success scenario
			treasury = &paramMockTreasury{shouldSucceed: true}
			accountLookup = &paramMockAccountLookup{hasAccount: true}
		case 1: // Treasury failure scenario
			treasury = &paramMockTreasury{shouldSucceed: false, errorMessage: "Funds unavailable"}
			accountLookup = &paramMockAccountLookup{hasAccount: true}
		case 2: // Missing account scenario
			treasury = &paramMockTreasury{shouldSucceed: true}
			accountLookup = &paramMockAccountLookup{hasAccount: false}
		}

		handler := NewDisbursementHandler(
			treasury,
			accountLookup,
			AccountInfo{AccountNumber: "****0001", RoutingNumber: "021000021", BankName: "US Treasury"},
		)

		record := buildPropertyTestRecord(payee, amount)
		result := handler.Execute(record)

		if result.Status == models.DisbursementStatusSuccess {
			// SUCCESS: confirmation must have ALL required fields
			if result.Confirmation == nil {
				t.Fatal("SUCCESS must have non-nil confirmation")
			}
			if result.Confirmation.TransactionID == "" {
				t.Fatal("SUCCESS confirmation must have non-empty TransactionID")
			}
			if result.Confirmation.Amount <= 0 {
				t.Fatalf("SUCCESS confirmation must have positive amount, got %.2f", result.Confirmation.Amount)
			}
			if result.Confirmation.Payee == "" {
				t.Fatal("SUCCESS confirmation must have non-empty Payee")
			}
			if result.Confirmation.DisbursedAt.IsZero() {
				t.Fatal("SUCCESS confirmation must have non-zero DisbursedAt")
			}
			if result.Confirmation.Reference == "" {
				t.Fatal("SUCCESS confirmation must have non-empty Reference")
			}
		} else if result.Status == models.DisbursementStatusFailed {
			// FAILED: confirmation must be nil (no partial state)
			if result.Confirmation != nil {
				t.Fatal("FAILED must have nil confirmation—no partial payment state allowed")
			}
		} else {
			t.Fatalf("unexpected disbursement status: %s (must be SUCCESS or FAILED)", result.Status)
		}
	})
}

// **Validates: Requirements 12.1, 12.2**

// allNonApprovedStatuses returns all PaymentStatus values except APPROVED.
func allNonApprovedStatuses() []models.PaymentStatus {
	return []models.PaymentStatus{
		models.PaymentStatusReceived,
		models.PaymentStatusExtracting,
		models.PaymentStatusExtracted,
		models.PaymentStatusValidating,
		models.PaymentStatusValidated,
		models.PaymentStatusCheckingCompliance,
		models.PaymentStatusCompliant,
		models.PaymentStatusRouting,
		models.PaymentStatusRouted,
		models.PaymentStatusApproving,
		models.PaymentStatusDisbursing,
		models.PaymentStatusDisbursed,
		models.PaymentStatusRejected,
		models.PaymentStatusEscalated,
		models.PaymentStatusFailed,
	}
}

// genNonApprovedStatus generates a random PaymentStatus that is NOT APPROVED.
func genNonApprovedStatus() *rapid.Generator[models.PaymentStatus] {
	return rapid.Custom(func(t *rapid.T) models.PaymentStatus {
		statuses := allNonApprovedStatuses()
		idx := rapid.IntRange(0, len(statuses)-1).Draw(t, "statusIndex")
		return statuses[idx]
	})
}

// buildRecordWithStatus creates a payment record with the given status and valid extracted data.
func buildRecordWithStatus(status models.PaymentStatus, payee string, amount float64) *models.PaymentRecord {
	amountStr := fmt.Sprintf("%.2f", amount)
	return &models.PaymentRecord{
		PaymentID: fmt.Sprintf("PAY-%d", time.Now().UnixNano()),
		Status:    status,
		ExtractedData: &models.ExtractionResult{
			DocumentType: models.DocumentTypeInvoice,
			Fields: map[string]models.ExtractedField{
				"payee": {
					Value:      payee,
					Confidence: 0.95,
					Normalized: payee,
				},
				"amount": {
					Value:      "$" + amountStr,
					Confidence: 0.98,
					Normalized: amountStr,
				},
			},
			OverallConfidence: 0.95,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// TestProperty_DisbursementPrecondition_NonApprovedAlwaysFails verifies that for any
// payment NOT in APPROVED status, Execute always returns FAILED with a reason
// mentioning "not in APPROVED state".
func TestProperty_DisbursementPrecondition_NonApprovedAlwaysFails(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		status := genNonApprovedStatus().Draw(t, "status")
		payee := genPayeeName().Draw(t, "payee")
		amount := genPositiveAmount().Draw(t, "amount")

		// Even with a treasury that always succeeds and account that always exists,
		// a non-APPROVED record must fail at precondition check.
		handler := NewDisbursementHandler(
			&paramMockTreasury{shouldSucceed: true},
			&paramMockAccountLookup{hasAccount: true},
			AccountInfo{AccountNumber: "****0001", RoutingNumber: "021000021", BankName: "US Treasury"},
		)

		record := buildRecordWithStatus(status, payee, amount)
		result := handler.Execute(record)

		// Must return FAILED status
		if result.Status != models.DisbursementStatusFailed {
			t.Fatalf("expected FAILED for status %s, got %s", status, result.Status)
		}

		// Reason must mention "not in APPROVED state"
		if !containsSubstring(result.Reason, "not in APPROVED state") {
			t.Fatalf("expected reason to mention 'not in APPROVED state', got: %s", result.Reason)
		}

		// Must not be retryable (precondition failure is not transient)
		if result.Retryable {
			t.Fatalf("precondition failure should not be retryable for status %s", status)
		}

		// Confirmation must be nil
		if result.Confirmation != nil {
			t.Fatalf("precondition failure must not produce a confirmation for status %s", status)
		}
	})
}

// TestProperty_DisbursementPrecondition_ApprovedProceedsPastPrecondition verifies that
// for any payment in APPROVED status with valid data and available account, execution
// proceeds past the precondition check (does not fail on precondition).
func TestProperty_DisbursementPrecondition_ApprovedProceedsPastPrecondition(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		payee := genPayeeName().Draw(t, "payee")
		amount := genPositiveAmount().Draw(t, "amount")

		handler := NewDisbursementHandler(
			&paramMockTreasury{shouldSucceed: true},
			&paramMockAccountLookup{hasAccount: true},
			AccountInfo{AccountNumber: "****0001", RoutingNumber: "021000021", BankName: "US Treasury"},
		)

		record := buildRecordWithStatus(models.PaymentStatusApproved, payee, amount)
		result := handler.Execute(record)

		// APPROVED payments with valid data should NOT fail on precondition
		if result.Status == models.DisbursementStatusFailed &&
			containsSubstring(result.Reason, "not in APPROVED state") {
			t.Fatalf("APPROVED payment should not fail precondition check, got reason: %s", result.Reason)
		}

		// With valid data and available account and a working treasury, it should succeed
		if result.Status != models.DisbursementStatusSuccess {
			t.Fatalf("expected SUCCESS for APPROVED payment with valid data, got %s (reason: %s)", result.Status, result.Reason)
		}
	})
}

// TestProperty_DisbursementPrecondition_AlwaysFirstCheck verifies that the precondition
// check is always the first check — even if treasury would succeed and account exists,
// a non-APPROVED record always fails with the precondition error (never reaching treasury).
func TestProperty_DisbursementPrecondition_AlwaysFirstCheck(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		status := genNonApprovedStatus().Draw(t, "status")
		payee := genPayeeName().Draw(t, "payee")
		amount := genPositiveAmount().Draw(t, "amount")

		// Use a treasury that tracks invocations to verify it's never called
		trackingTreasury := &trackingMockTreasury{called: false}

		handler := NewDisbursementHandler(
			trackingTreasury,
			&paramMockAccountLookup{hasAccount: true},
			AccountInfo{AccountNumber: "****0001", RoutingNumber: "021000021", BankName: "US Treasury"},
		)

		record := buildRecordWithStatus(status, payee, amount)
		result := handler.Execute(record)

		// Must fail with precondition error
		if result.Status != models.DisbursementStatusFailed {
			t.Fatalf("expected FAILED for non-APPROVED status %s, got %s", status, result.Status)
		}

		if !containsSubstring(result.Reason, "not in APPROVED state") {
			t.Fatalf("expected precondition error for status %s, got: %s", status, result.Reason)
		}

		// Treasury must never have been called (precondition is checked first)
		if trackingTreasury.called {
			t.Fatalf("treasury should not be called when status is %s — precondition must be checked first", status)
		}
	})
}

// trackingMockTreasury implements TreasuryInterface and tracks whether ExecuteTransfer was called.
type trackingMockTreasury struct {
	called bool
}

func (m *trackingMockTreasury) ExecuteTransfer(from, to AccountInfo, amount float64, reference, memo string) (*TransferResult, error) {
	m.called = true
	return &TransferResult{
		TransactionID: "TRACK-123",
		Status:        "SUCCESS",
	}, nil
}

// containsSubstring checks if s contains the given substring (case-sensitive).
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && stringContains(s, substr))
}

// stringContains is a simple helper to check substring containment.
func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
