// Proposal Evaluation Service
// Calls the Bedrock-powered Lambda for AI evaluation of proposals against SOW

export interface CLINBreakdown {
  clinNumber: string
  description: string
  type: string
  ceiling: number
  obligated: number
  expended: number
}

export interface EvaluationResult {
  summary: string
  clinBreakdown: CLINBreakdown[]
  boeAllocation: string
  score: number
  recommendation: 'APPROVE' | 'REVIEW' | 'REJECT'
}

interface EvaluationRequest {
  proposalText: string
  documentBase64: string
  documentName: string
  solicitationSOW: string
  priceProposal: number
  companyName: string
}

// The Lambda Function URL - calls Amazon Nova Pro for AI evaluation
const EVALUATE_PROPOSAL_URL = import.meta.env.VITE_EVALUATE_PROPOSAL_URL || ''

/**
 * Convert a File to base64 string
 */
async function fileToBase64(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => {
      const result = reader.result as string
      // Remove data:...;base64, prefix
      const base64 = result.split(',')[1] || ''
      resolve(base64)
    }
    reader.onerror = reject
    reader.readAsDataURL(file)
  })
}

/**
 * Call the Bedrock Lambda to evaluate a proposal against the SOW
 */
export async function evaluateProposal(params: {
  proposalText: string
  documentFile?: File
  solicitationSOW: string
  priceProposal: number
  companyName: string
}): Promise<EvaluationResult> {
  let documentBase64 = ''
  let documentName = ''

  if (params.documentFile) {
    documentBase64 = await fileToBase64(params.documentFile)
    documentName = params.documentFile.name
  }

  const request: EvaluationRequest = {
    proposalText: params.proposalText,
    documentBase64,
    documentName,
    solicitationSOW: params.solicitationSOW,
    priceProposal: params.priceProposal,
    companyName: params.companyName,
  }

  // If Lambda URL is configured, call real backend
  if (EVALUATE_PROPOSAL_URL) {
    const response = await fetch(EVALUATE_PROPOSAL_URL, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(request),
    })

    if (!response.ok) {
      const err = await response.json().catch(() => ({ error: 'Unknown error' }))
      throw new Error(err.error || `Evaluation failed: ${response.status}`)
    }

    return await response.json()
  }

  // Fallback: simulated evaluation (when Lambda URL not configured)
  return simulateEvaluation(params.priceProposal, params.proposalText, params.solicitationSOW)
}

/**
 * Simulated evaluation — produces realistic, contextual demo data
 * Uses proposal text and SOW to generate relevant summaries
 */
function simulateEvaluation(priceProposal: number, proposalText: string, _sow: string): Promise<EvaluationResult> {
  return new Promise((resolve) => {
    // Simulate processing time (3-5 seconds for realistic feel)
    const delay = 3000 + Math.random() * 2000
    setTimeout(() => {
      let price = priceProposal
      if (!price || price === 0) {
        price = 2500000 // Default estimate
      }

      const rdAmount = Math.round(price * 0.55)
      const integrationAmount = Math.round(price * 0.30)
      const pmAmount = price - rdAmount - integrationAmount

      const score = Math.floor(Math.random() * 16) + 80 // 80-95 for demo (always favorable)

      let recommendation: 'APPROVE' | 'REVIEW' | 'REJECT'
      if (score > 85) recommendation = 'APPROVE'
      else if (score >= 70) recommendation = 'REVIEW'
      else recommendation = 'REJECT'

      // Generate contextual summary from proposal text
      const techSnippet = proposalText && proposalText.length > 10
        ? proposalText.split('.')[0].trim()
        : 'the proposed technical solution'

      const strengths = score > 85
        ? 'The proposal exhibits strong technical merit and clear alignment with SOW objectives.'
        : 'The proposal demonstrates adequate technical understanding with some areas requiring clarification.'

      const summary = `${strengths} Proposed approach: "${techSnippet.slice(0, 120)}${techSnippet.length > 120 ? '...' : ''}". Total contract value of $${price.toLocaleString()} is allocated across 3 CLINs with appropriate cost-type designations. R&D effort (CLIN 0001, $${rdAmount.toLocaleString()}) represents 55% — consistent with Phase II SBIR standards. Integration & Testing (CLIN 0002, $${integrationAmount.toLocaleString()}) at 30% and Program Management (CLIN 0003, $${pmAmount.toLocaleString()}) at 15% reflect industry norms.`

      resolve({
        summary,
        clinBreakdown: [
          { clinNumber: '0001', description: 'Research & Development', type: 'CPFF', ceiling: rdAmount, obligated: 0, expended: 0 },
          { clinNumber: '0002', description: 'System Integration & Testing', type: 'CPFF', ceiling: integrationAmount, obligated: 0, expended: 0 },
          { clinNumber: '0003', description: 'Program Management', type: 'FFP', ceiling: pmAmount, obligated: 0, expended: 0 },
        ],
        boeAllocation: `R&D: 55% ($${rdAmount.toLocaleString()}) | Integration: 30% ($${integrationAmount.toLocaleString()}) | PM: 15% ($${pmAmount.toLocaleString()}) | Total: $${price.toLocaleString()}`,
        score,
        recommendation,
      })
    }, delay)
  })
}

/**
 * Check if the real Lambda backend is configured
 */
export function isBackendConfigured(): boolean {
  return !!EVALUATE_PROPOSAL_URL
}
