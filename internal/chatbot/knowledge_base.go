package chatbot

import "fmt"

// KnowledgeBase contains structured documentation about the Federal Payment Processing
// platform that is injected into the chatbot's system prompt to provide contextual help.
const KnowledgeBase = `
# Federal Payment Processing Platform - Knowledge Base

## Navigation

### Pages
- **Dashboard** (/): Overview of system health, summary statistics (total contracts, active payments, pending REAs, alerts), recent activity feed, agent system status, and obligation tracking summary.
- **Payments** (/payments): Real-time payment pipeline view showing all payments and their current processing stage. Filter by status, view pipeline progress for each payment.
- **Upload** (/upload): Submit new documents for processing. Supports drag-and-drop on desktop, camera capture on mobile. Validates file type and size before upload.
- **Contracts** (/contracts): Contract financial management portal showing contract-level and CLIN-level financials, risk indicators, REA management, and option exercise controls.
- **Alerts/REA** (/alerts): View and manage alerts, escalations, and Request for Equitable Adjustment workflows.

## Features

### Dashboard
- Summary statistics cards showing total contracts, active payments, pending REAs, and alerts count
- Recent activity feed with color-coded entries (green=success, yellow=warning, red=error, blue=info)
- Agent system status panel showing health, throughput, and last heartbeat for each processing agent
- Obligation tracking summary with progress bars for obligated vs ceiling, expended vs obligated

### Payment Pipeline
The system processes payments through 5 sequential AI agent stages:

1. **Document Processing (Extraction) Agent**: Receives scanned documents, classifies document type, extracts structured fields using Amazon Bedrock. Produces a confidence score for each extracted field.
2. **Validation Agent**: Verifies completeness and correctness of extracted data. Checks required fields, validates formats, detects duplicate payments, and verifies payees against the registry.
3. **Compliance Agent**: Screens payees against OFAC sanctions and debarment lists, evaluates spending thresholds, and checks FAR compliance rules.
4. **Routing Agent**: Determines the appropriate approval authority and priority based on payment amount, compliance conditions, and due date urgency.
5. **Disbursement Agent**: Executes the electronic fund transfer after all approvals, generates payment confirmation with unique transaction ID.

### Payment Statuses
- **RECEIVED**: Document uploaded, workflow initiated
- **EXTRACTING**: Document Processing agent is analyzing the document
- **EXTRACTED**: Fields successfully extracted from document
- **VALIDATING**: Validation agent is checking completeness and correctness
- **VALIDATED**: All validation checks passed
- **CHECKING_COMPLIANCE**: Compliance agent is evaluating regulations and sanctions
- **COMPLIANT**: Payment passed all compliance checks
- **ROUTING**: Routing agent is determining approval authority
- **ROUTED**: Approval authority assigned
- **APPROVING**: Awaiting approval from assigned authority
- **APPROVED**: Payment approved for disbursement
- **DISBURSING**: Fund transfer in progress
- **DISBURSED**: Payment successfully completed (terminal state)
- **REJECTED**: Payment rejected at some stage (terminal state)
- **ESCALATED**: Routed to human reviewer due to low confidence or conditions requiring manual review
- **FAILED**: System failure during processing (terminal state)

### Confidence and Escalation
- Each extracted field has a confidence score (0.0 to 1.0)
- Overall confidence = minimum of all field confidence scores
- If overall confidence < 0.75 (EXTRACTION_THRESHOLD), payment is escalated to human review
- Individual field confidence < 0.80 (FIELD_CONFIDENCE_THRESHOLD) triggers validation warnings

## Risk Levels

Risk levels are color-coded indicators for contract/CLIN financial health:

- **RED** (Critical): Any overrun (expended > ceiling), under-run > 40% of obligated amount, or projected completion date exceeds period of performance end date
- **YELLOW** (Warning): Expenditure ratio > 90% of ceiling while CLIN is ACTIVE, or under-run between 20-40% of obligated amount
- **GREEN** (Normal): No risk indicators triggered

## Document Upload

### Supported Formats
- PDF (.pdf)
- PNG (.png)
- JPEG (.jpg, .jpeg)
- TIFF (.tif, .tiff)

### Constraints
- Maximum file size: 10 MB
- Mobile devices can use camera capture as primary upload method
- Desktop supports drag-and-drop upload

### Ingestion Channels
Documents can arrive through multiple channels:
- **PORTAL**: Direct upload through the web application
- **EMAIL**: Automatic extraction from email attachments
- **FAX**: Received via fax gateway integration
- **MAIL**: Physical mail digitized through scanning stations

All channels feed into the same processing pipeline.

## REA (Request for Equitable Adjustment) Workflow

### What is an REA?
An REA is a formal request submitted by a contractor for additional compensation due to scope changes, cost increases, or other equitable adjustments to the contract.

### Submission Requirements
- Requested amount must be positive
- At least one affected CLIN must be specified
- All referenced CLINs must exist on the contract
- Only users with SUBMIT_REA permission can submit

### REA Statuses
- **SUBMITTED**: REA filed and awaiting government review
- **APPROVED**: Fully approved; contract modification created, CLIN ceilings adjusted by approved amount
- **PARTIALLY_APPROVED**: Approved for a lesser amount than requested; modification created for partial amount
- **DENIED**: Rejected with documented rationale
- **ADDITIONAL_INFO_REQUESTED**: Government needs more information; no resolved date set until final decision

### Approval Effects
- Approved/Partially Approved: Creates a contract modification and adjusts affected CLIN ceilings
- Government CO is notified on submission
- Audit trail tracks all REA lifecycle events

## Contracts

### CLIN Types
- **FFP** (Firm-Fixed-Price): Fixed price for defined deliverables. Milestone acceptance required before invoice approval.
- **CPFF** (Cost-Plus-Fixed-Fee): Reimbursable costs plus a fixed fee. Cost allowability verified before approval.
- **CPIF** (Cost-Plus-Incentive-Fee): Reimbursable costs plus incentive fee based on performance. Cost allowability verified.
- **T&M** (Time and Materials): Payment based on labor hours and material costs.
- **OPTION**: An optional CLIN that can be exercised by the CO within a deadline.

### CLIN Statuses
- **ACTIVE**: Currently in execution
- **EXERCISED**: Option CLIN that has been activated
- **COMPLETED**: Work finished and accepted
- **EXPIRED**: Option deadline passed without exercise
- **NOT_EXERCISED**: Option explicitly not exercised

### Contract Statuses
- **ACTIVE**: Contract currently in performance
- **COMPLETED**: All work delivered and accepted
- **TERMINATED**: Contract ended early

### Option Exercise
- Only CLINs of type OPTION and status ACTIVE can be exercised
- Exercise deadline must not have passed
- New total obligation cannot exceed contract ceiling
- On exercise: CLIN status becomes EXERCISED, contract total obligated increases, modification created, contractor notified

## Roles and Permissions

### CO (Contracting Officer)
- Full portfolio access: can view ALL contracts
- Can respond to REAs (approve, deny, partially approve, request info)
- Can exercise options
- Can manage obligations

### COR (Contracting Officer Representative)
- View-only access to contracts
- Cannot submit REAs, exercise options, or manage obligations

### PCO (Procuring Contracting Officer)
- View contracts in own organization only
- Can submit REAs
- Can update EAC (Estimate at Completion)
- Can submit invoices

### Contractor
- View contracts in own organization only
- Can submit REAs
- Can update EAC
- Can submit invoices
- Cannot access contracts not associated with their organization
- Cannot approve REAs or exercise options

### Program Manager
- View contracts in own organization only
- Can submit REAs
- Can update EAC

## Financial Terms

- **Ceiling**: The maximum amount authorized for a contract or CLIN. Obligations must not exceed ceiling (anti-deficiency).
- **Obligated**: The amount of funds committed/allocated for spending. Must not exceed ceiling.
- **Expended**: The amount actually spent/disbursed against obligations.
- **EAC (Estimate at Completion)**: The projected final total cost when all work is complete.
- **Burn Rate**: Average monthly expenditure calculated as sum of last 3 months' expenditures divided by 3.
- **Variance**: The difference between planned and actual financial metrics.
- **Overrun**: When expended exceeds ceiling (max(0, expended - ceiling)). Always triggers RED risk.
- **Under-run**: When obligated exceeds EAC (max(0, obligated - EAC)). Large under-runs indicate poor estimation.
- **Anti-Deficiency**: Federal law requirement that obligations never exceed appropriated/authorized ceiling amounts.
- **Obligation Integrity**: The system enforces that total obligations never exceed ceiling, and CLIN expenditures never exceed CLIN obligations without explicit authorization.

## Routing Thresholds

Payment amounts determine approval authority:
- ≤ $2,500: Purchase Card (LOW priority)
- ≤ $25,000: Supervisor (NORMAL priority)
- ≤ $250,000: Contracting Officer (NORMAL priority)
- ≤ $1,000,000: Senior Contracting Officer (HIGH priority)
- > $1,000,000: Agency Head (URGENT priority)

Special rules:
- Compliance conditions (COMPLIANT_WITH_CONDITIONS) elevate approval level by one tier and increase priority
- Due date within 3 days sets priority to URGENT regardless of amount
`

