# Implementation Plan: Federal Payment Processing Agentic AI Platform

## Overview

This implementation plan breaks down the Federal Payment Processing Platform into incremental coding tasks using Go for the backend agent pipeline, AWS CDK (TypeScript) for infrastructure, and React (TypeScript) with Tailwind CSS for the frontend portal. Each task builds on previous steps and ends with fully wired, integrated code. The platform orchestrates five specialized AI agents through AWS Step Functions, with a Contract Financial Management Portal providing unified financial visibility.

## Tasks

- [x] 1. Set up project structure, shared types, and core interfaces
  - [x] 1.1 Initialize Go module and project directory structure
    - Create Go module with `cmd/`, `internal/`, `pkg/` layout
    - Directories: `internal/agents/`, `internal/models/`, `internal/coordinator/`, `internal/portal/`, `pkg/messaging/`, `pkg/config/`
    - Set up `go.mod` with AWS SDK v2 dependencies (dynamodb, s3, sfn, bedrockruntime, lambda)
    - _Requirements: 13.1, 15.4_

  - [x] 1.2 Define core data models and enumerations in Go
    - Implement `PaymentRecord`, `PaymentStatus` enum with valid state transitions map
    - Implement `ExtractionResult`, `ExtractedField`, `DocumentType` enum
    - Implement `AgentMessage`, `Decision` enum, `TraceContext`
    - Implement `ValidationResult`, `ValidationIssue`, `Severity` enum
    - Implement `ComplianceResult`, `ComplianceFlag`, `ComplianceStatus` enum
    - Implement `RoutingDecision`, `ApprovalLevel`, `Priority` enum
    - Implement `DisbursementResult`, `PaymentConfirmation`
    - Implement `AuditEntry` struct with timestamp, actor, previous/new status, reason
    - _Requirements: 1.2, 1.3, 13.1, 14.1, 14.2_

  - [x] 1.3 Implement the payment state machine with transition validation
    - Define allowed transitions map: `RECEIVED → EXTRACTING → EXTRACTED → ...`
    - Implement `ValidateTransition(current, next PaymentStatus) error` function
    - Implement `TransitionPayment(record *PaymentRecord, newStatus PaymentStatus, actor string, reason string) error` that validates + appends audit entry
    - Reject invalid transitions with descriptive errors
    - _Requirements: 13.1, 13.2, 13.3_

  - [x] 1.4 Write property tests for payment state machine integrity
    - **Property 1: Payment State Machine Integrity**
    - **Validates: Requirements 13.1, 13.3**

  - [x] 1.5 Write property tests for audit trail completeness
    - **Property 2: Audit Trail Completeness**
    - **Validates: Requirements 14.1, 14.2, 14.4**

- [x] 2. Implement Document Extraction Agent
  - [x] 2.1 Implement the OCR/Extraction Agent Lambda handler
    - Create `internal/agents/extraction/handler.go` with Lambda handler
    - Implement document classification using Bedrock Claude Sonnet
    - Implement field extraction with per-field confidence scoring
    - Implement required field checking per document type (INVOICE, PURCHASE_ORDER, TRAVEL_VOUCHER, GRANT_PAYMENT, CONTRACT_PAYMENT)
    - Set overall confidence to minimum of all field confidences
    - Set missing required fields to confidence 0.0 and overall to 0.0
    - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 1.6, 1.7_

  - [x] 2.2 Write property test for overall confidence calculation
    - **Property 3: Overall Confidence is Minimum Field Confidence**
    - **Validates: Requirements 1.3, 1.7**

  - [x] 2.3 Implement confidence threshold escalation logic in coordinator
    - In `internal/coordinator/`, implement escalation check after extraction
    - If overall confidence < EXTRACTION_THRESHOLD (0.75), set status to ESCALATED
    - Record escalation reason, agent name, and timestamp in audit trail
    - Halt further automated processing for escalated payments
    - _Requirements: 2.1, 2.2, 2.4_

  - [x] 2.4 Write property test for confidence threshold escalation
    - **Property 4: Confidence Threshold Escalation**
    - **Validates: Requirements 2.1, 2.2**

