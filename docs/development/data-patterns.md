# Data Storage Patterns

Best practices and patterns for structuring data on Podoru Chain.

## Key Naming Conventions

### Hierarchical Structure

Use colons to create hierarchical keys:

```
pattern: "namespace:entity:id:field"

Examples:
user:alice:name
user:alice:email
post:12345:content
product:999:price
file:abc123:hash
```

### Best Practices

**Good Key Naming**:
```
✅ user:alice:profile:name
✅ user:alice:profile:bio
✅ post:12345:metadata:likes
✅ product:electronics:999:details:price
```

**Bad Key Naming**:
```
❌ alice_name              // No namespace
❌ userAliceProfileName    // Hard to query
❌ user-alice-name         // Use colons, not dashes
❌ user.alice.name         // Use colons, not dots
```

### Namespace Organization

Organize by domain:

```
# User namespace
user:{address}:username
user:{address}:bio
user:{address}:avatar
user:{address}:created

# Post namespace
post:{id}:author
post:{id}:content
post:{id}:timestamp
post:{id}:likes

# Product namespace
product:{id}:name
product:{id}:price
product:{id}:stock
product:{id}:seller
```

## Common Data Patterns

### 1. User Profiles

Store user information:

```javascript
// Data structure
user:{address}:username     → "alice"
user:{address}:email        → "alice@example.com"
user:{address}:bio          → "Blockchain developer"
user:{address}:avatar       → "ipfs://Qm..."
user:{address}:created      → "1704556800"
user:{address}:verified     → "true"

// Save profile
async function saveProfile(address, profile) {
  const operations = [
    { type: 'SET', key: `user:${address}:username`, value: btoa(profile.username) },
    { type: 'SET', key: `user:${address}:email`, value: btoa(profile.email) },
    { type: 'SET', key: `user:${address}:bio`, value: btoa(profile.bio) },
    { type: 'SET', key: `user:${address}:avatar`, value: btoa(profile.avatar) },
    { type: 'SET', key: `user:${address}:created`, value: btoa(String(Date.now())) }
  ]

  await submitTransaction(operations)
}

// Load profile
async function loadProfile(address) {
  const data = await queryPrefix(`user:${address}:`)

  const profile = {}
  for (const [key, value] of Object.entries(data.results)) {
    const field = key.split(':')[2]
    profile[field] = atob(value)
  }

  return profile
}
```

### 2. Social Posts

Store posts with metadata:

```javascript
// Data structure
post:{id}:author     → "0xAliceAddress"
post:{id}:content    → "Hello blockchain!"
post:{id}:timestamp  → "1704556800"
post:{id}:likes      → "42"
post:{id}:replies    → "5"

// Create post
async function createPost(content, author) {
  const postId = Date.now()

  const operations = [
    { type: 'SET', key: `post:${postId}:author`, value: btoa(author) },
    { type: 'SET', key: `post:${postId}:content`, value: btoa(content) },
    { type: 'SET', key: `post:${postId}:timestamp`, value: btoa(String(Date.now())) },
    { type: 'SET', key: `post:${postId}:likes`, value: btoa('0') },
    { type: 'SET', key: `post:${postId}:replies`, value: btoa('0') }
  ]

  await submitTransaction(operations)
  return postId
}

// Get post
async function getPost(postId) {
  const data = await queryPrefix(`post:${postId}:`)

  const post = { id: postId }
  for (const [key, value] of Object.entries(data.results)) {
    const field = key.split(':')[2]
    post[field] = atob(value)
  }

  return post
}
```

### 3. E-commerce Products

Product catalog:

```javascript
// Data structure
product:{id}:name        → "Laptop"
product:{id}:price       → "999.99"
product:{id}:stock       → "50"
product:{id}:seller      → "0xSellerAddress"
product:{id}:category    → "electronics"
product:{id}:description → "High-performance laptop"

// Save product
async function saveProduct(productId, product) {
  const operations = Object.entries(product).map(([field, value]) => ({
    type: 'SET',
    key: `product:${productId}:${field}`,
    value: btoa(String(value))
  }))

  await submitTransaction(operations)
}

// Get product
async function getProduct(productId) {
  const data = await queryPrefix(`product:${productId}:`)

  const product = { id: productId }
  for (const [key, value] of Object.entries(data.results)) {
    const field = key.split(':')[2]
    product[field] = atob(value)
  }

  return product
}
```