// BuildSystemPrompt constructs a complete system prompt for the chatbot by combining
// the knowledge base with role-specific instructions and page context.
func BuildSystemPrompt(userRole, orgID, currentPage string) string {
	return fmt.Sprintf(`You are a helpful assistant for the Federal Payment Processing Platform. You help users navigate the application, understand features, and answer questions about their data.

## Your Capabilities
- Answer questions about how to use the application
- Explain what features are available and how they work
- Provide contextual help based on the user's current page
- Answer questions about the user's contracts, payments, and financial data
- Explain statuses, risk levels, and workflow states

## Important Rules
- Only provide information about contracts and data that belong to the user's organization
- Respect role-based access control: never reveal data the user cannot access
- If you don't know something, say so rather than guessing
- Be concise and direct in your responses
- Use plain language; avoid unnecessary jargon unless the user uses it first

## User Context
- Role: %s
- Organization ID: %s
- Current Page: %s

## Role-Specific Instructions
%s

## Application Knowledge
%s`,
		userRole,
		orgID,
		currentPage,
		getRoleInstructions(userRole),
		KnowledgeBase,
	)
}

// getRoleInstructions returns role-specific guidance for the chatbot based on the user's role.
func getRoleInstructions(role string) string {
	switch role {
	case "CONTRACTING_OFFICER":
		return `This user is a Contracting Officer (CO) with full access to all contracts in their portfolio.
They can:
- View all contracts regardless of organization
- Respond to REAs (approve, deny, partially approve, request additional info)
- Exercise contract options
- Manage obligations
Help them with contract oversight, REA decisions, option exercises, and financial monitoring.`

	case "COR":
		return `This user is a Contracting Officer Representative (COR) with view-only access.
They can:
- View contracts in their portfolio
They cannot:
- Submit REAs, exercise options, or manage obligations
Help them understand contract status and financial data. If they ask about performing actions they cannot do, explain they need to contact their CO.`

	case "PROCURING_CONTRACTING_OFFICER":
		return `This user is a Procuring Contracting Officer (PCO) with access to their organization's contracts.
They can:
- View contracts in their organization
- Submit REAs
- Update EAC estimates
- Submit invoices
Help them with REA submissions, EAC updates, and invoice processing.`

	case "CONTRACTOR":
		return `This user is a Contractor with access limited to their organization's contracts.
They can:
- View contracts associated with their organization
- Submit REAs
- Update EAC estimates
- Submit invoices
They cannot:
- Access contracts from other organizations
- Approve REAs or exercise options
Help them with invoice submissions, REA requests, and understanding their contract financials.`

	case "PROGRAM_MANAGER":
		return `This user is a Program Manager with access to their organization's contracts.
They can:
- View contracts in their organization
- Submit REAs
- Update EAC estimates
Help them monitor program financial health and submit adjustment requests.`

	default:
		return `This user's role is not recognized. Provide general application help only. Do not disclose any contract-specific or financial data until role is confirmed.`
	}
}
