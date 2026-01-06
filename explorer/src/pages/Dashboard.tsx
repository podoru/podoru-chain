import { useEffect, useState } from 'react'
import { api, ChainInfo } from '~/lib/api'
import { useWebSocket, BlockEvent, TransactionEvent } from '~/hooks/useWebSocket'
import { formatHash, formatTimeAgo } from '~/lib/utils'

export default function Dashboard() {
  const [chainInfo, setChainInfo] = useState<ChainInfo | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const { connected, events } = useWebSocket(['new_block', 'new_transaction'])

  useEffect(() => {
    loadChainInfo()
    const interval = setInterval(loadChainInfo, 10000) // Refresh every 10s
    return () => clearInterval(interval)
  }, [])

  async function loadChainInfo() {
    try {
      const info = await api.getChainInfo()
      setChainInfo(info)
      setError(null)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load chain info')
    } finally {
      setLoading(false)
    }
  }

  if (loading) {
    return <div className="card">Loading...</div>
  }

  if (error) {
    return (
      <div className="card">
        <h2>Error</h2>
        <p style={{ color: '#e53e3e' }}>{error}</p>
        <button onClick={loadChainInfo} style={{ marginTop: '1rem' }}>
          Retry
        </button>
      </div>
    )
  }

  return (
    <div>
      {/* Chain Stats */}
      <div className="grid" style={{ gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))' }}>
        <div className="card">
          <div style={{ fontSize: '0.875rem', color: '#718096', marginBottom: '0.5rem' }}>
            Block Height
          </div>
          <div style={{ fontSize: '2rem', fontWeight: '700', color: '#2d3748' }}>
            {chainInfo?.height.toLocaleString()}
          </div>
        </div>

        <div className="card">
          <div style={{ fontSize: '0.875rem', color: '#718096', marginBottom: '0.5rem' }}>
            Authorities
          </div>
          <div style={{ fontSize: '2rem', fontWeight: '700', color: '#2d3748' }}>
            {chainInfo?.authorities.length}
          </div>
        </div>

        <div className="card">
          <div style={{ fontSize: '0.875rem', color: '#718096', marginBottom: '0.5rem' }}>
            WebSocket
          </div>
          <span className={`status ${connected ? 'connected' : 'disconnected'}`}>
            {connected ? '● Connected' : '○ Disconnected'}
          </span>
        </div>
      </div>

      {/* Activity Feed */}
      <div className="card">
        <h2>🔔 Live Activity Feed</h2>

        {events.length === 0 && (
          <p style={{ color: '#718096', padding: '2rem', textAlign: 'center' }}>
            Waiting for new blocks and transactions...
          </p>
        )}

        <div>
          {events.map((event, index) => (
            <div
              key={index}
              style={{
                padding: '1rem',
                borderLeft: event.type === 'new_block' ? '3px solid #48bb78' : '3px solid #ed8936',
                backgroundColor: '#f7fafc',
                marginBottom: '0.5rem',
                borderRadius: '4px',
              }}
            >
              {event.type === 'new_block' ? (
                <BlockEventCard event={event.data as BlockEvent} timestamp={event.timestamp} />
              ) : (
                <TransactionEventCard event={event.data as TransactionEvent} timestamp={event.timestamp} />
              )}
            </div>
          ))}
        </div>
      </div>

      {/* Current Hash */}
      <div className="card">
        <h2>Chain Info</h2>
        <table style={{ width: '100%' }}>
          <tbody>
            <tr>
              <td style={{ padding: '0.5rem 0', fontWeight: '600', width: '180px' }}>
                Current Hash
              </td>
              <td className="hash">{chainInfo?.current_hash}</td>
            </tr>
            <tr>
              <td style={{ padding: '0.5rem 0', fontWeight: '600' }}>
                Genesis Hash
              </td>
              <td className="hash">{chainInfo?.genesis_hash}</td>
            </tr>
            <tr>
              <td style={{ padding: '0.5rem 0', fontWeight: '600' }}>
                Authorities
              </td>
              <td>
                {chainInfo?.authorities.map((addr, i) => (
                  <div key={i} className="hash" style={{ marginBottom: '0.25rem' }}>
                    {addr}
                  </div>
                ))}
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  )
}

function BlockEventCard({ event, timestamp }: { event: BlockEvent; timestamp: number }) {
  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '0.5rem' }}>
        <div>
          <span style={{ fontWeight: '600', color: '#48bb78' }}>
            🟢 New Block
          </span>
          <span style={{ marginLeft: '1rem', color: '#718096' }}>
            Height {event.height}
          </span>
        </div>
        <div style={{ fontSize: '0.875rem', color: '#a0aec0' }}>
          {formatTimeAgo(timestamp)}
        </div>
      </div>
      <div style={{ fontSize: '0.875rem', color: '#4a5568' }}>
        <div className="hash">Hash: {formatHash(event.hash, 16)}</div>
        <div style={{ marginTop: '0.25rem' }}>
          Producer: <span className="hash">{formatHash(event.producer, 12)}</span>
          {' • '}
          Transactions: {event.transaction_count}
        </div>
      </div>
    </div>
  )
}

function TransactionEventCard({ event, timestamp }: { event: TransactionEvent; timestamp: number }) {
  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '0.5rem' }}>
        <div>
          <span style={{ fontWeight: '600', color: '#ed8936' }}>
            📝 New Transaction
          </span>
          <span style={{ marginLeft: '1rem' }}>
            <span className="status" style={{ backgroundColor: '#fef5e7', color: '#d97706' }}>
              {event.status}
            </span>
          </span>
        </div>
        <div style={{ fontSize: '0.875rem', color: '#a0aec0' }}>
          {formatTimeAgo(timestamp)}
        </div>
      </div>
      <div style={{ fontSize: '0.875rem', color: '#4a5568' }}>
        <div className="hash">Hash: {formatHash(event.hash, 16)}</div>
        <div style={{ marginTop: '0.25rem' }}>
          From: <span className="hash">{formatHash(event.from, 12)}</span>
        </div>
      </div>
    </div>
  )
}
