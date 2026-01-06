import { createFileRoute } from '@tanstack/react-router'
import { useState } from 'react'
import { api, StateEntry } from '~/lib/api'

export const Route = createFileRoute('/state')({
  component: StateBrowserPage,
})

function StateBrowserPage() {
  const [prefix, setPrefix] = useState('')
  const [results, setResults] = useState<StateEntry[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  async function handleSearch(e: React.FormEvent) {
    e.preventDefault()
    if (!prefix.trim()) return

    setLoading(true)
    setError(null)

    try {
      const data = await api.queryByPrefix(prefix)
      setResults(data)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to query state')
    } finally {
      setLoading(false)
    }
  }

  async function loadAllState() {
    setPrefix('')
    setLoading(true)
    setError(null)

    try {
      const data = await api.queryByPrefix('')
      setResults(data)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load state')
    } finally {
      setLoading(false)
    }
  }

  function tryParseJSON(value: string): any {
    try {
      return JSON.parse(value)
    } catch {
      return value
    }
  }

  return (
    <div>
      <div className="card">
        <h2>🗄️ State Browser</h2>
        <p style={{ color: '#718096', marginBottom: '1.5rem' }}>
          Browse the key-value state stored on the blockchain
        </p>

        <form onSubmit={handleSearch} style={{ marginBottom: '1.5rem' }}>
          <div style={{ display: 'flex', gap: '1rem' }}>
            <input
              type="text"
              placeholder="Enter key prefix (e.g., 'user:', 'chain:', or leave empty for all)"
              value={prefix}
              onChange={(e) => setPrefix(e.target.value)}
            />
            <button type="submit" disabled={loading}>
              {loading ? 'Searching...' : 'Search'}
            </button>
            <button type="button" onClick={loadAllState} disabled={loading}>
              Load All
            </button>
          </div>
        </form>

        {error && (
          <div style={{
            padding: '1rem',
            backgroundColor: '#fed7d7',
            color: '#742a2a',
            borderRadius: '4px',
            marginBottom: '1rem'
          }}>
            {error}
          </div>
        )}

        {results.length === 0 && !loading && (
          <div style={{ padding: '3rem', textAlign: 'center', color: '#a0aec0' }}>
            {prefix ? 'No results found for this prefix' : 'Enter a prefix to search or click "Load All"'}
          </div>
        )}

        {results.length > 0 && (
          <>
            <div style={{ marginBottom: '1rem', color: '#4a5568', fontWeight: '500' }}>
              Found {results.length} entries
            </div>

            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
              {results.map((entry, index) => {
                const parsedValue = tryParseJSON(entry.value)
                const isJSON = typeof parsedValue === 'object'

                return (
                  <div
                    key={index}
                    style={{
                      padding: '1rem',
                      backgroundColor: '#f7fafc',
                      borderRadius: '4px',
                      borderLeft: '3px solid #4299e1'
                    }}
                  >
                    <div style={{
                      fontWeight: '600',
                      color: '#2d3748',
                      marginBottom: '0.5rem',
                      fontFamily: 'monospace',
                      fontSize: '0.875rem'
                    }}>
                      🔑 {entry.key}
                    </div>
                    <div style={{
                      color: '#4a5568',
                      fontFamily: 'monospace',
                      fontSize: '0.875rem',
                      whiteSpace: isJSON ? 'pre-wrap' : 'normal',
                      wordBreak: 'break-word'
                    }}>
                      {isJSON ? JSON.stringify(parsedValue, null, 2) : entry.value}
                    </div>
                  </div>
                )
              })}
            </div>
          </>
        )}
      </div>
    </div>
  )
}
