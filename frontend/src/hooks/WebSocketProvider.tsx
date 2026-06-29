import React, { createContext, useContext, useMemo } from 'react'
import type { ConnectionStatus, WebSocketMessage } from '../types/websocket'
import { useWebSocket } from './useWebSocket'

interface WebSocketContextValue {
  /** Current connection status */
  status: ConnectionStatus
  /** Most recently received message */
  lastMessage: WebSocketMessage | null
  /** All received messages since connection (or last clear) */
  messages: WebSocketMessage[]
  /** Send a JSON message over the WebSocket */
  sendMessage: (data: Record<string, unknown>) => void
  /** Manually disconnect */
  disconnect: () => void
  /** Manually reconnect */
  connect: () => void
  /** Clear the message history */
  clearMessages: () => void
}

const WebSocketContext = createContext<WebSocketContextValue | null>(null)

interface WebSocketProviderProps {
  children: React.ReactNode
  /** Override the WebSocket URL (defaults to VITE_WS_URL env or localhost) */
  url?: string
  /** Enable demo simulation mode for hackathon demonstrations */
  simulatedMode?: boolean
}

/**
 * React context provider that wraps the app and provides access
 * to the latest WebSocket events for real-time pipeline updates.
 *
 * Validates: Requirements 25.1, 25.2, 25.3, 25.4
 */
export function WebSocketProvider({
  children,
  url,
  simulatedMode = true,
}: WebSocketProviderProps) {
  const ws = useWebSocket({ url, simulatedMode })

  const value = useMemo<WebSocketContextValue>(
    () => ({
      status: ws.status,
      lastMessage: ws.lastMessage,
      messages: ws.messages,
      sendMessage: ws.sendMessage,
      disconnect: ws.disconnect,
      connect: ws.connect,
      clearMessages: ws.clearMessages,
    }),
    [ws.status, ws.lastMessage, ws.messages, ws.sendMessage, ws.disconnect, ws.connect, ws.clearMessages],
  )

  return (
    <WebSocketContext.Provider value={value}>
      {children}
    </WebSocketContext.Provider>
  )
}

/**
 * Hook to consume the WebSocket context from any component in the tree.
 * Must be used within a WebSocketProvider.
 */
export function useWebSocketContext(): WebSocketContextValue {
  const context = useContext(WebSocketContext)
  if (!context) {
    throw new Error('useWebSocketContext must be used within a WebSocketProvider')
  }
  return context
}
