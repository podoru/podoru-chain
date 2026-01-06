import { useEffect, useState } from 'react'
import { api, MempoolInfo } from '~/lib/api'
import { formatHash, formatTimestamp, formatTimeAgo } from '~/lib/utils'

export default function Transactions() {
  const [mempool, setMempool] = useState<MempoolInfo | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    loadMempool()
    const interval = setInterval(loadMempool, 5000) // Refresh every 5s
    return () => clearInterval(interval)
  }, [])

  async function loadMempool() {
    try {
      const data = await api.getMempool()
      setMempool(data)
      setError(null)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load mempool')
    } finally {
      setLoading(false)
    }
  }

  if (loading) {
    return <div className="card">Loading transactions...</div>
  }

  if (error) {
    return (
      <div className="card">
        <h2>Error</h2>
        <p style={{ color: '#e53e3e' }}>{error}</p>
        <button onClick={loadMempool} style={{ marginTop: '1rem' }}>
          Retry
        </button>
      </div>
    )
  }

  return (
    <div>
      <div className="card">
        <h2>📝 Mempool Transactions</h2>
        <p style={{ color: '#718096', marginBottom: '1.5rem' }}>
          {mempool?.transaction_count || 0} pending transactions
        </p>

        {!mempool?.transactions || mempool.transactions.length === 0 ? (
          <div style={{ padding: '3rem', textAlign: 'center', color: '#a0aec0' }}>
            No pending transactions
          </div>
        ) : (
          <table className="table">
            <thead>
              <tr>
                <th>Hash</th>
                <th>From</th>
                <th>Timestamp</th>
                <th>Operations</th>
                <th>Nonce</th>
              </tr>
            </thead>
            <tbody>
              {mempool.transactions.map((tx) => (
                <tr key={tx.id}>
                  <td className="hash">
                    {formatHash(tx.id, 16)}
                  </td>
                  <td className="hash">
                    {formatHash(tx.from, 12)}
                  </td>
                  <td>
                    <div>{formatTimestamp(tx.timestamp)}</div>
                    <div style={{ fontSize: '0.75rem', color: '#a0aec0' }}>
                      {formatTimeAgo(tx.timestamp)}
                    </div>
                  </td>
                  <td>
                    <div style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem' }}>
                      {tx.data.operations.map((op, i) => (
                        <div key={i} style={{ fontSize: '0.875rem' }}>
                          <span style={{
                            backgroundColor: op.type === 'SET' ? '#bee3f8' : '#feebc8',
                            color: op.type === 'SET' ? '#2c5282' : '#7c2d12',
                            padding: '0.125rem 0.5rem',
                            borderRadius: '3px',
                            fontSize: '0.75rem',
                            fontWeight: '500',
                            marginRight: '0.5rem'
                          }}>
                            {op.type}
                          </span>
                          <span className="hash">{op.key}</span>
                        </div>
                      ))}
                    </div>
                  </td>
                  <td style={{ fontFamily: 'monospace' }}>
                    {tx.nonce}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  )
}
