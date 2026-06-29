/**
 * Simulated chatbot service that provides contextual help based on the
 * knowledge base content. Uses pattern matching to respond to common questions.
 *
 * Validates: Requirements 26.1, 26.2, 26.5
 */

interface PatternResponse {
  patterns: RegExp[]
  response: string
}

const knowledgeResponses: PatternResponse[] = [
  {
    patterns: [/upload.*document/i, /how.*upload/i, /submit.*document/i, /drag.*drop/i],
    response:
      'To upload a document, navigate to the Upload page (/upload). You can:\n\n' +
      '• **Desktop**: Drag and drop files onto the upload area, or click to browse\n' +
      '• **Mobile**: Use the camera capture button for quick document scanning\n\n' +
      'Supported formats: PDF, PNG, JPEG, and TIFF (max 10 MB).\n\n' +
      'Once uploaded, documents enter the 5-stage payment processing pipeline automatically.',
  },
  {
    patterns: [/red.*risk/i, /risk.*red/i, /what.*red/i, /critical.*risk/i],
    response:
      'A **RED (Critical)** risk level indicates one of these conditions:\n\n' +
      '• **Overrun**: Expended amount exceeds the ceiling\n' +
      '• **Large under-run**: Under-run exceeds 40% of the obligated amount\n' +
      '• **Schedule breach**: Projected completion date exceeds period of performance end date\n\n' +
      'RED items require immediate attention from the Contracting Officer.',
  },
  {
    patterns: [/risk.*level/i, /what.*risk/i, /yellow.*risk/i, /green.*risk/i, /explain.*risk/i],
    response:
      'The system uses three risk levels:\n\n' +
      '• 🔴 **RED (Critical)**: Overrun, large under-run (>40%), or schedule breach\n' +
      '• 🟡 **YELLOW (Warning)**: Expenditure >90% of ceiling while active, or under-run 20-40%\n' +
      '• 🟢 **GREEN (Normal)**: No risk indicators triggered\n\n' +
      'Risk levels are calculated automatically based on contract and CLIN financials.',
  },
  {
    patterns: [/payment.*stage/i, /pipeline.*stage/i, /processing.*stage/i, /5.*stage/i, /five.*stage/i, /what.*pipeline/i],
    response:
      'Payments flow through 5 sequential AI agent stages:\n\n' +
      '1. **Document Processing (Extraction)**: Classifies document type and extracts structured fields using AI\n' +
      '2. **Validation**: Verifies completeness, checks required fields, detects duplicates\n' +
      '3. **Compliance**: Screens against OFAC sanctions, checks FAR compliance rules\n' +
      '4. **Routing**: Determines approval authority based on amount and conditions\n' +
      '5. **Disbursement**: Executes the electronic fund transfer after all approvals\n\n' +
      'You can track progress on the Payments page.',
  },
  {
    patterns: [/submit.*rea/i, /how.*rea/i, /rea.*workflow/i, /request.*equitable/i, /file.*rea/i],
    response:
      'To submit an REA (Request for Equitable Adjustment):\n\n' +
      '1. Navigate to the Contracts page and select the relevant contract\n' +
      '2. Click "Submit REA" (requires SUBMIT_REA permission)\n' +
      '3. Enter the requested amount (must be positive)\n' +
      '4. Select at least one affected CLIN\n' +
      '5. Provide justification and submit\n\n' +
      'The REA will go through review stages: SUBMITTED → APPROVED/PARTIALLY_APPROVED/DENIED.\n' +
      'The Government CO is notified upon submission.',
  },
  {
    patterns: [/contractor.*do/i, /contractor.*permission/i, /contractor.*can/i, /contractor.*access/i, /what.*contractor/i],
    response:
      'As a **Contractor**, you can:\n\n' +
      '• View contracts associated with your organization\n' +
      '• Submit REAs (Request for Equitable Adjustment)\n' +
      '• Update EAC (Estimate at Completion) values\n' +
      '• Submit invoices\n\n' +
      'You **cannot**:\n' +
      '• Access contracts from other organizations\n' +
      '• Approve REAs or exercise options\n' +
      '• Manage obligations',
  },
  {
    patterns: [/payment.*status/i, /status.*mean/i, /what.*status/i],
    response:
      'Key payment statuses:\n\n' +
      '• **RECEIVED** → **EXTRACTING** → **EXTRACTED**: Document is being processed\n' +
      '• **VALIDATING** → **VALIDATED**: Data completeness/correctness check\n' +
      '• **CHECKING_COMPLIANCE** → **COMPLIANT**: Regulatory screening\n' +
      '• **ROUTING** → **ROUTED** → **APPROVING** → **APPROVED**: Approval workflow\n' +
      '• **DISBURSING** → **DISBURSED**: Payment complete ✓\n\n' +
      'Terminal states: DISBURSED (success), REJECTED, ESCALATED, FAILED',
  },
  {
    patterns: [/clin.*type/i, /what.*clin/i, /ffp|cpff|cpif|t&m/i, /contract.*type/i],
    response:
      'CLIN (Contract Line Item Number) types:\n\n' +
      '• **FFP** (Firm-Fixed-Price): Fixed price for defined deliverables\n' +
      '• **CPFF** (Cost-Plus-Fixed-Fee): Reimbursable costs + fixed fee\n' +
      '• **CPIF** (Cost-Plus-Incentive-Fee): Costs + performance-based incentive fee\n' +
      '• **T&M** (Time and Materials): Payment based on hours and material costs\n' +
      '• **OPTION**: Optional CLIN exercisable by the CO within a deadline',
  },
  {
    patterns: [/role/i, /permission/i, /who.*can/i, /access.*control/i],
    response:
      'System roles and key permissions:\n\n' +
      '• **CO** (Contracting Officer): Full portfolio access, approve REAs, exercise options\n' +
      '• **COR** (Contracting Officer Representative): View-only access\n' +
      '• **PCO** (Procuring CO): Org-scoped, submit REAs, update EAC, submit invoices\n' +
      '• **Contractor**: Org-scoped, submit REAs, update EAC, submit invoices\n' +
      '• **Program Manager**: Org-scoped, submit REAs, update EAC',
  },
  {
    patterns: [/routing.*threshold/i, /approval.*amount/i, /who.*approve/i, /approval.*authority/i],
    response:
      'Payment approval routing by amount:\n\n' +
      '• ≤ $2,500 → Purchase Card (LOW priority)\n' +
      '• ≤ $25,000 → Supervisor (NORMAL priority)\n' +
      '• ≤ $250,000 → Contracting Officer (NORMAL priority)\n' +
      '• ≤ $1,000,000 → Senior CO (HIGH priority)\n' +
      '• > $1,000,000 → Agency Head (URGENT priority)\n\n' +
      'Special rules: Compliance conditions elevate by one tier; due in 3 days = URGENT.',
  },
  {
    patterns: [/option.*exercise/i, /exercise.*option/i, /how.*exercise/i],
    response:
      'To exercise a contract option:\n\n' +
      '1. Only CLINs of type OPTION with status ACTIVE can be exercised\n' +
      '2. The exercise deadline must not have passed\n' +
      '3. New total obligation cannot exceed contract ceiling\n\n' +
      'On exercise: CLIN status → EXERCISED, total obligated increases, a modification is created, and the contractor is notified.\n\n' +
      'Only Contracting Officers (CO) can exercise options.',
  },
  {
    patterns: [/dashboard/i, /home.*page/i, /overview/i],
    response:
      'The **Dashboard** (home page) shows:\n\n' +
      '• Summary statistics: Total contracts, active payments, pending REAs, alerts\n' +
      '• Recent activity feed with color-coded entries\n' +
      '• Agent system status panel (health, throughput, heartbeat)\n' +
      '• Obligation tracking summary with progress bars',
  },
  {
    patterns: [/confidence/i, /escalat/i, /threshold/i],
    response:
      'The system uses confidence scores for extracted fields:\n\n' +
      '• Each field gets a confidence score from 0.0 to 1.0\n' +
      '• Overall confidence = minimum of all field scores\n' +
      '• If overall confidence < **0.75**, payment is escalated to human review\n' +
      '• Individual field confidence < **0.80** triggers validation warnings\n\n' +
      'Escalated payments appear on the Alerts page for manual review.',
  },
]

