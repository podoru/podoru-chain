# Querying Data

Advanced techniques for querying blockchain state efficiently.

## Query Methods

Podoru Chain provides three query methods, each optimized for different use cases.

### 1. Single Key Query

Get one specific value.

**Endpoint**: `GET /api/v1/state/{key}`

**Use Case**: When you know the exact key

**Example**:
```javascript
const result = await fetch('http://localhost:8545/api/v1/state/user:alice:name')
const { data } = await result.json()
console.log(atob(data.value))  // "Alice"
```

**Performance**: O(1) - Fastest

### 2. Batch Query

Get multiple specific values in one request.

**Endpoint**: `POST /api/v1/state/batch`

**Use Case**: When you know multiple keys

**Example**:
```javascript
const result = await fetch('http://localhost:8545/api/v1/state/batch', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    keys: [
      'user:alice:name',
      'user:alice:email',
      'user:alice:bio'
    ]
  })
})

const { data } = await result.json()
// {
//   'user:alice:name': 'QWxpY2U=',
//   'user:alice:email': '...',
//   'user:alice:bio': '...'
// }
```

**Performance**: O(n) where n = number of keys

### 3. Prefix Query

Get all keys matching a prefix pattern.

**Endpoint**: `POST /api/v1/state/query/prefix`

**Use Case**: When you want all data in a namespace

**Example**:
```javascript
const result = await fetch('http://localhost:8545/api/v1/state/query/prefix', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    prefix: 'user:alice:',
    limit: 100
  })
})

const { data } = await result.json()
// {
//   prefix: 'user:alice:',
//   count: 4,
//   results: {
//     'user:alice:name': '...',
//     'user:alice:email': '...',
//     'user:alice:bio': '...',
//     'user:alice:avatar': '...'
//   }
// }
```

**Performance**: O(n) where n = matching keys

## Choosing the Right Query Method

| Scenario | Method | Example |
|----------|--------|---------|
| Get user's name | Single | `GET /state/user:alice:name` |
| Get user profile (known fields) | Batch | `POST /state/batch` with known keys |
| Get all user data | Prefix | `POST /state/query/prefix` with `user:alice:` |
| Get all posts | Prefix | `POST /state/query/prefix` with `post:` |
| Get specific post | Batch | `POST /state/batch` with post field keys |

## Advanced Query Patterns

### Loading Complete Objects

Retrieve all fields for an entity:

```javascript
async function loadUser(address) {
  const response = await fetch('http://localhost:8545/api/v1/state/query/prefix', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      prefix: `user:${address}:`,
      limit: 100
    })
  })

  const { data } = await response.json()

  // Convert to object
  const user = { address }
  for (const [key, value] of Object.entries(data.results)) {
    const field = key.split(':')[2]
    user[field] = atob(value)
  }

  return user
}

const alice = await loadUser('0xAlice...')
// {
//   address: '0xAlice...',
//   name: 'Alice',
//   email: 'alice@example.com',
//   bio: '...',
//   avatar: '...'
// }
```

### Querying Collections

Get all items in a collection:

```javascript
async function getAllPosts(limit = 100) {
  const response = await fetch('http://localhost:8545/api/v1/state/query/prefix', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      prefix: 'post:',
      limit: limit
    })
  })

  const { data } = await response.json()

  // Group by post ID
  const posts = {}
  for (const [key, value] of Object.entries(data.results)) {
    const postId = key.split(':')[1]
    const field = key.split(':')[2]

    if (!posts[postId]) {
      posts[postId] = { id: postId }
    }

    posts[postId][field] = atob(value)
  }

  return Object.values(posts)
}

const posts = await getAllPosts()
// [
//   { id: '12345', author: '...', content: '...', timestamp: '...' },
//   { id: '12346', author: '...', content: '...', timestamp: '...' }
// ]
```

### Filtering Results

Filter results client-side:

