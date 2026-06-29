package coordinator

import (
	"context"
	"fmt"
	"log"

	"federal-payment-processing/internal/models"
)

// WorkflowStage represents a named stage in the payment processing pipeline.
type WorkflowStage string

const (
	StageExtraction  WorkflowStage = "extraction"
	StageValidation  WorkflowStage = "validation"
	StageCompliance  WorkflowStage = "compliance"
	StageRouting     WorkflowStage = "routing"
	StageApproval    WorkflowStage = "approval"
	StageDisbursement WorkflowStage = "disbursement"
)

// ExtractionFunc is the function type for the extraction agent.
// It receives the document path and returns the extraction result.
type ExtractionFunc func(ctx context.Context, documentPath string) (*models.ExtractionResult, error)

// ValidationFunc is the function type for the validation agent.
// It receives the extraction result and returns the validation result.
type ValidationFunc func(ctx context.Context, extractionResult *models.ExtractionResult) (*models.ValidationResult, error)

// ComplianceFunc is the function type for the compliance agent.
// It receives the payment record and returns the compliance result.
type ComplianceFunc func(ctx context.Context, record *models.PaymentRecord) (*models.ComplianceResult, error)

// RoutingFunc is the function type for the routing agent.
// It receives the payment record and returns the routing decision.
type RoutingFunc func(ctx context.Context, record *models.PaymentRecord) (*models.RoutingDecision, error)

// DisbursementFunc is the function type for the disbursement agent.
// It receives the payment record and returns the disbursement result.
type DisbursementFunc func(ctx context.Context, record *models.PaymentRecord) (*models.DisbursementResult, error)

// WorkflowCoordinator orchestrates the full payment processing pipeline,
// managing state transitions and escalation at each stage.
type WorkflowCoordinator struct {
	Extraction   ExtractionFunc
	Validation   ValidationFunc
	Compliance   ComplianceFunc
	Routing      RoutingFunc
	Disbursement DisbursementFunc
}

// NewWorkflowCoordinator creates a new WorkflowCoordinator with the given agent functions.
func NewWorkflowCoordinator(
	extraction ExtractionFunc,
	validation ValidationFunc,
	compliance ComplianceFunc,
	routing RoutingFunc,
	disbursement DisbursementFunc,
) *WorkflowCoordinator {
	return &WorkflowCoordinator{
		Extraction:   extraction,
		Validation:   validation,
		Compliance:   compliance,
		Routing:      routing,
		Disbursement: disbursement,
	}
}

// ExecuteWorkflow runs the full payment processing pipeline from RECEIVED through DISBURSED.
// It transitions the payment through each stage, invoking agents and handling escalation/rejection
// at each step. The workflow halts on escalation, rejection, or error.
func (wc *WorkflowCoordinator) ExecuteWorkflow(ctx context.Context, record *models.PaymentRecord, documentPath string) error {
	if record == nil {
		return fmt.Errorf("payment record must not be nil")
	}

	// Stage 1: Extraction
	if err := wc.executeExtraction(ctx, record, documentPath); err != nil {
		return err
	}
	if record.Status == models.PaymentStatusEscalated {
		return nil
	}

	// Stage 2: Validation
	if err := wc.executeValidation(ctx, record); err != nil {
		return err
	}
	if record.Status == models.PaymentStatusRejected || record.Status == models.PaymentStatusEscalated {
		return nil
	}

	// Stage 3: Compliance
	if err := wc.executeCompliance(ctx, record); err != nil {
		return err
	}
	if record.Status == models.PaymentStatusRejected {
		return nil
	}

	// Stage 4: Routing
	if err := wc.executeRouting(ctx, record); err != nil {
		return err
	}

	// Stage 5: Approval (simulated for hackathon)
	if err := wc.executeApproval(ctx, record); err != nil {
		return err
	}

	// Stage 6: Disbursement
	if err := wc.executeDisbursement(ctx, record); err != nil {
		return err
	}

	return nil
}

