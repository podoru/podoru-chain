// Utility functions for the explorer

export function formatHash(hash: string, length: number = 10): string {
  if (!hash) return ''
  if (hash.length <= length) return hash
  return `${hash.slice(0, length)}...${hash.slice(-4)}`
}

export function formatTimestamp(timestamp: number): string {
  return new Date(timestamp * 1000).toLocaleString()
}

export function formatTimeAgo(timestamp: number): string {
  const now = Date.now()
  const diff = now - (timestamp * 1000)

  const seconds = Math.floor(diff / 1000)
  const minutes = Math.floor(seconds / 60)
  const hours = Math.floor(minutes / 60)
  const days = Math.floor(hours / 24)

  if (days > 0) return `${days}d ago`
  if (hours > 0) return `${hours}h ago`
  if (minutes > 0) return `${minutes}m ago`
  return `${seconds}s ago`
}

export function formatAddress(address: string, length: number = 10): string {
  return formatHash(address, length)
}

export function isValidHash(value: string): boolean {
  return /^0x[a-fA-F0-9]{64}$/.test(value)
}

export function isValidAddress(value: string): boolean {
  return /^0x[a-fA-F0-9]{40}$/.test(value)
}

export function isNumeric(value: string): boolean {
  return /^\d+$/.test(value)
}
