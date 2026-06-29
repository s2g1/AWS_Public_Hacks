package correspondence

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"pgregory.net/rapid"
)

// mockBedrockClientForProperty returns valid JSON correspondence responses.
type mockBedrockClientForProperty struct{}

func (m *mockBedrockClientForProperty) InvokeModel(_ context.Context, _ string, payload []byte) ([]byte, error) {
	// Parse the incoming request to extract details for a realistic response
	var req BedrockRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		return nil, fmt.Errorf("failed to parse request: %w", err)
	}

	// Return a valid Bedrock response with correspondence JSON
	responseJSON := `{"subject": "Payment Processing Notification", "body": "Dear recipient, this is regarding your payment submission. Please review the details and contact us if you have questions."}`

	bedrockResp := BedrockResponse{
		Content: []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		}{
			{Type: "text", Text: responseJSON},
		},
	}

	return json.Marshal(bedrockResp)
}

// genCorrespondenceType generates a random valid CorrespondenceType.
func genCorrespondenceType() *rapid.Generator[CorrespondenceType] {
	return rapid.SampledFrom([]CorrespondenceType{
		CorrespondenceTypeApprovalConfirmation,
		CorrespondenceTypeRejectionNotice,
		CorrespondenceTypeREAResponse,
		CorrespondenceTypeEscalationNotification,
	})
}

// genOutputFormat generates a random valid OutputFormat.
func genOutputFormat() *rapid.Generator[OutputFormat] {
	return rapid.SampledFrom([]OutputFormat{
		OutputFormatEmailHTML,
		OutputFormatPDFContent,
		OutputFormatPortalNotification,
	})
}

// genNonEmptyString generates a non-empty string suitable for field values.
func genNonEmptyString() *rapid.Generator[string] {
	return rapid.StringMatching(`[a-zA-Z0-9@._\-]{1,50}`)
}

func TestProperty_ValidRequestNeverReturnsError(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		corrType := genCorrespondenceType().Draw(t, "type")
		format := genOutputFormat().Draw(t, "format")
		recipient := genNonEmptyString().Draw(t, "recipient")
		paymentID := genNonEmptyString().Draw(t, "paymentId")
		contractNumber := genNonEmptyString().Draw(t, "contractNumber")
		amount := genNonEmptyString().Draw(t, "amount")

		req := CorrespondenceRequest{
			Type:           corrType,
			Recipient:      recipient,
			PaymentID:      paymentID,
			ContractNumber: contractNumber,
			Amount:         amount,
			Reason:         "Test reason for correspondence generation",
		}

		generator := NewCorrespondenceGenerator(&mockBedrockClientForProperty{}, "test-model")
		result, err := generator.GenerateCorrespondence(context.Background(), req, format)

		if err != nil {
			t.Fatalf("expected no error for valid request, got: %v", err)
		}
		if result == nil {
			t.Fatal("expected non-nil result for valid request")
		}
	})
}

func TestProperty_GeneratedCorrespondenceHasNonEmptySubjectAndBody(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		corrType := genCorrespondenceType().Draw(t, "type")
		format := genOutputFormat().Draw(t, "format")
		recipient := genNonEmptyString().Draw(t, "recipient")
		paymentID := genNonEmptyString().Draw(t, "paymentId")

		req := CorrespondenceRequest{
			Type:           corrType,
			Recipient:      recipient,
			PaymentID:      paymentID,
			ContractNumber: "CONTRACT-001",
			Amount:         "$10000",
			Reason:         "Automated test",
		}

		generator := NewCorrespondenceGenerator(&mockBedrockClientForProperty{}, "test-model")
		result, err := generator.GenerateCorrespondence(context.Background(), req, format)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.Subject == "" {
			t.Fatal("generated correspondence must have non-empty Subject")
		}
		if result.Body == "" {
			t.Fatal("generated correspondence must have non-empty Body")
		}
	})
}

func TestProperty_FreshlyGeneratedStatusIsDraft(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		corrType := genCorrespondenceType().Draw(t, "type")
		format := genOutputFormat().Draw(t, "format")
		recipient := genNonEmptyString().Draw(t, "recipient")
		paymentID := genNonEmptyString().Draw(t, "paymentId")

		req := CorrespondenceRequest{
			Type:           corrType,
			Recipient:      recipient,
			PaymentID:      paymentID,
			ContractNumber: "C-123",
			Amount:         "$5000",
			Reason:         "Property test",
		}

		generator := NewCorrespondenceGenerator(&mockBedrockClientForProperty{}, "test-model")
		result, err := generator.GenerateCorrespondence(context.Background(), req, format)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.Status != CorrespondenceStatusDraft {
			t.Fatalf("expected status DRAFT for freshly generated correspondence, got: %s", result.Status)
		}
	})
}