// executeExtraction transitions RECEIVED → EXTRACTING, invokes extraction, and checks confidence.
func (wc *WorkflowCoordinator) executeExtraction(ctx context.Context, record *models.PaymentRecord, documentPath string) error {
	log.Printf("[Workflow] payment=%s transitioning to EXTRACTING", record.PaymentID)
	if err := models.TransitionPayment(record, models.PaymentStatusExtracting, "coordinator", "Starting extraction"); err != nil {
		return fmt.Errorf("failed to transition to EXTRACTING: %w", err)
	}

	result, err := wc.Extraction(ctx, documentPath)
	if err != nil {
		return fmt.Errorf("extraction agent failed: %w", err)
	}

	record.ExtractedData = result

	// Transition to EXTRACTED
	if err := models.TransitionPayment(record, models.PaymentStatusExtracted, "extraction-agent", "Extraction complete"); err != nil {
		return fmt.Errorf("failed to transition to EXTRACTED: %w", err)
	}

	// Check confidence threshold — may escalate
	escalated, err := CheckExtractionConfidence(record, result)
	if err != nil {
		return fmt.Errorf("confidence check failed: %w", err)
	}
	if escalated {
		log.Printf("[Workflow] payment=%s escalated due to low extraction confidence (%.2f)", record.PaymentID, result.OverallConfidence)
		return nil
	}

	return nil
}

// executeValidation transitions EXTRACTED → VALIDATING, invokes validation, handles rejection/escalation.
func (wc *WorkflowCoordinator) executeValidation(ctx context.Context, record *models.PaymentRecord) error {
	log.Printf("[Workflow] payment=%s transitioning to VALIDATING", record.PaymentID)
	if err := models.TransitionPayment(record, models.PaymentStatusValidating, "coordinator", "Starting validation"); err != nil {
		return fmt.Errorf("failed to transition to VALIDATING: %w", err)
	}

	result, err := wc.Validation(ctx, record.ExtractedData)
	if err != nil {
		return fmt.Errorf("validation agent failed: %w", err)
	}

	record.ValidationResult = result

	switch result.Status {
	case models.ValidationStatusRejected:
		log.Printf("[Workflow] payment=%s rejected by validation", record.PaymentID)
		if err := models.TransitionPayment(record, models.PaymentStatusRejected, "validation-agent", "Validation rejected: critical issues found"); err != nil {
			return fmt.Errorf("failed to transition to REJECTED: %w", err)
		}
		return nil

	case models.ValidationStatusNeedsReview:
		log.Printf("[Workflow] payment=%s escalated for validation review", record.PaymentID)
		if err := models.TransitionPayment(record, models.PaymentStatusEscalated, "validation-agent", "Validation needs human review"); err != nil {
			return fmt.Errorf("failed to transition to ESCALATED: %w", err)
		}
		return nil

	default:
		// VALID — proceed
		if err := models.TransitionPayment(record, models.PaymentStatusValidated, "validation-agent", "Validation passed"); err != nil {
			return fmt.Errorf("failed to transition to VALIDATED: %w", err)
		}
	}

	return nil
}

// executeCompliance transitions VALIDATED → CHECKING_COMPLIANCE, invokes compliance, handles rejection.
func (wc *WorkflowCoordinator) executeCompliance(ctx context.Context, record *models.PaymentRecord) error {
	log.Printf("[Workflow] payment=%s transitioning to CHECKING_COMPLIANCE", record.PaymentID)
	if err := models.TransitionPayment(record, models.PaymentStatusCheckingCompliance, "coordinator", "Starting compliance check"); err != nil {
		return fmt.Errorf("failed to transition to CHECKING_COMPLIANCE: %w", err)
	}

	result, err := wc.Compliance(ctx, record)
	if err != nil {
		return fmt.Errorf("compliance agent failed: %w", err)
	}

	record.ComplianceResult = result

	if result.Status == models.ComplianceStatusNonCompliant {
		log.Printf("[Workflow] payment=%s rejected by compliance (NON_COMPLIANT)", record.PaymentID)
		if err := models.TransitionPayment(record, models.PaymentStatusRejected, "compliance-agent", "Non-compliant: blocking flags found"); err != nil {
			return fmt.Errorf("failed to transition to REJECTED: %w", err)
		}
		return nil
	}

	// COMPLIANT or COMPLIANT_WITH_CONDITIONS — proceed
	if err := models.TransitionPayment(record, models.PaymentStatusCompliant, "compliance-agent", "Compliance check passed"); err != nil {
		return fmt.Errorf("failed to transition to COMPLIANT: %w", err)
	}

	return nil
}