### 4. File Metadata

Decentralized file storage metadata:

```javascript
// Data structure
file:{hash}:owner       → "0xOwnerAddress"
file:{hash}:name        → "document.pdf"
file:{hash}:size        → "1048576"
file:{hash}:type        → "application/pdf"
file:{hash}:ipfs        → "ipfs://Qm..."
file:{hash}:uploaded    → "1704556800"

// Store file metadata
async function storeFileMetadata(fileHash, metadata) {
  const operations = [
    { type: 'SET', key: `file:${fileHash}:owner`, value: btoa(metadata.owner) },
    { type: 'SET', key: `file:${fileHash}:name`, value: btoa(metadata.name) },
    { type: 'SET', key: `file:${fileHash}:size`, value: btoa(String(metadata.size)) },
    { type: 'SET', key: `file:${fileHash}:type`, value: btoa(metadata.type) },
    { type: 'SET', key: `file:${fileHash}:ipfs`, value: btoa(metadata.ipfsHash) },
    { type: 'SET', key: `file:${fileHash}:uploaded`, value: btoa(String(Date.now())) }
  ]

  await submitTransaction(operations)
}
```

## Indexing Strategies

### Creating Indexes

Create secondary indexes for efficient queries:

```javascript
// When storing a post, also create indexes
const operations = [
  // Main data
  { type: 'SET', key: `post:${postId}:content`, value: btoa(content) },
  { type: 'SET', key: `post:${postId}:author`, value: btoa(author) },
  { type: 'SET', key: `post:${postId}:timestamp`, value: btoa(String(timestamp)) },

  // Index by author
  { type: 'SET', key: `index:author:${author}:${postId}`, value: btoa('1') },

  // Index by date
  { type: 'SET', key: `index:date:${date}:${postId}`, value: btoa('1') },

  // Index by category
  { type: 'SET', key: `index:category:${category}:${postId}`, value: btoa('1') }
]

// Query by author
const authorPosts = await queryPrefix(`index:author:${author}:`)
const postIds = Object.keys(authorPosts.results).map(key => key.split(':')[3])

// Query by date
const datePosts = await queryPrefix(`index:date:2024-01-06:`)

// Query by category
const categoryPosts = await queryPrefix(`index:category:tech:`)
```

### Multi-Level Indexes

Create hierarchical indexes:

```javascript
// Product indexes
index:category:electronics:999           → "1"
index:seller:0xSeller:999                → "1"
index:price:range:0-100:999              → "1"

// Query products by category
const electronics = await queryPrefix('index:category:electronics:')

// Query products by seller
const sellerProducts = await queryPrefix(`index:seller:${sellerAddress}:`)

// Query products in price range
const affordableProducts = await queryPrefix('index:price:range:0-100:')
```

### Composite Indexes

Combine multiple dimensions:

```javascript
// Composite index: category + price range
index:composite:electronics:0-100:999     → "1"
index:composite:electronics:100-500:999   → "1"

// Query electronics under $100
const results = await queryPrefix('index:composite:electronics:0-100:')
```

## Advanced Patterns

### Counters

Implement counters:

```javascript
async function incrementCounter(key) {
  // Get current value
  const current = await get(key)
  const value = current ? parseInt(atob(current.value)) : 0

  // Increment
  const newValue = value + 1

  // Save
  await submitTransaction([
    { type: 'SET', key: key, value: btoa(String(newValue)) }
  ])

  return newValue
}

// Usage
await incrementCounter('post:12345:likes')
await incrementCounter('user:alice:followers')
```

### Lists

Implement append-only lists:

