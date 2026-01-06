# State Endpoints

Endpoints for querying blockchain state (key-value data).

## GET /state/{key}

Get the value for a single key.

### Request

```http
GET /api/v1/state/{key}
```

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| key | string | Yes | State key to query |

### Response

```json
{
  "success": true,
  "data": {
    "key": "user:alice:name",
    "value": "QWxpY2U=",
    "decoded": "Alice"
  }
}
```

### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| key | string | The queried key |
| value | string | Base64-encoded value |
| decoded | string | UTF-8 decoded value (if valid UTF-8) |

### Example

```bash
# Get chain name
curl http://localhost:8545/api/v1/state/chain:name | jq

# Get user data
curl http://localhost:8545/api/v1/state/user:alice:name | jq
```

```javascript
const key = 'user:alice:name'
const response = await fetch(`http://localhost:8545/api/v1/state/${key}`)
const { data } = await response.json()

// Decode base64 value
const value = atob(data.value)
console.log(`${key} = ${value}`)
```

```python
import requests
import base64

key = 'user:alice:name'
response = requests.get(f'http://localhost:8545/api/v1/state/{key}')
data = response.json()['data']

# Decode base64 value
value = base64.b64decode(data['value']).decode('utf-8')
print(f'{key} = {value}')
```

### Error Responses

**404 Not Found**:
```json
{
  "success": false,
  "error": "Key not found: user:bob:name"
}
```

---

## POST /state/batch

Get multiple values in a single request.

### Request

```http
POST /api/v1/state/batch
Content-Type: application/json
```

### Request Body

```json
{
  "keys": [
    "user:alice:name",
    "user:alice:email",
    "user:alice:bio"
  ]
}
```

### Response

```json
{
  "success": true,
  "data": {
    "user:alice:name": "QWxpY2U=",
    "user:alice:email": "YWxpY2VAZXhhbXBsZS5jb20=",
    "user:alice:bio": "RGV2ZWxvcGVy"
  }
}
```

**Note**: Keys that don't exist are omitted from the response.

### Example

```bash
curl -X POST http://localhost:8545/api/v1/state/batch \
  -H "Content-Type: application/json" \
  -d '{
    "keys": [
      "chain:name",
      "chain:version",
      "chain:description"
    ]
  }' | jq
```

```javascript
async function batchGet(keys) {
  const response = await fetch('http://localhost:8545/api/v1/state/batch', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ keys })
  })

  const { data } = await response.json()

  // Decode all values
  const decoded = {}
  for (const [key, value] of Object.entries(data)) {
    decoded[key] = atob(value)
  }

  return decoded
}

// Usage
const userProfile = await batchGet([
  'user:alice:name',
  'user:alice:email',
  'user:alice:bio',
  'user:alice:avatar'
])

console.log(userProfile)
// {
//   'user:alice:name': 'Alice',
//   'user:alice:email': 'alice@example.com',
//   'user:alice:bio': 'Developer',
//   'user:alice:avatar': 'ipfs://...'
// }
```

### Performance

Batch queries are more efficient than individual queries:

```javascript
// ❌ Inefficient: 4 HTTP requests
const name = await get('user:alice:name')
const email = await get('user:alice:email')
const bio = await get('user:alice:bio')
const avatar = await get('user:alice:avatar')

// ✅ Efficient: 1 HTTP request
const profile = await batchGet([
  'user:alice:name',
  'user:alice:email',
  'user:alice:bio',
  'user:alice:avatar'
])
```

---

## POST /state/query/prefix

Query all keys matching a prefix.

### Request

```http
POST /api/v1/state/query/prefix
Content-Type: application/json
```

### Request Body

```json
{
  "prefix": "user:alice:",
  "limit": 100
}
```

### Parameters

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| prefix | string | Yes | Key prefix to match |
| limit | integer | No | Maximum results (default: 1000) |

### Response

```json
{
  "success": true,
  "data": {
    "prefix": "user:alice:",
    "count": 4,
    "results": {
      "user:alice:name": "QWxpY2U=",
      "user:alice:email": "YWxpY2VAZXhhbXBsZS5jb20=",
      "user:alice:bio": "RGV2ZWxvcGVy",
      "user:alice:avatar": "aXBmczovL1FtLi4u"
    }
  }
}
```

### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| prefix | string | The queried prefix |
| count | integer | Number of matching keys |
| results | object | Key-value pairs (base64-encoded values) |

### Example

```bash
# Get all user data for Alice
curl -X POST http://localhost:8545/api/v1/state/query/prefix \
  -H "Content-Type: application/json" \
  -d '{
    "prefix": "user:alice:",
    "limit": 100
  }' | jq

# Get all posts
curl -X POST http://localhost:8545/api/v1/state/query/prefix \
  -H "Content-Type: application/json" \
  -d '{
    "prefix": "post:",
    "limit": 1000
  }' | jq
