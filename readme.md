# Federal Payment Processing Platform (FedPay)
## DC Summit 2026 Global Government Hackathon — Team S2G

### Live Demo: https://d2wbk4dmt2edww.cloudfront.net
### GitHub: https://github.com/s2g1/AWS_Public_Hacks

---

## Overview

An AI-powered federal payment processing platform that automates the full acquisition lifecycle — from solicitation through proposal evaluation, contract award, invoice processing, compliance validation, and disbursement. Built with multi-agent AI architecture using Amazon Bedrock, deployed on AWS serverless infrastructure.

---

## Screenshots

### Dashboard
![Dashboard](docs/images/Dashboard%20page.PNG)

### Solicitations
![Solicitations](docs/images/solicitations%20page.PNG)

### AI Proposal Evaluation
![AI Summary](docs/images/AI%20summary.PNG)

### Contracts Management
![Contracts](docs/images/contracts%20management%20page.PNG)

### History / Audit Trail
![History](docs/images/history%20page.PNG)

---

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    PRESENTATION LAYER                         │
│  React SPA (CloudFront) ←→ Lambda Function URLs              │
└──────────────────────────────┬──────────────────────────────┘
                               │
┌──────────────────────────────┼──────────────────────────────┐
│                    AI / ORCHESTRATION LAYER                   │
│  Amazon Bedrock (Nova Pro / Claude Sonnet 4.6)               │
│  AWS Step Functions (Payment Pipeline)                        │
│  WebSocket API Gateway (Real-time Event Bus)                 │
│  ┌─────────┐ ┌──────────┐ ┌──────────┐ ┌────────┐ ┌──────┐ │
│  │Extract  │→│Validate  │→│Compliance│→│Route   │→│Disburse│ │
│  │Agent    │ │Agent     │ │Agent     │ │Agent   │ │Agent  │ │
│  └─────────┘ └──────────┘ └──────────┘ └────────┘ └──────┘ │
└──────────────────────────────┬──────────────────────────────┘
                               │
┌──────────────────────────────┼──────────────────────────────┐
│                    DATA LAYER                                 │
│  DynamoDB (Payments, Contracts, CLINs, REAs, WS Connections) │
│  S3 (Documents, Shared State, Frontend Assets)               │
└─────────────────────────────────────────────────────────────┘
```

---

## Key Features

- **5 Specialized AI Agents** — Extraction, Validation, Compliance, Routing, Disbursement
- **Real AI Evaluation** — Amazon Bedrock evaluates proposals against SOW in real-time
- **Multi-Browser Sync** — WebSocket event bus propagates state between GOV and VENDOR browsers
- **Payment Pipeline** — Per-contract visual step-by-step disbursement tracking
- **SBIR Lifecycle** — Full Phase I → Phase II → Phase III option management
- **Role-Based Access** — GOV/VENDOR views with vendor isolation
- **Contract Modifications** — REA, ECP, and Government-initiated mods
- **In-App Chatbot** — Pattern-matched assistant with data reset command

---

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Frontend | React 18, TypeScript, Tailwind CSS, Vite |
| Backend (Agents) | Go 1.21 |
| AI/ML | Amazon Bedrock (Nova Pro, Claude Sonnet 4.6) |
| Infrastructure | AWS CDK (TypeScript) |
| Database | Amazon DynamoDB |
| Storage | Amazon S3 |
| CDN | Amazon CloudFront |
| Real-time | API Gateway WebSocket |
| Serverless | AWS Lambda (Go) |
| Testing | Property-based (rapid), Vitest |

---

## Deployment

```powershell
# Build frontend
cd frontend && npm run build

# Build Lambdas
$env:GOOS="linux"; $env:GOARCH="amd64"; $env:CGO_ENABLED="0"
cd lambda/evaluate-proposal && go build -o dist/bootstrap .
cd ../state-sync && go build -o dist/bootstrap .
cd ../ws-connect && go build -o dist/bootstrap .

# Deploy infrastructure
cd infra && npx cdk deploy --require-approval never
```

---

## Documentation

| Document | Path |
|----------|------|
| Software Development Plan | `docs/SDP-Software-Development-Plan.md` |
| Architecture Design Document | `docs/ADD-Architecture-Design-Document.md` |
| Version Description Document | `docs/VDD-Version-Description-Document.md` |
| Interface Specification | `docs/ISA-Interface-Specification-Agreement.md` |
| PPSM (Security) | `docs/PPSM-Program-Protection-Security-Management.md` |
| STIG Compliance | `docs/STIG-Compliance-Report.md` |
| CMMI Compliance | `docs/CMMI-Process-Compliance.md` |
| User Manual | `docs/USER-MANUAL.md` |
| Presentation Datasheet | `PRESENTATION_DATASHEET.md` |
| Demo Script | `DEMO_SCRIPT.md` |

### Reference Documents
| Document | Path |
|----------|------|
| DARPA RFP (W15QKN-25-R-0042) | `docs/reference/RFP-W15QKN-25-R-0042.md` |
| S2G Proposal & BOE | `docs/reference/S2G-Proposal-BOE-W15QKN-25-R-0042.md` |
| S2G Invoice #001 | `docs/reference/S2G-Invoice-001-W15QKN-25-R-0042.md` |

---

## Demo Instructions

1. Open https://d2wbk4dmt2edww.cloudfront.net in two browsers
2. Browser 1: Set to **GOV** (Contracting Officer)
3. Browser 2: Set to **VENDOR** (Nexus AI Solutions LLC)
4. Type `reset` in the chatbot to initialize shared state
5. Walk through: Solicitation → Proposal → AI Eval → Award → Invoice → Disburse

---

## Team

**S2G Technologies** — DC Summit 2026 Global Government Hackathon

---

## AWS Services Used

CloudFront, S3, Lambda, DynamoDB, API Gateway (WebSocket), Bedrock, Step Functions, IAM, CloudWatch, CDK
