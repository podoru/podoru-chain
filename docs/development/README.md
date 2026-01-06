# Development Guide

Learn how to build decentralized applications on Podoru Chain using its flexible key-value storage system.

## Overview

Podoru Chain is designed for application development with:
- **Flexible key-value storage** for any data type
- **Fast batch and prefix queries** for efficient data retrieval
- **REST API** for easy integration
- **No smart contracts needed** - simpler development
- **Fully decentralized** data storage

## What Can You Build?

Podoru Chain is perfect for:

### Social Media Platforms
- Twitter/X clones
- Reddit-style forums
- Blogging platforms
- Social networks

### File Storage Systems
- IPFS metadata storage
- Decentralized file indexes
- Content delivery networks

### User Systems
- Authentication services
- Profile management
- Identity platforms

### E-commerce
- Product catalogs
- Order tracking
- Inventory management

### Gaming
- Leaderboards
- Achievements
- Game state storage

### IoT
- Sensor data logging
- Device registries
- Telemetry storage

### Any Application
Requiring decentralized key-value storage!

## Key Concepts

### 1. Key-Value Storage

Store data as simple key-value pairs:

```javascript
// Store
key: "user:alice:name"
value: "Alice"

// Retrieve
GET /api/v1/state/user:alice:name
→ "Alice"
```

### 2. Hierarchical Keys

Use structured keys with colons:

```
pattern: "namespace:entity:id:field"

Examples:
user:alice:name          → "Alice"
user:alice:email         → "alice@example.com"
post:12345:content       → "Hello World"
product:999:price        → "99.99"
```

### 3. Query Methods

Three ways to query data:

**Single Key**:
```bash
GET /api/v1/state/user:alice:name
```

**Batch Query** (multiple keys):
```bash
POST /api/v1/state/batch
{"keys": ["user:alice:name", "user:alice:email"]}
```

**Prefix Query** (all matching keys):
```bash
POST /api/v1/state/query/prefix
{"prefix": "user:alice:", "limit": 100}
```

### 4. Transactions

Submit data changes via transactions:

```javascript
{
  "transaction": {
    "from": "0xYourAddress",
    "timestamp": 1704556800,
    "nonce": 0,
    "data": {
      "operations": [
        {
          "type": "SET",
          "key": "user:alice:name",
          "value": "QWxpY2U="  // base64("Alice")
        }
      ]
    },
    "signature": "0x..."
  }
}
```

## Quick Start

### 1. Connect to Node

```javascript
const API_URL = 'http://localhost:8545/api/v1'

async function get(endpoint) {
  const response = await fetch(`${API_URL}${endpoint}`)
  return await response.json()
}

async function post(endpoint, data) {
  const response = await fetch(`${API_URL}${endpoint}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data)
  })
  return await response.json()
}
```

### 2. Read Data

```javascript
// Get single value
const result = await get('/state/chain:name')
console.log(result.data.value)  // "Podoru Chain"

// Get multiple values
const batch = await post('/state/batch', {
  keys: ['chain:name', 'chain:version', 'chain:description']
})
console.log(batch.data)

// Query by prefix
const userData = await post('/state/query/prefix', {
  prefix: 'user:alice:',
  limit: 100
})
console.log(userData.data.results)
```

### 3. Write Data

```javascript
import { ethers } from 'ethers'

// Create wallet
const wallet = new ethers.Wallet(privateKey)

// Build transaction
const tx = {
  from: await wallet.getAddress(),
  timestamp: Math.floor(Date.now() / 1000),
  nonce: 0,  // Get from API
  data: {
    operations: [
      {
        type: 'SET',
        key: 'user:alice:name',
        value: btoa('Alice')
      }
    ]
  }
}

// Sign transaction
const txString = JSON.stringify({
  from: tx.from,
  timestamp: tx.timestamp,
  nonce: tx.nonce,
  data: tx.data
})
const txHash = ethers.utils.keccak256(ethers.utils.toUtf8Bytes(txString))
tx.signature = await wallet.signMessage(ethers.utils.arrayify(txHash))

// Submit
const result = await post('/transaction', { transaction: tx })
console.log(`Transaction hash: ${result.data.hash}`)
```

## Development Workflow

### 1. Design Data Model

Plan your key structure:

```
# User profiles
user:{address}:username
user:{address}:bio
user:{address}:avatar

