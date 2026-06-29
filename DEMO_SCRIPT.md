# Federal Payment Processing Platform — Demo Walkthrough Script

## Overview

This demo walks through the complete SBIR lifecycle for **Quantum Federal Systems LLC**, from initial solicitation through Phase II execution, culminating with a Request for Equitable Adjustment (REA) and government approval.

**Live URL**: https://d2wbk4dmt2edww.cloudfront.net

---

## Scene 1: Dashboard Overview (/)

**Narrator**: "Welcome to the Federal Payment Processing Platform — an AI-powered system that automates end-to-end federal payment processing using five specialized AI agents."

**Show**:
- Summary statistics (contracts, active payments, pending REAs, alerts)
- Agent health indicators (all green except one degraded for realism)
- Obligation tracking summary showing $52.4M managed
- Recent activity feed with real entries

**Key Talking Points**:
- Multi-agent AI architecture on AWS
- Real-time monitoring of all payment activity
- Aligned with the March 2025 Executive Order to eliminate paper-based payments

---

## Scene 2: Contract Management (/contracts)

**Narrator**: "Let's look at our SBIR Phase II contract — a Task Order under an IDIQ for autonomous payment processing AI."

**Show**:
1. **Contract Summary Card**:
   - FA8750-25-F-0018 — "Autonomous Payment Processing AI - SBIR Phase II"
   - Contractor: Quantum Federal Systems LLC
   - Agency: AFRL / Air Force Research Laboratory
   - Parent IDIQ: GS-00F-0001A
   - SBIR Topic: AF241-0042 ("Agentic AI for Federal Payment Modernization")
   
2. **Period of Performance**:
   - 6 months (Jan 2 → Jul 1, 2025)
   - 33% elapsed (2 months in)
   - On Track ✓

3. **Financial Health**:
   - Ceiling: $1,249,800
   - Obligated: $749,880
   - Expended: $208,300 (2 months of burn)
   - EAC: $1,180,000 (under ceiling, GREEN)

4. **SBIR Lifecycle Timeline**:
   - Phase I (Feasibility): ✓ COMPLETED — Sep 2023 – Mar 2024
   - Phase II (R&D): ● IN PROGRESS — Jan 2025 – Jul 2025
   - Phase III (Commercialization): ○ PENDING — Option pending exercise

**Key Talking Points**:
- The system provides a single pane of glass for contract financial health
- SBIR phased approach from feasibility through commercialization
- Phase III option available for exercise before Jun 1, 2025 deadline

---

## Scene 3: CLIN Drill-Down (Click "View CLINs")

**Narrator**: "Drilling into the contract line items, we see the funded base period and an unexercised Phase III option."

**Show**:
- CLIN 0001: Phase II R&D — CPFF, ACTIVE, $749,880 ceiling, $208,300 expended, $104,150/mo burn rate, GREEN
- CLIN 0002: Phase III Option — CPFF, OPTION, $499,920 ceiling, not yet exercised, deadline Jun 1 2025

**Key Talking Points**:
- Cost-Plus-Fixed-Fee structure appropriate for R&D
- Burn rate on track ($104K/mo × 6 months ≈ $624K < $749K ceiling)
- Option exercise decision coming in 3 months

---

## Scene 4: Payment Pipeline (/payments)

**Narrator**: "Now let's watch a payment flow through our 5-agent AI pipeline in real-time."

**Show**:
- Pipeline tracker animating through stages:
  1. **Extraction** — Document classified as INVOICE, 8 fields extracted (92% confidence)
  2. **Validation** — All required fields present, no duplicates
  3. **Compliance** — OFAC clear, FAR compliant, no threshold exceeded
  4. **Routing** — Routed to SUPERVISOR level, NORMAL priority
  5. **Disbursement** — Payment complete ✓
- Channel badge showing "EMAIL" source
- Activity feed with agent results and confidence scores

**Key Talking Points**:
- 5 specialized AI agents, each with its own decision-making capability
- Real-time WebSocket updates — no manual refresh needed
- Confidence scoring at every step with automatic escalation below 75%
- Multi-channel ingestion (email, fax, mail, portal)

---

## Scene 5: Escalation Scenario

**Narrator**: "When confidence is low, the system automatically escalates to human review."

**Show** (automatic after first demo cycle):
- Second payment (demo-payment-002) arrives via FAX channel
- Extraction agent produces 62% confidence (below 75% threshold)
- RED escalation alert banner appears
- "Extraction confidence below threshold (0.62 < 0.75)"
- Pipeline halts — human review required

**Key Talking Points**:
- Human-in-the-loop for high-risk transactions
- Automatic escalation based on configurable thresholds
- No payment proceeds without meeting confidence standards
- Full audit trail for every decision

---

## Scene 6: Document Upload (/upload)

