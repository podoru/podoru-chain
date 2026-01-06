# Node Endpoints

Endpoints for querying node information, health status, and peer connections.

## GET /node/info

Get detailed information about the node.

### Request

```http
GET /api/v1/node/info
```

### Response

```json
{
  "success": true,
  "data": {
    "version": "1.0.0",
    "node_type": "producer",
    "address": "0x3D4b25CBdda1014F74F9C80f040ce1Bb69130CBB",
    "p2p_port": 9000,
    "api_port": 8545,
    "peer_count": 3,
    "uptime": 86400,
    "is_syncing": false,
    "sync_progress": 100.0
  }
}
```

### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| version | string | Node software version |
| node_type | string | "producer" or "full" |
| address | string | Node's address (producer only) |
| p2p_port | integer | P2P network port |
| api_port | integer | API server port |
| peer_count | integer | Number of connected peers |
| uptime | integer | Uptime in seconds |
| is_syncing | boolean | Whether node is syncing |
| sync_progress | float | Sync progress percentage |

### Example

```bash
curl http://localhost:8545/api/v1/node/info | jq
```

```javascript
const response = await fetch('http://localhost:8545/api/v1/node/info')
const { data } = await response.json()

console.log(`Node type: ${data.node_type}`)
console.log(`Peers: ${data.peer_count}`)
console.log(`Uptime: ${Math.floor(data.uptime / 3600)} hours`)
```

```python
import requests

response = requests.get('http://localhost:8545/api/v1/node/info')
data = response.json()['data']

print(f"Node type: {data['node_type']}")
print(f"Peers: {data['peer_count']}")
print(f"Uptime: {data['uptime'] // 3600} hours")
```

### Use Cases

- **Monitoring**: Check node health and status
- **Discovery**: Identify node type and configuration
- **Debugging**: Verify node settings
- **Load Balancing**: Select nodes based on peer count

---

## GET /node/peers

Get list of connected peers.

### Request

```http
GET /api/v1/node/peers
```

### Response

```json
{
  "success": true,
  "data": {
    "peer_count": 3,
    "peers": [
      {
        "address": "192.168.1.10:9000",
        "node_type": "producer",
        "version": "1.0.0",
        "height": 1234,
        "last_seen": 1704556800,
        "latency_ms": 50
      },
      {
        "address": "192.168.1.11:9001",
        "node_type": "producer",
        "version": "1.0.0",
        "height": 1234,
        "last_seen": 1704556800,
        "latency_ms": 45
      },
      {
        "address": "192.168.1.12:9002",
        "node_type": "full",
        "version": "1.0.0",
        "height": 1234,
        "last_seen": 1704556800,
        "latency_ms": 60
      }
    ]
  }
}
```

### Peer Fields

| Field | Type | Description |
|-------|------|-------------|
| address | string | Peer's network address |
| node_type | string | "producer" or "full" |
| version | string | Peer's software version |
| height | integer | Peer's blockchain height |
| last_seen | integer | Last communication timestamp |
| latency_ms | integer | Network latency in milliseconds |

### Example

```bash
curl http://localhost:8545/api/v1/node/peers | jq
```

```javascript
async function getPeers() {
  const response = await fetch('http://localhost:8545/api/v1/node/peers')
  const { data } = await response.json()

  console.log(`Connected to ${data.peer_count} peers:`)

  data.peers.forEach(peer => {
    console.log(`  ${peer.address} (${peer.node_type}) - Height: ${peer.height}, Latency: ${peer.latency_ms}ms`)
  })
}

await getPeers()
```

### Use Cases

- **Network Health**: Monitor peer connections
- **Debugging**: Diagnose connectivity issues
- **Load Balancing**: Choose peers with low latency
- **Synchronization**: Verify all peers have same height

---

## GET /node/health

Health check endpoint for monitoring systems.

### Request

```http
GET /api/v1/node/health
```

### Response

```json
{
  "success": true,
  "data": {
    "status": "healthy",
    "checks": {
      "database": "ok",
      "network": "ok",
      "sync": "ok",
      "api": "ok"
    },
    "timestamp": 1704556800
  }
}
```

### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| status | string | "healthy", "degraded", or "unhealthy" |
| checks | object | Individual component health |
| timestamp | integer | Health check timestamp |

### Health Status

| Status | Description |
|--------|-------------|
| healthy | All systems operational |
| degraded | Some non-critical issues |
| unhealthy | Critical failures |

### Example

```bash
# Simple health check
curl http://localhost:8545/api/v1/node/health

# With monitoring
curl -f http://localhost:8545/api/v1/node/health || echo "Node is down"
```

