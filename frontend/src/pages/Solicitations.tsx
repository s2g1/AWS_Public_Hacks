import { useState, useMemo } from 'react'
import { useAppContext, Solicitation, Proposal, AIEvaluation } from '../store/AppContext'

// --- Local Types ---

type SolicitationStatus = 'DRAFT' | 'OPEN' | 'CLOSED' | 'AWARDED' | 'CANCELLED'
type SolicitationType = 'SBIR Phase I' | 'SBIR Phase II' | 'Full & Open Competition' | 'Sole Source'

interface ProposalForm {
  companyName: string
  technicalApproach: string
  priceProposal: string
  pastPerformance: string
  keyPersonnel: string
  attachmentName: string
}

interface CreateSolicitationForm {
  title: string
  type: SolicitationType
  description: string
  naicsCode: string
  estimatedValue: string
  closeDate: string
  evaluationCriteria: string
  attachmentName: string
}

// --- Utility Functions ---

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
  })
}

function getDaysRemaining(closeDate: string): number {
  const now = new Date()
  const close = new Date(closeDate)
  const diff = close.getTime() - now.getTime()
  return Math.ceil(diff / (1000 * 60 * 60 * 24))
}

function getStatusBadgeClasses(status: SolicitationStatus): string {
  switch (status) {
    case 'DRAFT':
      return 'bg-gray-100 text-gray-700 border-gray-300'
    case 'OPEN':
      return 'bg-green-100 text-green-800 border-green-300'
    case 'CLOSED':
      return 'bg-blue-100 text-blue-800 border-blue-300'
    case 'AWARDED':
      return 'bg-purple-100 text-purple-800 border-purple-300'
    case 'CANCELLED':
      return 'bg-red-100 text-red-800 border-red-300'
  }
}

// --- Status Filter Tabs ---

function StatusTabs({
  active,
  onChange,
  counts,
}: {
  active: string
  onChange: (tab: string) => void
  counts: Record<string, number>
}) {
  const tabs = [
    { key: 'ALL', label: 'All' },
    { key: 'OPEN', label: 'Open' },
    { key: 'CLOSED', label: 'Closed' },
    { key: 'AWARDED', label: 'Awarded' },
    { key: 'DRAFT', label: 'Draft' },
  ]

  return (
    <div className="flex gap-1 bg-gray-100 rounded-lg p-1">
      {tabs.map((tab) => (
        <button
          key={tab.key}
          onClick={() => onChange(tab.key)}
          className={`px-3 py-1.5 rounded-md text-sm font-medium transition-colors ${
            active === tab.key
              ? 'bg-white text-gray-900 shadow-sm'
              : 'text-gray-600 hover:text-gray-900'
          }`}
        >
          {tab.label}
          {counts[tab.key] !== undefined && (
            <span className="ml-1.5 text-xs text-gray-400">
              {counts[tab.key]}
            </span>
          )}
        </button>
      ))}
    </div>
  )
}

// --- Create Solicitation Modal ---

