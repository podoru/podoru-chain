import { useEffect, useState, useRef } from 'react'

const WS_URL = import.meta.env.VITE_WS_URL || 'ws://localhost:8545/api/v1/ws'

export type EventType = 'new_block' | 'new_transaction' | 'chain_update' | 'mempool_update'

export interface BlockEvent {
  height: number
  hash: string
  timestamp: number
  transaction_count: number
  producer: string
  previous_hash: string
}

export interface TransactionEvent {
  hash: string
  from: string
  timestamp: number
  status: string
  nonce: number
}

export interface WebSocketEvent {
  type: EventType
  data: BlockEvent | TransactionEvent | any
  timestamp: number
}

export function useWebSocket(subscriptions: EventType[] = []) {
  const [connected, setConnected] = useState(false)
  const [lastEvent, setLastEvent] = useState<WebSocketEvent | null>(null)
  const [events, setEvents] = useState<WebSocketEvent[]>([])
  const wsRef = useRef<WebSocket | null>(null)
  const reconnectTimeoutRef = useRef<number>()
  const subscriptionsRef = useRef(subscriptions)

  useEffect(() => {
    subscriptionsRef.current = subscriptions
  }, [subscriptions])

  useEffect(() => {
    let ws: WebSocket | null = null

    function connect() {
      if (ws?.readyState === WebSocket.OPEN) {
        return
      }

      console.log('Connecting to WebSocket...')
      ws = new WebSocket(WS_URL)
      wsRef.current = ws

      ws.onopen = () => {
        console.log('WebSocket connected')
        setConnected(true)

        // Subscribe to events if specified
        if (subscriptionsRef.current.length > 0) {
          const msg = JSON.stringify({
            action: 'subscribe',
            events: subscriptionsRef.current
          })
          console.log('Sending subscription:', msg)
          ws!.send(msg)
        }
      }

      ws.onclose = (event) => {
        console.log('WebSocket disconnected', {
          code: event.code,
          reason: event.reason,
          wasClean: event.wasClean
        })
        setConnected(false)

        // Attempt to reconnect after 3 seconds
        reconnectTimeoutRef.current = window.setTimeout(() => {
          console.log('Attempting to reconnect...')
          connect()
        }, 3000)
      }

      ws.onerror = (error) => {
        console.error('WebSocket error:', error)
      }

      ws.onmessage = (event) => {
        try {
          const data: WebSocketEvent = JSON.parse(event.data)
          console.log('Received WebSocket event:', data.type, data)
          setLastEvent(data)

          // Deduplicate events by checking if the same event already exists
          setEvents(prev => {
            // Check if event already exists (same type, timestamp, and hash/height)
            const isDuplicate = prev.some(e => {
              if (e.type !== data.type || e.timestamp !== data.timestamp) return false

              // For block events, check height
              if (data.type === 'new_block' && 'height' in data.data && 'height' in e.data) {
                return data.data.height === e.data.height
              }

              // For transaction events, check hash
              if (data.type === 'new_transaction' && 'hash' in data.data && 'hash' in e.data) {
                return data.data.hash === e.data.hash
              }

              return false
            })

            if (isDuplicate) {
              console.log('Skipping duplicate event:', data.type)
              return prev
            }

            return [data, ...prev].slice(0, 100) // Keep last 100 events
          })
        } catch (err) {
          console.error('Failed to parse WebSocket message:', err, event.data)
        }
      }
    }

    connect()

    return () => {
      console.log('Cleaning up WebSocket connection')
      if (reconnectTimeoutRef.current) {
        window.clearTimeout(reconnectTimeoutRef.current)
      }
      if (ws) {
        ws.close()
        ws = null
      }
    }
  }, []) // Empty dependency array - only run once

  return {
    connected,
    lastEvent,
    events,
  }
}
