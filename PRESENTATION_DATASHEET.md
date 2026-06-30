# Federal Payment Processing Platform — Presentation Datasheet

## Development Metrics

### Token Usage & Cost Estimate

| Metric | Value |
|--------|-------|
| **Total User Prompts** | ~45 prompts |
| **Estimated Input Tokens** | ~120,000 tokens |
| **Estimated Output Tokens** | ~850,000 tokens |
| **Estimated Total Tokens** | ~970,000 tokens |
| **Estimated Cost (Claude Sonnet via Kiro)** | ~$12–$18 (based on Anthropic API pricing: $3/M input, $15/M output) |

*Note: Exact token counts are not available from the Kiro IDE. These are estimates based on conversation length, code generated (~35,000+ lines), and sub-agent dispatches (~80+ parallel task executions).*

### Development Time

| Phase | Duration | Tasks Completed |
|-------|----------|-----------------|
| Spec Creation (Requirements + Design + Tasks) | ~10 min | 3 documents |
| Phase 1: Core Backend + Frontend | ~2 hours | 94 tasks |
| Phase 2: Chatbot, Multichannel, Correspondence | ~45 min | 15 tasks |
| UX Iterations (Contracts, Solicitations, Roles) | ~1 hour | 5 major refactors |
| Infrastructure + Deployment | ~20 min | CDK bootstrap + 8 deploys |
| **Total Development Time** | **~4.5 hours** | **113 spec tasks + 5 ad-hoc** |

### Code Output

| Category | Files | Lines of Code |
|----------|-------|---------------|
| Go Backend (agents, coordinator, portal, chatbot, correspondence, ingestion) | 65+ | ~12,000 |
| Go Tests (unit + property-based) | 40+ | ~8,000 |
| React Frontend (TypeScript) | 25+ | ~9,000 |
| AWS CDK Infrastructure (TypeScript) | 5 | ~300 |
| Documentation & Scripts | 4 | ~600 |
| **Total** | **~140 files** | **~30,000 lines** |

---

## AWS Services Used

### Compute

| Service | Purpose | Estimated Monthly Cost (Nominal) |
|---------|---------|----------------------------------|
| **AWS Lambda** | 5 AI agent functions + API handlers + WebSocket handler | $5–$15 (1M invocations/mo @ 256MB, 500ms avg) |
| **AWS Step Functions** | Payment pipeline orchestration | $2–$8 (10K state transitions/mo) |

### Storage

| Service | Purpose | Estimated Monthly Cost (Nominal) |
|---------|---------|----------------------------------|
| **Amazon DynamoDB** | Payment records, contracts, CLINs, REAs (4 tables w/ GSIs) | $5–$25 (on-demand, ~100K reads/writes per month) |
| **Amazon S3** | Document ingestion bucket + frontend hosting | $1–$5 (10GB storage, 10K requests) |

### AI/ML

| Service | Purpose | Estimated Monthly Cost (Nominal) |
|---------|---------|----------------------------------|
| **Amazon Bedrock (Claude Sonnet)** | Document extraction, FAR compliance evaluation, chatbot, correspondence generation | $50–$200 (depends on document volume; ~$3/1M input tokens, $15/1M output tokens) |

### Networking & Delivery

| Service | Purpose | Estimated Monthly Cost (Nominal) |
|---------|---------|----------------------------------|
| **Amazon CloudFront** | CDN for React SPA (global HTTPS delivery) | $1–$5 (100GB transfer/mo) |
| **Amazon API Gateway** | REST API + WebSocket API for real-time updates | $3–$10 (1M API calls/mo) |

### Security & Identity

| Service | Purpose | Estimated Monthly Cost (Nominal) |
|---------|---------|----------------------------------|
| **AWS IAM** | Role-based access (scoped per Lambda, least privilege) | $0 (included) |
| **AWS WAF** | Request filtering on API Gateway | $5–$10 |

### Observability

| Service | Purpose | Estimated Monthly Cost (Nominal) |
|---------|---------|----------------------------------|
| **AWS X-Ray** | Distributed tracing for Step Functions pipeline | $1–$5 |
| **Amazon CloudWatch** | Logs, metrics, alarms for all services | $5–$15 |

---

## Cost Summary (Nominal Monthly)

| Tier | Monthly Cost | Use Case |
|------|-------------|----------|
| **Development/Demo** | $5–$15 | Current state (CloudFront + S3 hosting only) |
| **Low Volume (100 payments/day)** | $75–$150 | Small agency pilot |
| **Medium Volume (1,000 payments/day)** | $200–$500 | Department-level deployment |
| **High Volume (10,000+ payments/day)** | $1,000–$3,000 | Enterprise/agency-wide |

*Bedrock is the primary cost driver at scale. Lambda and DynamoDB are pay-per-use and scale efficiently.*

---

## Architecture Highlights

- **5 Specialized AI Agents**: Document Processing, Validation, Compliance, Routing, Disbursement
- **Multi-Agent Orchestration**: AWS Step Functions with retry/escalation logic
- **Property-Based Testing**: 35+ formal correctness properties validated with `pgregory.net/rapid`
- **Real-Time Pipeline**: WebSocket-based payment tracking with live agent progress
- **Role-Based Access Control**: GOV and VENDOR views with vendor isolation
- **Persistent State**: localStorage (frontend demo) + DynamoDB (production)
- **Multichannel Ingestion**: Email, fax, mail scanning, portal upload
- **AI Chatbot**: Bedrock-powered assistant with RBAC-scoped data access
- **Correspondence Generation**: Automated professional letters for all payment actions
- **Handwriting Recognition**: Textract-based fallback for handwritten forms

---

## Live Deployment

| Resource | URL/Identifier |
|----------|---------------|
| **Frontend (CloudFront)** | https://d2wbk4dmt2edww.cloudfront.net |
| **GitHub Repository** | https://github.com/s2g1/AWS_Public_Hacks |
| **AWS Account** | 361274344489 |
| **Region** | us-east-1 |
| **CDK Stack** | InfraStack |

---

## Key Differentiators

1. **Spec-Driven Development**: Full requirements → design → tasks → implementation lifecycle
2. **Formal Correctness**: Property-based tests prove mathematical invariants (state machine integrity, obligation ceilings, confidence thresholds)
3. **End-to-End Workflow**: Solicitation → Proposal → Contract → Invoice → Compliance → Disbursement
4. **Human-in-the-Loop**: Configurable escalation thresholds, mandatory review for flagged items
5. **Zero Paper**: Aligned with March 2025 Executive Order to eliminate paper-based payments
6. **SBIR Lifecycle**: Full Phase I → Phase II → Phase III option management
7. **Anti-Deficiency Enforcement**: Obligations never exceed ceilings, programmatically enforced
