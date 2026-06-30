# Architecture Design Document (ADD)
## Federal Payment Processing Platform
### CDRL: DI-IPSC-81435A

---

## 1. System Architecture Overview

### 1.1 Architecture Style
The FedPay platform employs a **serverless multi-agent architecture** with the following key patterns:
- Event-driven microservices
- AI agent orchestration via state machine
- CQRS (Command Query Responsibility Segregation)
- Separation of concerns by domain boundary

### 1.2 High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    PRESENTATION LAYER                         │
│  React SPA (CloudFront) ←→ Lambda Function URLs              │
└──────────────────────────────┬──────────────────────────────┘
                               │
┌──────────────────────────────┼──────────────────────────────┐
│                    ORCHESTRATION LAYER                        │
│  AWS Step Functions (Payment Pipeline Coordinator)            │
│  ┌─────────┐ ┌──────────┐ ┌──────────┐ ┌────────┐ ┌──────┐ │
│  │Extract  │→│Validate  │→│Compliance│→│Route   │→│Disburse│ │
│  │Agent    │ │Agent     │ │Agent     │ │Agent   │ │Agent  │ │
│  └─────────┘ └──────────┘ └──────────┘ └────────┘ └──────┘ │
└──────────────────────────────┬──────────────────────────────┘
                               │
┌──────────────────────────────┼──────────────────────────────┐
│                    DATA LAYER                                 │
│  DynamoDB (Payments, Contracts, CLINs, REAs)                 │
│  S3 (Documents, Ingestion)                                   │
│  Bedrock (AI Inference — Nova Pro / Claude)                  │
└─────────────────────────────────────────────────────────────┘
```

---

## 2. Component Architecture

### 2.1 Frontend (Presentation Layer)

| Component | Purpose | Technology |
|-----------|---------|-----------|
| MainLayout | App shell, navigation, role switcher | React + Tailwind |
| Dashboard | KPIs, notifications, quick actions | React Context |
| Solicitations | Create/browse/respond to RFPs | React + evaluationService |
| Contracts | CLIN management, invoicing, mods, payment pipeline | React + AppContext |
| History | Audit trail of all user actions | React Context |
| ChatWidget | In-app assistant with data reset | Pattern-matching service |

### 2.2 Backend (Agent Layer)

| Agent | Responsibility | Inputs | Outputs |
|-------|---------------|--------|---------|
| Extraction Agent | Document classification + field extraction | Raw document (S3) | Structured fields, confidence scores |
| Validation Agent | Completeness check, duplicate detection | Extracted fields | Validation result, warnings |
| Compliance Agent | OFAC screening, FAR rule check | Payment record | Compliance status, flags |
| Routing Agent | Approval authority determination | Amount, risk level | Routing decision, priority |
| Disbursement Agent | EFT execution, notification | Approved payment | Disbursement confirmation |
| Evaluation Agent | Proposal-to-SOW analysis | Proposal text, SOW, price | CLIN breakdown, score, recommendation |

### 2.3 Infrastructure Layer

| Service | Configuration | Purpose |
|---------|--------------|---------|
| CloudFront | Global edge, HTTPS, SPA routing | Frontend delivery |
| S3 | Versioned, encrypted, lifecycle rules | Document storage |
| DynamoDB | On-demand, GSIs, PITR | Transactional data |
| Lambda | 256MB, 60s timeout, AL2023 runtime | Compute |
| Step Functions | Standard workflow, retry policies | Orchestration |
| IAM | Least-privilege, scoped per function | Access control |

---

## 3. Data Architecture

### 3.1 Data Models

**Payments Table** (Partition: paymentId)
- GSI: status-index (status + updatedAt)
- GSI: payee-index (payee + createdAt)

**Contracts Table** (Partition: contractId)
- Contains: ceiling, obligated, expended, POP dates

**CLINs Table** (Partition: contractId, Sort: clinId)
- Contains: type, ceiling, obligated, expended, status

**REAs Table** (Partition: reaId, Sort: contractId)
- Contains: amount, affected CLINs, status, justification

### 3.2 State Management (Frontend)
- Centralized React Context (`AppContext`)
- localStorage persistence under key `fedpay_app_state`
- Seed data for demo reset capability

---

## 4. Security Architecture

### 4.1 Authentication & Authorization
- Role-based access: GOV and VENDOR roles
- Vendor isolation: contractors see only their contracts
- Action-based permissions per role

### 4.2 Data Protection
- S3 encryption (SSE-S3)
- DynamoDB encryption at rest
- HTTPS everywhere (CloudFront → Lambda)
- No secrets in source code

### 4.3 Network Security
- S3 Block Public Access
- CloudFront Origin Access Control
- Lambda Function URL with CORS restrictions
- WAF-ready API Gateway (production)

---

## 5. Integration Architecture

### 5.1 External Interfaces

| Interface | Protocol | Purpose |
|-----------|----------|---------|
| Amazon Bedrock | HTTPS (SDK) | AI model inference |
| S3 | HTTPS (SDK) | Document ingestion/retrieval |
| DynamoDB | HTTPS (SDK) | Transactional data |
| CloudFront | HTTPS | Frontend delivery |

### 5.2 Internal Interfaces

| From | To | Method | Data |
|------|----|--------|------|
| Frontend | Evaluation Lambda | HTTPS POST | Proposal + SOW |
| Step Functions | Agent Lambdas | Invoke | Payment record |
| Agent Lambda | Bedrock | SDK InvokeModel | Prompt + document |
| Agent Lambda | DynamoDB | SDK PutItem/UpdateItem | State updates |

---

## 6. Scalability & Performance

### 6.1 Scaling Strategy
- Lambda: automatic concurrency scaling (up to 1000 concurrent)
- DynamoDB: on-demand capacity (auto-scales to workload)
- CloudFront: global edge caching
- S3: unlimited storage with lifecycle tiering

### 6.2 Performance Targets

| Metric | Target | Measured |
|--------|--------|----------|
| Page load time | <2s | ~1.5s (CloudFront cached) |
| AI evaluation latency | <15s | ~5-8s (Nova Pro) |
| Invoice compliance check | <500ms | ~200ms (client-side) |
| Deployment time | <3 min | ~90s (CDK) |

---

## 7. Availability & Disaster Recovery

| Component | RTO | RPO | Strategy |
|-----------|-----|-----|----------|
| Frontend | <5 min | 0 | CloudFront multi-AZ, S3 versioning |
| Lambda | <1 min | 0 | Multi-AZ by default |
| DynamoDB | <1 min | ~1s | PITR enabled, on-demand backup |
| S3 Documents | <5 min | 0 | Versioning, cross-region replication (production) |