- [~] 3. Checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [x] 4. Implement Validation Agent
  - [x] 4.1 Implement the Validation Agent Lambda handler
    - Create `internal/agents/validation/handler.go`
    - Implement completeness check: verify required fields present with confidence >= FIELD_CONFIDENCE_THRESHOLD (0.80)
    - Implement format validation: currency format + positive amount check
    - Implement date validation: valid format + future date warning
    - Determine status: CRITICAL → REJECTED, ERROR → NEEDS_REVIEW, else VALID
    - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5, 3.6_

  - [x] 4.2 Write property test for validation status determination
    - **Property 5: Validation Status Determination**
    - **Validates: Requirements 3.4, 3.5, 3.6**

  - [x] 4.3 Implement duplicate payment detection
    - Query DynamoDB for payments matching same payee, amount, date within 30-day lookback
    - Add WARNING severity issue with matching payment ID reference
    - Route duplicates to human reviewer (do not auto-reject)
    - _Requirements: 4.1, 4.2, 4.3_

  - [x] 4.4 Implement payee verification against registry
    - Cross-reference payee against registered payee registry (DynamoDB table)
    - Add WARNING severity issue if payee not found in registry
    - _Requirements: 5.1, 5.2_

- [x] 5. Implement Compliance Agent
  - [x] 5.1 Implement OFAC sanctions screening
    - Create `internal/agents/compliance/handler.go`
    - Implement fuzzy matching against OFAC sanctions list (threshold 0.85)
    - On match: set BLOCKING flag with rule OFAC_SANCTIONS, return NON_COMPLIANT immediately
    - Set Payment_Record status to REJECTED and generate security alert
    - _Requirements: 6.1, 6.2, 6.3_

  - [x] 5.2 Implement debarment screening
    - Check payee against federal debarment list
    - On match: set BLOCKING flag with rule DEBARMENT, return NON_COMPLIANT
    - _Requirements: 7.1, 7.2_

  - [x] 5.3 Implement spending threshold checks
    - Check payment amount against single transaction maximum for spend category
    - Calculate cumulative spend for payee in current fiscal year
    - Set REQUIRES_REVIEW flags for THRESHOLD_EXCEEDED and ANNUAL_LIMIT
    - _Requirements: 8.1, 8.2, 8.3, 8.4_

  - [x] 5.4 Implement FAR compliance evaluation using Bedrock
    - Build FAR check prompt with payment details, category, and amount
    - Invoke Bedrock Claude Sonnet for rule evaluation
    - Parse response into compliance flags with appropriate severity
    - _Requirements: 9.1, 9.2_

  - [x] 5.5 Implement compliance status determination logic
    - BLOCKING flags → NON_COMPLIANT
    - REQUIRES_REVIEW (no BLOCKING) → COMPLIANT_WITH_CONDITIONS
    - No flags → COMPLIANT
    - _Requirements: 9.3, 9.4, 9.5_

  - [x] 5.6 Write property test for compliance blocking enforcement
    - **Property 6: Compliance Blocking Enforcement**
    - **Validates: Requirements 6.2, 7.2, 9.3**

  - [x] 5.7 Write property test for compliance status determination
    - **Property 7: Compliance Status Determination**
    - **Validates: Requirements 9.3, 9.4, 9.5**

  - [x] 5.8 Write property test for spending threshold flagging
    - **Property 28: Spending Threshold Flagging**
    - **Validates: Requirements 8.2, 8.4**

- [x] 6. Implement Routing Agent
  - [x] 6.1 Implement amount-based routing logic
    - Create `internal/agents/routing/handler.go`
    - Implement threshold-based approval level assignment:
      - ≤ $2,500 → PURCHASE_CARD / LOW
      - ≤ $25,000 → SUPERVISOR / NORMAL
      - ≤ $250,000 → CONTRACTING_OFFICER / NORMAL
      - ≤ $1,000,000 → SENIOR_CONTRACTING_OFFICER / HIGH
      - > $1,000,000 → AGENCY_HEAD / URGENT
    - _Requirements: 10.1, 10.2, 10.3, 10.4, 10.5_

  - [x] 6.2 Implement compliance condition elevation and urgency override
    - If COMPLIANT_WITH_CONDITIONS: elevate approval level by one tier, increase priority by one level
    - If payment due date within 3 days: set priority to URGENT regardless of amount
    - _Requirements: 10.6, 11.3_

  - [x] 6.3 Implement delegation of authority fallback
    - Check if assigned approver is on leave or has expired delegation
    - Route to designated delegate if available
    - If no delegate: set status ESCALATED with URGENT priority, notify admin
    - _Requirements: 11.1, 11.2_

  - [x] 6.4 Write property test for routing authority matches amount
    - **Property 8: Routing Authority Matches Amount**
    - **Validates: Requirements 10.1, 10.2, 10.3, 10.4, 10.5**

  - [x] 6.5 Write property test for compliance conditions elevate routing
    - **Property 9: Compliance Conditions Elevate Routing**
    - **Validates: Requirement 10.6**

  - [x] 6.6 Write property test for urgency override
    - **Property 10: Urgency Override for Due Dates**
    - **Validates: Requirement 11.3**

  - [x] 6.7 Write property test for delegation fallback
    - **Property 11: Delegation Fallback**
    - **Validates: Requirements 11.1, 11.2**

