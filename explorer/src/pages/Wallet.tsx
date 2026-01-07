import { useEffect, useState } from 'react'
import { api, BalanceInfo, TokenInfo, GasConfig, GasEstimate } from '~/lib/api'
import { formatBalanceShort, formatGasFee, isValidAddress } from '~/lib/utils'

export default function Wallet() {
  const [tokenInfo, setTokenInfo] = useState<TokenInfo | null>(null)
  const [gasConfig, setGasConfig] = useState<GasConfig | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  // Balance lookup
  const [address, setAddress] = useState('')
  const [balance, setBalance] = useState<BalanceInfo | null>(null)
  const [balanceLoading, setBalanceLoading] = useState(false)
  const [balanceError, setBalanceError] = useState<string | null>(null)

  // Gas estimator
  const [txSize, setTxSize] = useState(500)
  const [gasEstimate, setGasEstimate] = useState<GasEstimate | null>(null)

  useEffect(() => {
    loadConfig()
  }, [])

  useEffect(() => {
    if (gasConfig?.enabled && txSize > 0) {
      estimateGas()
    }
  }, [txSize, gasConfig])

  async function loadConfig() {
    try {
      const [token, gas] = await Promise.all([
        api.getTokenInfo(),
        api.getGasConfig(),
      ])
      setTokenInfo(token)
      setGasConfig(gas)
      setError(null)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load config')
    } finally {
      setLoading(false)
    }
  }

  async function lookupBalance() {
    if (!isValidAddress(address)) {
      setBalanceError('Invalid address format')
      return
    }

    setBalanceLoading(true)
    setBalanceError(null)

    try {
      const result = await api.getBalance(address)
      setBalance(result)
    } catch (err) {
      setBalanceError(err instanceof Error ? err.message : 'Failed to get balance')
      setBalance(null)
    } finally {
      setBalanceLoading(false)
    }
  }

  async function estimateGas() {
    try {
      const estimate = await api.estimateGas(txSize)
      setGasEstimate(estimate)
    } catch {
      setGasEstimate(null)
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
        <button onClick={loadConfig} style={{ marginTop: '1rem' }}>
          Retry
        </button>
      </div>
    )
  }

  return (
    <div>
      {/* Token Info Card */}
      <div className="card">
        <h2>Token Information</h2>
        <div className="grid" style={{ gridTemplateColumns: 'repeat(auto-fit, minmax(150px, 1fr))', gap: '1rem' }}>
          <div>
            <div style={{ fontSize: '0.875rem', color: '#718096', marginBottom: '0.25rem' }}>
              Name
            </div>
            <div style={{ fontSize: '1.25rem', fontWeight: '600' }}>
              {tokenInfo?.name}
            </div>
          </div>
          <div>
            <div style={{ fontSize: '0.875rem', color: '#718096', marginBottom: '0.25rem' }}>
              Symbol
            </div>
            <div style={{ fontSize: '1.25rem', fontWeight: '600' }}>
              {tokenInfo?.symbol}
            </div>
          </div>
          <div>
            <div style={{ fontSize: '0.875rem', color: '#718096', marginBottom: '0.25rem' }}>
              Decimals
            </div>
            <div style={{ fontSize: '1.25rem', fontWeight: '600' }}>
              {tokenInfo?.decimals}
            </div>
          </div>
          {tokenInfo?.total_supply && (
            <div>
              <div style={{ fontSize: '0.875rem', color: '#718096', marginBottom: '0.25rem' }}>
                Initial Supply
              </div>
              <div style={{ fontSize: '1.25rem', fontWeight: '600' }}>
                {formatBalanceShort(tokenInfo.total_supply)}
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Balance Lookup Card */}
      <div className="card">
        <h2>Balance Lookup</h2>
        <div style={{ display: 'flex', gap: '0.5rem', marginBottom: '1rem' }}>
          <input
            type="text"
            value={address}
            onChange={(e) => setAddress(e.target.value)}
            placeholder="Enter wallet address (0x...)"
            style={{
              flex: 1,
              padding: '0.75rem',
              fontSize: '1rem',
              fontFamily: 'monospace',
              border: '1px solid #e2e8f0',
              borderRadius: '4px',
            }}
            onKeyPress={(e) => e.key === 'Enter' && lookupBalance()}
          />
          <button
            onClick={lookupBalance}
            disabled={balanceLoading}
            style={{
              padding: '0.75rem 1.5rem',
              backgroundColor: '#5a67d8',
              color: 'white',
              border: 'none',
              borderRadius: '4px',
              cursor: balanceLoading ? 'not-allowed' : 'pointer',
              opacity: balanceLoading ? 0.7 : 1,
            }}
          >
            {balanceLoading ? 'Loading...' : 'Lookup'}
          </button>
        </div>

        {balanceError && (
          <div style={{ color: '#e53e3e', marginBottom: '1rem' }}>
            {balanceError}
          </div>
        )}

        {balance && (
          <div style={{
            backgroundColor: '#f7fafc',
            padding: '1.5rem',
            borderRadius: '8px',
            border: '1px solid #e2e8f0',
          }}>
            <div style={{ fontSize: '0.875rem', color: '#718096', marginBottom: '0.5rem' }}>
              Address
            </div>
            <div className="hash" style={{ marginBottom: '1rem' }}>
              {balance.address}
            </div>
            <div style={{ fontSize: '0.875rem', color: '#718096', marginBottom: '0.5rem' }}>
              Balance
            </div>
            <div style={{ fontSize: '2rem', fontWeight: '700', color: '#2d3748' }}>
              {balance.balance_formatted}
            </div>
            <div style={{ fontSize: '0.875rem', color: '#718096', marginTop: '0.5rem' }}>
              {balance.balance} wei
            </div>
          </div>
        )}
      </div>

      {/* Gas Configuration Card */}
      <div className="card">
        <h2>Gas Configuration</h2>
        <div style={{
          display: 'flex',
          alignItems: 'center',
          gap: '0.5rem',
          marginBottom: '1rem',
        }}>
          <span className={`status ${gasConfig?.enabled ? 'connected' : 'disconnected'}`}>
            {gasConfig?.enabled ? '● Gas Fees Enabled' : '○ Gas Fees Disabled'}
          </span>
        </div>

        {gasConfig?.enabled && (
          <>
            <div className="grid" style={{ gridTemplateColumns: 'repeat(auto-fit, minmax(150px, 1fr))', gap: '1rem', marginBottom: '1.5rem' }}>
              <div>
                <div style={{ fontSize: '0.875rem', color: '#718096', marginBottom: '0.25rem' }}>
                  Base Fee
                </div>
                <div style={{ fontSize: '1.25rem', fontWeight: '600' }}>
                  {formatGasFee(gasConfig.base_fee)}
                </div>
              </div>
              <div>
                <div style={{ fontSize: '0.875rem', color: '#718096', marginBottom: '0.25rem' }}>
                  Per-Byte Fee
                </div>
                <div style={{ fontSize: '1.25rem', fontWeight: '600' }}>
                  {formatGasFee(gasConfig.per_byte_fee)}
                </div>
              </div>
            </div>

            {/* Gas Estimator */}
            <div style={{
              backgroundColor: '#f7fafc',
              padding: '1.5rem',
              borderRadius: '8px',
              border: '1px solid #e2e8f0',
            }}>
              <h3 style={{ marginTop: 0, marginBottom: '1rem' }}>Gas Fee Estimator</h3>
              <div style={{ marginBottom: '1rem' }}>
                <label style={{ display: 'block', marginBottom: '0.5rem', color: '#718096' }}>
                  Transaction Size (bytes)
                </label>
                <input
                  type="range"
                  min="100"
                  max="10000"
                  value={txSize}
                  onChange={(e) => setTxSize(parseInt(e.target.value))}
                  style={{ width: '100%' }}
                />
                <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.875rem', color: '#718096' }}>
                  <span>100 bytes</span>
                  <span style={{ fontWeight: '600', color: '#2d3748' }}>{txSize} bytes</span>
                  <span>10,000 bytes</span>
                </div>
              </div>

              {gasEstimate && (
                <div style={{ marginTop: '1rem' }}>
                  <table style={{ width: '100%' }}>
                    <tbody>
                      <tr>
                        <td style={{ padding: '0.5rem 0', color: '#718096' }}>Base Fee</td>
                        <td style={{ textAlign: 'right', fontWeight: '500' }}>{formatGasFee(gasEstimate.base_fee)}</td>
                      </tr>
                      <tr>
                        <td style={{ padding: '0.5rem 0', color: '#718096' }}>Size Fee ({txSize} bytes)</td>
                        <td style={{ textAlign: 'right', fontWeight: '500' }}>{formatGasFee(gasEstimate.size_fee)}</td>
                      </tr>
                      <tr style={{ borderTop: '1px solid #e2e8f0' }}>
                        <td style={{ padding: '0.75rem 0', fontWeight: '600' }}>Total Fee</td>
                        <td style={{ textAlign: 'right', fontWeight: '700', fontSize: '1.25rem', color: '#5a67d8' }}>
                          {gasEstimate.total_fee_formatted}
                        </td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              )}
            </div>
          </>
        )}
      </div>
    </div>
  )
}
