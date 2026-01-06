# Transaction Endpoints

Endpoints for submitting and querying transactions.

## POST /transaction

Submit a new transaction to the blockchain.

### Request

```http
POST /api/v1/transaction
Content-Type: application/json
```

### Request Body

```json
{
  "transaction": {
    "from": "0x3D4b25CBdda1014F74F9C80f040ce1Bb69130CBB",
    "timestamp": 1704556800,
    "nonce": 0,
    "data": {
      "operations": [
        {
          "type": "SET",
          "key": "user:alice:name",
          "value": "QWxpY2U="
        },
        {
          "type": "SET",
          "key": "user:alice:email",
          "value": "YWxpY2VAZXhhbXBsZS5jb20="
        }
      ]
    },
    "signature": "0x1234567890abcdef..."
  }
}
```

### Transaction Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| from | string | Yes | Sender's address (0x-prefixed) |
| timestamp | integer | Yes | Unix timestamp |
| nonce | integer | Yes | Transaction count for this address |
| data.operations | array | Yes | List of SET/DELETE operations |
| signature | string | Yes | ECDSA signature of transaction |

### Response

```json
{
  "success": true,
  "data": {
    "hash": "0xtx123abc456def...",
    "message": "Transaction submitted successfully"
  }
}
```

### Example

```bash
curl -X POST http://localhost:8545/api/v1/transaction \
  -H "Content-Type: application/json" \
  -d '{
    "transaction": {
      "from": "0x3D4b25CBdda1014F74F9C80f040ce1Bb69130CBB",
      "timestamp": 1704556800,
      "nonce": 0,
      "data": {
        "operations": [
          {
            "type": "SET",
            "key": "mykey",
            "value": "bXl2YWx1ZQ=="
          }
        ]
      },
      "signature": "0x..."
    }
  }' | jq
```

### JavaScript Example

```javascript
async function submitTransaction(operations, privateKey) {
  // Build transaction
  const tx = {
    from: myAddress,
    timestamp: Math.floor(Date.now() / 1000),
    nonce: await getNonce(myAddress),
    data: { operations }
  }

  // Sign transaction
  const signature = await signTransaction(tx, privateKey)
  tx.signature = signature

  // Submit
  const response = await fetch('http://localhost:8545/api/v1/transaction', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ transaction: tx })
  })

  return await response.json()
}

// Usage
const result = await submitTransaction([
  {
    type: 'SET',
    key: 'user:alice:name',
    value: btoa('Alice')
  }
], privateKey)

console.log(`Transaction hash: ${result.data.hash}`)
```

### Error Responses

**400 Bad Request - Invalid Format**:
```json
{
  "success": false,
  "error": "Invalid transaction format"
}
```

**400 Bad Request - Invalid Signature**:
```json
{
  "success": false,
  "error": "Invalid transaction signature"
}
```

**400 Bad Request - Invalid Nonce**:
```json
{
  "success": false,
  "error": "Invalid nonce: expected 5, got 3"
}
```

---

## GET /transaction/{hash}

Get transaction details by hash.

### Request

```http
GET /api/v1/transaction/{hash}
```

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| hash | string | Yes | Transaction hash (0x-prefixed) |

### Response

```json
{
  "success": true,
  "data": {
    "hash": "0xtx123...",
    "from": "0x3D4b25CBdda1014F74F9C80f040ce1Bb69130CBB",
    "timestamp": 1704556800,
    "nonce": 0,
    "data": {
      "operations": [
        {
          "type": "SET",
          "key": "user:alice:name",
          "value": "QWxpY2U="
        }
      ]
    },
    "signature": "0x...",
    "block_height": 1234,
    "block_hash": "0xblock123..."
  }
}
```

### Example

```bash
curl http://localhost:8545/api/v1/transaction/0xtx123... | jq
```