```javascript
async function appendToList(listKey, item) {
  const index = Date.now()  // Or use counter

  await submitTransaction([
    {
      type: 'SET',
      key: `${listKey}:${index}`,
      value: btoa(JSON.stringify(item))
    }
  ])

  return index
}

async function getList(listKey) {
  const data = await queryPrefix(listKey)

  const items = []
  for (const [key, value] of Object.entries(data.results)) {
    items.push(JSON.parse(atob(value)))
  }

  return items.sort((a, b) => a.timestamp - b.timestamp)
}

// Usage
await appendToList('comments:post123', {
  author: 'alice',
  text: 'Great post!',
  timestamp: Date.now()
})

const comments = await getList('comments:post123')
```

### Sets

Implement sets using keys:

```javascript
async function addToSet(setKey, member) {
  await submitTransaction([
    {
      type: 'SET',
      key: `${setKey}:${member}`,
      value: btoa('1')
    }
  ])
}

async function removeFromSet(setKey, member) {
  await submitTransaction([
    {
      type: 'DELETE',
      key: `${setKey}:${member}`
    }
  ])
}

async function getSetMembers(setKey) {
  const data = await queryPrefix(setKey)
  return Object.keys(data.results).map(key => key.split(':').pop())
}

async function isInSet(setKey, member) {
  try {
    await get(`${setKey}:${member}`)
    return true
  } catch {
    return false
  }
}

// Usage: Following system
await addToSet('followers:alice', '0xBobAddress')
await addToSet('followers:alice', '0xCarolAddress')

const followers = await getSetMembers('followers:alice')
// ['0xBobAddress', '0xCarolAddress']

const isFollowing = await isInSet('followers:alice', '0xBobAddress')
// true
```

### Soft Deletion

Mark items as deleted instead of removing:

```javascript
// Instead of DELETE
{
  type: 'DELETE',
  key: 'post:12345:content'
}

// Use soft delete
{
  type: 'SET',
  key: 'post:12345:deleted',
  value: btoa('true')
}

// Filter deleted items when querying
async function getActivePosts() {
  const data = await queryPrefix('post:')

  const posts = []
  // Group by post ID
  const grouped = groupByPostId(data.results)

  for (const post of Object.values(grouped)) {
    if (post.deleted !== 'true') {
      posts.push(post)
    }
  }

  return posts
}
```

## Data Size Limits

### Current Limits

- **Max key size**: 1 KB
- **Max value size**: 1 MB
- **Max transaction size**: Recommended < 100 KB

### Handling Large Data

**For Large Files**:
```javascript
// ❌ Bad: Store entire file
{
  type: 'SET',
  key: 'file:abc:data',
  value: '<10MB of data>'  // Too large!
}

// ✅ Good: Store reference
{
  type: 'SET',
  key: 'file:abc:metadata',
  value: btoa(JSON.stringify({
    ipfs: 'Qm...',
    size: 10485760,
    type: 'image/png'
  }))
}
```

**For Large Objects**:
```javascript
// ❌ Bad: Single large value
{
  type: 'SET',
  key: 'user:alice:data',
  value: btoa(JSON.stringify(veryLargeObject))
}

// ✅ Good: Split into fields
const operations = Object.entries(largeObject).map(([field, value]) => ({
  type: 'SET',
  key: `user:alice:${field}`,
  value: btoa(String(value))
}))
```

## Best Practices Summary

### Do's

✅ Use hierarchical keys with colons
✅ Create indexes for common queries
✅ Batch related operations in one transaction
✅ Store references to large files (IPFS, Arweave)
✅ Use meaningful, descriptive key names
✅ Plan key structure before development
✅ Use soft deletion when you might need recovery

### Don'ts

❌ Store large binary data directly
❌ Use inconsistent key patterns
❌ Create deeply nested keys (> 5 levels)
❌ Store sensitive data unencrypted
❌ Use special characters in keys (except colons)
❌ Create unbounded lists without limits
❌ Forget to create necessary indexes

## Further Reading

- [Querying Data](querying.md) - Advanced query techniques
- [Example Applications](examples.md) - Complete examples
- [API Reference](../api-reference/state.md) - State API details
