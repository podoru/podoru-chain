# Block Endpoints

Endpoints for querying block data from the blockchain.

## GET /block/latest

Get the most recent block in the blockchain.

### Request

```http
GET /api/v1/block/latest
```

### Response

```json
{
  "success": true,
  "data": {
    "height": 1234,
    "hash": "0xabc123def456...",
    "timestamp": 1704556800,
    "previous_hash": "0xdef456abc123...",
    "producer": "0x3D4b25CBdda1014F74F9C80f040ce1Bb69130CBB",
    "transaction_root": "0x789abc...",
    "state_root": "0x456def...",
    "transactions": [
      {
        "hash": "0xtx123...",
        "from": "0xAlice...",
        "timestamp": 1704556800,
        "nonce": 5,
        "data": {
          "operations": [
            {
              "type": "SET",
              "key": "user:alice:name",
              "value": "QWxpY2U="
            }
          ]
        },
        "signature": "0xsig..."
      }
    ],
    "signature": "0xblock_sig..."
  }
}
```

### Example

```bash
curl http://localhost:8545/api/v1/block/latest | jq
```

---

## GET /block/{hash}

Get a specific block by its hash.

### Request

```http
GET /api/v1/block/{hash}
```

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| hash | string | Yes | Block hash (0x-prefixed hex string) |

### Response

Same structure as `/block/latest`.

### Example

```bash
curl http://localhost:8545/api/v1/block/0xabc123def456... | jq
```

```javascript
const blockHash = '0xabc123def456...'
const response = await fetch(`http://localhost:8545/api/v1/block/${blockHash}`)
const { data } = await response.json()
console.log(`Block height: ${data.height}`)
```

### Error Responses

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
  "error": "Invalid block hash format"
}
```

---

## GET /block/height/{height}

Get a specific block by its height (block number).

### Request

```http
GET /api/v1/block/height/{height}
```

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| height | integer | Yes | Block height (0 for genesis) |

### Response

Same structure as `/block/latest`.

### Example

```bash
# Get genesis block
curl http://localhost:8545/api/v1/block/height/0 | jq

# Get block at height 100
curl http://localhost:8545/api/v1/block/height/100 | jq
```

```python
import requests

height = 100
response = requests.get(f'http://localhost:8545/api/v1/block/height/{height}')
block = response.json()['data']
print(f"Block {height} hash: {block['hash']}")
```

### Error Responses

**404 Not Found**:
```json
{
  "success": false,
  "error": "Block not found at height 1000000"
}
```

**400 Bad Request**:
```json
{
  "success": false,
  "error": "Invalid height format"
}
```

---

## Block Structure Reference

### Block Fields

| Field | Type | Description |
|-------|------|-------------|
| height | integer | Block number (0 for genesis) |
| hash | string | Unique block identifier |
| timestamp | integer | Unix timestamp when block was created |
| previous_hash | string | Hash of parent block |
| producer | string | Address of block producer |
| transaction_root | string | Merkle root of transactions |
| state_root | string | Merkle root of state |
| transactions | array | List of transactions in block |
| signature | string | Producer's signature |

### Transaction Fields

| Field | Type | Description |
|-------|------|-------------|
| hash | string | Transaction hash |
| from | string | Sender address |
| timestamp | integer | Unix timestamp |
| nonce | integer | Sender's transaction count |
| data | object | Transaction data (operations) |
| signature | string | Sender's signature |

### Operation Types

**SET Operation**:
```json
{
  "type": "SET",
  "key": "user:alice:name",
  "value": "QWxpY2U="
}
```

**DELETE Operation**:
```json
{
  "type": "DELETE",
  "key": "user:alice:old_field"
}
```

---

## Use Cases

### Monitor Blockchain Progress

```javascript
async function watchBlocks() {
  let lastHeight = 0

  setInterval(async () => {
    const response = await fetch('http://localhost:8545/api/v1/block/latest')
    const { data } = await response.json()

    if (data.height > lastHeight) {
      console.log(`New block: ${data.height}`)
      lastHeight = data.height
    }
  }, 5000)
}
```

### Retrieve Block Range

```javascript
async function getBlockRange(start, end) {
  const blocks = []

  for (let height = start; height <= end; height++) {
    const response = await fetch(`http://localhost:8545/api/v1/block/height/${height}`)
    const { data } = await response.json()
    blocks.push(data)
  }

  return blocks
}

// Get blocks 100-110
const blocks = await getBlockRange(100, 110)
```

### Verify Block Chain

```javascript
async function verifyBlockchain(startHeight, endHeight) {
  let previousHash = null

  for (let h = startHeight; h <= endHeight; h++) {
    const response = await fetch(`http://localhost:8545/api/v1/block/height/${h}`)
    const { data: block } = await response.json()

    if (previousHash && block.previous_hash !== previousHash) {
      console.error(`Chain broken at block ${h}`)
      return false
    }

    previousHash = block.hash
  }

  console.log('Blockchain verified')
  return true
}
```

### Extract All Transactions

```javascript
async function getAllTransactions(blockHeight) {
  const response = await fetch(`http://localhost:8545/api/v1/block/height/${blockHeight}`)
  const { data: block } = await response.json()

  return block.transactions.map(tx => ({
    hash: tx.hash,
    from: tx.from,
    timestamp: tx.timestamp,
    operations: tx.data.operations
  }))
}
```

---

## Related Endpoints

- [GET /chain/info](chain.md) - Get blockchain summary
- [GET /transaction/{hash}](transactions.md) - Get transaction details
- [GET /state/{key}](state.md) - Query state data
