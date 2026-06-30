# Interface Specification Agreement (ISA)
## Federal Payment Processing Platform
### CDRL: DI-IPSC-81436A

---

## 1. Purpose

This document defines all external and internal interfaces for the Federal Payment Processing Platform, including data formats, protocols, error handling, and performance requirements.

---

## 2. External Interfaces

### 2.1 Proposal Evaluation API (Lambda Function URL)

| Field | Value |
|-------|-------|
| **Endpoint** | `https://yoviof6vsz5k6kzevbhazlmehy0joqii.lambda-url.us-east-1.on.aws/` |
| **Protocol** | HTTPS (TLS 1.2+) |
| **Method** | POST |
| **Authentication** | None (public, CORS-restricted) |
| **Content-Type** | application/json |
| **Max Payload** | 6 MB (Lambda limit) |
| **Timeout** | 60 seconds |

#### Request Schema
```json
{
  "proposalText": "string — technical approach narrative",
  "documentBase64": "string — base64-encoded file (optional)",
  "documentName": "string — original filename (optional)",
  "solicitationSOW": "string — statement of work description",
  "priceProposal": "number — total proposed price in USD",
  "companyName": "string — proposing company name"
}
```

#### Response Schema (200 OK)
```json
{
  "summary": "string — 2-3 sentence evaluation",
  "clinBreakdown": [
    {
      "clinNumber": "string — e.g., '0001'",
      "description": "string — CLIN description",
      "type": "string — 'CPFF' | 'FFP' | 'T&M'",
      "ceiling": "number — dollar amount",
      "obligated": "number — always 0 on creation",
      "expended": "number — always 0 on creation"
    }
  ],
  "boeAllocation": "string — human-readable BOE breakdown",
  "score": "integer — 70-98",
  "recommendation": "string — 'APPROVE' | 'REVIEW' | 'REJECT'"
}
```

#### Error Response (4xx/5xx)
```json
{
  "error": "string — human-readable error message"
}
```

---

### 2.2 Amazon Bedrock (AI Inference)

| Field | Value |
|-------|-------|
| **Service** | Amazon Bedrock Runtime |
| **Model** | us.amazon.nova-pro-v1:0 |
| **API** | InvokeModel |
| **Region** | us-east-1 |
| **Auth** | IAM Role (Lambda execution role) |

#### Request Format (Nova Pro)
```json
{
  "messages": [
    {
      "role": "user",
      "content": [{"text": "prompt string"}]
    }
  ],
  "inferenceConfig": {
    "maxTokens": 4096,
    "temperature": 0.3
  }
}
```

#### Response Format
```json
{
  "output": {
    "message": {
      "content": [{"text": "model response"}],
      "role": "assistant"
    }
  },
  "stopReason": "max_tokens | end_turn",
  "usage": {
    "inputTokens": 100,
    "outputTokens": 500
  }
}
```

---

### 2.3 Amazon DynamoDB

| Table | Partition Key | Sort Key | Purpose |
|-------|--------------|----------|---------|
| fedpay-payments-361274344489 | paymentId (S) | — | Payment lifecycle records |
| fedpay-contracts-361274344489 | contractId (S) | — | Contract financial data |
| fedpay-clins-361274344489 | contractId (S) | clinId (S) | Line item details |
| fedpay-reas-361274344489 | reaId (S) | contractId (S) | Equitable adjustments |

---

### 2.4 Amazon S3

| Bucket | Purpose | Access |
|--------|---------|--------|
| fedpay-portal-361274344489-us-east-1 | Frontend static assets | CloudFront OAC |
| fedpay-ingestion-361274344489-us-east-1 | Document uploads | Lambda role |

---

### 2.5 Amazon CloudFront

| Setting | Value |
|---------|-------|
| Distribution | d2wbk4dmt2edww.cloudfront.net |
| Origin | S3 (fedpay-portal bucket) |
| Protocol | HTTPS redirect |
| Cache Policy | CachingOptimized |
| Error Pages | 403/404 → /index.html (SPA routing) |

---

## 3. Internal Interfaces

### 3.1 Frontend ↔ AppContext (State Management)

| Action | Signature | Description |
|--------|-----------|-------------|
| submitProposal | `(solicitationId, data, file?) → void` | Creates proposal, triggers AI eval |
| approveProposal | `(proposalId) → void` | Awards contract from proposal |
| submitInvoice | `(contractId, clin, amount, desc) → void` | Submits + auto-compliance |
| approveInvoice | `(invoiceId, justification?) → void` | Disburses flagged invoice |
| submitContractMod | `(contractId, type, title, desc, amt) → void` | REA/ECP/GOV_MOD |

### 3.2 Frontend ↔ Evaluation Service

| Function | Input | Output | Fallback |
|----------|-------|--------|----------|
| evaluateProposal | Proposal details + file | EvaluationResult | Simulated eval (3-5s delay) |
| isBackendConfigured | — | boolean | — |

---

## 4. Error Handling

| Error Code | Meaning | Frontend Behavior |
|-----------|---------|-------------------|
| 400 | Invalid request | Show error toast |
| 403 | Model access denied | Fallback to simulated eval |
| 404 | Model not found | Fallback to simulated eval |
| 500 | Server error | Fallback to simulated eval, manual approval available |
| Network error | Lambda unreachable | Fallback to simulated eval |

---

## 5. Performance Requirements

| Interface | Latency Target | Throughput |
|-----------|---------------|-----------|
| Evaluation API | <15s (AI inference) | 10 concurrent |
| CloudFront | <2s page load | Unlimited (CDN) |
| DynamoDB | <10ms per operation | 25K RCU/WCU (on-demand) |
| S3 upload | <5s per document | 3500 PUT/s per prefix |
