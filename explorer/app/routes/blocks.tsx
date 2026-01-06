import { createFileRoute, Link } from '@tanstack/react-router'
import { useEffect, useState } from 'react'
import { api, Block } from '~/lib/api'
import { formatHash, formatTimestamp, formatTimeAgo } from '~/lib/utils'

export const Route = createFileRoute('/blocks')({
  component: BlocksPage,
})

function BlocksPage() {
  const [blocks, setBlocks] = useState<Block[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    loadLatestBlocks()
  }, [])

  async function loadLatestBlocks() {
    try {
      const info = await api.getChainInfo()
      const latestBlocks: Block[] = []

      // Load last 20 blocks
      const count = Math.min(20, info.height + 1)
      for (let i = 0; i < count; i++) {
        const height = info.height - i
        if (height >= 0) {
          const block = await api.getBlockByHeight(height)
          latestBlocks.push(block)
        }
      }

      setBlocks(latestBlocks)
      setError(null)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load blocks')
    } finally {
      setLoading(false)
    }
  }

  if (loading) {
    return <div className="card">Loading blocks...</div>
  }

  if (error) {
    return (
      <div className="card">
        <h2>Error</h2>
        <p style={{ color: '#e53e3e' }}>{error}</p>
        <button onClick={loadLatestBlocks} style={{ marginTop: '1rem' }}>
          Retry
        </button>
      </div>
    )
  }

  return (
    <div>
      <div className="card">
        <h2>📦 Recent Blocks</h2>
        <p style={{ color: '#718096', marginBottom: '1.5rem' }}>
          Showing the latest {blocks.length} blocks
        </p>

        <table className="table">
          <thead>
            <tr>
              <th>Height</th>
              <th>Hash</th>
              <th>Timestamp</th>
              <th>Producer</th>
              <th>Txs</th>
            </tr>
          </thead>
          <tbody>
            {blocks.map((block) => (
              <tr key={block.header.height}>
                <td style={{ fontWeight: '600' }}>
                  {block.header.height}
                </td>
                <td className="hash">
                  {formatHash(block.header.previous_hash ? `0x${block.header.previous_hash}` : '0x00', 16)}
                </td>
                <td>
                  <div>{formatTimestamp(block.header.timestamp)}</div>
                  <div style={{ fontSize: '0.75rem', color: '#a0aec0' }}>
                    {formatTimeAgo(block.header.timestamp)}
                  </div>
                </td>
                <td className="hash">
                  {formatHash(block.header.producer_addr, 12)}
                </td>
                <td>
                  <span style={{
                    backgroundColor: block.transactions.length > 0 ? '#c6f6d5' : '#e2e8f0',
                    color: block.transactions.length > 0 ? '#22543d' : '#4a5568',
                    padding: '0.25rem 0.5rem',
                    borderRadius: '4px',
                    fontSize: '0.875rem',
                    fontWeight: '500'
                  }}>
                    {block.transactions.length}
                  </span>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}