- [x] 7. Implement Disbursement Agent
  - [x] 7.1 Implement the Disbursement Agent Lambda handler
    - Create `internal/agents/disbursement/handler.go`
    - Verify Payment_Record status is APPROVED before executing transfer
    - Return FAILED with reason if not APPROVED
    - Look up payee account information; return FAILED if missing
    - Generate unique transaction reference
    - Execute transfer (simulated treasury interface)
    - On success: generate PaymentConfirmation with transaction ID, amount, payee, timestamp, reference
    - On failure: record reason and retryable flag
    - _Requirements: 12.1, 12.2, 12.3, 12.4, 12.5, 12.6_

  - [x] 7.2 Write property test for disbursement atomicity
    - **Property 12: Disbursement Atomicity**
    - **Validates: Requirements 23.1, 23.2, 23.3**

  - [x] 7.3 Write property test for disbursement precondition enforcement
    - **Property 13: Disbursement Precondition Enforcement**
    - **Validates: Requirements 12.1, 12.2**

- [x] 8. Implement Agent Coordinator and Retry Logic
  - [x] 8.1 Implement agent invocation with exponential backoff retry
    - Create `internal/coordinator/invoker.go`
    - Implement retry logic: up to 3 retries with exponential backoff
    - Backoff formula: (2^retryCount) * 100ms + random jitter (0-100ms), capped at 10,000ms
    - On all retries exhausted: set decision to ESCALATE, route to human review
    - Log every invocation attempt with agent name, message ID, attempt number
    - _Requirements: 15.1, 15.2, 15.3, 15.4_

  - [x] 8.2 Implement the main orchestration workflow coordinator
    - Create `internal/coordinator/workflow.go`
    - Wire the full pipeline: Extraction → Validation → Compliance → Routing → Disbursement
    - Handle escalation at each stage based on confidence/status
    - Implement human review resumption from escalation point
    - _Requirements: 2.3, 13.1_

  - [x] 8.3 Write property test for exponential backoff bounds
    - **Property 14: Exponential Backoff Bounds**
    - **Validates: Requirement 15.2**

- [~] 9. Checkpoint - Ensure all agent tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [x] 10. Implement Contract Financial Management Portal - Data Layer
  - [x] 10.1 Define portal data models in Go
    - Implement `Contract`, `ContractLineItem`, `ContractType`, `CLINType`, `CLINStatus` structs/enums
    - Implement `REA`, `REAStatus`, `REAResponse` structs/enums
    - Implement `VarianceAnalysis`, `RiskLevel` struct/enum
    - Implement `PortalRoles` with permission sets for CO, COR, PCO, PM
    - _Requirements: 16.1, 16.2, 17.1, 18.1, 20.1, 20.2_

  - [x] 10.2 Implement CLIN summation integrity and financial calculations
    - Implement function to calculate contract-level totals from CLIN data
    - Enforce: sum of CLIN obligated = contract total obligated
    - Enforce: sum of CLIN expended = contract total expended
    - _Requirements: 16.4, 16.5_

  - [x] 10.3 Write property test for CLIN summation integrity
    - **Property 15: CLIN Summation Integrity**
    - **Validates: Requirements 16.4, 16.5**

  - [x] 10.4 Write property test for financial data consistency across roles
    - **Property 16: Financial Data Consistency Across Roles**
    - **Validates: Requirement 16.3**