// executeRouting transitions COMPLIANT → ROUTING → ROUTED.
func (wc *WorkflowCoordinator) executeRouting(ctx context.Context, record *models.PaymentRecord) error {
	log.Printf("[Workflow] payment=%s transitioning to ROUTING", record.PaymentID)
	if err := models.TransitionPayment(record, models.PaymentStatusRouting, "coordinator", "Starting routing"); err != nil {
		return fmt.Errorf("failed to transition to ROUTING: %w", err)
	}

	decision, err := wc.Routing(ctx, record)
	if err != nil {
		return fmt.Errorf("routing agent failed: %w", err)
	}

	record.RoutingDecision = decision

	if err := models.TransitionPayment(record, models.PaymentStatusRouted, "routing-agent", fmt.Sprintf("Routed to %s", decision.ApprovalLevel)); err != nil {
		return fmt.Errorf("failed to transition to ROUTED: %w", err)
	}

	return nil
}

// executeApproval transitions ROUTED → APPROVING → APPROVED (simulated for hackathon).
func (wc *WorkflowCoordinator) executeApproval(ctx context.Context, record *models.PaymentRecord) error {
	log.Printf("[Workflow] payment=%s transitioning to APPROVING (simulated)", record.PaymentID)
	if err := models.TransitionPayment(record, models.PaymentStatusApproving, "coordinator", "Starting approval"); err != nil {
		return fmt.Errorf("failed to transition to APPROVING: %w", err)
	}

	// Simulated approval for hackathon — auto-approve
	if err := models.TransitionPayment(record, models.PaymentStatusApproved, "approval-agent", "Auto-approved (hackathon simulation)"); err != nil {
		return fmt.Errorf("failed to transition to APPROVED: %w", err)
	}

	return nil
}

// executeDisbursement transitions APPROVED → DISBURSING, invokes disbursement agent.
func (wc *WorkflowCoordinator) executeDisbursement(ctx context.Context, record *models.PaymentRecord) error {
	log.Printf("[Workflow] payment=%s transitioning to DISBURSING", record.PaymentID)
	if err := models.TransitionPayment(record, models.PaymentStatusDisbursing, "coordinator", "Starting disbursement"); err != nil {
		return fmt.Errorf("failed to transition to DISBURSING: %w", err)
	}

	result, err := wc.Disbursement(ctx, record)
	if err != nil {
		return fmt.Errorf("disbursement agent failed: %w", err)
	}

	record.DisbursementResult = result

	if result.Status == models.DisbursementStatusFailed {
		if err := models.TransitionPayment(record, models.PaymentStatusFailed, "disbursement-agent", fmt.Sprintf("Disbursement failed: %s", result.Reason)); err != nil {
			return fmt.Errorf("failed to transition to FAILED: %w", err)
		}
		return nil
	}

	if err := models.TransitionPayment(record, models.PaymentStatusDisbursed, "disbursement-agent", "Disbursement complete"); err != nil {
		return fmt.Errorf("failed to transition to DISBURSED: %w", err)
	}

	log.Printf("[Workflow] payment=%s completed successfully", record.PaymentID)
	return nil
}