```

```javascript
async function queryPrefix(prefix, limit = 100) {
  const response = await fetch('http://localhost:8545/api/v1/state/query/prefix', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ prefix, limit })
  })

  const { data } = await response.json()

  // Decode all values
  const decoded = {}
  for (const [key, value] of Object.entries(data.results)) {
    decoded[key] = atob(value)
  }

  return {
    prefix: data.prefix,
    count: data.count,
    results: decoded
  }
}

// Usage: Get all data for user Alice
const aliceData = await queryPrefix('user:alice:')
console.log(`Found ${aliceData.count} keys`)
console.log(aliceData.results)
```

### Use Cases

#### Get User Profile

```javascript
async function getUserProfile(username) {
  const data = await queryPrefix(`user:${username}:`)

  const profile = {}
  for (const [key, value] of Object.entries(data.results)) {
    const field = key.split(':')[2]  // Extract field name
    profile[field] = value
  }

  return profile
}

const alice = await getUserProfile('alice')
// { name: 'Alice', email: 'alice@...', bio: '...', avatar: '...' }
```

#### Get All Posts

```javascript
async function getAllPosts(limit = 100) {
  const data = await queryPrefix('post:', limit)

  // Group by post ID
  const posts = {}

  for (const [key, value] of Object.entries(data.results)) {
    const parts = key.split(':')
    const postId = parts[1]
    const field = parts[2]

    if (!posts[postId]) {
      posts[postId] = { id: postId }
    }

    posts[postId][field] = value
  }

  return Object.values(posts)
}

const posts = await getAllPosts()
// [
//   { id: '123', author: '0x...', content: '...', timestamp: '...' },
//   { id: '124', author: '0x...', content: '...', timestamp: '...' }
// ]
```

#### Search by Category

```javascript
async function getProductsByCategory(category) {
  const data = await queryPrefix(`product:${category}:`)

  const products = {}

  for (const [key, value] of Object.entries(data.results)) {
    const productId = key.split(':')[2]
    const field = key.split(':')[3]

    if (!products[productId]) {
      products[productId] = { id: productId }
    }

    products[productId][field] = value
  }

  return Object.values(products)
}

const electronics = await getProductsByCategory('electronics')
```

---

## Query Patterns

### Key Naming Conventions

Use hierarchical keys for efficient prefix queries:

```
namespace:entity:id:field

Examples:
user:alice:name
user:alice:email
post:12345:content
post:12345:author
product:electronics:999:name
product:electronics:999:price
```

### Indexing Strategy

Create index keys for efficient lookups:

```javascript
// When storing a post, create indexes
const operations = [
  // Main data
  { type: 'SET', key: 'post:12345:content', value: btoa('Hello') },
  { type: 'SET', key: 'post:12345:author', value: btoa('alice') },

  // Index by author
  { type: 'SET', key: 'index:author:alice:12345', value: btoa('1') },

  // Index by date
  { type: 'SET', key: 'index:date:2024-01-06:12345', value: btoa('1') }
]

// Query posts by author
const alicePosts = await queryPrefix('index:author:alice:')

// Query posts by date
const todayPosts = await queryPrefix('index:date:2024-01-06:')
```

### Pagination (Future)

For large result sets, use limit and offset:

```javascript
// Future feature
async function getPaginatedResults(prefix, page = 1, pageSize = 100) {
  const offset = (page - 1) * pageSize

  const response = await fetch('http://localhost:8545/api/v1/state/query/prefix', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      prefix,
      limit: pageSize,
      offset: offset
    })
  })

  return await response.json()
}
```

---

## Performance Considerations

### Query Efficiency

| Query Type | Complexity | Use Case |
|------------|-----------|----------|
| Single key | O(1) | Get specific value |
| Batch | O(n) | Get multiple known keys |
| Prefix | O(n) | Scan by namespace |

### Best Practices

**Use Batch Queries**:
```javascript
// ✅ Good: Single request
const data = await batchGet(['key1', 'key2', 'key3'])

// ❌ Bad: Multiple requests
const val1 = await get('key1')
const val2 = await get('key2')
const val3 = await get('key3')
```

**Limit Prefix Results**:
```javascript
// ✅ Good: Limit results
const data = await queryPrefix('posts:', 100)

// ❌ Bad: Unlimited (could be huge)
const data = await queryPrefix('posts:')
```

**Cache Results**:
```javascript
const cache = new Map()

async function getCached(key) {
  if (cache.has(key)) {
    return cache.get(key)
  }

  const value = await get(key)
  cache.set(key, value)
  return value
}
```

---

## Error Responses

**400 Bad Request - Invalid Prefix**:
```json
{
  "success": false,
  "error": "Prefix cannot be empty"
}
```

**400 Bad Request - Limit Too Large**:
```json
{
  "success": false,
  "error": "Limit exceeds maximum (1000)"
}
```

---

## Related Endpoints

- [POST /transaction](transactions.md) - Submit state changes
- [GET /block/latest](blocks.md) - Get latest state root
- [GET /chain/info](chain.md) - Get blockchain info
