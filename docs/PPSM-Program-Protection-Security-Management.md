# Program Protection & Security Management Plan (PPSM)
## Federal Payment Processing Platform
### Contract: FA8750-25-F-0018

---

## 1. Purpose

This Program Protection & Security Management Plan establishes the security framework, threat posture, and protective measures for the Federal Payment Processing Platform. It addresses anti-tamper, cybersecurity, supply chain risk, and information assurance requirements per DoDI 5200.39 and DoDI 5000.83.

---

## 2. Critical Program Information (CPI)

### 2.1 CPI Identification

| CPI Element | Classification | Protection Level |
|-------------|---------------|-----------------|
| AI Agent Orchestration Logic | CUI | HIGH |
| Compliance Rule Engine | CUI | HIGH |
| Disbursement Routing Algorithms | CUI | HIGH |
| Vendor Proposal Data | CUI/PII | HIGH |
| Contract Financial Data | CUI | MEDIUM |
| CLIN Structures | UNCLASSIFIED | MEDIUM |

### 2.2 CPI Threats

| Threat | Likelihood | Impact | Countermeasure |
|--------|-----------|--------|---------------|
| Unauthorized access to payment data | Medium | Critical | RBAC, IAM least privilege |
| AI model manipulation/injection | Low | High | Input validation, prompt hardening |
| Supply chain compromise (dependencies) | Low | High | Pinned versions, vulnerability scanning |
| Insider threat (data exfiltration) | Low | Critical | Audit logging, DLP controls |
| DDoS against evaluation API | Medium | Medium | WAF, Lambda concurrency limits |

---

## 3. Security Architecture

### 3.1 Defense in Depth Layers

```
┌─────────────────────────────────────────────┐
│ Layer 1: Edge Security                       │
│ CloudFront + WAF + TLS 1.2                  │
├─────────────────────────────────────────────┤
│ Layer 2: Application Security                │
│ Input validation, CORS, CSP headers          │
├─────────────────────────────────────────────┤
│ Layer 3: Identity & Access                   │
│ IAM roles (least privilege), RBAC            │
├─────────────────────────────────────────────┤
│ Layer 4: Data Security                       │
│ Encryption at rest (AES-256), in transit     │
├─────────────────────────────────────────────┤
│ Layer 5: Monitoring & Response               │
│ CloudWatch, CloudTrail, X-Ray               │
└─────────────────────────────────────────────┘
```

### 3.2 IAM Role Inventory

| Role | Permissions | Scope |
|------|------------|-------|
| fedpay-agent-execution | DynamoDB RW, S3 RW, Bedrock Invoke | All tables, ingestion bucket |
| EvaluateProposalFn-ServiceRole | Bedrock Invoke, Marketplace View, CloudWatch Logs | Global models |
| CDK Deploy Role | CloudFormation, S3, Lambda, IAM | Stack resources only |

### 3.3 Data Classification & Handling

| Data Type | Classification | Storage | Encryption | Retention |
|-----------|---------------|---------|-----------|-----------|
| Payment Records | CUI | DynamoDB | AES-256 | 7 years |
| Contract Data | CUI | DynamoDB | AES-256 | Contract life + 6 years |
| Uploaded Documents | CUI | S3 | SSE-S3 | 90 days (lifecycle) |
| Frontend State | UNCLASSIFIED | localStorage | None (demo) | Session |
| Audit Logs | CUI | CloudWatch | AES-256 | 1 year |

---

## 4. Anti-Tamper Measures

| Control | Implementation |
|---------|---------------|
| Code integrity | Git signed commits, branch protection |
| Infrastructure immutability | CDK-managed, no manual AWS Console changes |
| Build reproducibility | Pinned dependencies, deterministic Go builds |
| Deployment verification | CDK diff review before apply |
| AI model integrity | Bedrock-managed models (AWS controls versioning) |

---

## 5. Supply Chain Risk Management (SCRM)

### 5.1 Dependency Management

| Ecosystem | Tool | Policy |
|-----------|------|--------|
| Go | go.mod | Pinned versions, `go mod verify` |
| Node.js | package-lock.json | Exact versions, npm audit |
| CDK | package-lock.json | Exact versions |

### 5.2 Third-Party Services

| Service | Provider | Risk Level | Mitigation |
|---------|----------|-----------|-----------|
| Amazon Bedrock | AWS | Low | AWS responsibility for model hosting |
| GitHub | Microsoft | Low | Private repo option, branch protection |
| CloudFront | AWS | Low | AWS managed infrastructure |

---

## 6. Incident Response

### 6.1 Response Procedures

| Severity | Response Time | Notification | Action |
|----------|--------------|-------------|--------|
| Critical (data breach) | <1 hour | CO + ISSM + ISSO | Isolate, investigate, report to DIB |
| High (unauthorized access) | <4 hours | CO + PM | Revoke access, review logs |
| Medium (service disruption) | <24 hours | PM | Restore service, root cause |
| Low (informational) | <72 hours | Development team | Log, fix in next sprint |

### 6.2 Audit Trail
- All state transitions logged with actor, timestamp, and details
- CloudTrail captures all AWS API calls
- Application-level history maintained per user action
- Immutable audit entries (append-only)

---

## 7. Compliance Mapping

| Requirement | Standard | Status |
|-------------|----------|--------|
| Access Control | NIST 800-53 AC-1 through AC-25 | Implemented (RBAC) |
| Audit & Accountability | NIST 800-53 AU-1 through AU-16 | Implemented (History tab) |
| Identification & Auth | NIST 800-53 IA-1 through IA-12 | Partial (demo mode) |
| System & Comms Protection | NIST 800-53 SC-1 through SC-44 | Implemented (TLS, encryption) |
| Configuration Management | NIST 800-53 CM-1 through CM-11 | Implemented (CDK, Git) |

---

## 8. Continuous Monitoring

| Activity | Frequency | Tool |
|----------|-----------|------|
| Vulnerability scanning | Weekly | npm audit, govulncheck |
| Access review | Monthly | IAM Access Analyzer |
| Penetration testing | Quarterly | Third-party assessment |
| Security metrics review | Monthly | CloudWatch dashboards |
| Dependency updates | Bi-weekly | Dependabot / manual |