# Posts
post:{id}:author
post:{id}:content
post:{id}:timestamp
post:{id}:likes

# Indexes
index:user:{address}:posts:{postId}
index:date:{date}:posts:{postId}
```

### 2. Build Client Library

Create reusable functions:

```javascript
class PodoruClient {
  constructor(apiURL, wallet) {
    this.apiURL = apiURL
    this.wallet = wallet
  }

  async get(key) { /* ... */ }
  async getBatch(keys) { /* ... */ }
  async queryPrefix(prefix, limit) { /* ... */ }
  async submitTransaction(operations) { /* ... */ }
}
```

### 3. Implement Features

Build application logic:

```javascript
async function createPost(content) {
  const postId = Date.now()

  await client.submitTransaction([
    {
      type: 'SET',
      key: `post:${postId}:author`,
      value: btoa(userAddress)
    },
    {
      type: 'SET',
      key: `post:${postId}:content`,
      value: btoa(content)
    },
    {
      type: 'SET',
      key: `post:${postId}:timestamp`,
      value: btoa(String(Date.now()))
    }
  ])

  return postId
}
```

### 4. Query Data

Retrieve and display:

```javascript
async function getPost(postId) {
  const data = await client.queryPrefix(`post:${postId}:`)

  return {
    id: postId,
    author: atob(data.results[`post:${postId}:author`]),
    content: atob(data.results[`post:${postId}:content`]),
    timestamp: parseInt(atob(data.results[`post:${postId}:timestamp`]))
  }
}
```

### 5. Test

Verify functionality:

```javascript
// Create post
const postId = await createPost('Hello Podoru!')

// Retrieve post
const post = await getPost(postId)
console.log(post)

// Verify
assert.equal(post.content, 'Hello Podoru!')
```

## Next Steps

Learn more about building applications on Podoru Chain:

- [Data Storage Patterns](data-patterns.md) - Best practices for data modeling
- [Querying Data](querying.md) - Advanced query techniques
- [Example Applications](examples.md) - Complete application examples

## Additional Resources

- [API Reference](../api-reference/README.md) - Complete API documentation
- [Architecture](../architecture/README.md) - Understanding how Podoru Chain works
- [Configuration](../configuration/README.md) - Node configuration

## Getting Help

- Check the [Troubleshooting Guide](../troubleshooting/README.md)
- Search [GitHub Issues](https://github.com/podoru/podoru-chain/issues)
- Ask questions in GitHub Discussions

## Example: Simple To-Do App

Here's a complete example to get you started:

```javascript
class TodoApp {
  constructor(client, userAddress) {
    this.client = client
    this.userAddress = userAddress
  }

  async addTodo(text) {
    const todoId = Date.now()

    await this.client.submitTransaction([
      {
        type: 'SET',
        key: `todo:${this.userAddress}:${todoId}:text`,
        value: btoa(text)
      },
      {
        type: 'SET',
        key: `todo:${this.userAddress}:${todoId}:done`,
        value: btoa('false')
      }
    ])

    return todoId
  }

  async getTodos() {
    const data = await this.client.queryPrefix(`todo:${this.userAddress}:`)

    const todos = {}
    for (const [key, value] of Object.entries(data.results)) {
      const todoId = key.split(':')[2]
      const field = key.split(':')[3]

      if (!todos[todoId]) {
        todos[todoId] = { id: todoId }
      }

      todos[todoId][field] = atob(value)
    }

    return Object.values(todos)
  }

  async toggleTodo(todoId) {
    const todos = await this.getTodos()
    const todo = todos.find(t => t.id === todoId)

    await this.client.submitTransaction([
      {
        type: 'SET',
        key: `todo:${this.userAddress}:${todoId}:done`,
        value: btoa(todo.done === 'true' ? 'false' : 'true')
      }
    ])
  }

  async deleteTodo(todoId) {
    await this.client.submitTransaction([
      {
        type: 'DELETE',
        key: `todo:${this.userAddress}:${todoId}:text`
      },
      {
        type: 'DELETE',
        key: `todo:${this.userAddress}:${todoId}:done`
      }
    ])
  }
}

// Usage
const app = new TodoApp(client, userAddress)

await app.addTodo('Learn Podoru Chain')
await app.addTodo('Build amazing dapp')

const todos = await app.getTodos()
console.log(todos)
```

Start building your decentralized application today!