```javascript
async function healthCheck() {
  try {
    const response = await fetch('http://localhost:8545/api/v1/node/health')
    const { data } = await response.json()

    if (data.status !== 'healthy') {
      console.error('Node is unhealthy:', data.checks)
      return false
    }

    console.log('Node is healthy')
    return true
  } catch (error) {
    console.error('Health check failed:', error)
    return false
  }
}

setInterval(healthCheck, 60000)  // Check every minute
```

### Use Cases

- **Load Balancers**: Health check for routing
- **Monitoring**: Alerting on failures
- **Docker**: HEALTHCHECK directive
- **Kubernetes**: Liveness/readiness probes

### Docker Health Check

```dockerfile
HEALTHCHECK --interval=30s --timeout=5s --retries=3 \
  CMD curl -f http://localhost:8545/api/v1/node/health || exit 1
```

### Kubernetes Probes

```yaml
livenessProbe:
  httpGet:
    path: /api/v1/node/health
    port: 8545
  initialDelaySeconds: 30
  periodSeconds: 10

readinessProbe:
  httpGet:
    path: /api/v1/node/health
    port: 8545
  initialDelaySeconds: 5
  periodSeconds: 5
```

---

## Monitoring Examples

### Check Node Synchronization

```javascript
async function checkSync() {
  const [nodeInfo, chainInfo] = await Promise.all([
    fetch('http://localhost:8545/api/v1/node/info').then(r => r.json()),
    fetch('http://localhost:8545/api/v1/chain/info').then(r => r.json())
  ])

  const isSyncing = nodeInfo.data.is_syncing
  const height = chainInfo.data.height

  console.log(`Height: ${height}, Syncing: ${isSyncing}`)

  return !isSyncing
}
```

### Monitor Network Health

```javascript
async function monitorNetwork() {
  const response = await fetch('http://localhost:8545/api/v1/node/peers')
  const { data } = await response.json()

  // Check peer count
  if (data.peer_count < 2) {
    console.warn('Low peer count:', data.peer_count)
  }

  // Check peer heights
  const heights = data.peers.map(p => p.height)
  const maxHeight = Math.max(...heights)
  const minHeight = Math.min(...heights)

  if (maxHeight - minHeight > 10) {
    console.warn('Peers out of sync:', { maxHeight, minHeight })
  }

  // Check latencies
  const avgLatency = data.peers.reduce((sum, p) => sum + p.latency_ms, 0) / data.peer_count

  if (avgLatency > 500) {
    console.warn('High network latency:', avgLatency)
  }

  return {
    peerCount: data.peer_count,
    heightDelta: maxHeight - minHeight,
    avgLatency
  }
}

// Run every minute
setInterval(monitorNetwork, 60000)
```

### Dashboard Example

```javascript
async function getDashboardData() {
  const [nodeInfo, chainInfo, peers, health] = await Promise.all([
    fetch('http://localhost:8545/api/v1/node/info').then(r => r.json()),
    fetch('http://localhost:8545/api/v1/chain/info').then(r => r.json()),
    fetch('http://localhost:8545/api/v1/node/peers').then(r => r.json()),
    fetch('http://localhost:8545/api/v1/node/health').then(r => r.json())
  ])

  return {
    node: {
      version: nodeInfo.data.version,
      type: nodeInfo.data.node_type,
      uptime: nodeInfo.data.uptime,
      status: health.data.status
    },
    blockchain: {
      height: chainInfo.data.height,
      latestBlock: chainInfo.data.latest_block_hash,
      authorities: chainInfo.data.authorities.length
    },
    network: {
      peers: peers.data.peer_count,
      avgLatency: peers.data.peers.reduce((sum, p) => sum + p.latency_ms, 0) / peers.data.peer_count
    }
  }
}

// Update dashboard every 5 seconds
setInterval(async () => {
  const data = await getDashboardData()
  updateDashboard(data)
}, 5000)
```

---

## Prometheus Metrics (Future)

Future releases will support Prometheus metrics:

```
GET /metrics

# HELP podoru_block_height Current blockchain height
# TYPE podoru_block_height gauge
podoru_block_height 1234

# HELP podoru_peer_count Number of connected peers
# TYPE podoru_peer_count gauge
podoru_peer_count 3

# HELP podoru_uptime_seconds Node uptime in seconds
# TYPE podoru_uptime_seconds counter
podoru_uptime_seconds 86400
```

---

## Error Responses

**500 Internal Server Error**:
```json
{
  "success": false,
  "error": "Failed to retrieve node information"
}
```

---

## Related Endpoints

- [GET /chain/info](chain.md) - Get blockchain info
- [GET /block/latest](blocks.md) - Get latest block
- [GET /mempool](transactions.md#get-mempool) - Get pending transactions
