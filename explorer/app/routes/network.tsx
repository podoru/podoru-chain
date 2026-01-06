import { createFileRoute } from '@tanstack/react-router'
import { useEffect, useState } from 'react'
import { api, NodeInfo, Peer, ChainInfo } from '~/lib/api'
import { formatTimeAgo } from '~/lib/utils'

export const Route = createFileRoute('/network')({
  component: NetworkPage,
})

function NetworkPage() {
  const [nodeInfo, setNodeInfo] = useState<NodeInfo | null>(null)
  const [peers, setPeers] = useState<Peer[]>([])
  const [chainInfo, setChainInfo] = useState<ChainInfo | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    loadNetworkInfo()
    const interval = setInterval(loadNetworkInfo, 10000) // Refresh every 10s
    return () => clearInterval(interval)
  }, [])

  async function loadNetworkInfo() {
    try {
      const [node, peersList, chain] = await Promise.all([
        api.getNodeInfo(),
        api.getPeers(),
        api.getChainInfo(),
      ])

      setNodeInfo(node)
      setPeers(peersList)
      setChainInfo(chain)
      setError(null)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load network info')
    } finally {
      setLoading(false)
    }
  }

  if (loading) {
    return <div className="card">Loading network information...</div>
  }

  if (error) {
    return (
      <div className="card">
        <h2>Error</h2>
        <p style={{ color: '#e53e3e' }}>{error}</p>
        <button onClick={loadNetworkInfo} style={{ marginTop: '1rem' }}>
          Retry
        </button>
      </div>
    )
  }

  return (
    <div>
      {/* Node Info */}
      <div className="card">
        <h2>🖥️ Node Information</h2>
        <table style={{ width: '100%', marginTop: '1rem' }}>
          <tbody>
            <tr>
              <td style={{ padding: '0.5rem 0', fontWeight: '600', width: '180px' }}>
                Node Type
              </td>
              <td>
                <span style={{
                  backgroundColor: nodeInfo?.node_type === 'producer' ? '#c6f6d5' : '#bee3f8',
                  color: nodeInfo?.node_type === 'producer' ? '#22543d' : '#2c5282',
                  padding: '0.25rem 0.75rem',
                  borderRadius: '12px',
                  fontSize: '0.875rem',
                  fontWeight: '500'
                }}>
                  {nodeInfo?.node_type}
                </span>
              </td>
            </tr>
            <tr>
              <td style={{ padding: '0.5rem 0', fontWeight: '600' }}>
                Address
              </td>
              <td className="hash">{nodeInfo?.address}</td>
            </tr>
            <tr>
              <td style={{ padding: '0.5rem 0', fontWeight: '600' }}>
                P2P Address
              </td>
              <td className="hash">{nodeInfo?.p2p_address}</td>
            </tr>
            <tr>
              <td style={{ padding: '0.5rem 0', fontWeight: '600' }}>
                Version
              </td>
              <td>{nodeInfo?.version || '1.0.0'}</td>
            </tr>
          </tbody>
        </table>
      </div>

      {/* Network Stats */}
      <div className="grid">
        <div className="card">
          <div style={{ fontSize: '0.875rem', color: '#718096', marginBottom: '0.5rem' }}>
            Connected Peers
          </div>
          <div style={{ fontSize: '2rem', fontWeight: '700', color: '#2d3748' }}>
            {peers.length}
          </div>
        </div>

        <div className="card">
          <div style={{ fontSize: '0.875rem', color: '#718096', marginBottom: '0.5rem' }}>
            Total Authorities
          </div>
          <div style={{ fontSize: '2rem', fontWeight: '700', color: '#2d3748' }}>
            {chainInfo?.authorities.length || 0}
          </div>
        </div>

        <div className="card">
          <div style={{ fontSize: '0.875rem', color: '#718096', marginBottom: '0.5rem' }}>
            Block Height
          </div>
          <div style={{ fontSize: '2rem', fontWeight: '700', color: '#2d3748' }}>
            {chainInfo?.height.toLocaleString()}
          </div>
        </div>
      </div>

      {/* Peers List */}
      <div className="card">
        <h2>👥 Connected Peers</h2>

        {peers.length === 0 ? (
          <div style={{ padding: '3rem', textAlign: 'center', color: '#a0aec0' }}>
            No connected peers
          </div>
        ) : (
          <table className="table">
            <thead>
              <tr>
                <th>Peer ID</th>
                <th>Address</th>
                <th>Connected Since</th>
              </tr>
            </thead>
            <tbody>
              {peers.map((peer) => (
                <tr key={peer.id}>
                  <td className="hash">{peer.id}</td>
                  <td className="hash">{peer.address}</td>
                  <td>
                    <div style={{ fontSize: '0.75rem', color: '#718096' }}>
                      {formatTimeAgo(peer.connected_at)}
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      {/* Authorities */}
      <div className="card">
        <h2>⚡ Block Authorities</h2>
        <p style={{ color: '#718096', marginBottom: '1rem' }}>
          Authorized nodes that can produce blocks
        </p>

        <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
          {chainInfo?.authorities.map((addr, index) => (
            <div
              key={index}
              style={{
                padding: '0.75rem 1rem',
                backgroundColor: '#f7fafc',
                borderRadius: '4px',
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center'
              }}
            >
              <div>
                <div style={{ fontSize: '0.75rem', color: '#718096', marginBottom: '0.25rem' }}>
                  Authority #{index + 1}
                </div>
                <div className="hash" style={{ fontSize: '0.875rem' }}>
                  {addr}
                </div>
              </div>
              {addr === nodeInfo?.address && (
                <span style={{
                  backgroundColor: '#c6f6d5',
                  color: '#22543d',
                  padding: '0.25rem 0.75rem',
                  borderRadius: '12px',
                  fontSize: '0.75rem',
                  fontWeight: '500'
                }}>
                  This Node
                </span>
              )}
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