/** Page-specific context hints based on current page path */
function getPageContext(currentPage: string): string {
  if (currentPage === '/' || currentPage === '') {
    return 'You are on the Dashboard. I can help you understand the summary stats, activity feed, or agent status.'
  }
  if (currentPage.includes('/payments')) {
    return 'You are on the Payments page. I can explain payment statuses, pipeline stages, or help with tracking a specific payment.'
  }
  if (currentPage.includes('/upload')) {
    return 'You are on the Upload page. I can help with supported formats, file size limits, or the upload process.'
  }
  if (currentPage.includes('/contracts')) {
    return 'You are on the Contracts page. I can help with CLIN types, risk levels, REA submissions, or option exercises.'
  }
  if (currentPage.includes('/alerts')) {
    return 'You are on the Alerts page. I can help with REA workflows, escalations, or alert management.'
  }
  return ''
}

/**
 * Sends a message to the simulated chatbot and returns a contextual response.
 * Uses pattern matching on the knowledge base to provide relevant answers.
 */
export async function sendMessage(message: string, currentPage: string): Promise<string> {
  // Simulate network latency (300-800ms)
  await new Promise((resolve) => setTimeout(resolve, 300 + Math.random() * 500))

  const trimmed = message.trim().toLowerCase()

  if (!trimmed) {
    return "I didn't catch that. Could you rephrase your question?"
  }

  // Check for greetings
  if (/^(hi|hello|hey|help)(\s|$|!|\?)/i.test(trimmed)) {
    const context = getPageContext(currentPage)
    return (
      'Hello! I\'m your Federal Payment Processing assistant. I can help you with:\n\n' +
      '• Uploading and processing documents\n' +
      '• Understanding payment pipeline stages and statuses\n' +
      '• Contract management, CLINs, and risk levels\n' +
      '• REA workflows and submissions\n' +
      '• Roles, permissions, and routing thresholds\n\n' +
      (context ? context + '\n\n' : '') +
      'What would you like to know?'
    )
  }

  // Match against knowledge patterns
  for (const entry of knowledgeResponses) {
    for (const pattern of entry.patterns) {
      if (pattern.test(message)) {
        return entry.response
      }
    }
  }

  // Fallback response with page context
  const context = getPageContext(currentPage)
  return (
    "I'm not sure I understand that question. Here are some things I can help with:\n\n" +
    '• "How do I upload a document?"\n' +
    '• "What does RED risk mean?"\n' +
    '• "What are the payment stages?"\n' +
    '• "How do I submit an REA?"\n' +
    '• "What can a contractor do?"\n' +
    '• "What are the routing thresholds?"\n\n' +
    (context ? context : 'Try asking about a specific feature or workflow!')
  )
}
