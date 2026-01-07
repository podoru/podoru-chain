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

// PDR token constants
const PDR_DECIMALS = 18n
const ONE_PDR = 10n ** PDR_DECIMALS

// Format wei amount to PDR with proper decimals
export function formatPDR(weiAmount: string | bigint, decimals: number = 6): string {
  try {
    const wei = typeof weiAmount === 'string' ? BigInt(weiAmount) : weiAmount

    if (wei === 0n) return '0 PDR'

    // Calculate whole and fractional parts
    const whole = wei / ONE_PDR
    const fraction = wei % ONE_PDR

    // Format fractional part with leading zeros
    let fractionStr = fraction.toString().padStart(18, '0')
    // Trim to requested decimals
    fractionStr = fractionStr.slice(0, decimals)
    // Remove trailing zeros
    fractionStr = fractionStr.replace(/0+$/, '')

    if (fractionStr) {
      return `${whole}.${fractionStr} PDR`
    }
    return `${whole} PDR`
  } catch {
    return `${weiAmount} wei`
  }
}

// Format balance with short notation (e.g., "33.33M PDR")
export function formatBalanceShort(weiAmount: string | bigint): string {
  try {
    const wei = typeof weiAmount === 'string' ? BigInt(weiAmount) : weiAmount

    if (wei === 0n) return '0 PDR'

    // Convert to float for display
    const pdr = Number(wei) / Number(ONE_PDR)

    if (pdr >= 1_000_000_000) {
      return `${(pdr / 1_000_000_000).toFixed(2)}B PDR`
    }
    if (pdr >= 1_000_000) {
      return `${(pdr / 1_000_000).toFixed(2)}M PDR`
    }
    if (pdr >= 1_000) {
      return `${(pdr / 1_000).toFixed(2)}K PDR`
    }
    if (pdr >= 1) {
      return `${pdr.toFixed(2)} PDR`
    }
    return `${pdr.toFixed(6)} PDR`
  } catch {
    return `${weiAmount} wei`
  }
}

// Format gas fee (wei) to readable string
export function formatGasFee(weiAmount: string | bigint): string {
  try {
    const wei = typeof weiAmount === 'string' ? BigInt(weiAmount) : weiAmount

    if (wei === 0n) return '0 wei'

    // If less than 1 PDR, show in wei
    if (wei < ONE_PDR / 1000n) {
      return `${wei.toString()} wei`
    }

    return formatPDR(wei, 8)
  } catch {
    return `${weiAmount} wei`
  }
}

// Parse PDR amount to wei
export function parsePDR(pdrAmount: string): bigint {
  try {
    const amount = parseFloat(pdrAmount)
    if (isNaN(amount)) return 0n

    // Multiply by 10^18
    const wei = BigInt(Math.floor(amount * 1e18))
    return wei
  } catch {
    return 0n
  }
}