- [x] 11. Implement Variance Analysis and Risk Detection
  - [x] 11.1 Implement variance calculation algorithm
    - Compute overrun: max(0, expended - ceiling)
    - Compute under-run: max(0, obligated - EAC)
    - Calculate burn rate: sum of last 3 months expenditures / 3
    - Project completion date based on burn rate
    - _Requirements: 17.1, 17.2, 17.8_

  - [x] 11.2 Implement risk level determination
    - RED: any overrun, under-run > 40%, projected completion > PoP end
    - YELLOW: expenditure ratio > 90% of ceiling (ACTIVE), under-run 20-40%
    - GREEN: otherwise
    - _Requirements: 17.3, 17.4, 17.5, 17.6, 17.7_

  - [x] 11.3 Write property test for variance calculation correctness
    - **Property 17: Variance Calculation Correctness**
    - **Validates: Requirements 17.1, 17.2**

  - [x] 11.4 Write property test for risk level determination
    - **Property 18: Risk Level Determination**
    - **Validates: Requirements 17.3, 17.4, 17.5, 17.6, 17.7**

  - [x] 11.5 Write property test for burn rate calculation
    - **Property 27: Burn Rate Calculation**
    - **Validates: Requirement 17.8**

- [ ] 12. Implement REA Workflow Management
  - [x] 12.1 Implement REA submission and validation
    - Validate requested amount is positive
    - Validate at least one affected CLIN specified
    - Validate all referenced CLINs exist on contract
    - Create REA record with SUBMITTED status
    - Notify government CO and log audit trail
    - _Requirements: 18.1, 18.2_

  - [x] 12.2 Implement REA response handling (approve, partial, deny, info request)
    - APPROVED: create contract modification, adjust CLIN ceilings by approved amount
    - PARTIALLY_APPROVED: create modification for partial amount, adjust CLINs
    - DENIED: record rationale, set status DENIED
    - ADDITIONAL_INFO_REQUESTED: set status without resolved date
    - _Requirements: 18.3, 18.4, 18.5, 18.6_

  - [x] 12.3 Write property test for REA validation rules
    - **Property 20: REA Validation Rules**
    - **Validates: Requirement 18.1**

  - [~] 12.4 Write property test for REA approval adjusts ceilings
    - **Property 21: REA Approval Adjusts Ceilings**
    - **Validates: Requirements 18.3, 18.4**

- [ ] 13. Implement Option Exercise and Obligation Management
  - [x] 13.1 Implement option exercise with validation
    - Verify CLIN is an option and in ACTIVE status
    - Verify exercise deadline has not passed
    - Verify new total obligation does not exceed contract ceiling
    - Update CLIN status to EXERCISED, increase contract total obligated
    - Create contract modification and notify contractor
    - _Requirements: 19.1, 19.2, 19.3, 19.4_

  - [x] 13.2 Implement obligation integrity enforcement
    - Enforce total contract obligations never exceed total ceiling
    - Enforce CLIN expenditures never exceed CLIN obligations without authorization
    - Hold payments and notify CO when expenditure would exceed CLIN obligation
    - _Requirements: 22.1, 22.2, 22.3_

  - [x] 13.3 Write property test for option exercise constraints
    - **Property 22: Option Exercise Constraints**
    - **Validates: Requirements 19.1, 19.2, 19.3**

  - [~] 13.4 Write property test for obligation cannot exceed ceiling
    - **Property 19: Obligation Cannot Exceed Ceiling**
    - **Validates: Requirements 22.1, 22.2**

- [x] 14. Implement Role-Based Access Control
  - [x] 14.1 Implement RBAC enforcement for portal access
    - CO: view all contracts in portfolio, perform REA response, option exercise, obligation management
    - PCO: view own organization contracts, submit REAs, update EAC, submit invoices
    - Deny contractor access to non-associated contracts
    - Deny actions outside role permissions
    - _Requirements: 20.1, 20.2, 20.3, 20.4_

  - [x] 14.2 Write property test for role-based access enforcement
    - **Property 23: Role-Based Access Enforcement**
    - **Validates: Requirements 20.3, 20.4**

