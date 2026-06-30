# User Manual
## Federal Payment Processing Platform (FedPay)
### Version 2.3.0

---

## Table of Contents
1. [Getting Started](#1-getting-started)
2. [Dashboard](#2-dashboard)
3. [Solicitations](#3-solicitations)
4. [Contracts Management](#4-contracts-management)
5. [History](#5-history)
6. [Chatbot Assistant](#6-chatbot-assistant)
7. [Troubleshooting](#7-troubleshooting)

---

## 1. Getting Started

### 1.1 Accessing the Platform
Navigate to: **https://d2wbk4dmt2edww.cloudfront.net**

The platform works in any modern browser (Chrome, Firefox, Edge, Safari).

### 1.2 Role Selection
The platform supports two roles, toggled via the switcher in the top-right corner:

| Role | Description |
|------|------------|
| **GOV** | Government Contracting Officer — creates solicitations, reviews proposals, approves contracts, reviews invoices |
| **VENDOR** | Contractor — responds to solicitations, manages contracts, submits invoices |

Click the toggle to switch between GOV and VENDOR views. The interface adapts to show role-appropriate actions and data.

### 1.3 Demo Company
- **Demo Vendor**: Nexus AI Solutions LLC
- **Seed Vendor (GOV-visible only)**: Quantum Federal Systems LLC
- **Seed Vendor 2 (GOV-visible only)**: Atlas Defense Technologies

### 1.4 Resetting Data
To reset the application to its default state:
1. Open the chatbot (floating button, bottom-right)
2. Type **`reset`** and press Enter
3. The page will reload with fresh seed data

---

## 2. Dashboard

The Dashboard is the home page showing an overview of activity.

### 2.1 Sections (top to bottom)
1. **Obligation & Tracking** — Financial summary of contracts
2. **Notifications** — Role-targeted alerts requiring action
3. **Statistics** — Contract counts, pending items
4. **Recent Activity + Quick Actions** — History feed and actionable shortcuts
5. **System Status** — Agent health indicators

### 2.2 Notifications
- **GOV** sees vendor submissions (proposals, invoices, mod requests)
- **VENDOR** sees government actions (approvals, rejections, disbursements)
- Click notification buttons to take action directly

---

## 3. Solicitations

### 3.1 Government Actions

#### Create a Solicitation
1. Click **"Create Solicitation"** (top-right)
2. Fill in: Title, Type, Description, NAICS code, Value range, Close date
3. Optionally upload a PWS/SOW/RFP document
4. Click **"Create as Draft"**
5. Open the draft and click **"Publish"** to make it visible to vendors

#### Review Proposals
1. Click on a solicitation with proposals
2. Click **"Review Proposals"**
3. Wait for AI Evaluation to appear (processing animation shows)
4. Review: Score, CLIN breakdown, BOE allocation, recommendation
5. Click **"Approve & Award"** or **"Reject"**

#### Manual Approval (without AI Eval)
If the AI evaluation hasn't completed or failed:
1. Click **"Approve Manually (with Justification)"**
2. Enter justification (required)
3. Click **"Confirm Approval"**

### 3.2 Vendor Actions

#### Browse Solicitations
- **OPEN** solicitations: Available for proposal submission
- **AWARDED** solicitations: Shows "Awarded to You" or "Not Awarded" badge

#### Submit a Proposal
1. Click on an OPEN solicitation
2. Click **"Submit Proposal"**
3. Fill in:
   - Technical Approach (text) OR upload a proposal document
   - Price Proposal (optional — AI will estimate if blank)
   - Past Performance, Key Personnel
4. Click **"Submit Proposal"**
5. A processing animation shows while AI evaluates your proposal

#### Download RFP
Click the **"📥 Download RFP"** button on any solicitation with an attached document.

---

## 4. Contracts Management

### 4.1 Viewing Contracts
- **GOV** sees all contracts across vendors
- **VENDOR** sees only their company's contracts
- Contracts are sorted by upcoming POP (Period of Performance) end date

### 4.2 Contract Card
Click any contract card header to expand. The expanded view shows:
- **Period of Performance** — Progress bar showing elapsed time
- **Financial Summary** — Ceiling, Obligated, Expended, Remaining
- **Actions** — View CLINs, Submit Invoice (vendor), Request/Issue Mod
- **Payment Pipeline** — Visual step-by-step disbursement tracker

### 4.3 View CLINs
Click **"View CLINs"** to see detailed Contract Line Item information:
- CLIN Number, Description, Type (CPFF/FFP)
- Ceiling, Obligated, Expended, Remaining
- Risk indicator (GREEN/YELLOW/RED)
- Expenditure progress bar

### 4.4 Submit Invoice (Vendor)
1. Click **"Submit Invoice"** on your contract
2. Choose input method:
   - **Manual**: Select CLIN, enter amount and description
   - **File Upload**: Upload PDF/DOC, system simulates OCR extraction
3. Click **"Submit Invoice"**
4. System runs automated compliance check:
   - **Pass**: Invoice auto-disbursed, payment pipeline shows ✅
   - **Flagged**: GOV notified for review

### 4.5 Payment Pipeline
After submitting an invoice, the pipeline visualization shows:
```
📄 Submitted → 🔍 Compliance → ⚠️ Review → ✅ Approved → 💰 Disbursed
```
- Green checkmarks = completed steps
- Blue ring = current step
- Red X = rejected at review stage
- Shows compliance issues, rejection reasons, or disbursement confirmation

### 4.6 Contract Modifications
Click **"Request Mod"** (vendor) or **"Issue Mod"** (GOV):
- **REA** (Request for Equitable Adjustment) — Vendor-initiated
- **ECP** (Engineering Change Proposal) — Vendor-initiated
- **GOV_MOD** — Government-initiated modification

Pending mods show at the top of the contract card with approve/reject buttons.

### 4.7 Invoice Review (GOV)
Flagged invoices appear in a yellow panel at the top of the Contracts page:
1. Review compliance issues
2. Enter justification
3. Click **"Approve & Disburse"** or **"Reject"**

---

## 5. History

The History tab shows a chronological audit trail of all actions:
- Solicitation created/published
- Proposal submitted/approved/rejected
- Invoice submitted/approved/disbursed
- Contract modifications
- Payment events

Each entry shows: Actor, Action, Details, Timestamp.

---

## 6. Chatbot Assistant

### 6.1 Opening the Chatbot
Click the floating chat button (bottom-right corner of any page).

### 6.2 Available Commands

| Command | Action |
|---------|--------|
| `help` or `hello` | Shows help menu |
| `reset` | Clears all data, reloads with fresh seed |
| `clear` | Same as reset |
| `fresh start` | Same as reset |

### 6.3 Questions You Can Ask
- "How do I upload a document?"
- "What does RED risk mean?"
- "What are the payment stages?"
- "How do I submit an REA?"
- "What can a contractor do?"
- "What are the routing thresholds?"
- "What are the CLIN types?"
- "Explain risk levels"

---

## 7. Troubleshooting

| Issue | Solution |
|-------|----------|
| Page shows stale data | Type `reset` in chatbot |
| Proposal eval shows error | GOV can still approve manually with justification |
| Vendor doesn't see new contract | Switch to VENDOR role; contract appears after proposal approval |
| Can't submit proposal | Ensure solicitation is OPEN (not CLOSED or AWARDED) |
| CLIN shows RED risk | Expended exceeds ceiling — submit a mod request |
| Invoice rejected | Check compliance issues; resubmit with corrected amount |

### 7.1 Browser Requirements
- JavaScript enabled
- localStorage available (not in private/incognito in some browsers)
- Modern browser (ES2020+ support)

### 7.2 Support
For technical issues, contact the development team or use the in-app chatbot for guidance.