// ResumeFromEscalation resumes the payment workflow from the point where it was escalated.
// The fromStage parameter indicates which stage the payment should resume from.
// Valid stages: "extraction", "validation", "compliance", "routing", "approval".
func (wc *WorkflowCoordinator) ResumeFromEscalation(ctx context.Context, record *models.PaymentRecord, fromStage string) error {
	if record == nil {
		return fmt.Errorf("payment record must not be nil")
	}
	if record.Status != models.PaymentStatusEscalated {
		return fmt.Errorf("cannot resume: payment status is %q, expected ESCALATED", record.Status)
	}

	stage := WorkflowStage(fromStage)

	switch stage {
	case StageExtraction:
		// Resume from extraction: re-run validation onward
		log.Printf("[Workflow] payment=%s resuming from extraction stage", record.PaymentID)
		if err := models.TransitionPayment(record, models.PaymentStatusValidating, "coordinator", "Resuming from escalation: extraction reviewed"); err != nil {
			return fmt.Errorf("failed to resume from extraction: %w", err)
		}
		return wc.resumeFromValidation(ctx, record)

	case StageValidation:
		// Resume from validation: re-run compliance onward
		log.Printf("[Workflow] payment=%s resuming from validation stage", record.PaymentID)
		if err := models.TransitionPayment(record, models.PaymentStatusCheckingCompliance, "coordinator", "Resuming from escalation: validation reviewed"); err != nil {
			return fmt.Errorf("failed to resume from validation: %w", err)
		}
		return wc.resumeFromCompliance(ctx, record)

	case StageCompliance:
		log.Printf("[Workflow] payment=%s resuming from compliance stage", record.PaymentID)
		if err := models.TransitionPayment(record, models.PaymentStatusRouting, "coordinator", "Resuming from escalation: compliance reviewed"); err != nil {
			return fmt.Errorf("failed to resume from compliance: %w", err)
		}
		return wc.resumeFromRouting(ctx, record)

	case StageRouting:
		log.Printf("[Workflow] payment=%s resuming from routing stage", record.PaymentID)
		if err := models.TransitionPayment(record, models.PaymentStatusApproving, "coordinator", "Resuming from escalation: routing reviewed"); err != nil {
			return fmt.Errorf("failed to resume from routing: %w", err)
		}
		return wc.resumeFromApproval(ctx, record)

	case StageApproval:
		log.Printf("[Workflow] payment=%s resuming from approval stage", record.PaymentID)
		if err := models.TransitionPayment(record, models.PaymentStatusDisbursing, "coordinator", "Resuming from escalation: approval reviewed"); err != nil {
			return fmt.Errorf("failed to resume from approval: %w", err)
		}
		return wc.resumeFromDisbursement(ctx, record)

	default:
		return fmt.Errorf("unknown escalation stage: %q", fromStage)
	}
}

// resumeFromValidation runs the pipeline from validation onward (after extraction escalation).
func (wc *WorkflowCoordinator) resumeFromValidation(ctx context.Context, record *models.PaymentRecord) error {
	result, err := wc.Validation(ctx, record.ExtractedData)
	if err != nil {
		return fmt.Errorf("validation agent failed: %w", err)
	}
	record.ValidationResult = result

	switch result.Status {
	case models.ValidationStatusRejected:
		return models.TransitionPayment(record, models.PaymentStatusRejected, "validation-agent", "Validation rejected: critical issues found")
	case models.ValidationStatusNeedsReview:
		return models.TransitionPayment(record, models.PaymentStatusEscalated, "validation-agent", "Validation needs human review")
	default:
		if err := models.TransitionPayment(record, models.PaymentStatusValidated, "validation-agent", "Validation passed"); err != nil {
			return err
		}
	}

	if record.Status == models.PaymentStatusRejected || record.Status == models.PaymentStatusEscalated {
		return nil
	}

	return wc.continueFromCompliance(ctx, record)
}