- [ ] 15. Implement SBIR Payment Processing Integration
  - [x] 15.1 Implement SBIR invoice validation against CLIN obligations
    - Verify referenced CLIN is in ACTIVE or EXERCISED status
    - Hold payment if invoice amount would cause CLIN expenditure to exceed obligation
    - For CPFF/CPIF: verify cost allowability before approval
    - For FFP: verify associated milestone accepted before approval
    - _Requirements: 21.1, 21.2, 21.3, 21.4_

  - [~] 15.2 Write property test for SBIR CLIN status gate
    - **Property 24: SBIR CLIN Status Gate**
    - **Validates: Requirements 21.1, 21.4**

  - [~] 15.3 Write property test for SBIR expenditure ceiling enforcement
    - **Property 25: SBIR Expenditure Ceiling Enforcement**
    - **Validates: Requirements 21.2, 22.3**

- [~] 16. Checkpoint - Ensure all backend and portal logic tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [x] 17. Implement Frontend - Responsive UI Shell and Document Upload
  - [x] 17.1 Set up React frontend with TypeScript, Tailwind CSS, and routing
    - Initialize React app with Vite + TypeScript
    - Configure Tailwind CSS with responsive breakpoints (mobile: 0-767, tablet: 768-1023, desktop: 1024+)
    - Set up React Router with routes for Dashboard, Payments, Upload, Contracts, Alerts
    - Configure TanStack Query for server state management
    - _Requirements: 24.1, 24.2, 24.3_

  - [x] 17.2 Implement responsive layout components
    - Mobile: single-column stacked cards, bottom tab navigation, floating action button
    - Tablet: two-column split view, collapsible sidebar
    - Desktop: full multi-panel dashboard, persistent sidebar, grid layout
    - _Requirements: 24.1, 24.2, 24.3_

  - [x] 17.3 Implement document upload with file validation
    - Mobile: camera capture as primary upload option
    - Desktop: drag-and-drop with react-dropzone
    - Client-side validation: max 10MB, supported formats (PDF, PNG, JPEG, TIFF)
    - Get presigned URL from API and upload directly to S3
    - _Requirements: 24.4, 24.5_

  - [x] 17.4 Write property test for file upload validation
    - **Property 26: File Upload Validation**
    - **Validates: Requirement 24.5**

- [x] 18. Implement Frontend - Real-Time Updates and Payment Pipeline View
  - [x] 18.1 Implement WebSocket connection for real-time updates
    - Connect to API Gateway WebSocket endpoint
    - Handle message types: STATUS_CHANGE, AGENT_RESULT, ESCALATION, COMPLETE
    - Update pipeline visualization in real-time on status changes
    - Push escalation notifications to appropriate reviewers
    - _Requirements: 25.1, 25.2, 25.3, 25.4_

  - [x] 18.2 Implement payment pipeline tracker and activity feed
    - Horizontal step visualization showing agent progress
    - Agent activity feed appending results in real-time
    - Payment detail view with document preview, extraction results, decision timeline
    - Escalation panel with human review form
    - _Requirements: 25.1, 25.2_

- [ ] 19. Implement Frontend - Contract Financial Portal Views
  - [x] 19.1 Implement contract financial dashboard
    - Display contract-level financials: ceiling, obligated, expended, EAC
    - Display CLIN-level financials in expandable table/accordion
    - Color-coded risk indicators (RED/YELLOW/GREEN)
    - Burn rate sparkline charts using Recharts
    - _Requirements: 16.1, 16.2, 17.3, 17.4, 17.5, 17.6, 17.7, 17.8_

  - [~] 19.2 Implement REA management views
    - Contractor: REA submission form with validation
    - Government: REA review and response workflow
    - Status badges, audit trail display
    - _Requirements: 18.1, 18.2, 18.3, 18.4, 18.5, 18.6_

  - [~] 19.3 Implement option exercise and obligation views
    - Option exercise interface with constraint validation feedback
    - Obligation tracking with ceiling enforcement indicators
    - Contract modification history
    - _Requirements: 19.1, 19.2, 19.3, 19.4, 22.1, 22.2_

