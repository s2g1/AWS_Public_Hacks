import { useState } from 'react'

// --- Types ---

type REAStatus =
  | 'SUBMITTED'
  | 'APPROVED'
  | 'PARTIALLY_APPROVED'
  | 'DENIED'
  | 'ADDITIONAL_INFO_REQUESTED'

interface AuditEntry {
  timestamp: string
  actor: string
  action: string
  details: string
}

interface REARecord {
  id: string
  contractNumber: string
  requestedAmount: number
  approvedAmount?: number
  affectedCLINs: string[]
  justification: string
  status: REAStatus
  submittedBy: string
  submittedDate: string
  resolvedDate?: string
  auditTrail: AuditEntry[]
}

// --- Sample Data ---

const sampleREAs: REARecord[] = [
  {
    id: 'REA-2024-001',
    contractNumber: 'FA8750-23-C-0042',
    requestedAmount: 185_000,
    approvedAmount: 185_000,
    affectedCLINs: ['0001', '0003'],
    justification: 'Scope change due to updated security requirements mandated by NIST SP 800-171 Rev 3.',
    status: 'APPROVED',
    submittedBy: 'J. Martinez (NovaTech Solutions)',
    submittedDate: '2024-08-15',
    resolvedDate: '2024-09-02',
    auditTrail: [
      { timestamp: '2024-08-15T09:30:00Z', actor: 'J. Martinez', action: 'Submitted REA', details: 'Initial submission with supporting documentation.' },
      { timestamp: '2024-08-16T14:00:00Z', actor: 'System', action: 'Notification sent', details: 'CO M. Thompson notified of new REA submission.' },
      { timestamp: '2024-08-22T10:15:00Z', actor: 'M. Thompson (CO)', action: 'Review started', details: 'Technical evaluation initiated.' },
      { timestamp: '2024-09-02T16:45:00Z', actor: 'M. Thompson (CO)', action: 'Approved', details: 'Full amount approved. Contract mod #003 created.' },
    ],
  },
  {
    id: 'REA-2024-002',
    contractNumber: 'FA8750-23-C-0042',
    requestedAmount: 320_000,
    approvedAmount: 210_000,
    affectedCLINs: ['0001'],
    justification: 'Additional labor hours required for Phase II data migration complexity beyond original SOW.',
    status: 'PARTIALLY_APPROVED',
    submittedBy: 'J. Martinez (NovaTech Solutions)',
    submittedDate: '2024-10-03',
    resolvedDate: '2024-11-15',
    auditTrail: [
      { timestamp: '2024-10-03T08:00:00Z', actor: 'J. Martinez', action: 'Submitted REA', details: 'Requested $320,000 for additional labor.' },
      { timestamp: '2024-10-04T09:00:00Z', actor: 'System', action: 'Notification sent', details: 'CO M. Thompson notified.' },
      { timestamp: '2024-10-10T11:30:00Z', actor: 'M. Thompson (CO)', action: 'Info requested', details: 'Requested detailed breakdown of labor categories.' },
      { timestamp: '2024-10-18T14:00:00Z', actor: 'J. Martinez', action: 'Info provided', details: 'Submitted labor category breakdown and historical actuals.' },
      { timestamp: '2024-11-15T10:00:00Z', actor: 'M. Thompson (CO)', action: 'Partially approved', details: 'Approved $210,000. Senior engineer hours reduced per negotiation.' },
    ],
  },
  {
    id: 'REA-2024-003',
    contractNumber: 'FA8750-23-C-0042',
    requestedAmount: 95_000,
    affectedCLINs: ['0002'],
    justification: 'Cloud infrastructure cost increase due to AWS pricing changes effective Jan 2025.',
    status: 'DENIED',
    submittedBy: 'J. Martinez (NovaTech Solutions)',
    submittedDate: '2024-11-20',
    resolvedDate: '2024-12-05',
    auditTrail: [
      { timestamp: '2024-11-20T10:00:00Z', actor: 'J. Martinez', action: 'Submitted REA', details: 'Requested $95,000 for cloud cost increases.' },
      { timestamp: '2024-11-21T08:30:00Z', actor: 'System', action: 'Notification sent', details: 'CO M. Thompson notified.' },
      { timestamp: '2024-12-05T15:30:00Z', actor: 'M. Thompson (CO)', action: 'Denied', details: 'FFP CLIN - price risk borne by contractor per contract terms.' },
    ],
  },
  {
    id: 'REA-2024-004',
    contractNumber: 'FA8750-23-C-0042',
    requestedAmount: 150_000,
    affectedCLINs: ['0001', '0004'],
    justification: 'Integration testing scope expansion required by updated DoDI 5000.87 guidance.',
    status: 'ADDITIONAL_INFO_REQUESTED',
    submittedBy: 'J. Martinez (NovaTech Solutions)',
    submittedDate: '2025-01-10',
    auditTrail: [
      { timestamp: '2025-01-10T09:00:00Z', actor: 'J. Martinez', action: 'Submitted REA', details: 'Requested $150,000 for expanded integration testing.' },
      { timestamp: '2025-01-11T10:00:00Z', actor: 'System', action: 'Notification sent', details: 'CO M. Thompson notified.' },
      { timestamp: '2025-01-18T13:00:00Z', actor: 'M. Thompson (CO)', action: 'Additional info requested', details: 'Please provide DoDI reference and mapping to specific test events.' },
    ],
  },
  {
    id: 'REA-2025-001',
    contractNumber: 'FA8750-23-C-0042',
    requestedAmount: 275_000,
    affectedCLINs: ['0001', '0002', '0003'],
    justification: 'Supply chain disruptions requiring alternate vendor qualification and retest.',
    status: 'SUBMITTED',
    submittedBy: 'J. Martinez (NovaTech Solutions)',
    submittedDate: '2025-02-01',
    auditTrail: [
      { timestamp: '2025-02-01T08:30:00Z', actor: 'J. Martinez', action: 'Submitted REA', details: 'Initial submission for supply chain impact.' },
      { timestamp: '2025-02-01T08:31:00Z', actor: 'System', action: 'Notification sent', details: 'CO M. Thompson notified of new REA submission.' },
    ],
  },
]

