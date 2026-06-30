import { useAppContext } from '../store/AppContext'

function formatTimestamp(ts: string): string {
  return new Date(ts).toLocaleString('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function getActionIcon(action: string): string {
  const lower = action.toLowerCase()
  if (lower.includes('solicitation')) return '📋'
  if (lower.includes('proposal')) return '📝'
  if (lower.includes('invoice')) return '🧾'
  if (lower.includes('payment')) return '💳'
  if (lower.includes('mod')) return '🔧'
  if (lower.includes('approved') || lower.includes('approve')) return '✅'
  if (lower.includes('rejected') || lower.includes('reject')) return '❌'
  return '📄'
}

function History() {
  const { state } = useAppContext()
  const isGov = state.currentRole === 'GOV'

  // Filter history entries by role
  const entries = state.history.filter(h => h.actor === state.currentRole)

  return (
    <div className="p-4 sm:p-6 lg:p-8 max-w-7xl mx-auto">
      {/* Header */}
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-900">History</h1>
        <p className="mt-1 text-sm text-gray-600">
          {isGov
            ? 'Your government actions — solicitations, reviews, approvals, and modifications'
            : 'Your vendor actions — proposals, invoices, mods, and payments received'}
        </p>
      </div>

      {entries.length === 0 ? (
        <div className="text-center py-16 bg-white rounded-lg border border-gray-200">
          <svg className="mx-auto h-16 w-16 text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <p className="mt-4 text-lg font-medium text-gray-700">No History Yet</p>
          <p className="mt-2 text-sm text-gray-500 max-w-md mx-auto">
            Actions you take will appear here in chronological order.
          </p>
        </div>
      ) : (
        <div className="bg-white rounded-lg shadow-sm border border-gray-200">
          <div className="divide-y divide-gray-100">
            {entries.map((entry) => (
              <div key={entry.id} className="p-4 sm:p-5 hover:bg-gray-50 transition-colors">
                <div className="flex items-start gap-3">
                  <span className="text-xl flex-shrink-0 mt-0.5" aria-hidden="true">
                    {getActionIcon(entry.action)}
                  </span>
                  <div className="min-w-0 flex-1">
                    <div className="flex items-center gap-2 flex-wrap">
                      <span className="text-sm font-semibold text-gray-900">{entry.action}</span>
                      <span className="text-xs text-gray-400">•</span>
                      <span className="text-xs text-gray-500">{entry.actorName}</span>
                    </div>
                    <p className="mt-0.5 text-sm text-gray-600">{entry.details}</p>
                    <p className="mt-1 text-xs text-gray-400">{formatTimestamp(entry.timestamp)}</p>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}

export default History
