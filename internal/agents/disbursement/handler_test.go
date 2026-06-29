package disbursement

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"federal-payment-processing/internal/models"
)

// mockTreasury implements TreasuryInterface for testing.
type mockTreasury struct {
	result *TransferResult
	err    error
}

func (m *mockTreasury) ExecuteTransfer(from, to AccountInfo, amount float64, reference, memo string) (*TransferResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.result, nil
}

// mockAccountLookup implements AccountLookup for testing.
type mockAccountLookup struct {
	account *AccountInfo
	err     error
}

func (m *mockAccountLookup) GetPayeeAccount(payee string) (*AccountInfo, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.account, nil
}

// helper to build a valid approved payment record for tests.
func buildApprovedPaymentRecord() *models.PaymentRecord {
	return &models.PaymentRecord{
		PaymentID: "PAY-001",
		Status:    models.PaymentStatusApproved,
		ExtractedData: &models.ExtractionResult{
			DocumentType: models.DocumentTypeInvoice,
			Fields: map[string]models.ExtractedField{
				"payee": {
					Value:      "Acme Corp",
					Confidence: 0.95,
					Normalized: "Acme Corp",
				},
				"amount": {
					Value:      "$5,000.00",
					Confidence: 0.98,
					Normalized: "5000.00",
				},
			},
			OverallConfidence: 0.95,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func newTestHandler(treasury TreasuryInterface, lookup AccountLookup) *DisbursementHandler {
	return NewDisbursementHandler(
		treasury,
		lookup,
		AccountInfo{
			AccountNumber: "****0001",
			RoutingNumber: "021000021",
			BankName:      "US Treasury",
		},
	)
}

func TestExecute_NotApproved_ReturnsFailed(t *testing.T) {
	handler := newTestHandler(
		&mockTreasury{result: &TransferResult{Status: "SUCCESS", TransactionID: "TX-123"}},
		&mockAccountLookup{account: &AccountInfo{AccountNumber: "1234", RoutingNumber: "5678", BankName: "Test Bank"}},
	)

	statuses := []models.PaymentStatus{
		models.PaymentStatusReceived,
		models.PaymentStatusExtracted,
		models.PaymentStatusValidated,
		models.PaymentStatusCompliant,
		models.PaymentStatusRouted,
		models.PaymentStatusDisbursing,
		models.PaymentStatusRejected,
		models.PaymentStatusEscalated,
		models.PaymentStatusFailed,
	}

	for _, status := range statuses {
		t.Run(string(status), func(t *testing.T) {
			record := buildApprovedPaymentRecord()
			record.Status = status

			result := handler.Execute(record)

			if result.Status != models.DisbursementStatusFailed {
				t.Errorf("expected FAILED status for payment in %s state, got %s", status, result.Status)
			}
			if result.Reason == "" {
				t.Error("expected a failure reason when payment is not APPROVED")
			}
			if !strings.Contains(result.Reason, "not in APPROVED state") {
				t.Errorf("expected reason to mention 'not in APPROVED state', got: %s", result.Reason)
			}
			if result.Retryable {
				t.Error("expected retryable to be false for precondition failure")
			}
			if result.Confirmation != nil {
				t.Error("expected no confirmation for failed disbursement")
			}
		})
	}
}

func TestExecute_MissingPayeeAccount_ReturnsFailed(t *testing.T) {
	handler := newTestHandler(
		&mockTreasury{result: &TransferResult{Status: "SUCCESS", TransactionID: "TX-123"}},
		&mockAccountLookup{account: nil}, // no account found
	)

	record := buildApprovedPaymentRecord()
	result := handler.Execute(record)

	if result.Status != models.DisbursementStatusFailed {
		t.Errorf("expected FAILED status when payee account missing, got %s", result.Status)
	}
	if !strings.Contains(result.Reason, "account") {
		t.Errorf("expected reason to mention account issue, got: %s", result.Reason)
	}
	if result.Retryable {
		t.Error("expected retryable to be false for missing account")
	}
}

func TestExecute_AccountLookupError_ReturnsFailed(t *testing.T) {
	handler := newTestHandler(
		&mockTreasury{result: &TransferResult{Status: "SUCCESS", TransactionID: "TX-123"}},
		&mockAccountLookup{err: fmt.Errorf("database connection failed")},
	)

	record := buildApprovedPaymentRecord()
	result := handler.Execute(record)

	if result.Status != models.DisbursementStatusFailed {
		t.Errorf("expected FAILED status when account lookup errors, got %s", result.Status)
	}
	if !strings.Contains(result.Reason, "look up payee account") {
		t.Errorf("expected reason to mention lookup failure, got: %s", result.Reason)
	}
}

func TestExecute_SuccessfulTransfer(t *testing.T) {
	handler := newTestHandler(
		&mockTreasury{result: &TransferResult{
			TransactionID: "TREAS-12345",
			Status:        "SUCCESS",
		}},
		&mockAccountLookup{account: &AccountInfo{
			AccountNumber: "9876543210",
			RoutingNumber: "021000021",
			BankName:      "Federal Reserve Bank",
		}},
	)

	record := buildApprovedPaymentRecord()
	result := handler.Execute(record)

	if result.Status != models.DisbursementStatusSuccess {
		t.Errorf("expected SUCCESS status, got %s", result.Status)
	}
	if result.Confirmation == nil {
		t.Fatal("expected confirmation to be present on success")
	}
	if result.Confirmation.TransactionID != "TREAS-12345" {
		t.Errorf("expected transaction ID 'TREAS-12345', got %s", result.Confirmation.TransactionID)
	}
	if result.Confirmation.Amount != 5000.00 {
		t.Errorf("expected amount 5000.00, got %.2f", result.Confirmation.Amount)
	}
	if result.Confirmation.Payee != "Acme Corp" {
		t.Errorf("expected payee 'Acme Corp', got %s", result.Confirmation.Payee)
	}
	if result.Confirmation.Reference == "" {
		t.Error("expected non-empty transaction reference")
	}
	if !strings.HasPrefix(result.Confirmation.Reference, "TXN-PAY-001-") {
		t.Errorf("expected reference to start with 'TXN-PAY-001-', got %s", result.Confirmation.Reference)
	}
	if result.Confirmation.DisbursedAt.IsZero() {
		t.Error("expected non-zero disbursed timestamp")
	}
	if result.Reason != "" {
		t.Errorf("expected empty reason on success, got: %s", result.Reason)
	}
	if result.Retryable {
		t.Error("expected retryable to be false on success")
	}
}

func TestExecute_FailedTransfer_WithRetryableFlag(t *testing.T) {
	handler := newTestHandler(
		&mockTreasury{result: &TransferResult{
			TransactionID: "",
			Status:        "FAILED",
			ErrorMessage:  "Insufficient funds in source account",
			IsRetryable:   true,
		}},
		&mockAccountLookup{account: &AccountInfo{
			AccountNumber: "9876543210",
			RoutingNumber: "021000021",
			BankName:      "Federal Reserve Bank",
		}},
	)

	record := buildApprovedPaymentRecord()
	result := handler.Execute(record)

	if result.Status != models.DisbursementStatusFailed {
		t.Errorf("expected FAILED status, got %s", result.Status)
	}
	if result.Reason != "Insufficient funds in source account" {
		t.Errorf("expected specific failure reason, got: %s", result.Reason)
	}
	if !result.Retryable {
		t.Error("expected retryable to be true for retryable transfer failure")
	}
	if result.Confirmation != nil {
		t.Error("expected no confirmation for failed transfer")
	}
}

func TestExecute_FailedTransfer_NotRetryable(t *testing.T) {
	handler := newTestHandler(
		&mockTreasury{result: &TransferResult{
			TransactionID: "",
			Status:        "FAILED",
			ErrorMessage:  "Account closed permanently",
			IsRetryable:   false,
		}},
		&mockAccountLookup{account: &AccountInfo{
			AccountNumber: "9876543210",
			RoutingNumber: "021000021",
			BankName:      "Federal Reserve Bank",
		}},
	)

	record := buildApprovedPaymentRecord()
	result := handler.Execute(record)

	if result.Status != models.DisbursementStatusFailed {
		t.Errorf("expected FAILED status, got %s", result.Status)
	}
	if result.Retryable {
		t.Error("expected retryable to be false for non-retryable failure")
	}
}

func TestExecute_TreasuryError_RetryableTrue(t *testing.T) {
	handler := newTestHandler(
		&mockTreasury{err: fmt.Errorf("network timeout")},
		&mockAccountLookup{account: &AccountInfo{
			AccountNumber: "9876543210",
			RoutingNumber: "021000021",
			BankName:      "Federal Reserve Bank",
		}},
	)

	record := buildApprovedPaymentRecord()
	result := handler.Execute(record)

	if result.Status != models.DisbursementStatusFailed {
		t.Errorf("expected FAILED status on treasury error, got %s", result.Status)
	}
	if !strings.Contains(result.Reason, "Treasury transfer error") {
		t.Errorf("expected reason to mention treasury error, got: %s", result.Reason)
	}
	if !result.Retryable {
		t.Error("expected retryable to be true when treasury interface returns an error")
	}
}

func TestExecute_NoExtractedData_ReturnsFailed(t *testing.T) {
	handler := newTestHandler(
		&mockTreasury{result: &TransferResult{Status: "SUCCESS", TransactionID: "TX-123"}},
		&mockAccountLookup{account: &AccountInfo{AccountNumber: "1234", RoutingNumber: "5678", BankName: "Test Bank"}},
	)

	record := buildApprovedPaymentRecord()
	record.ExtractedData = nil

	result := handler.Execute(record)

	if result.Status != models.DisbursementStatusFailed {
		t.Errorf("expected FAILED status when no extracted data, got %s", result.Status)
	}
	if !strings.Contains(result.Reason, "extract payment details") {
		t.Errorf("expected reason to mention payment details, got: %s", result.Reason)
	}
}

func TestGenerateTransactionReference_UniquePerCall(t *testing.T) {
	ref1 := generateTransactionReference("PAY-001")
	ref2 := generateTransactionReference("PAY-001")

	if ref1 == ref2 {
		t.Error("expected unique transaction references per call")
	}
	if !strings.HasPrefix(ref1, "TXN-PAY-001-") {
		t.Errorf("expected reference to start with 'TXN-PAY-001-', got %s", ref1)
	}
}

func TestParseCurrencyString(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
		hasError bool
	}{
		{"5000.00", 5000.00, false},
		{"$5,000.00", 5000.00, false},
		{"$1,234,567.89", 1234567.89, false},
		{"100", 100.00, false},
		{"$0.99", 0.99, false},
		{"", 0, true},
		{"abc", 0, true},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result, err := parseCurrencyString(tc.input)
			if tc.hasError {
				if err == nil {
					t.Errorf("expected error for input %q, got %.2f", tc.input, result)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for input %q: %v", tc.input, err)
				}
				if result != tc.expected {
					t.Errorf("expected %.2f for input %q, got %.2f", tc.expected, tc.input, result)
				}
			}
		})
	}
}
