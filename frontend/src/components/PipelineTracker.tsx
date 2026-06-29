import { useMemo } from 'react'
import type { WebSocketMessage } from '../types/websocket'

/**
 * Pipeline stages in processing order.
 * Each stage maps to a status from the WebSocket STATUS_CHANGE events.
 */
const PIPELINE_STAGES: { id: string; label: string; statuses: string[] }[] = [
  { id: 'extraction', label: 'Extraction', statuses: ['EXTRACTING'] },
  { id: 'validation', label: 'Validation', statuses: ['VALIDATING'] },
  { id: 'compliance', label: 'Compliance', statuses: ['CHECKING_COMPLIANCE'] },
  { id: 'routing', label: 'Routing', statuses: ['ROUTING'] },
  { id: 'disbursement', label: 'Disbursement', statuses: ['DISBURSING', 'DISBURSED'] },
]

/** Ingestion channel type */
type IngestionChannel = 'EMAIL' | 'FAX' | 'MAIL' | 'PORTAL'

/** Channel badge color mapping */
const CHANNEL_STYLES: Record<IngestionChannel, string> = {
  EMAIL: 'bg-blue-100 text-blue-800',
  FAX: 'bg-purple-100 text-purple-800',
  MAIL: 'bg-amber-100 text-amber-800',
  PORTAL: 'bg-green-100 text-green-800',
}

/** Channel badge icons */
const CHANNEL_ICONS: Record<IngestionChannel, string> = {
  EMAIL: '📧',
  FAX: '📠',
  MAIL: '📬',
  PORTAL: '🌐',
}

interface PipelineTrackerProps {
  messages: WebSocketMessage[]
  /** The ingestion channel source for this payment */
  channel?: IngestionChannel
}

/**
 * Horizontal step visualization showing agent progress stages.
 * Each step lights up as real-time WebSocket events arrive.
 * Displays a channel badge showing the ingestion source (EMAIL/FAX/MAIL/PORTAL).
 *
 * Validates: Requirements 25.1, 25.2, 24.3
 */
export function PipelineTracker({ messages, channel }: PipelineTrackerProps) {
  const activeStageIndex = useMemo(() => {
    // Find the furthest stage reached by examining STATUS_CHANGE messages
    let maxIndex = -1

    for (const msg of messages) {
      if (msg.type === 'STATUS_CHANGE') {
        for (let i = 0; i < PIPELINE_STAGES.length; i++) {
          if (PIPELINE_STAGES[i].statuses.includes(msg.newStatus)) {
            maxIndex = Math.max(maxIndex, i)
          }
        }
      }
      if (msg.type === 'COMPLETE') {
        maxIndex = PIPELINE_STAGES.length - 1
      }
    }

    return maxIndex
  }, [messages])

  return (
    <div className="w-full rounded-lg border border-gray-200 bg-white p-4 sm:p-6">
      <div className="mb-4 flex items-center justify-between">
        <h2 className="text-lg font-semibold text-gray-900">Pipeline Progress</h2>
        {channel && <ChannelBadge channel={channel} />}
      </div>
      <div className="flex items-center justify-between">
        {PIPELINE_STAGES.map((stage, index) => {
          const isCompleted = index < activeStageIndex
          const isActive = index === activeStageIndex
          const isPending = index > activeStageIndex

          return (
            <div key={stage.id} className="flex flex-1 items-center">
              {/* Step circle + label */}
              <div className="flex flex-col items-center">
                <div
                  className={`flex h-8 w-8 items-center justify-center rounded-full border-2 text-sm font-medium transition-colors sm:h-10 sm:w-10 ${
                    isCompleted
                      ? 'border-green-500 bg-green-500 text-white'
                      : isActive
                        ? 'border-blue-500 bg-blue-500 text-white animate-pulse'
                        : 'border-gray-300 bg-white text-gray-400'
                  }`}
                  aria-label={`${stage.label}: ${isCompleted ? 'completed' : isActive ? 'in progress' : 'pending'}`}
                >
                  {isCompleted ? (
                    <svg className="h-4 w-4 sm:h-5 sm:w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                    </svg>
                  ) : (
                    <span>{index + 1}</span>
                  )}
                </div>
                <span
                  className={`mt-2 text-xs font-medium sm:text-sm ${
                    isCompleted
                      ? 'text-green-600'
                      : isActive
                        ? 'text-blue-600'
                        : 'text-gray-400'
                  }`}
                >
                  {stage.label}
                </span>
              </div>

              {/* Connector line between steps */}
              {index < PIPELINE_STAGES.length - 1 && (
                <div
                  className={`mx-1 h-0.5 flex-1 transition-colors sm:mx-2 ${
                    index < activeStageIndex
                      ? 'bg-green-500'
                      : isPending
                        ? 'bg-gray-200'
                        : 'bg-blue-300'
                  }`}
                  aria-hidden="true"
                />
              )}
            </div>
          )
        })}
      </div>
    </div>
  )
}

/** Badge component displaying the ingestion channel source */
function ChannelBadge({ channel }: { channel: IngestionChannel }) {
  const style = CHANNEL_STYLES[channel] || 'bg-gray-100 text-gray-800'
  const icon = CHANNEL_ICONS[channel] || '📄'

  return (
    <span
      className={`inline-flex items-center gap-1 rounded-full px-2.5 py-0.5 text-xs font-medium ${style}`}
      aria-label={`Ingestion channel: ${channel}`}
    >
      <span aria-hidden="true">{icon}</span>
      {channel}
    </span>
  )
}