- [ ] 20. Implement Infrastructure as Code (AWS CDK)
  - [~] 20.1 Set up AWS CDK stack with core infrastructure
    - VPC with private subnets and VPC endpoints (S3, DynamoDB, Bedrock)
    - DynamoDB tables: payments (with GSIs for status-index, payee-index), contracts, CLINs, REAs
    - S3 buckets: document ingestion (with event notification), static website hosting
    - IAM roles scoped per agent Lambda (least privilege)
    - _Requirements: 1.1, 13.1, 14.1_

  - [~] 20.2 Deploy Lambda functions and Step Functions workflow
    - Lambda functions for each agent (extraction, validation, compliance, routing, disbursement)
    - Lambda functions for portal API handlers and WebSocket handler
    - Step Functions Standard Workflow with logging and X-Ray tracing
    - S3 event notification to trigger Step Functions on document upload
    - _Requirements: 1.1, 15.1, 25.1_

  - [~] 20.3 Deploy API Gateway, CloudFront, and frontend hosting
    - REST API Gateway for portal endpoints with Cognito authorizer
    - WebSocket API Gateway for real-time updates
    - CloudFront distribution with S3 origin for React SPA
    - AWS WAF rules for request filtering
    - _Requirements: 24.1, 25.1, 20.1_

- [ ] 21. Integration wiring and end-to-end flow
  - [~] 21.1 Wire the full agent pipeline with Step Functions state machine definition
    - Define Step Functions ASL with states for each agent
    - Wire retry/catch logic at each step for escalation
    - Connect S3 event trigger → Step Functions → agent sequence
    - Wire DynamoDB updates at each state transition
    - _Requirements: 1.1, 2.1, 2.2, 13.1, 15.1_

  - [~] 21.2 Wire portal API handlers to data layer and payment pipeline
    - Connect contract finance API endpoints to DynamoDB operations
    - Connect portal to payment pipeline for SBIR invoice processing
    - Wire DynamoDB Streams → aggregation Lambda for real-time financial updates
    - Connect notification service for REA/option exercise events
    - _Requirements: 16.1, 18.2, 21.1, 25.1_

  - [~] 21.3 Implement synthetic data generator for hackathon demo
    - Generate realistic invoice PDFs with varying quality
    - Include known OFAC test entries and debarment matches
    - Span all approval thresholds
    - Include edge cases: duplicates, future dates, missing fields
    - Generate sample contracts with CLINs, options, and REAs for portal demo
    - _Requirements: 1.2, 4.1, 6.1_

  - [~] 21.4 Write integration tests for end-to-end payment flow
    - Happy path: clean invoice → DISBURSED
    - Rejection path: sanctioned payee → NON_COMPLIANT
    - Escalation path: low confidence → ESCALATED
    - Portal flow: REA submission → approval → ceiling adjustment
    - _Requirements: 1.1, 6.2, 2.1, 18.3_

- [~] 22. Final checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

## Notes

- Tasks marked with `*` are optional and can be skipped for faster MVP
- Each task references specific requirements for traceability
- Checkpoints ensure incremental validation
- Property tests validate universal correctness properties from the design document
- Unit tests validate specific examples and edge cases
- Go is used for all backend agent logic and portal API handlers
- TypeScript is used for AWS CDK infrastructure and React frontend
- The platform uses synthetic data for hackathon demonstration — no real PII

## Task Dependency Graph

```json
{
  "waves": [
    { "id": 0, "tasks": ["1.1"] },
    { "id": 1, "tasks": ["1.2"] },
    { "id": 2, "tasks": ["1.3", "10.1"] },
    { "id": 3, "tasks": ["1.4", "1.5", "2.1", "4.1", "5.1", "5.2", "6.1", "17.1"] },
    { "id": 4, "tasks": ["2.2", "2.3", "4.2", "4.3", "4.4", "5.3", "5.4", "6.2", "6.3", "7.1", "17.2"] },
    { "id": 5, "tasks": ["2.4", "5.5", "5.6", "5.7", "5.8", "6.4", "6.5", "6.6", "6.7", "7.2", "7.3", "17.3"] },
    { "id": 6, "tasks": ["8.1", "10.2", "11.1", "17.4"] },
    { "id": 7, "tasks": ["8.2", "8.3", "10.3", "10.4", "11.2", "12.1", "18.1"] },
    { "id": 8, "tasks": ["11.3", "11.4", "11.5", "12.2", "13.1", "13.2", "14.1", "18.2"] },
    { "id": 9, "tasks": ["12.3", "12.4", "13.3", "13.4", "14.2", "15.1", "19.1"] },
    { "id": 10, "tasks": ["15.2", "15.3", "19.2", "19.3", "20.1"] },
    { "id": 11, "tasks": ["20.2", "20.3"] },
    { "id": 12, "tasks": ["21.1", "21.2"] },
    { "id": 13, "tasks": ["21.3", "21.4"] }
  ]
}
```
