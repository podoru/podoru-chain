# Building Applications on Podoru Chain

This guide shows how to build decentralized applications using Podoru Chain's flexible key-value storage.

## 📋 Table of Contents

1. [Overview](#overview)
2. [Data Storage Patterns](#data-storage-patterns)
3. [Query API](#query-api)
4. [Example Applications](#example-applications)
5. [Best Practices](#best-practices)

## Overview

Podoru Chain stores data as **key-value pairs**, making it flexible for any application type:

- **Social Media** (Twitter, Reddit)
- **File Metadata** (IPFS, storage systems)
- **User Profiles** (authentication, settings)
- **E-commerce** (products, orders)
- **Gaming** (scores, achievements)
- **IoT Data** (sensor readings, logs)

## Data Storage Patterns

### Naming Convention

Use structured keys with colons as separators:

```
pattern: "namespace:entity:id:field"

Examples:
user:alice:name
user:alice:email
tweet:12345:content
product:999:price
file:abc123:hash
```

### Example Schemas

**User System:**
```
user:{address}:username     → "alice"
user:{address}:bio          → "Blockchain developer"
user:{address}:avatar       → "ipfs://Qm..."
user:{address}:created      → "1704556800"
```

**Social Posts:**
```
post:{id}:author     → "0xAliceAddress"
post:{id}:content    → "Hello blockchain!"
post:{id}:timestamp  → "1704556800"
post:{id}:likes      → "42"
```

**E-commerce:**
```
product:{id}:name        → "Laptop"
product:{id}:price       → "999.99"
product:{id}:stock       → "50"
product:{id}:seller      → "0xSellerAddress"
```

## Query API

### 1. Single Key Query

Get one value:

```bash
GET /api/v1/state/{key}

# Example
curl http://localhost:8545/api/v1/state/user:alice:username
```

**Response:**
```json
{
  "success": true,
  "data": {
    "key": "user:alice:username",
    "value": "YWxpY2U="  // base64 encoded "alice"
  }
}
```

### 2. Batch Query (NEW!)

Get multiple values at once:

```bash
POST /api/v1/state/batch

# Example
curl -X POST http://localhost:8545/api/v1/state/batch \
  -H "Content-Type: application/json" \
  -d '{
    "keys": [
      "user:alice:username",
      "user:alice:bio",
      "user:alice:avatar"
    ]
  }'
```

**Response:**
```json
{
  "success": true,
  "data": {
    "user:alice:username": "alice",
    "user:alice:bio": "Blockchain developer",
    "user:alice:avatar": "ipfs://Qm..."
  }
}
```

### 3. Prefix Query (NEW!)

Get all keys matching a prefix (like SQL `LIKE 'prefix%'`):

```bash
POST /api/v1/state/query/prefix

# Get all of Alice's data
curl -X POST http://localhost:8545/api/v1/state/query/prefix \
  -H "Content-Type: application/json" \
  -d '{
    "prefix": "user:alice:",
    "limit": 100
  }'
```

**Response:**
```json
{
  "success": true,
  "data": {
    "prefix": "user:alice:",
    "count": 3,
    "results": {
      "user:alice:username": "alice",
      "user:alice:bio": "Blockchain developer",
      "user:alice:avatar": "ipfs://Qm..."
    }
  }
}
```

### 4. Write Data (Transactions)

Submit transactions to store data:

```bash
POST /api/v1/transaction

curl -X POST http://localhost:8545/api/v1/transaction \
  -H "Content-Type: application/json" \
  -d '{
    "transaction": {
      "from": "0xYourAddress",
      "timestamp": 1704556800,
      "nonce": 0,
      "data": {
        "operations": [
          {
            "type": "SET",
            "key": "user:alice:username",
            "value": "YWxpY2U="
          },
          {
            "type": "SET",
            "key": "user:alice:bio",
            "value": "QmxvY2tjaGFpbiBkZXZlbG9wZXI="
          }
        ]
      },
      "signature": "0xYourSignature..."
    }
  }'
```

## Example Applications

### Example 1: Twitter Clone

**Store a Tweet:**

```javascript
// Frontend JavaScript
async function postTweet(content) {
  const tweetId = Date.now(); // Simple ID generation

  const transaction = {
    from: userAddress,
    timestamp: Math.floor(Date.now() / 1000),
    nonce: await getNonce(userAddress),
    data: {
      operations: [
        {
          type: "SET",
          key: `tweet:${tweetId}:author`,
          value: btoa(userAddress)
        },
        {
          type: "SET",
          key: `tweet:${tweetId}:content`,
          value: btoa(content)
        },
        {
          type: "SET",
          key: `tweet:${tweetId}:timestamp`,
          value: btoa(String(Date.now()))
        }
      ]
    }
  };

  // Sign transaction
  const signature = await signTransaction(transaction);
  transaction.signature = signature;

  // Submit
  const response = await fetch('http://localhost:8545/api/v1/transaction', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ transaction })
  });

  return await response.json();
}
```

**Query User's Tweets:**

```javascript
async function getUserTweets(userAddress) {
  // Query all tweets by this user
  const response = await fetch('http://localhost:8545/api/v1/state/query/prefix', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      prefix: `tweet:`,
      limit: 1000
    })
  });

  const data = await response.json();

  // Filter tweets by author
  const userTweets = Object.entries(data.data.results)
    .filter(([key, value]) => key.includes(':author') && value === userAddress)
    .map(([key]) => {
      const tweetId = key.split(':')[1];
      return getTweetById(tweetId);
    });

  return userTweets;
}

async function getTweetById(tweetId) {
  const response = await fetch('http://localhost:8545/api/v1/state/batch', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      keys: [
        `tweet:${tweetId}:author`,
        `tweet:${tweetId}:content`,
        `tweet:${tweetId}:timestamp`
      ]
    })
  });

  return await response.json();
}
```

### Example 2: Decentralized File Storage Metadata

**Store File Metadata:**

```javascript
async function storeFileMetadata(fileHash, metadata) {
  const transaction = {
    from: userAddress,
    timestamp: Math.floor(Date.now() / 1000),
    nonce: await getNonce(userAddress),
    data: {
      operations: [
        {
          type: "SET",
          key: `file:${fileHash}:owner`,
          value: btoa(userAddress)
        },
        {
          type: "SET",
          key: `file:${fileHash}:name`,
          value: btoa(metadata.name)
        },
        {
          type: "SET",
          key: `file:${fileHash}:size`,
          value: btoa(String(metadata.size))
        },
        {
          type: "SET",
          key: `file:${fileHash}:type`,
          value: btoa(metadata.type)
        },
        {
          type: "SET",
          key: `file:${fileHash}:ipfs`,
          value: btoa(metadata.ipfsHash)
        }
      ]
    }
  };

  const signature = await signTransaction(transaction);
  transaction.signature = signature;

  return await submitTransaction(transaction);
}
```

**Query File Metadata:**

```javascript
async function getFileMetadata(fileHash) {
  const response = await fetch('http://localhost:8545/api/v1/state/query/prefix', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      prefix: `file:${fileHash}:`,
      limit: 100
    })
  });

  const data = await response.json();
  return data.data.results;
}
```

### Example 3: User Profile System

**Complete User Profile:**

```javascript
class UserProfile {
  constructor(address) {
    this.address = address;
  }

  async save(profileData) {
    const operations = Object.entries(profileData).map(([field, value]) => ({
      type: "SET",
      key: `user:${this.address}:${field}`,
      value: btoa(String(value))
    }));

    const transaction = {
      from: this.address,
      timestamp: Math.floor(Date.now() / 1000),
      nonce: await getNonce(this.address),
      data: { operations }
    };

    const signature = await signTransaction(transaction);
    transaction.signature = signature;

    return await submitTransaction(transaction);
  }

  async load() {
    const response = await fetch('http://localhost:8545/api/v1/state/query/prefix', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        prefix: `user:${this.address}:`,
        limit: 100
      })
    });

    const data = await response.json();

    // Convert back to object
    const profile = {};
    for (const [key, value] of Object.entries(data.data.results)) {
      const field = key.split(':')[2];
      profile[field] = atob(value); // Decode base64
    }

    return profile;
  }

  async update(field, value) {
    const transaction = {
      from: this.address,
      timestamp: Math.floor(Date.now() / 1000),
      nonce: await getNonce(this.address),
      data: {
        operations: [{
          type: "SET",
          key: `user:${this.address}:${field}`,
          value: btoa(String(value))
        }]
      }
    };

    const signature = await signTransaction(transaction);
    transaction.signature = signature;

    return await submitTransaction(transaction);
  }
}

// Usage
const profile = new UserProfile('0xAliceAddress');
await profile.save({
  username: 'alice',
  bio: 'Blockchain developer',
  avatar: 'ipfs://Qm...',
  email: 'alice@example.com'
});

const data = await profile.load();
console.log(data); // { username: 'alice', bio: '...', ... }

await profile.update('bio', 'Senior blockchain developer');
```

## Best Practices

### 1. Key Design

✅ **Good:**
```
user:alice:profile:name
user:alice:profile:bio
tweet:12345:content
product:999:price
```

❌ **Bad:**
```
alice_name          // No namespace
userAliceProfileName // Hard to query
user-alice-name     // Use colons, not dashes
```

### 2. Data Organization

**Hierarchical Structure:**
```
app:module:entity:id:field

Examples:
social:user:alice:posts:count
social:post:12345:likes:count
shop:product:999:inventory:stock
shop:order:777:status:value
```

### 3. Indexing Strategy

For efficient queries, create index keys:

```javascript
// When storing a post, also create an index
operations: [
  // Main data
  { type: "SET", key: "post:12345:content", value: "..." },
  { type: "SET", key: "post:12345:author", value: "alice" },

  // Index: user's posts
  { type: "SET", key: "index:user:alice:posts:12345", value: "1" },

  // Index: posts by date
  { type: "SET", key: "index:posts:2024-01-06:12345", value: "1" }
]
```

Then query efficiently:
```javascript
// Get all of Alice's posts
const posts = await queryPrefix("index:user:alice:posts:");

// Get posts from a specific date
const todayPosts = await queryPrefix("index:posts:2024-01-06:");
```

### 4. Data Size Limits

Current limits:
- **Max key size**: 1 KB
- **Max value size**: 1 MB
- **Max transaction size**: Recommended < 100 KB

For large data:
- Store on IPFS/Arweave
- Store only hash/reference on blockchain

```javascript
// Good: Store reference
{
  type: "SET",
  key: "file:abc:metadata",
  value: JSON.stringify({
    ipfs: "Qm...",
    size: 1000000,
    type: "image/png"
  })
}

// Bad: Store entire file
{
  type: "SET",
  key: "file:abc:data",
  value: "<10MB of binary data>" // ❌ Too large!
}
```

### 5. Transaction Batching

Batch related operations in a single transaction:

```javascript
// ✅ Good: All user data in one transaction
const transaction = {
  data: {
    operations: [
      { type: "SET", key: "user:alice:name", value: "Alice" },
      { type: "SET", key: "user:alice:email", value: "alice@example.com" },
      { type: "SET", key: "user:alice:bio", value: "Developer" }
    ]
  }
};

// ❌ Bad: Three separate transactions
// This wastes gas and may leave inconsistent state
await submitTx({ operations: [{ type: "SET", key: "user:alice:name", ... }] });
await submitTx({ operations: [{ type: "SET", key: "user:alice:email", ... }] });
await submitTx({ operations: [{ type: "SET", key: "user:alice:bio", ... }] });
```

## Advanced Patterns

### Counters

```javascript
// Increment a counter
async function increment(key) {
  const current = await getState(key) || "0";
  const newValue = String(parseInt(current) + 1);
  await submitTx({
    operations: [{ type: "SET", key, value: btoa(newValue) }]
  });
}

// Usage
await increment("post:12345:likes:count");
```

### Lists (Append-Only)

```javascript
// Add item to list
async function appendToList(listKey, item) {
  const index = Date.now(); // Or use counter
  await submitTx({
    operations: [{
      type: "SET",
      key: `${listKey}:${index}`,
      value: btoa(JSON.stringify(item))
    }]
  });
}

// Get entire list
async function getList(listKey) {
  const result = await queryPrefix(listKey);
  return Object.values(result.results).map(v => JSON.parse(atob(v)));
}
```

### Soft Delete

```javascript
// Instead of DELETE, mark as deleted
{
  type: "SET",
  key: "post:12345:deleted",
  value: btoa("true")
}

// Filter deleted items when querying
const posts = await getPosts();
const activePosts = posts.filter(p => !p.deleted);
```

## Client Libraries

### JavaScript/TypeScript SDK (Example)

```javascript
class PodoruClient {
  constructor(apiUrl) {
    this.apiUrl = apiUrl;
  }

  async get(key) {
    const res = await fetch(`${this.apiUrl}/api/v1/state/${key}`);
    return await res.json();
  }

  async getBatch(keys) {
    const res = await fetch(`${this.apiUrl}/api/v1/state/batch`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ keys })
    });
    return await res.json();
  }

  async queryPrefix(prefix, limit = 100) {
    const res = await fetch(`${this.apiUrl}/api/v1/state/query/prefix`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ prefix, limit })
    });
    return await res.json();
  }

  async submitTx(transaction) {
    const res = await fetch(`${this.apiUrl}/api/v1/transaction`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ transaction })
    });
    return await res.json();
  }
}

// Usage
const client = new PodoruClient('http://localhost:8545');
const userData = await client.queryPrefix('user:alice:');
```

## Summary

Podoru Chain is **perfect for building decentralized apps** because:

✅ **Flexible key-value storage** - store any data structure
✅ **Fast queries** - single key, batch, and prefix scans
✅ **Decentralized** - data replicated across all nodes
✅ **Immutable** - all changes recorded in blockchain
✅ **REST API** - easy integration with any language
✅ **No smart contracts needed** - simpler development

**Start building today!** 🚀
