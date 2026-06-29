import { useCallback, useEffect, useRef, useState } from 'react'
import type { ConnectionStatus, WebSocketMessage } from '../types/websocket'

const DEFAULT_WS_URL =
  import.meta.env.VITE_WS_URL || 'ws://localhost:3001'

const MAX_RECONNECT_DELAY = 30_000
const INITIAL_RECONNECT_DELAY = 1_000

interface UseWebSocketOptions {
  url?: string
  /** Enable simulated demo mode that generates sample events */
  simulatedMode?: boolean
  onMessage?: (message: WebSocketMessage) => void
}

/**
 * Custom React hook for WebSocket connection with exponential backoff reconnection.
 * Connects to the API Gateway WebSocket endpoint for real-time payment updates.
 *
 * Validates: Requirements 25.1, 25.2, 25.3, 25.4
 */
export function useWebSocket(options: UseWebSocketOptions = {}) {
  const { url = DEFAULT_WS_URL, simulatedMode = false, onMessage } = options

  const [status, setStatus] = useState<ConnectionStatus>('disconnected')
  const [lastMessage, setLastMessage] = useState<WebSocketMessage | null>(null)
  const [messages, setMessages] = useState<WebSocketMessage[]>([])

  const wsRef = useRef<WebSocket | null>(null)
  const reconnectAttemptRef = useRef(0)
  const reconnectTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const simulationIntervalRef = useRef<ReturnType<typeof setInterval> | null>(null)
  const onMessageRef = useRef(onMessage)

  // Keep callback ref up to date
  useEffect(() => {
    onMessageRef.current = onMessage
  }, [onMessage])

  const handleIncomingMessage = useCallback((msg: WebSocketMessage) => {
    setLastMessage(msg)
    setMessages((prev) => [...prev, msg])
    onMessageRef.current?.(msg)
  }, [])

  const getReconnectDelay = useCallback(() => {
    const delay = INITIAL_RECONNECT_DELAY * Math.pow(2, reconnectAttemptRef.current)
    return Math.min(delay, MAX_RECONNECT_DELAY)
  }, [])

  const connect = useCallback(() => {
    if (simulatedMode) {
      // In simulated mode, skip actual WebSocket and generate demo events
      setStatus('connected')
      return
    }

    setStatus('connecting')

    try {
      const ws = new WebSocket(url)

      ws.onopen = () => {
        setStatus('connected')
        reconnectAttemptRef.current = 0
      }

      ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data) as WebSocketMessage
          handleIncomingMessage(data)
        } catch {
          // Ignore malformed messages
        }
      }

      ws.onclose = () => {
        setStatus('disconnected')
        wsRef.current = null
        scheduleReconnect()
      }

      ws.onerror = () => {
        ws.close()
      }

      wsRef.current = ws
    } catch {
      setStatus('disconnected')
      scheduleReconnect()
    }
  }, [url, simulatedMode, handleIncomingMessage])

  const scheduleReconnect = useCallback(() => {
    const delay = getReconnectDelay()
    reconnectAttemptRef.current += 1

    reconnectTimeoutRef.current = setTimeout(() => {
      connect()
    }, delay)
  }, [connect, getReconnectDelay])

  const disconnect = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current)
      reconnectTimeoutRef.current = null
    }
    if (wsRef.current) {
      wsRef.current.close()
      wsRef.current = null
    }
    if (simulationIntervalRef.current) {
      clearInterval(simulationIntervalRef.current)
      simulationIntervalRef.current = null
    }
    setStatus('disconnected')
  }, [])

  const sendMessage = useCallback((data: Record<string, unknown>) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify(data))
    }
  }, [])

  const clearMessages = useCallback(() => {
    setMessages([])
    setLastMessage(null)
  }, [])

  // Start simulated demo events when in simulated mode
  useEffect(() => {
    if (!simulatedMode || status !== 'connected') return

    const samplePaymentId = 'demo-payment-001'
    const stages: WebSocketMessage[] = [
      { type: 'STATUS_CHANGE', paymentId: samplePaymentId, previousStatus: 'RECEIVED', newStatus: 'EXTRACTING', timestamp: new Date().toISOString() },
      { type: 'AGENT_RESULT', paymentId: samplePaymentId, agentName: 'OCR/Extraction', result: 'Document classified as INVOICE. 8 fields extracted.', confidence: 0.92 },
      { type: 'STATUS_CHANGE', paymentId: samplePaymentId, previousStatus: 'EXTRACTING', newStatus: 'VALIDATING', timestamp: new Date().toISOString() },
      { type: 'AGENT_RESULT', paymentId: samplePaymentId, agentName: 'Validation', result: 'All required fields present. No duplicates detected.', confidence: 0.95 },
      { type: 'STATUS_CHANGE', paymentId: samplePaymentId, previousStatus: 'VALIDATING', newStatus: 'CHECKING_COMPLIANCE', timestamp: new Date().toISOString() },
      { type: 'AGENT_RESULT', paymentId: samplePaymentId, agentName: 'Compliance', result: 'OFAC clear. FAR compliant. No threshold exceeded.', confidence: 0.98 },
      { type: 'STATUS_CHANGE', paymentId: samplePaymentId, previousStatus: 'CHECKING_COMPLIANCE', newStatus: 'ROUTING', timestamp: new Date().toISOString() },
      { type: 'AGENT_RESULT', paymentId: samplePaymentId, agentName: 'Routing', result: 'Routed to SUPERVISOR level. Priority: NORMAL.', confidence: 0.99 },
      { type: 'STATUS_CHANGE', paymentId: samplePaymentId, previousStatus: 'ROUTING', newStatus: 'DISBURSING', timestamp: new Date().toISOString() },
      { type: 'COMPLETE', paymentId: samplePaymentId, finalStatus: 'DISBURSED' },
    ]

    let index = 0
    simulationIntervalRef.current = setInterval(() => {
      if (index < stages.length) {
        handleIncomingMessage(stages[index])
        index++
      } else {
        // Restart simulation loop with escalation scenario
        index = 0
        const escalationStages: WebSocketMessage[] = [
          { type: 'STATUS_CHANGE', paymentId: 'demo-payment-002', previousStatus: 'RECEIVED', newStatus: 'EXTRACTING', timestamp: new Date().toISOString() },
          { type: 'AGENT_RESULT', paymentId: 'demo-payment-002', agentName: 'OCR/Extraction', result: 'Low confidence extraction. Multiple fields uncertain.', confidence: 0.62 },
          { type: 'ESCALATION', paymentId: 'demo-payment-002', reason: 'Extraction confidence below threshold (0.62 < 0.75)', agentName: 'OCR/Extraction' },
        ]
        escalationStages.forEach((msg) => handleIncomingMessage(msg))
        if (simulationIntervalRef.current) {
          clearInterval(simulationIntervalRef.current)
          simulationIntervalRef.current = null
        }
      }
    }, 2000)

    return () => {
      if (simulationIntervalRef.current) {
        clearInterval(simulationIntervalRef.current)
        simulationIntervalRef.current = null
      }
    }
  }, [simulatedMode, status, handleIncomingMessage])

  // Connect on mount
  useEffect(() => {
    connect()
    return () => {
      disconnect()
    }
  }, [connect, disconnect])

  return {
    status,
    lastMessage,
    messages,
    sendMessage,
    disconnect,
    connect,
    clearMessages,
  }
}
