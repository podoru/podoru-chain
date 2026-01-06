# API Reference

Podoru Chain provides a comprehensive REST API for interacting with the blockchain.

## Base URL

```
http://localhost:8545/api/v1
```

Replace `localhost:8545` with your node's address and port.

## API Endpoints Overview

### Chain Endpoints

Get blockchain information and metadata.

- `GET /chain/info` - Get blockchain summary
- `GET /block/latest` - Get latest block
- `GET /block/{hash}` - Get block by hash
- `GET /block/height/{height}` - Get block by height

[View Chain Endpoints](chain.md)

### Block Endpoints

Query and retrieve block data.

- `GET /block/{hash}` - Get block by hash
- `GET /block/height/{height}` - Get block by height
- `GET /block/latest` - Get latest block

[View Block Endpoints](blocks.md)

### Transaction Endpoints

Submit and query transactions.

- `POST /transaction` - Submit new transaction
- `GET /transaction/{hash}` - Get transaction by hash
- `GET /mempool` - Get pending transactions

[View Transaction Endpoints](transactions.md)

### State Endpoints

Query blockchain state (key-value data).

- `GET /state/{key}` - Get single value
- `POST /state/batch` - Get multiple values
- `POST /state/query/prefix` - Query keys by prefix

[View State Endpoints](state.md)

### Node Endpoints

Get node information and health status.

- `GET /node/info` - Get node information
- `GET /node/peers` - Get connected peers
- `GET /node/health` - Health check

[View Node Endpoints](node.md)

## Request Format

All requests use standard HTTP methods:
- `GET` for queries
- `POST` for submissions and batch operations

### Headers

```http
Content-Type: application/json
Accept: application/json
```

### Example Request

```bash
curl -X GET http://localhost:8545/api/v1/chain/info \
  -H "Content-Type: application/json"
```

## Response Format

All API responses follow a consistent format:

### Success Response

```json
{
  "success": true,
  "data": {
    // Response data
  }
}
```

### Error Response

```json
{
  "success": false,
  "error": "Error message description"
}
```

## HTTP Status Codes

| Status Code | Meaning |
|-------------|---------|
| 200 | Success |
| 400 | Bad Request (invalid parameters) |
| 404 | Not Found (resource doesn't exist) |
| 500 | Internal Server Error |

## Data Encoding

### Base64 Encoding

All binary data is base64-encoded in JSON responses:

```json
{
  "key": "user:alice:name",
  "value": "QWxpY2U="  // base64("Alice")
}
```

**Encoding/Decoding**:

```javascript
// Encode
const encoded = btoa("Alice")  // "QWxpY2U="

// Decode
const decoded = atob("QWxpY2U=")  // "Alice"
```

```python
# Encode
import base64
encoded = base64.b64encode(b"Alice").decode()  # "QWxpY2U="

# Decode
decoded = base64.b64decode("QWxpY2U=").decode()  # "Alice"
```

### Timestamps

All timestamps are Unix timestamps (seconds since epoch):

```json
{
  "timestamp": 1704556800  // 2024-01-06 16:00:00 UTC
}
```

### Addresses

Ethereum-compatible addresses (42 characters, 0x-prefixed):

```
0x3D4b25CBdda1014F74F9C80f040ce1Bb69130CBB
```

## Authentication

Currently, the API does not require authentication for read operations.

**Write operations** (submitting transactions) require:
- Valid transaction signature
- Correct nonce
- Proper formatting

## Rate Limiting

No rate limiting is currently enforced. For production deployments, consider:
- Using a reverse proxy (nginx, caddy)
- Implementing rate limiting at the proxy level
- Monitoring API usage

## CORS

CORS is enabled by default for all origins in development. For production:

```go
// Configure CORS in your reverse proxy
Access-Control-Allow-Origin: https://yourdomain.com
```

## Pagination

Currently, endpoints return all results. Future versions will support:

```
GET /state/query/prefix?limit=100&offset=0
```

## Versioning

API version is included in the URL path: `/api/v1/...`

Breaking changes will result in a new version: `/api/v2/...`

## Client Libraries

### JavaScript/TypeScript

```javascript
class PodoruClient {
  constructor(baseURL) {
    this.baseURL = baseURL || 'http://localhost:8545/api/v1'
  }

  async get(endpoint) {
    const response = await fetch(`${this.baseURL}${endpoint}`)
    return await response.json()
  }

  async post(endpoint, data) {
    const response = await fetch(`${this.baseURL}${endpoint}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data)
    })
    return await response.json()
  }

  // Convenience methods
  async getChainInfo() {
    return this.get('/chain/info')
  }

  async getState(key) {
    return this.get(`/state/${key}`)
  }

  async submitTransaction(tx) {
    return this.post('/transaction', { transaction: tx })
  }
}

