import { useState, useMemo } from 'react'

// --- Types ---

type SolicitationStatus = 'DRAFT' | 'OPEN' | 'CLOSED' | 'AWARDED' | 'CANCELLED'
type SolicitationType = 'SBIR Phase I' | 'SBIR Phase II' | 'Full & Open Competition' | 'Sole Source'

interface Solicitation {
  id: string
  solicitationNumber: string
  title: string
  type: SolicitationType
  agency: string
  postedDate: string
  closeDate: string
  status: SolicitationStatus
  description: string
  naicsCode: string
  estimatedValue: string
  evaluationCriteria: string
  awardee?: string
  responsesCount: number
}

interface ProposalForm {
  companyName: string
  technicalApproach: string
  priceProposal: string
  pastPerformance: string
  keyPersonnel: string
}

interface CreateSolicitationForm {
  title: string
  type: SolicitationType
  description: string
  naicsCode: string
  estimatedValue: string
  closeDate: string
  evaluationCriteria: string
}

// --- Initial Mock Data ---

const initialSolicitations: Solicitation[] = [
  {
    id: '1',
    solicitationNumber: 'FA8750-25-SBIR-0042',
    title: 'Agentic AI for Federal Payment Modernization',
    type: 'SBIR Phase II',
    agency: 'AFRL / Air Force Research Laboratory',
    postedDate: '2024-12-15',
    closeDate: '2025-01-15',
    status: 'AWARDED',
    description:
      'The Air Force Research Laboratory seeks innovative solutions leveraging agentic AI architectures to modernize federal payment processing systems. The solution must demonstrate autonomous document extraction, intelligent validation, and adaptive disbursement routing capabilities while maintaining FedRAMP High compliance.',
    naicsCode: '541512',
    estimatedValue: '$750K - $1.5M',
    evaluationCriteria:
      'Technical Approach (40%), Past Performance (25%), Price (20%), Key Personnel (15%)',
    awardee: 'Quantum Federal Systems LLC',
    responsesCount: 7,
  },
  {
    id: '2',
    solicitationNumber: 'FA8750-25-RFP-0098',
    title: 'Next-Generation Document Intelligence Platform',
    type: 'Full & Open Competition',
    agency: 'AFRL / Air Force Research Laboratory',
    postedDate: '2025-06-01',
    closeDate: '2025-07-15',
    status: 'OPEN',
    description:
      'AFRL requires a cloud-native document intelligence platform capable of processing unstructured federal financial documents at scale. The platform must support OCR, NLP-based entity extraction, automated classification, and integration with existing Treasury systems. Must achieve 99.5% accuracy on standard government form types (SF-1034, SF-1035, DD-250).',
    naicsCode: '541511',
    estimatedValue: '$2.5M - $5M',
    evaluationCriteria:
      'Technical Capability (35%), Management Approach (20%), Past Performance (25%), Price (20%)',
    responsesCount: 0,
  },
  {
    id: '3',
    solicitationNumber: 'FA8750-25-SBIR-0103',
    title: 'Cloud-Native Financial Reporting System',
    type: 'SBIR Phase I',
    agency: 'AFRL / Air Force Research Laboratory',
    postedDate: '2025-06-10',
    closeDate: '2025-08-01',
    status: 'DRAFT',
    description:
      'Seeking innovative small businesses to develop a cloud-native financial reporting system that provides real-time visibility into federal payment pipelines, obligation tracking, and disbursement analytics with role-based access controls.',
    naicsCode: '541519',
    estimatedValue: '$150K - $250K',
    evaluationCriteria:
      'Technical Merit (45%), Feasibility (25%), Commercial Potential (15%), Price (15%)',
    responsesCount: 0,
  },
]

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
          {/* Title */}
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

          {/* Type */}
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

          {/* Description / SOW */}
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

          {/* NAICS Code and Estimated Value */}
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

          {/* Close Date */}
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

          {/* Evaluation Criteria */}
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

          {/* Actions */}
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
}: {
  solicitation: Solicitation
  onClose: () => void
  onSubmit: () => void
}) {
  const [form, setForm] = useState<ProposalForm>({
    companyName: '',
    technicalApproach: '',
    priceProposal: '',
    pastPerformance: '',
    keyPersonnel: '',
  })
  const [errors, setErrors] = useState<Record<string, string>>({})

  function validate(): boolean {
    const newErrors: Record<string, string> = {}
    if (!form.companyName.trim()) newErrors.companyName = 'Company name is required'
    if (!form.technicalApproach.trim()) newErrors.technicalApproach = 'Technical approach is required'
    if (!form.priceProposal.trim()) newErrors.priceProposal = 'Price proposal is required'
    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (validate()) {
      onSubmit()
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
              Company Name <span className="text-red-500">*</span>
            </label>
            <input
              type="text"
              value={form.companyName}
              onChange={(e) => setForm({ ...form, companyName: e.target.value })}
              className={`w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500 ${
                errors.companyName ? 'border-red-300' : 'border-gray-300'
              }`}
              placeholder="Your company legal name"
            />
            {errors.companyName && <p className="mt-1 text-xs text-red-600">{errors.companyName}</p>}
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Technical Approach <span className="text-red-500">*</span>
            </label>
            <textarea
              value={form.technicalApproach}
              onChange={(e) => setForm({ ...form, technicalApproach: e.target.value })}
              rows={5}
              className={`w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500 ${
                errors.technicalApproach ? 'border-red-300' : 'border-gray-300'
              }`}
              placeholder="Describe your technical approach to meeting the requirements..."
            />
            {errors.technicalApproach && <p className="mt-1 text-xs text-red-600">{errors.technicalApproach}</p>}
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Price Proposal ($) <span className="text-red-500">*</span>
            </label>
            <input
              type="text"
              value={form.priceProposal}
              onChange={(e) => setForm({ ...form, priceProposal: e.target.value })}
              className={`w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500 ${
                errors.priceProposal ? 'border-red-300' : 'border-gray-300'
              }`}
              placeholder="e.g., 3,500,000"
            />
            {errors.priceProposal && <p className="mt-1 text-xs text-red-600">{errors.priceProposal}</p>}
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

// --- Solicitation Detail Panel ---

function SolicitationDetail({
  solicitation,
  onClose,
  onSubmitProposal,
}: {
  solicitation: Solicitation
  onClose: () => void
  onSubmitProposal: () => void
}) {
  const daysRemaining = getDaysRemaining(solicitation.closeDate)

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
      <div className="bg-white rounded-xl shadow-xl w-full max-w-3xl max-h-[90vh] overflow-y-auto">
        <div className="px-6 py-4 border-b border-gray-200">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-semibold border ${getStatusBadgeClasses(solicitation.status)}`}>
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
          {/* Key Info Grid */}
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

          {/* Awardee info for awarded solicitations */}
          {solicitation.awardee && (
            <div className="bg-purple-50 border border-purple-200 rounded-lg p-4">
              <p className="text-xs font-medium text-purple-600 uppercase mb-1">Award Recipient</p>
              <p className="text-sm font-semibold text-purple-900">{solicitation.awardee}</p>
            </div>
          )}

          {/* Description */}
          <div>
            <h3 className="text-sm font-semibold text-gray-900 mb-2">Description / Statement of Work</h3>
            <p className="text-sm text-gray-700 leading-relaxed">{solicitation.description}</p>
          </div>

          {/* Evaluation Criteria */}
          {solicitation.evaluationCriteria && (
            <div>
              <h3 className="text-sm font-semibold text-gray-900 mb-2">Evaluation Criteria</h3>
              <p className="text-sm text-gray-700">{solicitation.evaluationCriteria}</p>
            </div>
          )}

          {/* Responses */}
          {solicitation.responsesCount > 0 && (
            <div className="bg-blue-50 border border-blue-200 rounded-lg p-3">
              <p className="text-sm text-blue-800">
                <span className="font-semibold">{solicitation.responsesCount}</span> {solicitation.responsesCount === 1 ? 'Response' : 'Responses'} Received
              </p>
            </div>
          )}

          {/* Submit Proposal Button - only for OPEN solicitations */}
          {solicitation.status === 'OPEN' && (
            <div className="pt-2">
              <button
                onClick={onSubmitProposal}
                className="w-full sm:w-auto px-6 py-3 text-sm font-medium text-white bg-green-600 rounded-lg hover:bg-green-700 focus:ring-2 focus:ring-green-500 focus:ring-offset-2 transition-colors"
              >
                Submit Proposal
              </button>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

// --- Solicitation Card ---

function SolicitationCard({
  solicitation,
  onClick,
}: {
  solicitation: Solicitation
  onClick: () => void
}) {
  const daysRemaining = getDaysRemaining(solicitation.closeDate)

  return (
    <button
      onClick={onClick}
      className="w-full text-left bg-white rounded-lg shadow-sm border border-gray-200 p-4 sm:p-5 hover:border-blue-300 hover:shadow-md transition-all"
    >
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-2 flex-wrap mb-1">
            <span className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-semibold border ${getStatusBadgeClasses(solicitation.status)}`}>
              {solicitation.status}
            </span>
            <span className="text-xs font-mono text-gray-500">{solicitation.solicitationNumber}</span>
            <span className="text-xs text-gray-400 bg-gray-100 px-2 py-0.5 rounded">{solicitation.type}</span>
          </div>
          <h3 className="text-sm sm:text-base font-semibold text-gray-900 mt-1">{solicitation.title}</h3>
          <p className="text-xs sm:text-sm text-gray-500 mt-1">{solicitation.agency}</p>
        </div>

        {/* Right side info */}
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

      {/* Bottom meta */}
      <div className="flex items-center gap-4 mt-3 pt-3 border-t border-gray-100 text-xs text-gray-500">
        <span>Posted: {formatDate(solicitation.postedDate)}</span>
        <span>Closes: {formatDate(solicitation.closeDate)}</span>
        {solicitation.responsesCount > 0 && (
          <span className="text-blue-600 font-medium">
            {solicitation.responsesCount} {solicitation.responsesCount === 1 ? 'Response' : 'Responses'}
          </span>
        )}
        {solicitation.awardee && (
          <span className="text-purple-600 font-medium">
            Awarded: {solicitation.awardee}
          </span>
        )}
      </div>

      {/* Timeline bar for OPEN */}
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
    </button>
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
  const [solicitations, setSolicitations] = useState<Solicitation[]>(initialSolicitations)
  const [activeTab, setActiveTab] = useState('ALL')
  const [searchQuery, setSearchQuery] = useState('')
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [selectedSolicitation, setSelectedSolicitation] = useState<Solicitation | null>(null)
  const [showProposalModal, setShowProposalModal] = useState(false)
  const [toast, setToast] = useState<string | null>(null)

  // Filter solicitations
  const filteredSolicitations = useMemo(() => {
    let filtered = solicitations
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
  }, [solicitations, activeTab, searchQuery])

  // Count by status
  const counts = useMemo(() => {
    const c: Record<string, number> = { ALL: solicitations.length }
    for (const s of solicitations) {
      c[s.status] = (c[s.status] || 0) + 1
    }
    return c
  }, [solicitations])

  // Show toast with auto-dismiss
  function showToast(message: string) {
    setToast(message)
    setTimeout(() => setToast(null), 4000)
  }

  // Handle create solicitation
  function handleCreateSolicitation(form: CreateSolicitationForm) {
    const newId = String(Date.now())
    const solNum = `FA8750-25-SBIR-${String(Math.floor(Math.random() * 9000) + 1000)}`
    const newSolicitation: Solicitation = {
      id: newId,
      solicitationNumber: solNum,
      title: form.title,
      type: form.type,
      agency: 'AFRL / Air Force Research Laboratory',
      postedDate: new Date().toISOString().split('T')[0],
      closeDate: form.closeDate,
      status: 'DRAFT',
      description: form.description,
      naicsCode: form.naicsCode,
      estimatedValue: form.estimatedValue,
      evaluationCriteria: form.evaluationCriteria,
      responsesCount: 0,
    }
    setSolicitations((prev) => [newSolicitation, ...prev])
    setShowCreateModal(false)
    showToast(`Solicitation "${form.title}" created as DRAFT`)
  }

  // Handle proposal submission
  function handleProposalSubmit() {
    if (selectedSolicitation) {
      setSolicitations((prev) =>
        prev.map((s) =>
          s.id === selectedSolicitation.id
            ? { ...s, responsesCount: s.responsesCount + 1 }
            : s
        )
      )
      setSelectedSolicitation({
        ...selectedSolicitation,
        responsesCount: selectedSolicitation.responsesCount + 1,
      })
      setShowProposalModal(false)
      showToast('Proposal submitted successfully! You will receive a confirmation email.')
    }
  }

  return (
    <div className="p-4 sm:p-6 lg:p-8 max-w-7xl mx-auto">
      {/* Page Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Solicitations</h1>
          <p className="mt-1 text-sm text-gray-500">
            Browse open opportunities and manage federal procurement solicitations
          </p>
        </div>
        <button
          onClick={() => setShowCreateModal(true)}
          className="inline-flex items-center gap-2 px-4 py-2.5 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 transition-colors focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
        >
          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
          </svg>
          Create Solicitation
        </button>
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
              onClick={() => setSelectedSolicitation(solicitation)}
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

      {selectedSolicitation && !showProposalModal && (
        <SolicitationDetail
          solicitation={selectedSolicitation}
          onClose={() => setSelectedSolicitation(null)}
          onSubmitProposal={() => setShowProposalModal(true)}
        />
      )}

      {showProposalModal && selectedSolicitation && (
        <ProposalModal
          solicitation={selectedSolicitation}
          onClose={() => setShowProposalModal(false)}
          onSubmit={handleProposalSubmit}
        />
      )}

      {/* Toast */}
      {toast && <Toast message={toast} onClose={() => setToast(null)} />}
    </div>
  )
}

export default Solicitations
