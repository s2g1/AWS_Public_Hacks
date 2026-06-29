/**
 * WebSocket message types for real-time payment pipeline updates.
 * Validates: Requirements 25.1, 25.2, 25.3, 25.4
 */

export type WebSocketMessageType =
  | 'STATUS_CHANGE'
  | 'AGENT_RESULT'
  | 'ESCALATION'
  | 'COMPLETE'

export type IngestionChannel = 'EMAIL' | 'FAX' | 'MAIL' | 'PORTAL'

export interface StatusChangeEvent {
  type: 'STATUS_CHANGE'
  paymentId: string
  previousStatus: string
  newStatus: string
  timestamp: string
  channel?: IngestionChannel
}

export interface AgentResultEvent {
  type: 'AGENT_RESULT'
  paymentId: string
  agentName: string
  result: string
  confidence: number
  channel?: IngestionChannel
}

export interface EscalationEvent {
  type: 'ESCALATION'
  paymentId: string
  reason: string
  agentName: string
  channel?: IngestionChannel
}

export interface CompleteEvent {
  type: 'COMPLETE'
  paymentId: string
  finalStatus: string
  channel?: IngestionChannel
}

export type WebSocketMessage =
  | StatusChangeEvent
  | AgentResultEvent
  | EscalationEvent
  | CompleteEvent

export type ConnectionStatus = 'connecting' | 'connected' | 'disconnected'
