package coordinator

import (
	"context"
	"fmt"
	"testing"
	"time"

	"federal-payment-processing/internal/models"
)

// --- Test Helpers ---

func newTestRecord() *models.PaymentRecord {
	return &models.PaymentRecord{
		PaymentID: "PAY-001",
		Status:    models.PaymentStatusReceived,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func successExtraction(_ context.Context, _ string) (*models.ExtractionResult, error) {
	return &models.ExtractionResult{
		DocumentType:      models.DocumentTypeInvoice,
		OverallConfidence: 0.95,
		Fields: map[string]models.ExtractedField{
			"payee":         {Value: "Acme Corp", Confidence: 0.95},
			"amount":        {Value: "$1000.00", Confidence: 0.98},
			"invoiceNumber": {Value: "INV-001", Confidence: 0.96},
			"date":          {Value: "2024-01-15", Confidence: 0.97},
		},
	}, nil
}

func lowConfidenceExtraction(_ context.Context, _ string) (*models.ExtractionResult, error) {
	return &models.ExtractionResult{
		DocumentType:      models.DocumentTypeInvoice,
		OverallConfidence: 0.50,
		Fields: map[string]models.ExtractedField{
			"payee":  {Value: "Blurry Corp", Confidence: 0.50},
			"amount": {Value: "$???", Confidence: 0.30},
		},
	}, nil
}

func successValidation(_ context.Context, _ *models.ExtractionResult) (*models.ValidationResult, error) {
	return &models.ValidationResult{
		Status:      models.ValidationStatusValid,
		Issues:      nil,
		ValidatedAt: time.Now(),
	}, nil
}

func rejectedValidation(_ context.Context, _ *models.ExtractionResult) (*models.ValidationResult, error) {
	return &models.ValidationResult{
		Status: models.ValidationStatusRejected,
		Issues: []models.ValidationIssue{
			{Severity: models.SeverityCritical, Field: "amount", Message: "Missing required field"},
		},
		ValidatedAt: time.Now(),
	}, nil
}

func needsReviewValidation(_ context.Context, _ *models.ExtractionResult) (*models.ValidationResult, error) {
	return &models.ValidationResult{
		Status: models.ValidationStatusNeedsReview,
		Issues: []models.ValidationIssue{
			{Severity: models.SeverityError, Field: "date", Message: "Invalid date format"},
		},
		ValidatedAt: time.Now(),
	}, nil
}

func successCompliance(_ context.Context, _ *models.PaymentRecord) (*models.ComplianceResult, error) {
	return &models.ComplianceResult{
		Status:    models.ComplianceStatusCompliant,
		Flags:     nil,
		CheckedAt: time.Now(),
	}, nil
}

func nonCompliant(_ context.Context, _ *models.PaymentRecord) (*models.ComplianceResult, error) {
	return &models.ComplianceResult{
		Status: models.ComplianceStatusNonCompliant,
		Flags: []models.ComplianceFlag{
			{Rule: "OFAC_SANCTIONS", Severity: models.FlagSeverityBlocking, Message: "Payee matched sanctions list"},
		},
		CheckedAt: time.Now(),
	}, nil
}

func successRouting(_ context.Context, _ *models.PaymentRecord) (*models.RoutingDecision, error) {
	return &models.RoutingDecision{
		Status:        models.RoutingStatusRouted,
		Approver:      "supervisor@agency.gov",
		ApprovalLevel: models.ApprovalLevelSupervisor,
		Priority:      models.PriorityNormal,
		Rationale:     "Amount within supervisor threshold",
		RoutedAt:      time.Now(),
	}, nil
}

func successDisbursement(_ context.Context, _ *models.PaymentRecord) (*models.DisbursementResult, error) {
	return &models.DisbursementResult{
		Status: models.DisbursementStatusSuccess,
		Confirmation: &models.PaymentConfirmation{
			TransactionID: "TXN-001",
			Amount:        1000.00,
			Payee:         "Acme Corp",
			DisbursedAt:   time.Now(),
			Reference:     "REF-001",
		},
	}, nil
}

func failedDisbursement(_ context.Context, _ *models.PaymentRecord) (*models.DisbursementResult, error) {
	return &models.DisbursementResult{
		Status:    models.DisbursementStatusFailed,
		Reason:    "Insufficient funds in treasury account",
		Retryable: true,
	}, nil
}

// --- Tests ---

func TestExecuteWorkflow_HappyPath(t *testing.T) {
	wc := NewWorkflowCoordinator(
		successExtraction,
		successValidation,
		successCompliance,
		successRouting,
		successDisbursement,
	)

	record := newTestRecord()
	err := wc.ExecuteWorkflow(context.Background(), record, "s3://bucket/doc.pdf")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if record.Status != models.PaymentStatusDisbursed {
		t.Errorf("expected status DISBURSED, got %s", record.Status)
	}
	if record.ExtractedData == nil {
		t.Error("expected ExtractedData to be set")
	}
	if record.ValidationResult == nil {
		t.Error("expected ValidationResult to be set")
	}
	if record.ComplianceResult == nil {
		t.Error("expected ComplianceResult to be set")
	}
	if record.RoutingDecision == nil {
		t.Error("expected RoutingDecision to be set")
	}
	if record.DisbursementResult == nil {
		t.Error("expected DisbursementResult to be set")
	}

	// Verify audit trail covers all transitions
	if len(record.AuditTrail) < 10 {
		t.Errorf("expected at least 10 audit entries for full pipeline, got %d", len(record.AuditTrail))
	}
}

func TestExecuteWorkflow_EscalationAtExtraction(t *testing.T) {
	wc := NewWorkflowCoordinator(
		lowConfidenceExtraction,
		successValidation,
		successCompliance,
		successRouting,
		successDisbursement,
	)

	record := newTestRecord()
	err := wc.ExecuteWorkflow(context.Background(), record, "s3://bucket/blurry.pdf")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if record.Status != models.PaymentStatusEscalated {
		t.Errorf("expected status ESCALATED, got %s", record.Status)
	}
	// Validation should NOT have been invoked
	if record.ValidationResult != nil {
		t.Error("expected ValidationResult to be nil when escalated at extraction")
	}
}

func TestExecuteWorkflow_RejectionAtValidation(t *testing.T) {
	wc := NewWorkflowCoordinator(
		successExtraction,
		rejectedValidation,
		successCompliance,
		successRouting,
		successDisbursement,
	)

	record := newTestRecord()
	err := wc.ExecuteWorkflow(context.Background(), record, "s3://bucket/doc.pdf")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if record.Status != models.PaymentStatusRejected {
		t.Errorf("expected status REJECTED, got %s", record.Status)
	}
	// Compliance should NOT have been invoked
	if record.ComplianceResult != nil {
		t.Error("expected ComplianceResult to be nil when rejected at validation")
	}
}

func TestExecuteWorkflow_NeedsReviewAtValidation(t *testing.T) {
	wc := NewWorkflowCoordinator(
		successExtraction,
		needsReviewValidation,
		successCompliance,
		successRouting,
		successDisbursement,
	)

	record := newTestRecord()
	err := wc.ExecuteWorkflow(context.Background(), record, "s3://bucket/doc.pdf")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if record.Status != models.PaymentStatusEscalated {
		t.Errorf("expected status ESCALATED, got %s", record.Status)
	}
	if record.ComplianceResult != nil {
		t.Error("expected ComplianceResult to be nil when escalated at validation")
	}
}

func TestExecuteWorkflow_RejectionAtCompliance(t *testing.T) {
	wc := NewWorkflowCoordinator(
		successExtraction,
		successValidation,
		nonCompliant,
		successRouting,
		successDisbursement,
	)

	record := newTestRecord()
	err := wc.ExecuteWorkflow(context.Background(), record, "s3://bucket/doc.pdf")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if record.Status != models.PaymentStatusRejected {
		t.Errorf("expected status REJECTED, got %s", record.Status)
	}
	// Routing should NOT have been invoked
	if record.RoutingDecision != nil {
		t.Error("expected RoutingDecision to be nil when rejected at compliance")
	}
}

func TestExecuteWorkflow_DisbursementFailure(t *testing.T) {
	wc := NewWorkflowCoordinator(
		successExtraction,
		successValidation,
		successCompliance,
		successRouting,
		failedDisbursement,
	)

	record := newTestRecord()
	err := wc.ExecuteWorkflow(context.Background(), record, "s3://bucket/doc.pdf")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if record.Status != models.PaymentStatusFailed {
		t.Errorf("expected status FAILED, got %s", record.Status)
	}
}

func TestExecuteWorkflow_NilRecord(t *testing.T) {
	wc := NewWorkflowCoordinator(
		successExtraction,
		successValidation,
		successCompliance,
		successRouting,
		successDisbursement,
	)

	err := wc.ExecuteWorkflow(context.Background(), nil, "s3://bucket/doc.pdf")
	if err == nil {
		t.Fatal("expected error for nil record")
	}
}

func TestExecuteWorkflow_ExtractionAgentError(t *testing.T) {
	failingExtraction := func(_ context.Context, _ string) (*models.ExtractionResult, error) {
		return nil, fmt.Errorf("bedrock timeout")
	}

	wc := NewWorkflowCoordinator(
		failingExtraction,
		successValidation,
		successCompliance,
		successRouting,
		successDisbursement,
	)

	record := newTestRecord()
	err := wc.ExecuteWorkflow(context.Background(), record, "s3://bucket/doc.pdf")

	if err == nil {
		t.Fatal("expected error when extraction agent fails")
	}
}

func TestResumeFromEscalation_Extraction(t *testing.T) {
	wc := NewWorkflowCoordinator(
		successExtraction,
		successValidation,
		successCompliance,
		successRouting,
		successDisbursement,
	)

	// Simulate a payment that was escalated during extraction
	record := &models.PaymentRecord{
		PaymentID: "PAY-002",
		Status:    models.PaymentStatusEscalated,
		ExtractedData: &models.ExtractionResult{
			DocumentType:      models.DocumentTypeInvoice,
			OverallConfidence: 0.90, // human corrected
			Fields: map[string]models.ExtractedField{
				"payee":         {Value: "Acme Corp", Confidence: 0.95},
				"amount":        {Value: "$5000.00", Confidence: 0.90},
				"invoiceNumber": {Value: "INV-002", Confidence: 0.92},
				"date":          {Value: "2024-02-01", Confidence: 0.95},
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := wc.ResumeFromEscalation(context.Background(), record, "extraction")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if record.Status != models.PaymentStatusDisbursed {
		t.Errorf("expected status DISBURSED after resuming from extraction, got %s", record.Status)
	}
}

func TestResumeFromEscalation_Validation(t *testing.T) {
	wc := NewWorkflowCoordinator(
		successExtraction,
		successValidation,
		successCompliance,
		successRouting,
		successDisbursement,
	)

	record := &models.PaymentRecord{
		PaymentID: "PAY-003",
		Status:    models.PaymentStatusEscalated,
		ExtractedData: &models.ExtractionResult{
			DocumentType:      models.DocumentTypeInvoice,
			OverallConfidence: 0.95,
			Fields: map[string]models.ExtractedField{
				"payee":  {Value: "Acme Corp", Confidence: 0.95},
				"amount": {Value: "$5000.00", Confidence: 0.90},
			},
		},
		ValidationResult: &models.ValidationResult{
			Status:      models.ValidationStatusValid,
			ValidatedAt: time.Now(),
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := wc.ResumeFromEscalation(context.Background(), record, "validation")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if record.Status != models.PaymentStatusDisbursed {
		t.Errorf("expected status DISBURSED after resuming from validation, got %s", record.Status)
	}
}

func TestResumeFromEscalation_InvalidStatus(t *testing.T) {
	wc := NewWorkflowCoordinator(
		successExtraction,
		successValidation,
		successCompliance,
		successRouting,
		successDisbursement,
	)

	record := &models.PaymentRecord{
		PaymentID: "PAY-004",
		Status:    models.PaymentStatusReceived, // not escalated
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := wc.ResumeFromEscalation(context.Background(), record, "extraction")
	if err == nil {
		t.Fatal("expected error when record is not in ESCALATED status")
	}
}

func TestResumeFromEscalation_UnknownStage(t *testing.T) {
	wc := NewWorkflowCoordinator(
		successExtraction,
		successValidation,
		successCompliance,
		successRouting,
		successDisbursement,
	)

	record := &models.PaymentRecord{
		PaymentID: "PAY-005",
		Status:    models.PaymentStatusEscalated,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := wc.ResumeFromEscalation(context.Background(), record, "unknown_stage")
	if err == nil {
		t.Fatal("expected error for unknown stage")
	}
}

func TestResumeFromEscalation_NilRecord(t *testing.T) {
	wc := NewWorkflowCoordinator(
		successExtraction,
		successValidation,
		successCompliance,
		successRouting,
		successDisbursement,
	)

	err := wc.ResumeFromEscalation(context.Background(), nil, "extraction")
	if err == nil {
		t.Fatal("expected error for nil record")
	}
}
