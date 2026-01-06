import { useEffect, useState, useRef, useCallback } from 'react'

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
  const reconnectTimeoutRef = useRef<NodeJS.Timeout>()

  const connect = useCallback(() => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      return
    }

    const ws = new WebSocket(WS_URL)

    ws.onopen = () => {
      console.log('WebSocket connected')
      setConnected(true)

      // Subscribe to events if specified
      if (subscriptions.length > 0) {
        ws.send(JSON.stringify({
          action: 'subscribe',
          events: subscriptions
        }))
      }
    }

    ws.onclose = () => {
      console.log('WebSocket disconnected')
      setConnected(false)

      // Attempt to reconnect after 3 seconds
      reconnectTimeoutRef.current = setTimeout(() => {
        connect()
      }, 3000)
    }

    ws.onerror = (error) => {
      console.error('WebSocket error:', error)
    }

    ws.onmessage = (event) => {
      try {
        const data: WebSocketEvent = JSON.parse(event.data)
        setLastEvent(data)
        setEvents(prev => [data, ...prev].slice(0, 100)) // Keep last 100 events
      } catch (err) {
        console.error('Failed to parse WebSocket message:', err)
      }
    }

    wsRef.current = ws
  }, [subscriptions])

  const disconnect = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current)
    }
    if (wsRef.current) {
      wsRef.current.close()
      wsRef.current = null
    }
  }, [])

  useEffect(() => {
    connect()

    return () => {
      disconnect()
    }
  }, [connect, disconnect])

  return {
    connected,
    lastEvent,
    events,
    connect,
    disconnect,
  }
}
