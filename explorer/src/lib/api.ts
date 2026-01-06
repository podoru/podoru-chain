// API client for Podoru Chain REST API

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8545'

export interface ChainInfo {
  height: number
  current_hash: string
  genesis_hash: string
  authorities: string[]
}

export interface Block {
  header: {
    version: number
    height: number
    previous_hash: string
    timestamp: number
    merkle_root: string
    state_root: string
    producer_addr: string
    nonce: number
  }
  transactions: Transaction[]
  signature: string
}

export interface Transaction {
  id: string
  from: string
  timestamp: number
  data: {
    operations: Operation[]
  }
  signature: string
  nonce: number
}

export interface Operation {
  type: 'SET' | 'DELETE'
  key: string
  value?: string
}

export interface NodeInfo {
  version: string
  node_type: string
  address: string
  p2p_address: string
}

export interface Peer {
  id: string
  address: string
  connected_at: number
}

export interface MempoolInfo {
  transaction_count: number
  transactions: Transaction[]
}

export interface StateEntry {
  key: string
  value: string
}

class APIClient {
  private baseURL: string

  constructor(baseURL: string = API_BASE_URL) {
    this.baseURL = baseURL
  }

  private async fetch<T>(path: string, options?: RequestInit): Promise<T> {
    const response = await fetch(`${this.baseURL}${path}`, {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        ...options?.headers,
      },
    })

    if (!response.ok) {
      const error = await response.text()
      throw new Error(`API Error: ${response.status} - ${error}`)
    }

    return response.json()
  }

  // Chain endpoints
  async getChainInfo(): Promise<ChainInfo> {
    return this.fetch<ChainInfo>('/api/v1/chain/info')
  }

  async getBlockByHash(hash: string): Promise<Block> {
    return this.fetch<Block>(`/api/v1/block/${hash}`)
  }

  async getBlockByHeight(height: number): Promise<Block> {
    return this.fetch<Block>(`/api/v1/block/height/${height}`)
  }

  async getLatestBlock(): Promise<Block> {
    return this.fetch<Block>('/api/v1/block/latest')
  }

  // Transaction endpoints
  async getTransaction(hash: string): Promise<Transaction> {
    return this.fetch<Transaction>(`/api/v1/transaction/${hash}`)
  }

  async submitTransaction(tx: Transaction): Promise<{ success: boolean }> {
    return this.fetch('/api/v1/transaction', {
      method: 'POST',
      body: JSON.stringify(tx),
    })
  }

  // State endpoints
  async getState(key: string): Promise<{ value: string }> {
    return this.fetch<{ value: string }>(`/api/v1/state/${encodeURIComponent(key)}`)
  }

  async batchGetState(keys: string[]): Promise<{ [key: string]: string }> {
    return this.fetch('/api/v1/state/batch', {
      method: 'POST',
      body: JSON.stringify({ keys }),
    })
  }

  async queryByPrefix(prefix: string): Promise<StateEntry[]> {
    return this.fetch('/api/v1/state/query/prefix', {
      method: 'POST',
      body: JSON.stringify({ prefix }),
    })
  }

  // Node endpoints
  async getNodeInfo(): Promise<NodeInfo> {
    return this.fetch<NodeInfo>('/api/v1/node/info')
  }

  async getPeers(): Promise<Peer[]> {
    return this.fetch<Peer[]>('/api/v1/node/peers')
  }

  async getHealth(): Promise<{ status: string }> {
    return this.fetch<{ status: string }>('/api/v1/node/health')
  }

  // Mempool endpoints
  async getMempool(): Promise<MempoolInfo> {
    return this.fetch<MempoolInfo>('/api/v1/mempool')
  }
}

export const api = new APIClient()