```javascript
async function getActiveUsers() {
  // Get all users
  const response = await fetch('http://localhost:8545/api/v1/state/query/prefix', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      prefix: 'user:',
      limit: 1000
    })
  })

  const { data } = await response.json()

  // Group by address
  const users = {}
  for (const [key, value] of Object.entries(data.results)) {
    const address = key.split(':')[1]
    const field = key.split(':')[2]

    if (!users[address]) {
      users[address] = { address }
    }

    users[address][field] = atob(value)
  }

  // Filter active users
  return Object.values(users).filter(user => user.status === 'active')
}
```

### Sorting Results

Sort results by a field:

```javascript
async function getPostsSortedByTimestamp() {
  const posts = await getAllPosts()

  // Sort by timestamp (newest first)
  return posts.sort((a, b) => parseInt(b.timestamp) - parseInt(a.timestamp))
}

async function getUsersSortedByName() {
  const users = await getAllUsers()

  // Sort by name (alphabetical)
  return users.sort((a, b) => a.name.localeCompare(b.name))
}
```

## Using Indexes for Efficient Queries

### Creating Indexes

When storing data, create indexes for common queries:

```javascript
async function createPost(postId, author, content, category) {
  const timestamp = Date.now()
  const date = new Date(timestamp).toISOString().split('T')[0]

  const operations = [
    // Main data
    { type: 'SET', key: `post:${postId}:author`, value: btoa(author) },
    { type: 'SET', key: `post:${postId}:content`, value: btoa(content) },
    { type: 'SET', key: `post:${postId}:timestamp`, value: btoa(String(timestamp)) },
    { type: 'SET', key: `post:${postId}:category`, value: btoa(category) },

    // Index by author
    { type: 'SET', key: `index:author:${author}:${postId}`, value: btoa('1') },

    // Index by date
    { type: 'SET', key: `index:date:${date}:${postId}`, value: btoa('1') },

    // Index by category
    { type: 'SET', key: `index:category:${category}:${postId}`, value: btoa('1') }
  ]

  await submitTransaction(operations)
}
```

### Querying by Index

Use indexes to find items:

```javascript
async function getPostsByAuthor(author) {
  // Query author index
  const response = await fetch('http://localhost:8545/api/v1/state/query/prefix', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      prefix: `index:author:${author}:`,
      limit: 100
    })
  })

  const { data } = await response.json()

  // Extract post IDs
  const postIds = Object.keys(data.results).map(key => key.split(':')[3])

  // Load full posts
  const posts = await Promise.all(postIds.map(id => getPost(id)))

  return posts
}

async function getPostsByDate(date) {
  const response = await fetch('http://localhost:8545/api/v1/state/query/prefix', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      prefix: `index:date:${date}:`,
      limit: 100
    })
  })

  const { data } = await response.json()
  const postIds = Object.keys(data.results).map(key => key.split(':')[3])

  return Promise.all(postIds.map(id => getPost(id)))
}

async function getPostsByCategory(category) {
  const response = await fetch('http://localhost:8545/api/v1/state/query/prefix', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      prefix: `index:category:${category}:`,
      limit: 100
    })
  })

  const { data } = await response.json()
  const postIds = Object.keys(data.results).map(key => key.split(':')[3])

  return Promise.all(postIds.map(id => getPost(id)))
}
```

## Performance Optimization

### Batch Loading

Load multiple items efficiently:

```javascript
// ❌ Bad: Multiple sequential queries
async function getPostsInefficient(postIds) {
  const posts = []
  for (const id of postIds) {
    const post = await getPost(id)  // Sequential HTTP requests
    posts.push(post)
  }
  return posts
}

// ✅ Good: Parallel batch query
async function getPostsEfficient(postIds) {
  // Build list of all keys
  const keys = postIds.flatMap(id => [
    `post:${id}:author`,
    `post:${id}:content`,
    `post:${id}:timestamp`
  ])

  // Single batch query
  const response = await fetch('http://localhost:8545/api/v1/state/batch', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ keys })
  })

  const { data } = await response.json()

  // Group by post ID
  const posts = {}
  for (const [key, value] of Object.entries(data)) {
    const postId = key.split(':')[1]
    const field = key.split(':')[2]

    if (!posts[postId]) {
      posts[postId] = { id: postId }
    }

    posts[postId][field] = atob(value)
  }

  return postIds.map(id => posts[id])
}
```