// Usage
const client = new PodoruClient()
const info = await client.getChainInfo()
```

### Python

```python
import requests
import base64

class PodoruClient:
    def __init__(self, base_url='http://localhost:8545/api/v1'):
        self.base_url = base_url

    def get(self, endpoint):
        response = requests.get(f'{self.base_url}{endpoint}')
        return response.json()

    def post(self, endpoint, data):
        response = requests.post(
            f'{self.base_url}{endpoint}',
            json=data
        )
        return response.json()

    # Convenience methods
    def get_chain_info(self):
        return self.get('/chain/info')

    def get_state(self, key):
        return self.get(f'/state/{key}')

    def submit_transaction(self, tx):
        return self.post('/transaction', {'transaction': tx})

# Usage
client = PodoruClient()
info = client.get_chain_info()
```

### Go

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
)

type PodoruClient struct {
    BaseURL string
}

func NewClient(baseURL string) *PodoruClient {
    if baseURL == "" {
        baseURL = "http://localhost:8545/api/v1"
    }
    return &PodoruClient{BaseURL: baseURL}
}

func (c *PodoruClient) Get(endpoint string) (map[string]interface{}, error) {
    resp, err := http.Get(c.BaseURL + endpoint)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&result)
    return result, nil
}

func (c *PodoruClient) Post(endpoint string, data interface{}) (map[string]interface{}, error) {
    jsonData, _ := json.Marshal(data)
    resp, err := http.Post(
        c.BaseURL+endpoint,
        "application/json",
        bytes.NewBuffer(jsonData),
    )
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&result)
    return result, nil
}

// Usage
client := NewClient("")
info, _ := client.Get("/chain/info")
fmt.Println(info)
```

## WebSocket Support

WebSocket support for real-time updates is planned for a future release:

```javascript
// Future feature
const ws = new WebSocket('ws://localhost:8545/api/v1/ws')

ws.on('block', (block) => {
  console.log('New block:', block)
})

ws.on('transaction', (tx) => {
  console.log('New transaction:', tx)
})
```

## Testing the API

### Using curl

```bash
# Get chain info
curl http://localhost:8545/api/v1/chain/info | jq

# Get state
curl http://localhost:8545/api/v1/state/chain:name | jq

# Submit transaction
curl -X POST http://localhost:8545/api/v1/transaction \
  -H "Content-Type: application/json" \
  -d @transaction.json | jq
```

### Using httpie

```bash
# Get chain info
http GET localhost:8545/api/v1/chain/info

# Get state
http GET localhost:8545/api/v1/state/chain:name

# Submit transaction
http POST localhost:8545/api/v1/transaction < transaction.json
```

### Using Postman

Import the Podoru Chain Postman collection (coming soon).

## Error Handling

### Common Errors

**404 Not Found**:
```json
{
  "success": false,
  "error": "Block not found"
}
```

**400 Bad Request**:
```json
{
  "success": false,
  "error": "Invalid transaction format"
}
```

**500 Internal Server Error**:
```json
{
  "success": false,
  "error": "Internal server error"
}
```

### Best Practices

```javascript
async function safeAPICall(endpoint) {
  try {
    const response = await fetch(endpoint)
    const data = await response.json()

    if (!data.success) {
      throw new Error(data.error)
    }

    return data.data
  } catch (error) {
    console.error('API Error:', error)
    throw error
  }
}
```

## Further Reading

- [Chain Endpoints](chain.md) - Blockchain information
- [Block Endpoints](blocks.md) - Block data
- [Transaction Endpoints](transactions.md) - Transaction operations
- [State Endpoints](state.md) - State queries
- [Node Endpoints](node.md) - Node information
- [Development Guide](../development/README.md) - Building applications