**Narrator**: "Documents can be submitted through multiple channels. Here's the portal upload experience."

**Show**:
- **Desktop**: Drag-and-drop zone
- **Mobile** (if phone available): Camera capture as primary option
- File validation: PDF, PNG, JPEG, TIFF up to 10MB
- Upload progress indicator

**Key Talking Points**:
- Multichannel ingestion: email, fax, mail scanning stations, portal upload
- Client-side validation prevents bad submissions
- Presigned URLs for secure direct-to-S3 upload
- Immediate pipeline initiation on upload

---

## Scene 7: REA Workflow (/alerts)

**Narrator**: "After 2 months of execution, the contractor encounters a scope change requiring a Request for Equitable Adjustment."

**Show**:
1. Click "+ New REA" button
2. Fill in:
   - Requested Amount: $85,000
   - Affected CLINs: Select "CLIN 0001"
   - Justification: "Additional security compliance requirements mandated by updated NIST SP 800-171 Rev 3 guidance published after contract award. Requires additional FTEs for 6 weeks of security testing."
3. Submit → REA appears with "SUBMITTED" status

4. Show existing sample REAs demonstrating the full lifecycle:
   - REA-2024-001: APPROVED ($185K)
   - REA-2024-002: PARTIALLY APPROVED ($210K of $320K requested)
   - REA-2024-003: DENIED
   - REA-2024-004: ADDITIONAL INFO REQUESTED (pending)
   - REA-2025-001: SUBMITTED (new)

5. Expand an approved REA to show the audit trail timeline

**Key Talking Points**:
- Contractors can self-serve REA submissions
- Government CO is notified automatically
- Full audit trail tracks every action
- Approval creates contract modification and adjusts CLIN ceilings
- Platform handles approve, partially approve, deny, and info request workflows

---

## Scene 8: Government Approval (Final)

**Narrator**: "The Government Contracting Officer reviews and approves the REA, triggering an automatic contract modification."

**Show** (describe the workflow):
1. CO receives notification of new REA
2. Reviews justification and supporting documentation
3. Approves for the full $85,000
4. System automatically:
   - Creates Contract Modification #001
   - Adjusts CLIN 0001 ceiling by $85,000
   - Updates contract total ceiling
   - Generates formal correspondence to contractor
   - Logs complete audit trail

**Key Talking Points**:
- Agentic AI generates the approval correspondence automatically
- Human reviews and approves before sending
- Contract modification created programmatically
- Anti-deficiency rules enforced (obligations never exceed ceiling)
- Complete financial audit trail for compliance

---

## Scene 9: AI Assistant (Chat Widget)

**Narrator**: "Any user can ask the AI assistant for help navigating the system or understanding their data."

**Show** (click the purple chat bubble):
- Ask: "What are the payment stages?"
- Ask: "What does my contract ceiling mean?"
- Ask: "How do I submit an REA?"
- Ask: "What roles can exercise options?"

**Key Talking Points**:
- Context-aware — knows which page you're on
- RBAC-enforced — only shows data you're authorized to see
- Powered by Amazon Bedrock (Claude)
- Reduces training burden and support tickets

---

## Scene 10: Correspondence (/correspondence)

**Narrator**: "The platform automatically generates professional correspondence for every payment action."

**Show**:
- List of generated letters with status badges (DRAFT, PENDING, SENT)
- Preview an approval confirmation letter
- Preview a rejection notice with specific reasons
- Preview an REA response in formal government style

**Key Talking Points**:
- AI-generated using Amazon Bedrock
- Multiple output formats (email, PDF, portal notification)
- Human review before sending (no auto-send)
- Professional government correspondence style

---

## Closing

**Narrator**: "This is the Federal Payment Processing Platform — transforming paper-based payment processing into an automated, AI-driven pipeline that maintains full compliance, complete audit trails, and human oversight where it matters most."

**Architecture Highlights**:
- 5 specialized AI agents (Extraction, Validation, Compliance, Routing, Disbursement)
- AWS native: Lambda, Step Functions, DynamoDB, S3, Bedrock, CloudFront
- 35+ property-based tests ensuring mathematical correctness
- RBAC with role-specific access controls
- Multi-channel document ingestion (email, fax, mail, portal)
- Real-time WebSocket pipeline monitoring
- SBIR lifecycle management with option exercise and REA workflows
- Automated correspondence generation with human-in-the-loop review

---

## Quick Demo Route (5 minutes)

If short on time, hit these pages in order:
1. **/** — Dashboard overview (30s)
2. **/contracts** — Contract card + SBIR timeline (60s)
3. **/payments** — Watch pipeline animate (90s, wait for escalation)
4. **/alerts** — Submit a new REA (60s)
5. **Chat** — Ask 2 questions (30s)