### Caching

Implement client-side caching:

```javascript
class CachedClient {
  constructor() {
    this.cache = new Map()
    this.cacheTTL = 60000  // 1 minute
  }

  async get(key) {
    const cached = this.cache.get(key)

    if (cached && Date.now() - cached.timestamp < this.cacheTTL) {
      return cached.value
    }

    const response = await fetch(`http://localhost:8545/api/v1/state/${key}`)
    const { data } = await response.json()

    this.cache.set(key, {
      value: data,
      timestamp: Date.now()
    })

    return data
  }

  invalidate(key) {
    this.cache.delete(key)
  }

  invalidatePrefix(prefix) {
    for (const key of this.cache.keys()) {
      if (key.startsWith(prefix)) {
        this.cache.delete(key)
      }
    }
  }
}

const client = new CachedClient()

// First call: hits API
const name1 = await client.get('user:alice:name')

// Second call: hits cache
const name2 = await client.get('user:alice:name')

// Invalidate when data changes
await updateUser('alice', { name: 'Alice Smith' })
client.invalidate('user:alice:name')
```

### Limiting Results

Always use limits for prefix queries:

```javascript
// ❌ Bad: Unlimited query
const data = await queryPrefix('post:')  // Could return millions

// ✅ Good: Limited query
const data = await queryPrefix('post:', 100)  // Maximum 100 results
```

### Pagination (Client-Side)

Implement pagination in the client:

```javascript
async function getPaginatedPosts(page = 1, pageSize = 20) {
  // Get all post IDs from index
  const response = await fetch('http://localhost:8545/api/v1/state/query/prefix', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      prefix: 'index:date:',
      limit: 1000
    })
  })

  const { data } = await response.json()
  const postIds = Object.keys(data.results)
    .map(key => key.split(':').pop())
    .sort((a, b) => b - a)  // Sort by ID (newest first)

  // Paginate
  const start = (page - 1) * pageSize
  const end = start + pageSize
  const pagePostIds = postIds.slice(start, end)

  // Load posts for this page
  const posts = await getPostsEfficient(pagePostIds)

  return {
    posts,
    page,
    pageSize,
    totalPosts: postIds.length,
    totalPages: Math.ceil(postIds.length / pageSize)
  }
}

// Usage
const page1 = await getPaginatedPosts(1, 20)
console.log(`Page 1 of ${page1.totalPages}`)
console.log(page1.posts)
```

## Query Patterns Summary

| Pattern | Method | Complexity | Best For |
|---------|--------|-----------|----------|
| Single value | GET | O(1) | Known exact key |
| Multiple values | Batch POST | O(n) | Known multiple keys |
| All in namespace | Prefix POST | O(n) | All data for entity |
| Indexed lookup | Prefix POST | O(n) | Find by attribute |
| Filtered results | Prefix + filter | O(n) | Complex conditions |

## Best Practices

### Do's

✅ Use batch queries for multiple known keys
✅ Create indexes for common query patterns
✅ Always set limits on prefix queries
✅ Implement client-side caching
✅ Load data in parallel when possible
✅ Filter and sort on the client side

### Don'ts

❌ Query in loops (use batch queries)
❌ Load unlimited data
❌ Query without caching frequently-accessed data
❌ Create indexes for every possible query
❌ Forget to decode base64 values
❌ Query full collections when you need one item

## Further Reading

- [Data Storage Patterns](data-patterns.md) - Key design and indexing
- [Example Applications](examples.md) - Complete query examples
- [State API Reference](../api-reference/state.md) - Detailed API documentation
