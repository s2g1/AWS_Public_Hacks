/**
 * State Sync Service — shared state via S3 + WebSocket event bus
 * 
 * Architecture:
 * - State stored in S3 (read/write via Lambda Function URL)
 * - WebSocket connection receives "state_changed" events
 * - On event: frontend fetches latest state from S3
 * - On local change: write to S3 (which broadcasts to all other clients)
 */

const STATE_SYNC_URL = import.meta.env.VITE_STATE_SYNC_URL || ''
const WS_URL = import.meta.env.VITE_WS_URL || ''

type StateChangeCallback = () => void

let ws: WebSocket | null = null
let reconnectTimer: ReturnType<typeof setTimeout> | null = null
let changeCallback: StateChangeCallback | null = null

/**
 * Connect to the WebSocket event bus
 * Calls onStateChanged when another client modifies the shared state
 */
export function connectEventBus(onStateChanged: StateChangeCallback): void {
  if (!WS_URL) return
  changeCallback = onStateChanged

  function connect() {
    try {
      ws = new WebSocket(WS_URL)

      ws.onopen = () => {
        console.log('[StateSync] WebSocket connected')
        if (reconnectTimer) {
          clearTimeout(reconnectTimer)
          reconnectTimer = null
        }
      }

      ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data)
          if (data.event === 'state_changed' && changeCallback) {
            changeCallback()
          }
        } catch {
          // Ignore non-JSON messages
        }
      }

      ws.onclose = () => {
        console.log('[StateSync] WebSocket disconnected, reconnecting...')
        reconnectTimer = setTimeout(connect, 3000)
      }

      ws.onerror = () => {
        ws?.close()
      }
    } catch {
      reconnectTimer = setTimeout(connect, 3000)
    }
  }

  connect()
}

/**
 * Disconnect from the event bus
 */
export function disconnectEventBus(): void {
  if (reconnectTimer) clearTimeout(reconnectTimer)
  if (ws) {
    ws.onclose = null // Prevent reconnect
    ws.close()
    ws = null
  }
}

/**
 * Read shared state from S3
 */
export async function readSharedState<T>(): Promise<T | null> {
  if (!STATE_SYNC_URL) return null

  try {
    const response = await fetch(STATE_SYNC_URL, { method: 'GET' })
    if (!response.ok) return null
    const body = await response.json()
    if (body.exists === false) return null
    return body as T
  } catch {
    return null
  }
}

/**
 * Write shared state to S3 (triggers WebSocket notification to all other clients)
 */
export async function writeSharedState(state: unknown): Promise<boolean> {
  if (!STATE_SYNC_URL) return false

  try {
    const response = await fetch(STATE_SYNC_URL, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(state),
    })
    return response.ok
  } catch {
    return false
  }
}

/**
 * Check if sync is enabled (both URLs configured)
 */
export function isSyncEnabled(): boolean {
  return !!(STATE_SYNC_URL && WS_URL)
}
