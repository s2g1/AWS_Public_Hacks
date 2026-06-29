import { useState } from 'react'

type CorrespondenceStatus = 'DRAFT' | 'PENDING' | 'SENT'

interface CorrespondenceItem {
  id: string
  subject: string
  recipient: string
  type: string
  status: CorrespondenceStatus
  createdAt: string
  body: string
}

/** Demo data for correspondence items */
const SAMPLE_CORRESPONDENCE: CorrespondenceItem[] = [
  {
    id: 'corr-001',
    subject: 'Payment Approval Confirmation - INV-2024-0847',
    recipient: 'contractor@acmecorp.com',
    type: 'APPROVAL_CONFIRMATION',
    status: 'SENT',
    createdAt: '2024-01-15T10:30:00Z',
    body: 'Dear Acme Corp,\n\nWe are pleased to confirm that your payment of $45,000.00 for Invoice INV-2024-0847 under Contract FA8101-23-C-0042 has been approved.\n\nDisbursement is expected within 5 business days.\n\nRegards,\nContracting Office',
  },
  {
    id: 'corr-002',
    subject: 'Payment Rejection Notice - INV-2024-0912',
    recipient: 'billing@defensetech.com',
    type: 'REJECTION_NOTICE',
    status: 'PENDING',
    createdAt: '2024-01-16T14:15:00Z',
    body: 'Dear DefenseTech Inc,\n\nWe regret to inform you that your payment submission INV-2024-0912 for $128,500.00 has been rejected.\n\nReason: Missing CLIN reference and incomplete cost breakdown.\n\nPlease resubmit with the required documentation within 10 business days.\n\nRegards,\nContracting Office',
  },
  {
    id: 'corr-003',
    subject: 'REA Response - REA-2024-003',
    recipient: 'contracts@globalservices.com',
    type: 'REA_RESPONSE',
    status: 'DRAFT',
    createdAt: '2024-01-17T09:00:00Z',
    body: 'Dear Global Services LLC,\n\nRe: Request for Equitable Adjustment REA-2024-003\n\nAfter careful review, the Government has partially approved your REA for $75,000 of the requested $120,000.\n\nA contract modification will be issued within 15 business days.\n\nRegards,\nContracting Officer',
  },
  {
    id: 'corr-004',
    subject: 'Escalation Notification - PMT-2024-1102',
    recipient: 'senior.co@agency.gov',
    type: 'ESCALATION_NOTIFICATION',
    status: 'SENT',
    createdAt: '2024-01-17T16:45:00Z',
    body: 'URGENT: Payment PMT-2024-1102 requires your immediate review.\n\nAmount: $1,250,000.00\nReason: Amount exceeds standard approval threshold.\n\nPlease review and take action within 24 hours.',
  },
  {
    id: 'corr-005',
    subject: 'Payment Approval Confirmation - INV-2024-0955',
    recipient: 'finance@techsolutions.com',
    type: 'APPROVAL_CONFIRMATION',
    status: 'DRAFT',
    createdAt: '2024-01-18T08:20:00Z',
    body: 'Dear TechSolutions,\n\nYour payment of $22,750.00 for Invoice INV-2024-0955 has been approved and is processing for disbursement.\n\nRegards,\nContracting Office',
  },
]

const STATUS_STYLES: Record<CorrespondenceStatus, string> = {
  DRAFT: 'bg-gray-100 text-gray-700',
  PENDING: 'bg-yellow-100 text-yellow-800',
  SENT: 'bg-green-100 text-green-800',
}

/**
 * Correspondence management page showing generated correspondence items
 * with status badges and preview capability.
 *
 * Validates: Requirement 26.3
 */
function Correspondence() {
  const [selectedItem, setSelectedItem] = useState<CorrespondenceItem | null>(null)

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Correspondence</h1>
        <p className="mt-1 text-sm text-gray-600">
          Manage generated correspondence, review drafts, and track sent communications.
        </p>
      </div>

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
        {/* Correspondence list */}
        <div className="lg:col-span-2">
          <div className="rounded-lg border border-gray-200 bg-white">
            <div className="border-b border-gray-200 px-4 py-3">
              <h2 className="text-sm font-semibold text-gray-900">All Correspondence</h2>
            </div>
            <ul className="divide-y divide-gray-100" role="list">
              {SAMPLE_CORRESPONDENCE.map((item) => (
                <li key={item.id}>
                  <button
                    className={`w-full px-4 py-3 text-left hover:bg-gray-50 transition-colors ${
                      selectedItem?.id === item.id ? 'bg-indigo-50' : ''
                    }`}
                    onClick={() => setSelectedItem(item)}
                    aria-label={`View correspondence: ${item.subject}`}
                  >
                    <div className="flex items-start justify-between gap-2">
                      <div className="min-w-0 flex-1">
                        <p className="truncate text-sm font-medium text-gray-900">
                          {item.subject}
                        </p>
                        <p className="mt-0.5 text-xs text-gray-500">
                          To: {item.recipient}
                        </p>
                        <p className="mt-0.5 text-xs text-gray-400">
                          {new Date(item.createdAt).toLocaleDateString()}
                        </p>
                      </div>
                      <span
                        className={`inline-flex shrink-0 items-center rounded-full px-2 py-0.5 text-xs font-medium ${STATUS_STYLES[item.status]}`}
                      >
                        {item.status}
                      </span>
                    </div>
                  </button>
                </li>
              ))}
            </ul>
          </div>
        </div>

        {/* Preview panel */}
        <div className="lg:col-span-1">
          <div className="sticky top-4 rounded-lg border border-gray-200 bg-white p-4">
            <h2 className="mb-3 text-sm font-semibold text-gray-900">Preview</h2>
            {selectedItem ? (
              <div className="space-y-3">
                <div>
                  <span
                    className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${STATUS_STYLES[selectedItem.status]}`}
                  >
                    {selectedItem.status}
                  </span>
                </div>
                <div>
                  <p className="text-xs font-medium text-gray-500">Subject</p>
                  <p className="text-sm text-gray-900">{selectedItem.subject}</p>
                </div>
                <div>
                  <p className="text-xs font-medium text-gray-500">Recipient</p>
                  <p className="text-sm text-gray-900">{selectedItem.recipient}</p>
                </div>
                <div>
                  <p className="text-xs font-medium text-gray-500">Type</p>
                  <p className="text-sm text-gray-900">{selectedItem.type.replace(/_/g, ' ')}</p>
                </div>
                <div>
                  <p className="text-xs font-medium text-gray-500">Body</p>
                  <pre className="mt-1 whitespace-pre-wrap rounded bg-gray-50 p-2 text-xs text-gray-700 font-sans">
                    {selectedItem.body}
                  </pre>
                </div>
              </div>
            ) : (
              <p className="text-sm text-gray-500">
                Select a correspondence item to preview its content.
              </p>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}

export default Correspondence