```javascript
const txHash = '0xtx123...'
const response = await fetch(`http://localhost:8545/api/v1/transaction/${txHash}`)
const { data } = await response.json()
console.log(`Transaction in block: ${data.block_height}`)
```

### Error Responses

**404 Not Found**:
```json
{
  "success": false,
  "error": "Transaction not found"
}
```

---

## GET /mempool

Get all pending transactions in the mempool.

### Request

```http
GET /api/v1/mempool
```

### Response

```json
{
  "success": true,
  "data": {
    "count": 5,
    "transactions": [
      {
        "hash": "0xtx1...",
        "from": "0xAlice...",
        "timestamp": 1704556800,
        "nonce": 0,
        "data": {
          "operations": [
            {
              "type": "SET",
              "key": "test",
              "value": "dGVzdA=="
            }
          ]
        },
        "signature": "0x..."
      }
    ]
  }
}
```

### Example

```bash
curl http://localhost:8545/api/v1/mempool | jq
```

```javascript
const response = await fetch('http://localhost:8545/api/v1/mempool')
const { data } = await response.json()
console.log(`Pending transactions: ${data.count}`)
```

---

## Transaction Signing

### Signature Algorithm

Podoru Chain uses ECDSA with secp256k1 curve (Ethereum-compatible).

### Signing Process

1. **Build Transaction Object**:
```javascript
const tx = {
  from: "0x3D4b25CBdda1014F74F9C80f040ce1Bb69130CBB",
  timestamp: 1704556800,
  nonce: 0,
  data: {
    operations: [
      {
        type: "SET",
        key: "mykey",
        value: "bXl2YWx1ZQ=="
      }
    ]
  }
}
```

2. **Calculate Transaction Hash**:
```javascript
import { keccak256 } from 'ethers'

function hashTransaction(tx) {
  const txString = JSON.stringify({
    from: tx.from,
    timestamp: tx.timestamp,
    nonce: tx.nonce,
    data: tx.data
  })
  return keccak256(txString)
}

const txHash = hashTransaction(tx)
```

3. **Sign Hash**:
```javascript
import { ethers } from 'ethers'

const wallet = new ethers.Wallet(privateKey)
const signature = await wallet.signMessage(ethers.utils.arrayify(txHash))
```

4. **Add Signature to Transaction**:
```javascript
tx.signature = signature
```

### Complete Example

```javascript
import { ethers } from 'ethers'

async function createAndSignTransaction(privateKey, operations) {
  const wallet = new ethers.Wallet(privateKey)
  const address = await wallet.getAddress()

  // Build transaction
  const tx = {
    from: address,
    timestamp: Math.floor(Date.now() / 1000),
    nonce: await getNonce(address),
    data: { operations }
  }

  // Hash transaction
  const txString = JSON.stringify({
    from: tx.from,
    timestamp: tx.timestamp,
    nonce: tx.nonce,
    data: tx.data
  })
  const txHash = ethers.utils.keccak256(ethers.utils.toUtf8Bytes(txString))

  // Sign
  tx.signature = await wallet.signMessage(ethers.utils.arrayify(txHash))

  return tx
}

// Usage
const operations = [
  {
    type: 'SET',
    key: 'user:alice:name',
    value: btoa('Alice')
  }
]

const signedTx = await createAndSignTransaction(privateKey, operations)
console.log(signedTx)
```

---

## Operation Types

### SET Operation

Creates or updates a key-value pair.

```json
{
  "type": "SET",
  "key": "user:alice:name",
  "value": "QWxpY2U="
}
```

**Fields**:
- `type`: Must be "SET"
- `key`: String key (max 1KB)
- `value`: Base64-encoded value (max 1MB)

### DELETE Operation

Removes a key-value pair.

```json
{
  "type": "DELETE",
  "key": "user:alice:old_field"
}
```

**Fields**:
- `type`: Must be "DELETE"
- `key`: String key to delete

---

## Best Practices

### Nonce Management

```javascript
// Get current nonce
async function getNonce(address) {
  const response = await fetch(`http://localhost:8545/api/v1/account/${address}`)
  const { data } = await response.json()
  return data.nonce || 0
}

// Track nonce locally
let localNonce = await getNonce(myAddress)

async function submitTxWithNonce(operations) {
  const tx = await createAndSignTransaction(privateKey, operations)
  tx.nonce = localNonce

  const result = await submitTransaction(tx)

  if (result.success) {
    localNonce++
  }

  return result
}
```

### Batch Operations

```javascript
// Combine multiple operations in one transaction
const operations = [
  { type: 'SET', key: 'user:alice:name', value: btoa('Alice') },
  { type: 'SET', key: 'user:alice:email', value: btoa('alice@example.com') },
  { type: 'SET', key: 'user:alice:bio', value: btoa('Developer') },
  { type: 'DELETE', key: 'user:alice:temp' }
]

const tx = await createAndSignTransaction(privateKey, operations)
await submitTransaction(tx)
```

### Error Handling

```javascript
async function safeSubmitTransaction(tx) {
  try {
    const response = await fetch('http://localhost:8545/api/v1/transaction', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ transaction: tx })
    })

    const result = await response.json()

    if (!result.success) {
      throw new Error(result.error)
    }

    return result.data
  } catch (error) {
    console.error('Transaction failed:', error)
    throw error
  }
}
```

---

## Related Endpoints

- [GET /block/{hash}](blocks.md) - Get block containing transaction
- [GET /state/{key}](state.md) - Query state after transaction
- [GET /mempool](#get-mempool) - Check pending transactions