// --- Utility Functions ---

function formatCurrency(amount: number): string {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(amount)
}

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  })
}

function formatTimestamp(ts: string): string {
  return new Date(ts).toLocaleString('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function getStatusBadge(status: REAStatus): { label: string; classes: string } {
  switch (status) {
    case 'SUBMITTED':
      return { label: 'Submitted', classes: 'bg-blue-100 text-blue-800 border-blue-200' }
    case 'APPROVED':
      return { label: 'Approved', classes: 'bg-green-100 text-green-800 border-green-200' }
    case 'PARTIALLY_APPROVED':
      return { label: 'Partially Approved', classes: 'bg-yellow-100 text-yellow-800 border-yellow-200' }
    case 'DENIED':
      return { label: 'Denied', classes: 'bg-red-100 text-red-800 border-red-200' }
    case 'ADDITIONAL_INFO_REQUESTED':
      return { label: 'Info Requested', classes: 'bg-purple-100 text-purple-800 border-purple-200' }
  }
}

// --- REA Submission Form Component ---

interface REAFormData {
  requestedAmount: string
  affectedCLINs: string[]
  justification: string
}

function REASubmissionForm({ onSubmit }: { onSubmit: (data: REAFormData) => void }) {
  const [formData, setFormData] = useState<REAFormData>({
    requestedAmount: '',
    affectedCLINs: [],
    justification: '',
  })
  const [errors, setErrors] = useState<Record<string, string>>({})

  const availableCLINs = ['0001', '0002', '0003', '0004']

  function toggleCLIN(clin: string) {
    setFormData((prev) => ({
      ...prev,
      affectedCLINs: prev.affectedCLINs.includes(clin)
        ? prev.affectedCLINs.filter((c) => c !== clin)
        : [...prev.affectedCLINs, clin],
    }))
  }

  function validate(): boolean {
    const newErrors: Record<string, string> = {}
    const amount = parseFloat(formData.requestedAmount)
    if (!formData.requestedAmount || isNaN(amount) || amount <= 0) {
      newErrors.requestedAmount = 'Requested amount must be a positive number.'
    }
    if (formData.affectedCLINs.length === 0) {
      newErrors.affectedCLINs = 'At least one affected CLIN must be selected.'
    }
    if (!formData.justification.trim()) {
      newErrors.justification = 'Justification is required.'
    }
    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (validate()) {
      onSubmit(formData)
      setFormData({ requestedAmount: '', affectedCLINs: [], justification: '' })
      setErrors({})
    }
  }

  return (
    <form onSubmit={handleSubmit} className="bg-white rounded-lg shadow-sm border border-gray-200 p-5">
      <h3 className="text-lg font-semibold text-gray-900 mb-4">Submit New REA</h3>

      {/* Requested Amount */}
      <div className="mb-4">
        <label htmlFor="rea-amount" className="block text-sm font-medium text-gray-700 mb-1">
          Requested Amount ($)
        </label>
        <input
          id="rea-amount"
          type="number"
          min="0"
          step="0.01"
          placeholder="e.g. 150000"
          value={formData.requestedAmount}
          onChange={(e) => setFormData((prev) => ({ ...prev, requestedAmount: e.target.value }))}
          className={`w-full border rounded-md px-3 py-2 text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500 ${
            errors.requestedAmount ? 'border-red-300 bg-red-50' : 'border-gray-300'
          }`}
        />
        {errors.requestedAmount && (
          <p className="mt-1 text-xs text-red-600">{errors.requestedAmount}</p>
        )}
      </div>

      {/* Affected CLINs */}
      <div className="mb-4">
        <fieldset>
          <legend className="block text-sm font-medium text-gray-700 mb-2">Affected CLINs</legend>
          <div className="flex flex-wrap gap-2">
            {availableCLINs.map((clin) => (
              <button
                key={clin}
                type="button"
                onClick={() => toggleCLIN(clin)}
                className={`px-3 py-1.5 text-sm rounded-md border transition-colors ${
                  formData.affectedCLINs.includes(clin)
                    ? 'bg-blue-600 text-white border-blue-600'
                    : 'bg-white text-gray-700 border-gray-300 hover:bg-gray-50'
                }`}
              >
                CLIN {clin}
              </button>
            ))}
          </div>
          {errors.affectedCLINs && (
            <p className="mt-1 text-xs text-red-600">{errors.affectedCLINs}</p>
          )}
        </fieldset>
      </div>

      {/* Justification */}
      <div className="mb-4">
        <label htmlFor="rea-justification" className="block text-sm font-medium text-gray-700 mb-1">
          Justification
        </label>
        <textarea
          id="rea-justification"
          rows={3}
          placeholder="Describe the scope change or cost increase requiring equitable adjustment..."
          value={formData.justification}
          onChange={(e) => setFormData((prev) => ({ ...prev, justification: e.target.value }))}
          className={`w-full border rounded-md px-3 py-2 text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500 ${
            errors.justification ? 'border-red-300 bg-red-50' : 'border-gray-300'
          }`}
        />
        {errors.justification && (
          <p className="mt-1 text-xs text-red-600">{errors.justification}</p>
        )}
      </div>

      <button
        type="submit"
        className="w-full sm:w-auto px-5 py-2 bg-blue-600 text-white text-sm font-medium rounded-md hover:bg-blue-700 transition-colors focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
      >
        Submit REA
      </button>
    </form>
  )
}

// --- Timeline Component ---

function AuditTimeline({ entries }: { entries: AuditEntry[] }) {
  return (
    <div className="flow-root">
      <ul className="-mb-4">
        {entries.map((entry, idx) => (
          <li key={idx} className="relative pb-4">
            {idx !== entries.length - 1 && (
              <span className="absolute left-3 top-5 -ml-px h-full w-0.5 bg-gray-200" aria-hidden="true" />
            )}
            <div className="relative flex items-start gap-3">
              <div className="flex-shrink-0">
                <span className="h-6 w-6 rounded-full bg-blue-100 border-2 border-blue-400 flex items-center justify-center">
                  <span className="h-2 w-2 rounded-full bg-blue-600" />
                </span>
              </div>
              <div className="min-w-0 flex-1">
                <div className="flex items-center gap-2 flex-wrap">
                  <span className="text-sm font-medium text-gray-900">{entry.action}</span>
                  <span className="text-xs text-gray-500">by {entry.actor}</span>
                </div>
                <p className="mt-0.5 text-xs text-gray-600">{entry.details}</p>
                <p className="mt-0.5 text-xs text-gray-400">{formatTimestamp(entry.timestamp)}</p>
              </div>
            </div>
          </li>
        ))}
      </ul>
    </div>
  )
}

// --- REA Card Component ---

function REACard({ rea, isExpanded, onToggle }: { rea: REARecord; isExpanded: boolean; onToggle: () => void }) {
  const badge = getStatusBadge(rea.status)

  return (
    <div className="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden">
      <button
        onClick={onToggle}
        className="w-full text-left p-4 sm:p-5 hover:bg-gray-50 transition-colors"
        aria-expanded={isExpanded}
      >
        <div className="flex items-start justify-between gap-3">
          <div className="min-w-0 flex-1">
            <div className="flex items-center gap-2 flex-wrap">
              <span className="text-sm font-semibold text-gray-900">{rea.id}</span>
              <span className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium border ${badge.classes}`}>
                {badge.label}
              </span>
            </div>
            <p className="mt-1 text-sm text-gray-600 line-clamp-1">{rea.justification}</p>
            <div className="mt-2 flex flex-wrap gap-3 text-xs text-gray-500">
              <span>Requested: <span className="font-medium text-gray-700">{formatCurrency(rea.requestedAmount)}</span></span>
              {rea.approvedAmount !== undefined && (
                <span>Approved: <span className="font-medium text-green-700">{formatCurrency(rea.approvedAmount)}</span></span>
              )}
              <span>CLINs: <span className="font-medium text-gray-700">{rea.affectedCLINs.join(', ')}</span></span>
            </div>
          </div>
          <span className="text-gray-400 transition-transform duration-200 flex-shrink-0 mt-1" style={{ transform: isExpanded ? 'rotate(90deg)' : 'rotate(0deg)' }}>
            ▶
          </span>
        </div>
      </button>

      {isExpanded && (
        <div className="border-t border-gray-200 bg-gray-50 p-4 sm:p-5">
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 mb-5">
            <div>
              <p className="text-xs font-medium text-gray-500 uppercase">Contract</p>
              <p className="text-sm text-gray-900">{rea.contractNumber}</p>
            </div>
            <div>
              <p className="text-xs font-medium text-gray-500 uppercase">Submitted By</p>
              <p className="text-sm text-gray-900">{rea.submittedBy}</p>
            </div>
            <div>
              <p className="text-xs font-medium text-gray-500 uppercase">Submitted Date</p>
              <p className="text-sm text-gray-900">{formatDate(rea.submittedDate)}</p>
            </div>
            <div>
              <p className="text-xs font-medium text-gray-500 uppercase">Resolved Date</p>
              <p className="text-sm text-gray-900">{rea.resolvedDate ? formatDate(rea.resolvedDate) : 'Pending'}</p>
            </div>
          </div>

          <div className="mb-4">
            <p className="text-xs font-medium text-gray-500 uppercase mb-1">Justification</p>
            <p className="text-sm text-gray-700 bg-white border border-gray-200 rounded p-3">{rea.justification}</p>
          </div>

          <div>
            <p className="text-xs font-medium text-gray-500 uppercase mb-3">Audit Trail / History</p>
            <AuditTimeline entries={rea.auditTrail} />
          </div>
        </div>
      )}
    </div>
  )
}

// --- Main Component ---

function Alerts() {
  const [expandedREAs, setExpandedREAs] = useState<Set<string>>(new Set())
  const [showForm, setShowForm] = useState(false)
  const [submissions, setSubmissions] = useState<REARecord[]>(sampleREAs)

  function toggleREA(id: string) {
    setExpandedREAs((prev) => {
      const next = new Set(prev)
      if (next.has(id)) {
        next.delete(id)
      } else {
        next.add(id)
      }
      return next
    })
  }

  function handleSubmit(data: REAFormData) {
    const newREA: REARecord = {
      id: `REA-2025-${String(submissions.length + 1).padStart(3, '0')}`,
      contractNumber: 'FA8750-23-C-0042',
      requestedAmount: parseFloat(data.requestedAmount),
      affectedCLINs: data.affectedCLINs,
      justification: data.justification,
      status: 'SUBMITTED',
      submittedBy: 'Current User (Contractor)',
      submittedDate: new Date().toISOString().split('T')[0],
      auditTrail: [
        {
          timestamp: new Date().toISOString(),
          actor: 'Current User',
          action: 'Submitted REA',
          details: 'Initial submission.',
        },
        {
          timestamp: new Date().toISOString(),
          actor: 'System',
          action: 'Notification sent',
          details: 'CO M. Thompson notified of new REA submission.',
        },
      ],
    }
    setSubmissions((prev) => [newREA, ...prev])
    setShowForm(false)
  }

  // Status summary counts
  const statusCounts = submissions.reduce(
    (acc, rea) => {
      acc[rea.status] = (acc[rea.status] || 0) + 1
      return acc
    },
    {} as Record<string, number>,
  )

  return (
    <div className="p-4 sm:p-6 lg:p-8 max-w-7xl mx-auto">
      {/* Header */}
      <div className="mb-6">
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
          <div>
            <h1 className="text-2xl font-bold text-gray-900">REA Management</h1>
            <p className="mt-1 text-sm text-gray-600">
              Requests for Equitable Adjustment — FA8750-23-C-0042
            </p>
          </div>
          <button
            onClick={() => setShowForm(!showForm)}
            className="inline-flex items-center px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-md hover:bg-blue-700 transition-colors focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
          >
            {showForm ? 'Cancel' : '+ New REA'}
          </button>
        </div>
      </div>

      {/* Status Summary */}
      <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-3 mb-6">
        {([
          ['SUBMITTED', 'Submitted', 'bg-blue-50 border-blue-200 text-blue-800'],
          ['APPROVED', 'Approved', 'bg-green-50 border-green-200 text-green-800'],
          ['PARTIALLY_APPROVED', 'Partial', 'bg-yellow-50 border-yellow-200 text-yellow-800'],
          ['DENIED', 'Denied', 'bg-red-50 border-red-200 text-red-800'],
          ['ADDITIONAL_INFO_REQUESTED', 'Info Req.', 'bg-purple-50 border-purple-200 text-purple-800'],
        ] as [REAStatus, string, string][]).map(([status, label, classes]) => (
          <div key={status} className={`rounded-lg border p-3 text-center ${classes}`}>
            <p className="text-2xl font-bold">{statusCounts[status] || 0}</p>
            <p className="text-xs font-medium mt-0.5">{label}</p>
          </div>
        ))}
      </div>

      {/* Submission Form */}
      {showForm && (
        <div className="mb-6">
          <REASubmissionForm onSubmit={handleSubmit} />
        </div>
      )}

      {/* REA List */}
      <div className="space-y-3">
        {submissions.map((rea) => (
          <REACard
            key={rea.id}
            rea={rea}
            isExpanded={expandedREAs.has(rea.id)}
            onToggle={() => toggleREA(rea.id)}
          />
        ))}
      </div>
    </div>
  )
}

export default Alerts
