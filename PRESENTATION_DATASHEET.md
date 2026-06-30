# Federal Payment Processing Platform — Presentation Datasheet

## Development Metrics

### Token Usage & Cost Estimate

| Metric | Value |
|--------|-------|
| **Total User Prompts** | ~65 prompts |
| **Estimated Input Tokens** | ~200,000 tokens |
| **Estimated Output Tokens** | ~1,400,000 tokens |
| **Estimated Total Tokens** | ~1,600,000 tokens |
| **Estimated Cost (Claude Sonnet via Kiro)** | ~$20–$28 (based on Anthropic API pricing: $3/M input, $15/M output) |

*Note: Exact token counts are not available from the Kiro IDE. These are estimates based on conversation length, code generated (~40,000+ lines), and sub-agent dispatches (~80+ parallel task executions). Includes Phase 2 MVP work: Bedrock Lambda integration, payment pipeline visualization, proposal evaluation service, and multiple deployment iterations.*

### Development Time

| Phase | Duration | Tasks Completed |
|-------|----------|-----------------|
| Spec Creation (Requirements + Design + Tasks) | ~10 min | 3 documents |
| Phase 1: Core Backend + Frontend | ~2 hours | 94 tasks |
| Phase 2: Chatbot, Multichannel, Correspondence | ~45 min | 15 tasks |
| UX Iterations (Contracts, Solicitations, Roles) | ~1 hour | 5 major refactors |
| Phase 3: AI Evaluation, Payment Pipeline, Bug Fixes | ~1.5 hours | Lambda + 8 iterations |
| Infrastructure + Deployment | ~30 min | CDK bootstrap + 15 deploys |
| **Total Development Time** | **~6 hours** | **113 spec tasks + 12 ad-hoc** |

### Code Output

| Category | Files | Lines of Code |
|----------|-------|---------------|
| Go Backend (agents, coordinator, portal, chatbot, correspondence, ingestion) | 65+ | ~12,000 |
| Go Lambda (proposal evaluation with Bedrock) | 3 | ~250 |
| Go Tests (unit + property-based) | 40+ | ~8,000 |
| React Frontend (TypeScript) | 28+ | ~11,000 |
| Frontend Services (evaluation, chat) | 3 | ~250 |
| AWS CDK Infrastructure (TypeScript) | 5 | ~350 |
| Documentation & Scripts | 4 | ~600 |
| **Total** | **~150 files** | **~32,500 lines** |

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
| **Amazon Bedrock (Nova Pro)** | Proposal-to-SOW evaluation, CLIN extraction, compliance analysis | $20–$100 (depends on proposal volume; ~$0.80/1M input tokens, $3.20/1M output tokens) |
| **Amazon Bedrock (Claude Sonnet 4.6)** | Document extraction, FAR compliance evaluation (when Marketplace access enabled) | $50–$200 ($3/1M input, $15/1M output) |
| **Lambda Function URL** | Serverless API for real-time proposal evaluation | Included in Lambda costs |

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
- **Real AI Evaluation**: Amazon Nova Pro (Bedrock) evaluates proposals against SOW in real-time
- **Multi-Agent Orchestration**: AWS Step Functions with retry/escalation logic
- **Property-Based Testing**: 35+ formal correctness properties validated with `pgregory.net/rapid`
- **Payment Pipeline Visualization**: Per-contract step-by-step disbursement tracking (Submit → Compliance → Review → Approve → Disburse)
- **Real-Time Pipeline**: WebSocket-based payment tracking with live agent progress
- **Role-Based Access Control**: GOV and VENDOR views with vendor isolation
- **Manual Approval with Justification**: GOV can approve proposals without AI eval using documented rationale
- **Persistent State**: localStorage (frontend demo) + DynamoDB (production)
- **Multichannel Ingestion**: Email, fax, mail scanning, portal upload
- **AI Chatbot**: Pattern-matched assistant with in-app reset command
- **Correspondence Generation**: Automated professional letters for all payment actions
- **Handwriting Recognition**: Textract-based fallback for handwritten forms
- **Lambda Function URL**: Serverless API endpoint for async AI evaluation (no API Gateway needed)

---

## Live Deployment

| Resource | URL/Identifier |
|----------|---------------|
| **Frontend (CloudFront)** | https://d2wbk4dmt2edww.cloudfront.net |
| **Evaluation Lambda URL** | https://yoviof6vsz5k6kzevbhazlmehy0joqii.lambda-url.us-east-1.on.aws/ |
| **GitHub Repository** | https://github.com/s2g1/AWS_Public_Hacks |
| **AWS Account** | 361274344489 |
| **Region** | us-east-1 |
| **CDK Stack** | InfraStack |
| **Bedrock Model** | us.amazon.nova-pro-v1:0 (Amazon Nova Pro) |

---

## Key Differentiators

1. **Spec-Driven Development**: Full requirements → design → tasks → implementation lifecycle
2. **Formal Correctness**: Property-based tests prove mathematical invariants (state machine integrity, obligation ceilings, confidence thresholds)
3. **End-to-End Workflow**: Solicitation → Proposal → Contract → Invoice → Compliance → Disbursement
4. **Human-in-the-Loop**: Configurable escalation thresholds, mandatory review for flagged items
5. **Zero Paper**: Aligned with March 2025 Executive Order to eliminate paper-based payments
6. **SBIR Lifecycle**: Full Phase I → Phase II → Phase III option management
7. **Anti-Deficiency Enforcement**: Obligations never exceed ceilings, programmatically enforced
