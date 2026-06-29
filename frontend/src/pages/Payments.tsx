import { useMemo } from 'react'
import { useWebSocketContext } from '../hooks'
import { PipelineTracker } from '../components/PipelineTracker'
import { ActivityFeed } from '../components/ActivityFeed'

/**
 * Payments page with real-time pipeline tracker, activity feed,
 * and escalation notification banner.
 *
 * Validates: Requirements 25.1, 25.2
 */
function Payments() {
  const { messages, status } = useWebSocketContext()

  // Detect escalation events to show the notification banner
  const escalations = useMemo(
    () => messages.filter((msg) => msg.type === 'ESCALATION'),
    [messages],
  )

  const latestEscalation = escalations.length > 0 ? escalations[escalations.length - 1] : null

  return (
    <div className="p-4 sm:p-6 lg:p-8">
      {/* Header */}
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Payments</h1>
          <p className="mt-1 text-gray-600">Payment processing pipeline</p>
        </div>
        <ConnectionBadge status={status} />
      </div>

      {/* Escalation notification banner */}
      {latestEscalation && latestEscalation.type === 'ESCALATION' && (
        <div
          className="mb-6 flex items-start gap-3 rounded-lg border border-red-200 bg-red-50 p-4"
          role="alert"
        >
          <svg className="mt-0.5 h-5 w-5 flex-shrink-0 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L4.082 16.5c-.77.833.192 2.5 1.732 2.5z" />
          </svg>
          <div className="flex-1">
            <h3 className="text-sm font-semibold text-red-800">Escalation Alert</h3>
            <p className="mt-1 text-sm text-red-700">{latestEscalation.reason}</p>
            <p className="mt-1 text-xs text-red-600">
              Agent: {latestEscalation.agentName} • Payment: {latestEscalation.paymentId}
            </p>
          </div>
        </div>
      )}

      {/* Pipeline Tracker */}
      <div className="mb-6">
        <PipelineTracker messages={messages} />
      </div>

      {/* Activity Feed */}
      <ActivityFeed messages={messages} />
    </div>
  )
}

function ConnectionBadge({ status }: { status: string }) {
  const config = {
    connected: { color: 'bg-green-100 text-green-800', dot: 'bg-green-500', label: 'Connected' },
    connecting: { color: 'bg-yellow-100 text-yellow-800', dot: 'bg-yellow-500', label: 'Connecting' },
    disconnected: { color: 'bg-red-100 text-red-800', dot: 'bg-red-500', label: 'Disconnected' },
  }[status] ?? { color: 'bg-gray-100 text-gray-800', dot: 'bg-gray-500', label: status }

  return (
    <span className={`inline-flex items-center gap-1.5 rounded-full px-3 py-1 text-xs font-medium ${config.color}`}>
      <span className={`h-2 w-2 rounded-full ${config.dot}`} aria-hidden="true" />
      {config.label}
    </span>
  )
}

export default Payments
