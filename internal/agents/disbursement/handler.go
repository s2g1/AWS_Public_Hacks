package disbursement

import (
	"fmt"
	"time"

	"federal-payment-processing/internal/models"

	"github.com/google/uuid"
)

// AccountInfo represents payee bank account information for fund transfers.
type AccountInfo struct {
	AccountNumber string `json:"accountNumber"`
	RoutingNumber string `json:"routingNumber"`
	BankName      string `json:"bankName"`
}

// TransferResult represents the outcome of a treasury transfer operation.
type TransferResult struct {
	TransactionID string `json:"transactionId"`
	Status        string `json:"status"`
	ErrorMessage  string `json:"errorMessage,omitempty"`
	IsRetryable   bool   `json:"isRetryable"`
}

// TreasuryInterface defines the interface for executing fund transfers
// against the treasury system.
type TreasuryInterface interface {
	ExecuteTransfer(from, to AccountInfo, amount float64, reference, memo string) (*TransferResult, error)
}

// AccountLookup defines the interface for retrieving payee account information.
type AccountLookup interface {
	GetPayeeAccount(payee string) (*AccountInfo, error)
}

// DisbursementHandler implements the disbursement agent logic.
// It verifies payment preconditions, looks up account info, and executes
// the fund transfer through the treasury interface.
type DisbursementHandler struct {
	Treasury      TreasuryInterface
	AccountLookup AccountLookup
	AgencyAccount AccountInfo
}

// NewDisbursementHandler creates a new DisbursementHandler with the given dependencies.
func NewDisbursementHandler(treasury TreasuryInterface, accountLookup AccountLookup, agencyAccount AccountInfo) *DisbursementHandler {
	return &DisbursementHandler{
		Treasury:      treasury,
		AccountLookup: accountLookup,
		AgencyAccount: agencyAccount,
	}
}

// Execute processes a payment record for disbursement.
// It verifies the payment is APPROVED, looks up the payee account,
// generates a transaction reference, and executes the transfer.
func (h *DisbursementHandler) Execute(record *models.PaymentRecord) *models.DisbursementResult {
	// Step 1: Verify payment is in APPROVED status
	if record.Status != models.PaymentStatusApproved {
		return &models.DisbursementResult{
			Status:    models.DisbursementStatusFailed,
			Reason:    fmt.Sprintf("Payment not in APPROVED state, current status: %s", record.Status),
			Retryable: false,
		}
	}

	// Step 2: Extract payee and amount from extracted data
	payee, amount, err := extractPayeeAndAmount(record)
	if err != nil {
		return &models.DisbursementResult{
			Status:    models.DisbursementStatusFailed,
			Reason:    fmt.Sprintf("Failed to extract payment details: %v", err),
			Retryable: false,
		}
	}

	// Step 3: Look up payee account information
	accountInfo, err := h.AccountLookup.GetPayeeAccount(payee)
	if err != nil || accountInfo == nil {
		reason := "No account on file for payee"
		if err != nil {
			reason = fmt.Sprintf("Failed to look up payee account: %v", err)
		}
		return &models.DisbursementResult{
			Status:    models.DisbursementStatusFailed,
			Reason:    reason,
			Retryable: false,
		}
	}

	// Step 4: Generate unique transaction reference
	transactionRef := generateTransactionReference(record.PaymentID)

	// Step 5: Build payment memo
	memo := buildPaymentMemo(record)

	// Step 6: Execute transfer via treasury interface
	transferResult, err := h.Treasury.ExecuteTransfer(
		h.AgencyAccount,
		*accountInfo,
		amount,
		transactionRef,
		memo,
	)
	if err != nil {
		return &models.DisbursementResult{
			Status:    models.DisbursementStatusFailed,
			Reason:    fmt.Sprintf("Treasury transfer error: %v", err),
			Retryable: true,
		}
	}

	// Step 7: Process transfer result
	if transferResult.Status == "SUCCESS" {
		confirmation := &models.PaymentConfirmation{
			TransactionID: transferResult.TransactionID,
			Amount:        amount,
			Payee:         payee,
			DisbursedAt:   time.Now(),
			Reference:     transactionRef,
		}
		return &models.DisbursementResult{
			Status:       models.DisbursementStatusSuccess,
			Confirmation: confirmation,
			Retryable:    false,
		}
	}

	// Transfer failed
	return &models.DisbursementResult{
		Status:    models.DisbursementStatusFailed,
		Reason:    transferResult.ErrorMessage,
		Retryable: transferResult.IsRetryable,
	}
}

// extractPayeeAndAmount retrieves the payee name and amount from the payment record's
// extracted data fields.
func extractPayeeAndAmount(record *models.PaymentRecord) (string, float64, error) {
	if record.ExtractedData == nil {
		return "", 0, fmt.Errorf("no extracted data available")
	}

	// Get payee
	payeeField, ok := record.ExtractedData.Fields["payee"]
	if !ok {
		// Try vendor field for purchase orders
		payeeField, ok = record.ExtractedData.Fields["vendor"]
		if !ok {
			return "", 0, fmt.Errorf("no payee or vendor field found")
		}
	}
	payee := payeeField.Normalized
	if payee == "" {
		payee = payeeField.Value
	}
	if payee == "" {
		return "", 0, fmt.Errorf("payee field is empty")
	}

	// Get amount
	amountField, ok := record.ExtractedData.Fields["amount"]
	if !ok {
		amountField, ok = record.ExtractedData.Fields["totalAmount"]
		if !ok {
			amountField, ok = record.ExtractedData.Fields["totalClaim"]
			if !ok {
				return "", 0, fmt.Errorf("no amount field found")
			}
		}
	}
	amountStr := amountField.Normalized
	if amountStr == "" {
		amountStr = amountField.Value
	}

	amount, err := parseCurrencyString(amountStr)
	if err != nil {
		return "", 0, fmt.Errorf("invalid amount %q: %w", amountStr, err)
	}

	if amount <= 0 {
		return "", 0, fmt.Errorf("amount must be positive, got %.2f", amount)
	}

	return payee, amount, nil
}

// parseCurrencyString converts a currency string (e.g., "$1,234.56" or "1234.56") to float64.
func parseCurrencyString(s string) (float64, error) {
	// Remove currency symbols and commas
	cleaned := ""
	for _, c := range s {
		if c >= '0' && c <= '9' || c == '.' || c == '-' {
			cleaned += string(c)
		}
	}
	if cleaned == "" {
		return 0, fmt.Errorf("no numeric value found in %q", s)
	}

	var amount float64
	_, err := fmt.Sscanf(cleaned, "%f", &amount)
	if err != nil {
		return 0, fmt.Errorf("failed to parse amount %q: %w", cleaned, err)
	}
	return amount, nil
}

// generateTransactionReference creates a unique reference for the disbursement transaction.
func generateTransactionReference(paymentID string) string {
	return fmt.Sprintf("TXN-%s-%s", paymentID, uuid.New().String()[:8])
}

// buildPaymentMemo constructs a memo string for the treasury transfer.
func buildPaymentMemo(record *models.PaymentRecord) string {
	return fmt.Sprintf("Federal payment disbursement for payment ID: %s", record.PaymentID)
}