// resumeFromCompliance runs the pipeline from compliance onward.
func (wc *WorkflowCoordinator) resumeFromCompliance(ctx context.Context, record *models.PaymentRecord) error {
	result, err := wc.Compliance(ctx, record)
	if err != nil {
		return fmt.Errorf("compliance agent failed: %w", err)
	}
	record.ComplianceResult = result

	if result.Status == models.ComplianceStatusNonCompliant {
		return models.TransitionPayment(record, models.PaymentStatusRejected, "compliance-agent", "Non-compliant: blocking flags found")
	}

	if err := models.TransitionPayment(record, models.PaymentStatusCompliant, "compliance-agent", "Compliance check passed"); err != nil {
		return err
	}

	return wc.continueFromRouting(ctx, record)
}

// resumeFromRouting runs the pipeline from routing onward.
func (wc *WorkflowCoordinator) resumeFromRouting(ctx context.Context, record *models.PaymentRecord) error {
	decision, err := wc.Routing(ctx, record)
	if err != nil {
		return fmt.Errorf("routing agent failed: %w", err)
	}
	record.RoutingDecision = decision

	if err := models.TransitionPayment(record, models.PaymentStatusRouted, "routing-agent", fmt.Sprintf("Routed to %s", decision.ApprovalLevel)); err != nil {
		return err
	}

	return wc.continueFromApproval(ctx, record)
}

// resumeFromApproval runs the pipeline from approval onward.
func (wc *WorkflowCoordinator) resumeFromApproval(ctx context.Context, record *models.PaymentRecord) error {
	if err := models.TransitionPayment(record, models.PaymentStatusApproved, "approval-agent", "Auto-approved (hackathon simulation)"); err != nil {
		return err
	}
	return wc.continueFromDisbursement(ctx, record)
}

// resumeFromDisbursement runs the disbursement stage.
func (wc *WorkflowCoordinator) resumeFromDisbursement(ctx context.Context, record *models.PaymentRecord) error {
	result, err := wc.Disbursement(ctx, record)
	if err != nil {
		return fmt.Errorf("disbursement agent failed: %w", err)
	}
	record.DisbursementResult = result

	if result.Status == models.DisbursementStatusFailed {
		return models.TransitionPayment(record, models.PaymentStatusFailed, "disbursement-agent", fmt.Sprintf("Disbursement failed: %s", result.Reason))
	}

	return models.TransitionPayment(record, models.PaymentStatusDisbursed, "disbursement-agent", "Disbursement complete")
}

// continueFromCompliance runs compliance and the remaining stages.
func (wc *WorkflowCoordinator) continueFromCompliance(ctx context.Context, record *models.PaymentRecord) error {
	if err := wc.executeCompliance(ctx, record); err != nil {
		return err
	}
	if record.Status == models.PaymentStatusRejected {
		return nil
	}
	if err := wc.executeRouting(ctx, record); err != nil {
		return err
	}
	if err := wc.executeApproval(ctx, record); err != nil {
		return err
	}
	return wc.executeDisbursement(ctx, record)
}

// continueFromRouting runs routing and the remaining stages.
func (wc *WorkflowCoordinator) continueFromRouting(ctx context.Context, record *models.PaymentRecord) error {
	if err := wc.executeRouting(ctx, record); err != nil {
		return err
	}
	if err := wc.executeApproval(ctx, record); err != nil {
		return err
	}
	return wc.executeDisbursement(ctx, record)
}

// continueFromApproval runs approval and the remaining stages.
func (wc *WorkflowCoordinator) continueFromApproval(ctx context.Context, record *models.PaymentRecord) error {
	if err := wc.executeApproval(ctx, record); err != nil {
		return err
	}
	return wc.executeDisbursement(ctx, record)
}

// continueFromDisbursement runs the disbursement stage.
func (wc *WorkflowCoordinator) continueFromDisbursement(ctx context.Context, record *models.PaymentRecord) error {
	return wc.executeDisbursement(ctx, record)
}
