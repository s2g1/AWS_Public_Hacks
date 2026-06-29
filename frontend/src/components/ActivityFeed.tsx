import { useEffect, useRef } from 'react'
import type { WebSocketMessage } from '../types/websocket'

interface ActivityFeedProps {
  messages: WebSocketMessage[]
}

/**
 * Scrollable feed that appends agent results in real-time as they arrive via WebSocket.
 * Shows agent name, result, confidence, and timestamp.
 *
 * Validates: Requirements 25.1, 25.2
 */
export function ActivityFeed({ messages }: ActivityFeedProps) {
  const feedEndRef = useRef<HTMLDivElement>(null)

  // Auto-scroll to bottom when new messages arrive
  useEffect(() => {
    feedEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages.length])

  if (messages.length === 0) {
    return (
      <div className="rounded-lg border border-gray-200 bg-white p-4 sm:p-6">
        <h2 className="mb-4 text-lg font-semibold text-gray-900">Activity Feed</h2>
        <p className="text-sm text-gray-500">Waiting for pipeline events...</p>
      </div>
    )
  }

  return (
    <div className="rounded-lg border border-gray-200 bg-white p-4 sm:p-6">
      <h2 className="mb-4 text-lg font-semibold text-gray-900">Activity Feed</h2>
      <div className="max-h-96 space-y-3 overflow-y-auto pr-1">
        {messages.map((msg, index) => (
          <ActivityItem key={index} message={msg} />
        ))}
        <div ref={feedEndRef} />
      </div>
    </div>
  )
}

function ActivityItem({ message }: { message: WebSocketMessage }) {
  switch (message.type) {
    case 'AGENT_RESULT':
      return (
        <div className="rounded-md border border-blue-100 bg-blue-50 p-3">
          <div className="flex items-center justify-between">
            <span className="text-sm font-semibold text-blue-800">
              {message.agentName}
            </span>
            <ConfidenceBadge confidence={message.confidence} />
          </div>
          <p className="mt-1 text-sm text-blue-700">{message.result}</p>
          <p className="mt-1 text-xs text-blue-500">Payment: {message.paymentId}</p>
        </div>
      )

    case 'STATUS_CHANGE':
      return (
        <div className="rounded-md border border-gray-100 bg-gray-50 p-3">
          <div className="flex items-center gap-2">
            <span className="inline-block h-2 w-2 rounded-full bg-gray-400" aria-hidden="true" />
            <span className="text-sm text-gray-700">
              Status changed: <span className="font-medium">{message.previousStatus}</span>
              {' → '}
              <span className="font-medium">{message.newStatus}</span>
            </span>
          </div>
          <p className="mt-1 text-xs text-gray-500">
            {message.paymentId} • {formatTimestamp(message.timestamp)}
          </p>
        </div>
      )

    case 'ESCALATION':
      return (
        <div className="rounded-md border border-amber-200 bg-amber-50 p-3">
          <div className="flex items-center gap-2">
            <svg className="h-4 w-4 text-amber-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L4.082 16.5c-.77.833.192 2.5 1.732 2.5z" />
            </svg>
            <span className="text-sm font-semibold text-amber-800">Escalation Required</span>
          </div>
          <p className="mt-1 text-sm text-amber-700">{message.reason}</p>
          <p className="mt-1 text-xs text-amber-600">
            Agent: {message.agentName} • {message.paymentId}
          </p>
        </div>
      )

    case 'COMPLETE':
      return (
        <div className="rounded-md border border-green-100 bg-green-50 p-3">
          <div className="flex items-center gap-2">
            <svg className="h-4 w-4 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            <span className="text-sm font-semibold text-green-800">Payment Complete</span>
          </div>
          <p className="mt-1 text-sm text-green-700">
            Final status: <span className="font-medium">{message.finalStatus}</span>
          </p>
          <p className="mt-1 text-xs text-green-600">{message.paymentId}</p>
        </div>
      )

    default:
      return null
  }
}

function ConfidenceBadge({ confidence }: { confidence: number }) {
  const percentage = Math.round(confidence * 100)
  const color =
    confidence >= 0.9
      ? 'bg-green-100 text-green-800'
      : confidence >= 0.75
        ? 'bg-yellow-100 text-yellow-800'
        : 'bg-red-100 text-red-800'

  return (
    <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${color}`}>
      {percentage}% confidence
    </span>
  )
}

function formatTimestamp(isoString: string): string {
  try {
    const date = new Date(isoString)
    return date.toLocaleTimeString(undefined, {
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    })
  } catch {
    return isoString
  }
}