function CreateSolicitationModal({
  onClose,
  onSubmit,
}: {
  onClose: () => void
  onSubmit: (form: CreateSolicitationForm) => void
}) {
  const [form, setForm] = useState<CreateSolicitationForm>({
    title: '',
    type: 'SBIR Phase I',
    description: '',
    naicsCode: '',
    estimatedValue: '',
    closeDate: '',
    evaluationCriteria: '',
    attachmentName: '',
  })
  const [errors, setErrors] = useState<Record<string, string>>({})

  function validate(): boolean {
    const newErrors: Record<string, string> = {}
    if (!form.title.trim()) newErrors.title = 'Title is required'
    if (!form.closeDate) {
      newErrors.closeDate = 'Close date is required'
    } else {
      const close = new Date(form.closeDate)
      if (close <= new Date()) {
        newErrors.closeDate = 'Close date must be in the future'
      }
    }
    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (validate()) {
      onSubmit(form)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
      <div className="bg-white rounded-xl shadow-xl w-full max-w-2xl max-h-[90vh] overflow-y-auto">
        <div className="flex items-center justify-between px-6 py-4 border-b border-gray-200">
          <h2 className="text-lg font-semibold text-gray-900">Create New Solicitation</h2>
          <button onClick={onClose} className="text-gray-400 hover:text-gray-600">
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <form onSubmit={handleSubmit} className="p-6 space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Title <span className="text-red-500">*</span>
            </label>
            <input
              type="text"
              value={form.title}
              onChange={(e) => setForm({ ...form, title: e.target.value })}
              className={`w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500 ${
                errors.title ? 'border-red-300' : 'border-gray-300'
              }`}
              placeholder="e.g., Autonomous Payment Processing System"
            />
            {errors.title && <p className="mt-1 text-xs text-red-600">{errors.title}</p>}
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Type</label>
            <select
              value={form.type}
              onChange={(e) => setForm({ ...form, type: e.target.value as SolicitationType })}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
            >
              <option value="SBIR Phase I">SBIR Phase I</option>
              <option value="SBIR Phase II">SBIR Phase II</option>
              <option value="Full & Open Competition">Full & Open Competition</option>
              <option value="Sole Source">Sole Source</option>
            </select>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Description / SOW Summary</label>
            <textarea
              value={form.description}
              onChange={(e) => setForm({ ...form, description: e.target.value })}
              rows={4}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
              placeholder="Provide a summary of the Statement of Work..."
            />
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">NAICS Code</label>
              <input
                type="text"
                value={form.naicsCode}
                onChange={(e) => setForm({ ...form, naicsCode: e.target.value })}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                placeholder="e.g., 541512"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Estimated Value Range</label>
              <input
                type="text"
                value={form.estimatedValue}
                onChange={(e) => setForm({ ...form, estimatedValue: e.target.value })}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                placeholder="e.g., $500K - $1M"
              />
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Close Date <span className="text-red-500">*</span>
            </label>
            <input
              type="date"
              value={form.closeDate}
              onChange={(e) => setForm({ ...form, closeDate: e.target.value })}
              className={`w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500 ${
                errors.closeDate ? 'border-red-300' : 'border-gray-300'
              }`}
            />
            {errors.closeDate && <p className="mt-1 text-xs text-red-600">{errors.closeDate}</p>}
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Evaluation Criteria</label>
            <textarea
              value={form.evaluationCriteria}
              onChange={(e) => setForm({ ...form, evaluationCriteria: e.target.value })}
              rows={3}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
              placeholder="e.g., Technical Approach (40%), Past Performance (30%), Price (30%)"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Upload PWS/SOW/RFP Document</label>
            <input
              type="file"
              accept=".pdf,.doc,.docx,.txt,.zip"
              onChange={(e) => {
                const file = e.target.files?.[0]
                if (file) setForm({ ...form, attachmentName: file.name })
              }}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500 file:mr-4 file:py-1 file:px-3 file:rounded file:border-0 file:text-sm file:font-medium file:bg-blue-50 file:text-blue-700 hover:file:bg-blue-100"
            />
            {form.attachmentName && (
              <p className="mt-1 text-xs text-green-600">📎 {form.attachmentName}</p>
            )}
          </div>

          <div className="flex justify-end gap-3 pt-4 border-t border-gray-200">
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50"
            >
              Cancel
            </button>
            <button
              type="submit"
              className="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-lg hover:bg-blue-700 focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
            >
              Create as Draft
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}

// --- Proposal Submission Modal ---

function ProposalModal({
  solicitation,
  onClose,
  onSubmit,
  vendorCompany,
  onFileSelected,
}: {
  solicitation: Solicitation
  onClose: () => void
  onSubmit: (form: ProposalForm) => void
  vendorCompany: string
  onFileSelected: (file: File | undefined) => void
}) {
  const [form, setForm] = useState<ProposalForm>({
    companyName: vendorCompany,
    technicalApproach: '',
    priceProposal: '',
    pastPerformance: '',
    keyPersonnel: '',
    attachmentName: '',
  })
  const [errors, setErrors] = useState<Record<string, string>>({})

  function validate(): boolean {
    const newErrors: Record<string, string> = {}
    // Require either technicalApproach text OR an attachment
    if (!form.technicalApproach.trim() && !form.attachmentName) {
      newErrors.technicalApproach = 'Either a technical approach or an uploaded proposal document is required'
    }
    // priceProposal is optional — defaults to 0 if empty
    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (validate()) {
      onSubmit({ ...form, companyName: vendorCompany })
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
      <div className="bg-white rounded-xl shadow-xl w-full max-w-2xl max-h-[90vh] overflow-y-auto">
        <div className="flex items-center justify-between px-6 py-4 border-b border-gray-200">
          <div>
            <h2 className="text-lg font-semibold text-gray-900">Submit Proposal</h2>
            <p className="text-xs text-gray-500 mt-0.5">{solicitation.solicitationNumber} — {solicitation.title}</p>
          </div>
          <button onClick={onClose} className="text-gray-400 hover:text-gray-600">
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <form onSubmit={handleSubmit} className="p-6 space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Company Name
            </label>
            <input
              type="text"
              value={vendorCompany}
              disabled
              className="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm bg-gray-50 text-gray-700 cursor-not-allowed"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Technical Approach {!form.attachmentName && <span className="text-red-500">*</span>}
            </label>
            <textarea
              value={form.technicalApproach}
              onChange={(e) => setForm({ ...form, technicalApproach: e.target.value })}
              rows={5}
              className={`w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500 ${
                errors.technicalApproach ? 'border-red-300' : 'border-gray-300'
              }`}
              placeholder="Describe your technical approach to meeting the requirements... (or upload a proposal document below)"
            />
            {errors.technicalApproach && <p className="mt-1 text-xs text-red-600">{errors.technicalApproach}</p>}
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Price Proposal ($)
            </label>
            <input
              type="text"
              value={form.priceProposal}
              onChange={(e) => setForm({ ...form, priceProposal: e.target.value })}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
              placeholder="e.g., 3500000 (optional if included in uploaded document)"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Past Performance Summary</label>
            <textarea
              value={form.pastPerformance}
              onChange={(e) => setForm({ ...form, pastPerformance: e.target.value })}
              rows={3}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
              placeholder="Describe relevant past contract performance..."
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Key Personnel</label>
            <textarea
              value={form.keyPersonnel}
              onChange={(e) => setForm({ ...form, keyPersonnel: e.target.value })}
              rows={3}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
              placeholder="List key personnel and their qualifications..."
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">📤 Upload Proposal Document</label>
            <input
              type="file"
              accept=".pdf,.doc,.docx,.txt,.zip"
              onChange={(e) => {
                const file = e.target.files?.[0]
                if (file) {
                  setForm({ ...form, attachmentName: file.name })
                  onFileSelected(file)
                } else {
                  onFileSelected(undefined)
                }
              }}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500 file:mr-4 file:py-1 file:px-3 file:rounded file:border-0 file:text-sm file:font-medium file:bg-green-50 file:text-green-700 hover:file:bg-green-100"
            />
            {form.attachmentName && (
              <p className="mt-1 text-xs text-green-600">📎 {form.attachmentName}</p>
            )}
          </div>

          <div className="flex justify-end gap-3 pt-4 border-t border-gray-200">
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50"
            >
              Cancel
            </button>
            <button
              type="submit"
              className="px-4 py-2 text-sm font-medium text-white bg-green-600 rounded-lg hover:bg-green-700 focus:ring-2 focus:ring-green-500 focus:ring-offset-2"
            >
              Submit Proposal
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}

// --- Proposal Actions (handles manual approval with justification) ---

function ProposalActions({
  proposal,
  isEvaluating,
  onApprove,
  onReject,
}: {
  proposal: Proposal
  isEvaluating: boolean
  onApprove: (proposalId: string) => void
  onReject: (proposalId: string) => void
}) {
  const [showJustification, setShowJustification] = useState(false)
  const [justification, setJustification] = useState('')
  const hasEval = !!proposal.aiEvaluation

  // If AI eval exists, allow direct approve
  if (hasEval) {
    return (
      <div className="flex gap-2 mt-3">
        <button
          onClick={() => onApprove(proposal.id)}
          className="px-3 py-1.5 text-xs font-medium text-white bg-green-600 rounded-lg hover:bg-green-700"
        >
          Approve & Award
        </button>
        <button
          onClick={() => onReject(proposal.id)}
          className="px-3 py-1.5 text-xs font-medium text-red-700 bg-red-50 border border-red-200 rounded-lg hover:bg-red-100"
        >
          Reject
        </button>
      </div>
    )
  }

  // No AI eval — require justification for manual approval
  return (
    <div className="mt-3 space-y-2">
      {isEvaluating ? (
        <p className="text-xs text-blue-600 italic">⏳ AI evaluation in progress — you can approve manually below</p>
      ) : (
        <div className="bg-amber-50 border border-amber-200 rounded-lg p-2">
          <p className="text-xs text-amber-800">⚠️ AI evaluation unavailable. Manual approval requires justification.</p>
        </div>
      )}

      {!showJustification ? (
        <div className="flex gap-2">
          <button
            onClick={() => setShowJustification(true)}
            className="px-3 py-1.5 text-xs font-medium text-white bg-green-600 rounded-lg hover:bg-green-700"
          >
            Approve Manually (with Justification)
          </button>
          <button
            onClick={() => onReject(proposal.id)}
            className="px-3 py-1.5 text-xs font-medium text-red-700 bg-red-50 border border-red-200 rounded-lg hover:bg-red-100"
          >
            Reject
          </button>
        </div>
      ) : (
        <div className="space-y-2">
          <textarea
            value={justification}
            onChange={(e) => setJustification(e.target.value)}
            rows={2}
            className="w-full px-3 py-2 border border-gray-300 rounded-lg text-xs focus:ring-2 focus:ring-green-500 focus:border-green-500"
            placeholder="Enter justification for manual approval (e.g., CO review completed, price fair & reasonable determination)..."
          />
          <div className="flex gap-2">
            <button
              onClick={() => {
                if (justification.trim()) onApprove(proposal.id)
              }}
              disabled={!justification.trim()}
              className={`px-3 py-1.5 text-xs font-medium text-white rounded-lg ${
                justification.trim() ? 'bg-green-600 hover:bg-green-700' : 'bg-gray-300 cursor-not-allowed'
              }`}
            >
              Confirm Approval
            </button>
            <button
              onClick={() => { setShowJustification(false); setJustification('') }}
              className="px-3 py-1.5 text-xs font-medium text-gray-600 bg-gray-100 rounded-lg hover:bg-gray-200"
            >
              Cancel
            </button>
          </div>
        </div>
      )}
    </div>
  )
}

// --- Proposals Review Panel (GOV) ---

function ProposalsReviewPanel({
  solicitation,
  proposals,
  onApprove,
  onReject,
  onClose,
  evaluatingProposals,
}: {
  solicitation: Solicitation
  proposals: Proposal[]
  onApprove: (proposalId: string) => void
  onReject: (proposalId: string) => void
  onClose: () => void
  evaluatingProposals: Set<string>
}) {
  const solProposals = proposals.filter(p => p.solicitationId === solicitation.id)

  function getRecommendationColor(rec: AIEvaluation['recommendation']) {
    switch (rec) {
      case 'APPROVE': return 'bg-green-100 text-green-800 border-green-200'
      case 'REVIEW': return 'bg-yellow-100 text-yellow-800 border-yellow-200'
      case 'REJECT': return 'bg-red-100 text-red-800 border-red-200'
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
      <div className="bg-white rounded-xl shadow-xl w-full max-w-4xl max-h-[90vh] overflow-y-auto">
        <div className="flex items-center justify-between px-6 py-4 border-b border-gray-200">
          <div>
            <h2 className="text-lg font-semibold text-gray-900">Review Proposals</h2>
            <p className="text-xs text-gray-500 mt-0.5">{solicitation.solicitationNumber} — {solicitation.title}</p>
          </div>
          <button onClick={onClose} className="text-gray-400 hover:text-gray-600">
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <div className="p-6">
          {solProposals.length === 0 ? (
            <div className="text-center py-8">
              <p className="text-sm text-gray-500">No proposals submitted yet</p>
            </div>
          ) : (
            <div className="space-y-6">
              {solProposals.map(proposal => (
                <div key={proposal.id} className="border border-gray-200 rounded-lg p-4">
                  <div className="flex items-center justify-between mb-2">
                    <h3 className="font-semibold text-gray-900">{proposal.companyName}</h3>
                    <div className="flex items-center gap-2">
                      {proposal.attachmentName && (
                        <span className="text-xs text-blue-600 bg-blue-50 px-2 py-0.5 rounded">📎 {proposal.attachmentName}</span>
                      )}
                      <span className={`px-2 py-0.5 rounded-full text-xs font-medium ${
                        proposal.status === 'SUBMITTED' ? 'bg-yellow-100 text-yellow-800' :
                        proposal.status === 'APPROVED' ? 'bg-green-100 text-green-800' :
                        proposal.status === 'REJECTED' ? 'bg-red-100 text-red-800' :
                        'bg-blue-100 text-blue-800'
                      }`}>
                        {proposal.status}
                      </span>
                    </div>
                  </div>
                  <p className="text-sm text-gray-600 mb-2">{proposal.technicalApproach}</p>
                  <div className="flex items-center gap-4 text-xs text-gray-500 mb-3">
                    <span>Price: ${proposal.priceProposal.toLocaleString()}</span>
                    <span>Submitted: {new Date(proposal.submittedAt).toLocaleDateString()}</span>
                  </div>

                  {/* AI Evaluation Section */}
                  {evaluatingProposals.has(proposal.id) && !proposal.aiEvaluation && (
                    <div className="mt-3 bg-blue-50 border border-blue-200 rounded-lg p-6">
                      <div className="flex flex-col items-center gap-3">
                        <div className="relative w-12 h-12">
                          <div className="absolute inset-0 rounded-full border-4 border-blue-200"></div>
                          <div className="absolute inset-0 rounded-full border-4 border-blue-600 border-t-transparent animate-spin"></div>
                        </div>
                        <div className="text-center">
                          <h4 className="text-sm font-semibold text-blue-900">🤖 AI Evaluation in Progress</h4>
                          <p className="text-xs text-blue-700 mt-1">Analyzing proposal against SOW requirements...</p>
                          <p className="text-xs text-blue-500 mt-0.5">Extracting CLIN structure & generating BOE allocation</p>
                        </div>
                        <div className="flex gap-1 mt-1">
                          <div className="w-2 h-2 bg-blue-400 rounded-full animate-bounce" style={{ animationDelay: '0ms' }}></div>
                          <div className="w-2 h-2 bg-blue-400 rounded-full animate-bounce" style={{ animationDelay: '150ms' }}></div>
                          <div className="w-2 h-2 bg-blue-400 rounded-full animate-bounce" style={{ animationDelay: '300ms' }}></div>
                        </div>
                      </div>
                    </div>
                  )}
                  {proposal.aiEvaluation && (
                    <div className="mt-3 bg-gray-50 border border-gray-200 rounded-lg p-4">
                      <div className="flex items-center justify-between mb-3">
                        <h4 className="text-sm font-semibold text-gray-900">🤖 AI Evaluation</h4>
                        <div className="flex items-center gap-2">
                          <span className="text-sm font-bold text-gray-900">Score: {proposal.aiEvaluation.score}/100</span>
                          <span className={`px-2 py-0.5 rounded-full text-xs font-semibold border ${getRecommendationColor(proposal.aiEvaluation.recommendation)}`}>
                            {proposal.aiEvaluation.recommendation}
                          </span>
                        </div>
                      </div>
                      <p className="text-sm text-gray-700 mb-3">{proposal.aiEvaluation.summary}</p>

                      <div className="mb-3">
                        <h5 className="text-xs font-semibold text-gray-600 uppercase mb-1">CLIN Breakdown</h5>
                        <div className="grid grid-cols-1 sm:grid-cols-3 gap-2">
                          {proposal.aiEvaluation.clinBreakdown.map(clin => (
                            <div key={clin.clinNumber} className="bg-white border border-gray-200 rounded p-2">
                              <p className="text-xs font-medium text-gray-900">CLIN {clin.clinNumber}</p>
                              <p className="text-xs text-gray-600">{clin.description}</p>
                              <p className="text-xs font-semibold text-gray-900">${clin.ceiling.toLocaleString()}</p>
                            </div>
                          ))}
                        </div>
                      </div>

                      <div>
                        <h5 className="text-xs font-semibold text-gray-600 uppercase mb-1">BOE Allocation</h5>
                        <p className="text-xs text-gray-700 bg-white border border-gray-200 rounded p-2">{proposal.aiEvaluation.boeAllocation}</p>
                      </div>
                    </div>
                  )}

                  {proposal.status === 'SUBMITTED' && (
                    <ProposalActions
                      proposal={proposal}
                      isEvaluating={evaluatingProposals.has(proposal.id)}
                      onApprove={onApprove}
                      onReject={onReject}
                    />
                  )}
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

// --- Solicitation Detail Panel ---

function SolicitationDetail({
  solicitation,
  proposalCount,
  onClose,
  onSubmitProposal,
  onPublish,
  onReviewProposals,
  isGov,
  onDownloadRFP,
}: {
  solicitation: Solicitation
  proposalCount: number
  onClose: () => void
  onSubmitProposal: () => void
  onPublish: () => void
  onReviewProposals: () => void
  isGov: boolean
  onDownloadRFP: (filename: string) => void
}) {
  const daysRemaining = getDaysRemaining(solicitation.closeDate)

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
      <div className="bg-white rounded-xl shadow-xl w-full max-w-3xl max-h-[90vh] overflow-y-auto">
        <div className="px-6 py-4 border-b border-gray-200">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-semibold border ${getStatusBadgeClasses(solicitation.status as SolicitationStatus)}`}>
                {solicitation.status}
              </span>
              <span className="text-sm font-mono text-gray-500">{solicitation.solicitationNumber}</span>
            </div>
            <button onClick={onClose} className="text-gray-400 hover:text-gray-600">
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
          <h2 className="text-xl font-bold text-gray-900 mt-2">{solicitation.title}</h2>
        </div>

        <div className="p-6 space-y-6">
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div className="space-y-3">
              <div>
                <p className="text-xs font-medium text-gray-500 uppercase">Type</p>
                <p className="text-sm text-gray-900">{solicitation.type}</p>
              </div>
              <div>
                <p className="text-xs font-medium text-gray-500 uppercase">Agency</p>
                <p className="text-sm text-gray-900">{solicitation.agency}</p>
              </div>
              <div>
                <p className="text-xs font-medium text-gray-500 uppercase">NAICS Code</p>
                <p className="text-sm font-mono text-gray-900">{solicitation.naicsCode}</p>
              </div>
            </div>
            <div className="space-y-3">
              <div>
                <p className="text-xs font-medium text-gray-500 uppercase">Posted Date</p>
                <p className="text-sm text-gray-900">{formatDate(solicitation.postedDate)}</p>
              </div>
              <div>
                <p className="text-xs font-medium text-gray-500 uppercase">Close Date</p>
                <p className="text-sm text-gray-900">
                  {formatDate(solicitation.closeDate)}
                  {solicitation.status === 'OPEN' && daysRemaining > 0 && (
                    <span className={`ml-2 text-xs font-medium ${daysRemaining <= 7 ? 'text-red-600' : daysRemaining <= 14 ? 'text-amber-600' : 'text-green-600'}`}>
                      ({daysRemaining} days remaining)
                    </span>
                  )}
                </p>
              </div>
              <div>
                <p className="text-xs font-medium text-gray-500 uppercase">Estimated Value</p>
                <p className="text-sm font-medium text-gray-900">{solicitation.estimatedValue}</p>
              </div>
            </div>
          </div>

          {solicitation.awardedTo && (
            <div className="bg-purple-50 border border-purple-200 rounded-lg p-4">
              <p className="text-xs font-medium text-purple-600 uppercase mb-1">Award Recipient</p>
              <p className="text-sm font-semibold text-purple-900">{solicitation.awardedTo}</p>
            </div>
          )}

          <div>
            <h3 className="text-sm font-semibold text-gray-900 mb-2">Description / Statement of Work</h3>
            <p className="text-sm text-gray-700 leading-relaxed">{solicitation.description}</p>
          </div>

          {solicitation.evaluationCriteria && (
            <div>
              <h3 className="text-sm font-semibold text-gray-900 mb-2">Evaluation Criteria</h3>
              <p className="text-sm text-gray-700">{solicitation.evaluationCriteria}</p>
            </div>
          )}

          {solicitation.attachmentName && (
            <div className="bg-blue-50 border border-blue-200 rounded-lg p-3">
              <p className="text-sm text-blue-800">
                📎 Attached Document: <span className="font-semibold">{solicitation.attachmentName}</span>
              </p>
            </div>
          )}

          {proposalCount > 0 && (
            <div className="bg-blue-50 border border-blue-200 rounded-lg p-3">
              <p className="text-sm text-blue-800">
                <span className="font-semibold">{proposalCount}</span> {proposalCount === 1 ? 'Proposal' : 'Proposals'} Received
              </p>
            </div>
          )}

          {/* Actions based on role and status */}
          <div className="pt-2 flex gap-3 flex-wrap">
            {/* Download RFP for vendors */}
            {!isGov && solicitation.attachmentName && (
              <button
                onClick={() => onDownloadRFP(solicitation.attachmentName!)}
                className="px-6 py-3 text-sm font-medium text-blue-700 bg-blue-50 border border-blue-200 rounded-lg hover:bg-blue-100 transition-colors"
              >
                📥 Download RFP ({solicitation.attachmentName})
              </button>
            )}

            {/* GOV: Publish draft */}
            {isGov && solicitation.status === 'DRAFT' && (
              <button
                onClick={onPublish}
                className="px-6 py-3 text-sm font-medium text-white bg-blue-600 rounded-lg hover:bg-blue-700 focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 transition-colors"
              >
                Publish Solicitation
              </button>
            )}

            {/* GOV: Review proposals */}
            {isGov && proposalCount > 0 && solicitation.status !== 'AWARDED' && (
              <button
                onClick={onReviewProposals}
                className="px-6 py-3 text-sm font-medium text-white bg-purple-600 rounded-lg hover:bg-purple-700 focus:ring-2 focus:ring-purple-500 focus:ring-offset-2 transition-colors"
              >
                Review Proposals ({proposalCount})
              </button>
            )}

            {/* VENDOR: Submit proposal on OPEN */}
            {!isGov && solicitation.status === 'OPEN' && (
              <button
                onClick={onSubmitProposal}
                className="px-6 py-3 text-sm font-medium text-white bg-green-600 rounded-lg hover:bg-green-700 focus:ring-2 focus:ring-green-500 focus:ring-offset-2 transition-colors"
              >
                Submit Proposal
              </button>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}

// --- Solicitation Card ---

function SolicitationCard({
  solicitation,
  proposalCount,
  onClick,
  isVendor,
  vendorCompany,
  onDownloadRFP,
}: {
  solicitation: Solicitation
  proposalCount: number
  onClick: () => void
  isVendor: boolean
  vendorCompany: string
  onDownloadRFP?: (filename: string) => void
}) {
  const daysRemaining = getDaysRemaining(solicitation.closeDate)
  const isAwardedToOther = solicitation.status === 'AWARDED' && solicitation.awardedTo !== vendorCompany

  return (
    <div className="w-full text-left bg-white rounded-lg shadow-sm border border-gray-200 p-4 sm:p-5 hover:border-blue-300 hover:shadow-md transition-all">
      <button
        onClick={onClick}
        className="w-full text-left"
      >
        <div className="flex items-start justify-between gap-3">
          <div className="min-w-0 flex-1">
            <div className="flex items-center gap-2 flex-wrap mb-1">
              <span className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-semibold border ${getStatusBadgeClasses(solicitation.status as SolicitationStatus)}`}>
                {solicitation.status}
              </span>
              {isVendor && isAwardedToOther && (
                <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-semibold border bg-red-100 text-red-800 border-red-300">
                  Not Awarded
                </span>
              )}
              {isVendor && solicitation.status === 'AWARDED' && solicitation.awardedTo === vendorCompany && (
                <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-semibold border bg-green-100 text-green-800 border-green-300">
                  ✓ Awarded to You
                </span>
              )}
              <span className="text-xs font-mono text-gray-500">{solicitation.solicitationNumber}</span>
              <span className="text-xs text-gray-400 bg-gray-100 px-2 py-0.5 rounded">{solicitation.type}</span>
            </div>
            <h3 className="text-sm sm:text-base font-semibold text-gray-900 mt-1">{solicitation.title}</h3>
            <p className="text-xs sm:text-sm text-gray-500 mt-1">{solicitation.agency}</p>
          </div>

          <div className="text-right flex-shrink-0">
            {solicitation.estimatedValue && (
              <p className="text-xs sm:text-sm font-medium text-gray-900">{solicitation.estimatedValue}</p>
            )}
            {solicitation.status === 'OPEN' && daysRemaining > 0 && (
              <div className={`mt-1 text-xs font-medium ${daysRemaining <= 7 ? 'text-red-600' : daysRemaining <= 14 ? 'text-amber-600' : 'text-green-600'}`}>
                {daysRemaining} days left
              </div>
            )}
          </div>
        </div>

        <div className="flex items-center gap-4 mt-3 pt-3 border-t border-gray-100 text-xs text-gray-500">
          <span>Posted: {formatDate(solicitation.postedDate)}</span>
          <span>Closes: {formatDate(solicitation.closeDate)}</span>
          {proposalCount > 0 && (
            <span className="text-blue-600 font-medium">
              {proposalCount} {proposalCount === 1 ? 'Proposal' : 'Proposals'}
            </span>
          )}
          {solicitation.awardedTo && (
            <span className="text-purple-600 font-medium">
              Awarded: {solicitation.awardedTo}
            </span>
          )}
          {solicitation.attachmentName && (
            <span className="text-blue-600">📎 RFP attached</span>
          )}
        </div>
      </button>

      {/* Download RFP button for vendors on OPEN solicitations */}
      {isVendor && solicitation.status === 'OPEN' && solicitation.attachmentName && onDownloadRFP && (
        <div className="mt-2 pt-2 border-t border-gray-100">
          <button
            onClick={(e) => { e.stopPropagation(); onDownloadRFP(solicitation.attachmentName!) }}
            className="inline-flex items-center gap-1 px-3 py-1.5 text-xs font-medium text-blue-700 bg-blue-50 border border-blue-200 rounded-lg hover:bg-blue-100 transition-colors"
          >
            📥 Download RFP
          </button>
        </div>
      )}

      {solicitation.status === 'OPEN' && daysRemaining > 0 && (
        <div className="mt-3">
          <div className="flex justify-between text-xs text-gray-400 mb-1">
            <span>Posted</span>
            <span>Closes</span>
          </div>
          <div className="w-full bg-gray-100 rounded-full h-1.5">
            <div
              className={`rounded-full h-1.5 transition-all ${daysRemaining <= 7 ? 'bg-red-400' : daysRemaining <= 14 ? 'bg-amber-400' : 'bg-green-400'}`}
              style={{
                width: `${Math.max(5, Math.min(95, 100 - (daysRemaining / Math.max(1, getDaysRemaining(solicitation.postedDate) * -1 + daysRemaining)) * 100))}%`,
              }}
            />
          </div>
        </div>
      )}
    </div>
  )
}

// --- Toast Notification ---

function Toast({ message, onClose }: { message: string; onClose: () => void }) {
  return (
    <div className="fixed top-4 right-4 z-[60] animate-in slide-in-from-top">
      <div className="bg-green-600 text-white px-4 py-3 rounded-lg shadow-lg flex items-center gap-3">
        <svg className="w-5 h-5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
        </svg>
        <span className="text-sm font-medium">{message}</span>
        <button onClick={onClose} className="ml-2 text-green-200 hover:text-white">
          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>
    </div>
  )
}

// --- Main Solicitations Component ---

function Solicitations() {
  const { state, createSolicitation, publishSolicitation, submitProposal, approveProposal, rejectProposal, evaluatingProposals } = useAppContext()
  const isGov = state.currentRole === 'GOV'

  const [activeTab, setActiveTab] = useState('ALL')
  const [searchQuery, setSearchQuery] = useState('')
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [selectedSolicitation, setSelectedSolicitation] = useState<Solicitation | null>(null)
  const [showProposalModal, setShowProposalModal] = useState(false)
  const [showReviewPanel, setShowReviewPanel] = useState(false)
  const [toast, setToast] = useState<string | null>(null)
  const [proposalFile, setProposalFile] = useState<File | undefined>(undefined)

  // VENDOR sees: OPEN (all) + AWARDED if awarded to their company + AWARDED to others (show "Not Awarded" badge)
  const visibleSolicitations = useMemo(() => {
    if (isGov) return state.solicitations
    return state.solicitations.filter(s =>
      s.status === 'OPEN' || s.status === 'AWARDED'
    )
  }, [state.solicitations, isGov])

  // Filter solicitations
  const filteredSolicitations = useMemo(() => {
    let filtered = visibleSolicitations
    if (activeTab !== 'ALL') {
      filtered = filtered.filter((s) => s.status === activeTab)
    }
    if (searchQuery.trim()) {
      const query = searchQuery.toLowerCase()
      filtered = filtered.filter(
        (s) =>
          s.title.toLowerCase().includes(query) ||
          s.solicitationNumber.toLowerCase().includes(query) ||
          s.agency.toLowerCase().includes(query)
      )
    }
    return filtered
  }, [visibleSolicitations, activeTab, searchQuery])

  // Count by status
  const counts = useMemo(() => {
    const c: Record<string, number> = { ALL: visibleSolicitations.length }
    for (const s of visibleSolicitations) {
      c[s.status] = (c[s.status] || 0) + 1
    }
    return c
  }, [visibleSolicitations])

  // Get proposal count for a solicitation
  function getProposalCount(solId: string): number {
    return state.proposals.filter(p => p.solicitationId === solId).length
  }

  // Show toast with auto-dismiss
  function showToast(message: string) {
    setToast(message)
    setTimeout(() => setToast(null), 4000)
  }

  // Handle create solicitation
  function handleCreateSolicitation(form: CreateSolicitationForm) {
    createSolicitation({
      title: form.title,
      type: form.type,
      agency: 'AFRL / Air Force Research Laboratory',
      description: form.description,
      naicsCode: form.naicsCode,
      estimatedValue: form.estimatedValue,
      closeDate: form.closeDate,
      evaluationCriteria: form.evaluationCriteria,
      attachmentName: form.attachmentName || undefined,
    })
    setShowCreateModal(false)
    showToast(`Solicitation "${form.title}" created as DRAFT`)
  }

  // Handle proposal submission
  function handleProposalSubmit(form: ProposalForm) {
    if (selectedSolicitation) {
      submitProposal(selectedSolicitation.id, {
        companyName: form.companyName,
        technicalApproach: form.technicalApproach,
        priceProposal: parseFloat(form.priceProposal.replace(/,/g, '')) || 0,
        pastPerformance: form.pastPerformance,
        keyPersonnel: form.keyPersonnel,
        attachmentName: form.attachmentName || undefined,
      }, proposalFile)
      setShowProposalModal(false)
      setProposalFile(undefined)
      showToast('Proposal submitted! AI evaluation in progress...')
    }
  }

  // Handle publish
  function handlePublish() {
    if (selectedSolicitation) {
      publishSolicitation(selectedSolicitation.id)
      setSelectedSolicitation(null)
      showToast('Solicitation published and now OPEN for proposals')
    }
  }

  // Handle approve proposal
  function handleApproveProposal(proposalId: string) {
    approveProposal(proposalId)
    setShowReviewPanel(false)
    setSelectedSolicitation(null)
    showToast('Proposal approved! Contract created and solicitation awarded.')
  }

  // Handle reject proposal
  function handleRejectProposal(proposalId: string) {
    rejectProposal(proposalId)
    showToast('Proposal rejected.')
  }

  // Handle download RFP (simulated)
  function handleDownloadRFP(filename: string) {
    showToast(`Downloading ${filename}...`)
  }

  return (
    <div className="p-4 sm:p-6 lg:p-8 max-w-7xl mx-auto">
      {/* Page Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Solicitations</h1>
          <p className="mt-1 text-sm text-gray-500">
            {isGov
              ? 'Manage federal procurement solicitations and review proposals'
              : 'Browse open opportunities and submit proposals'}
          </p>
        </div>
        {isGov && (
          <button
            onClick={() => setShowCreateModal(true)}
            className="inline-flex items-center gap-2 px-4 py-2.5 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 transition-colors focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
          >
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
            </svg>
            Create Solicitation
          </button>
        )}
      </div>

      {/* Filters */}
      <div className="flex flex-col sm:flex-row gap-4 mb-6">
        <StatusTabs active={activeTab} onChange={setActiveTab} counts={counts} />
        <div className="flex-1 sm:max-w-xs">
          <div className="relative">
            <svg className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
            </svg>
            <input
              type="text"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="Search solicitations..."
              className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
            />
          </div>
        </div>
      </div>

      {/* Solicitation List */}
      <div className="space-y-3">
        {filteredSolicitations.length === 0 ? (
          <div className="text-center py-12 bg-white rounded-lg border border-gray-200">
            <svg className="mx-auto h-12 w-12 text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
            </svg>
            <p className="mt-3 text-sm text-gray-500">No solicitations found</p>
            <p className="text-xs text-gray-400 mt-1">Try adjusting your filters or search query</p>
          </div>
        ) : (
          filteredSolicitations.map((solicitation) => (
            <SolicitationCard
              key={solicitation.id}
              solicitation={solicitation}
              proposalCount={getProposalCount(solicitation.id)}
              onClick={() => setSelectedSolicitation(solicitation)}
              isVendor={!isGov}
              vendorCompany={state.vendorCompany}
              onDownloadRFP={handleDownloadRFP}
            />
          ))
        )}
      </div>

      {/* Modals */}
      {showCreateModal && (
        <CreateSolicitationModal
          onClose={() => setShowCreateModal(false)}
          onSubmit={handleCreateSolicitation}
        />
      )}

      {selectedSolicitation && !showProposalModal && !showReviewPanel && (
        <SolicitationDetail
          solicitation={selectedSolicitation}
          proposalCount={getProposalCount(selectedSolicitation.id)}
          onClose={() => setSelectedSolicitation(null)}
          onSubmitProposal={() => setShowProposalModal(true)}
          onPublish={handlePublish}
          onReviewProposals={() => setShowReviewPanel(true)}
          isGov={isGov}
          onDownloadRFP={handleDownloadRFP}
        />
      )}

      {showProposalModal && selectedSolicitation && (
        <ProposalModal
          solicitation={selectedSolicitation}
          onClose={() => setShowProposalModal(false)}
          onSubmit={handleProposalSubmit}
          vendorCompany={state.vendorCompany}
          onFileSelected={(file) => setProposalFile(file)}
        />
      )}

      {showReviewPanel && selectedSolicitation && (
        <ProposalsReviewPanel
          solicitation={selectedSolicitation}
          proposals={state.proposals}
          onApprove={handleApproveProposal}
          onReject={handleRejectProposal}
          onClose={() => setShowReviewPanel(false)}
          evaluatingProposals={evaluatingProposals}
        />
      )}

      {/* Toast */}
      {toast && <Toast message={toast} onClose={() => setToast(null)} />}
    </div>
  )
}

export default Solicitations
