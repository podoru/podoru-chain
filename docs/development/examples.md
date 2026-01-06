# Example Applications

Complete examples of building decentralized applications on Podoru Chain.

## Example 1: Twitter Clone

A decentralized social media platform.

### Data Model

```
# Users
user:{address}:username
user:{address}:bio
user:{address}:avatar

# Posts
post:{id}:author
post:{id}:content
post:{id}:timestamp
post:{id}:likes

# Indexes
index:author:{address}:posts:{id}
index:timeline:{timestamp}:{id}
```

### Implementation

```javascript
class DecentralizedTwitter {
  constructor(apiURL, wallet) {
    this.apiURL = apiURL
    this.wallet = wallet
    this.address = wallet.address
  }

  // Create user profile
  async createProfile(username, bio, avatar) {
    const operations = [
      {
        type: 'SET',
        key: `user:${this.address}:username`,
        value: btoa(username)
      },
      {
        type: 'SET',
        key: `user:${this.address}:bio`,
        value: btoa(bio)
      },
      {
        type: 'SET',
        key: `user:${this.address}:avatar`,
        value: btoa(avatar)
      }
    ]

    const tx = await this.createTransaction(operations)
    return await this.submitTransaction(tx)
  }

  // Post a tweet
  async postTweet(content) {
    const tweetId = Date.now()
    const timestamp = Date.now()

    const operations = [
      {
        type: 'SET',
        key: `post:${tweetId}:author`,
        value: btoa(this.address)
      },
      {
        type: 'SET',
        key: `post:${tweetId}:content`,
        value: btoa(content)
      },
      {
        type: 'SET',
        key: `post:${tweetId}:timestamp`,
        value: btoa(String(timestamp))
      },
      {
        type: 'SET',
        key: `post:${tweetId}:likes`,
        value: btoa('0')
      },
      // Index by author
      {
        type: 'SET',
        key: `index:author:${this.address}:posts:${tweetId}`,
        value: btoa('1')
      },
      // Index by timeline
      {
        type: 'SET',
        key: `index:timeline:${timestamp}:${tweetId}`,
        value: btoa('1')
      }
    ]

    const tx = await this.createTransaction(operations)
    await this.submitTransaction(tx)

    return tweetId
  }

  // Get user's tweets
  async getUserTweets(address) {
    const response = await fetch(`${this.apiURL}/state/query/prefix`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        prefix: `index:author:${address}:posts:`,
        limit: 100
      })
    })

    const { data } = await response.json()

    // Extract tweet IDs
    const tweetIds = Object.keys(data.results)
      .map(key => key.split(':').pop())

    // Load full tweets
    const tweets = await Promise.all(
      tweetIds.map(id => this.getTweet(id))
    )

    return tweets.sort((a, b) => b.timestamp - a.timestamp)
  }

  // Get single tweet
  async getTweet(tweetId) {
    const response = await fetch(`${this.apiURL}/state/query/prefix`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        prefix: `post:${tweetId}:`,
        limit: 10
      })
    })

    const { data } = await response.json()

    const tweet = { id: tweetId }
    for (const [key, value] of Object.entries(data.results)) {
      const field = key.split(':')[2]
      tweet[field] = atob(value)
    }

    return tweet
  }

  // Get timeline (all tweets)
  async getTimeline(limit = 50) {
    const response = await fetch(`${this.apiURL}/state/query/prefix`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        prefix: 'index:timeline:',
        limit: limit
      })
    })

    const { data } = await response.json()

    // Extract tweet IDs and sort by timestamp (newest first)
    const entries = Object.keys(data.results)
      .map(key => {
        const parts = key.split(':')
        return {
          timestamp: parseInt(parts[2]),
          tweetId: parts[3]
        }
      })
      .sort((a, b) => b.timestamp - a.timestamp)

    // Load full tweets
    const tweets = await Promise.all(
      entries.slice(0, limit).map(entry => this.getTweet(entry.tweetId))
    )

    return tweets
  }

  // Like a tweet
  async likeTweet(tweetId) {
    const tweet = await this.getTweet(tweetId)
    const likes = parseInt(tweet.likes || '0') + 1

    const operations = [
      {
        type: 'SET',
        key: `post:${tweetId}:likes`,
        value: btoa(String(likes))
      }
    ]

    const tx = await this.createTransaction(operations)
    return await this.submitTransaction(tx)
  }

  // Helper methods
  async createTransaction(operations) {
    const tx = {
      from: this.address,
      timestamp: Math.floor(Date.now() / 1000),
      nonce: await this.getNonce(),
      data: { operations }
    }

    const txString = JSON.stringify({
      from: tx.from,
      timestamp: tx.timestamp,
      nonce: tx.nonce,
      data: tx.data
    })

    const txHash = ethers.utils.keccak256(ethers.utils.toUtf8Bytes(txString))
    tx.signature = await this.wallet.signMessage(ethers.utils.arrayify(txHash))

    return tx
  }

  async submitTransaction(tx) {
    const response = await fetch(`${this.apiURL}/transaction`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ transaction: tx })
    })

    return await response.json()
  }

  async getNonce() {
    // Implement nonce tracking
    return 0
  }
}

// Usage
const twitter = new DecentralizedTwitter('http://localhost:8545/api/v1', wallet)

// Create profile
await twitter.createProfile('alice', 'Blockchain developer', 'ipfs://Qm...')

// Post tweet
const tweetId = await twitter.postTweet('Hello decentralized world!')

// Get timeline
const timeline = await twitter.getTimeline(20)
console.log(timeline)

// Like tweet
await twitter.likeTweet(tweetId)
```

---

## Example 2: Decentralized File Storage

Store file metadata on-chain while files live on IPFS.

### Data Model

```
# Files
file:{hash}:owner
file:{hash}:name
file:{hash}:size
file:{hash}:type
file:{hash}:ipfs
file:{hash}:uploaded

# Indexes
index:owner:{address}:files:{hash}
index:type:{type}:files:{hash}
```

### Implementation

```javascript
class DecentralizedFileStorage {
  constructor(apiURL, wallet, ipfsClient) {
    this.apiURL = apiURL
    this.wallet = wallet
    this.ipfsClient = ipfsClient
    this.address = wallet.address
  }

  // Upload file
  async uploadFile(file) {
    // 1. Upload to IPFS
    const ipfsResult = await this.ipfsClient.add(file)
    const ipfsHash = ipfsResult.path

    // 2. Calculate file hash
    const fileHash = await this.calculateFileHash(file)

    // 3. Store metadata on blockchain
    const timestamp = Date.now()
    const dateStr = new Date(timestamp).toISOString().split('T')[0]

    const operations = [
      {
        type: 'SET',
        key: `file:${fileHash}:owner`,
        value: btoa(this.address)
      },
      {
        type: 'SET',
        key: `file:${fileHash}:name`,
        value: btoa(file.name)
      },
      {
        type: 'SET',
        key: `file:${fileHash}:size`,
        value: btoa(String(file.size))
      },
      {
        type: 'SET',
        key: `file:${fileHash}:type`,
        value: btoa(file.type)
      },
      {
        type: 'SET',
        key: `file:${fileHash}:ipfs`,
        value: btoa(ipfsHash)
      },
      {
        type: 'SET',
        key: `file:${fileHash}:uploaded`,
        value: btoa(String(timestamp))
      },
      // Index by owner
      {
        type: 'SET',
        key: `index:owner:${this.address}:files:${fileHash}`,
        value: btoa('1')
      },
      // Index by type
      {
        type: 'SET',
        key: `index:type:${file.type}:files:${fileHash}`,
        value: btoa('1')
      },
      // Index by date
      {
        type: 'SET',
        key: `index:date:${dateStr}:files:${fileHash}`,
        value: btoa('1')
      }
    ]

    const tx = await this.createTransaction(operations)
    await this.submitTransaction(tx)

    return {
      fileHash,
      ipfsHash,
      timestamp
    }
  }

  // Get file metadata
  async getFileMetadata(fileHash) {
    const response = await fetch(`${this.apiURL}/state/query/prefix`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        prefix: `file:${fileHash}:`,
        limit: 10
      })
    })

    const { data } = await response.json()

    const metadata = { fileHash }
    for (const [key, value] of Object.entries(data.results)) {
      const field = key.split(':')[2]
      metadata[field] = atob(value)
    }

    return metadata
  }

  // Get user's files
  async getUserFiles(address) {
    const response = await fetch(`${this.apiURL}/state/query/prefix`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        prefix: `index:owner:${address}:files:`,
        limit: 100
      })
    })

    const { data } = await response.json()

    const fileHashes = Object.keys(data.results)
      .map(key => key.split(':').pop())

    const files = await Promise.all(
      fileHashes.map(hash => this.getFileMetadata(hash))
    )

    return files.sort((a, b) => b.uploaded - a.uploaded)
  }

  // Download file
  async downloadFile(fileHash) {
    const metadata = await this.getFileMetadata(fileHash)

    // Get from IPFS
    const chunks = []
    for await (const chunk of this.ipfsClient.cat(metadata.ipfs)) {
      chunks.push(chunk)
    }

    return new Blob(chunks, { type: metadata.type })
  }

  // Helper methods
  async calculateFileHash(file) {
    const buffer = await file.arrayBuffer()
    const hashBuffer = await crypto.subtle.digest('SHA-256', buffer)
    const hashArray = Array.from(new Uint8Array(hashBuffer))
    return hashArray.map(b => b.toString(16).padStart(2, '0')).join('')
  }

  async createTransaction(operations) {
    // Same as Twitter example
  }

  async submitTransaction(tx) {
    // Same as Twitter example
  }
}

// Usage
const storage = new DecentralizedFileStorage(
  'http://localhost:8545/api/v1',
  wallet,
  ipfsClient
)

// Upload file
const file = document.getElementById('file-input').files[0]
const result = await storage.uploadFile(file)
console.log(`Uploaded: ${result.fileHash}`)

// Get user's files
const myFiles = await storage.getUserFiles(wallet.address)
console.log(myFiles)

// Download file
const blob = await storage.downloadFile(result.fileHash)
```

---

## Example 3: E-commerce Platform

Decentralized product catalog and orders.

### Data Model

```
# Products
product:{id}:name
product:{id}:price
product:{id}:stock
product:{id}:seller
product:{id}:category
product:{id}:description
product:{id}:image

# Orders
order:{id}:buyer
order:{id}:product
order:{id}:quantity
order:{id}:total
order:{id}:status
order:{id}:timestamp

# Indexes
index:seller:{address}:products:{id}
index:category:{category}:products:{id}
index:buyer:{address}:orders:{id}
```

### Implementation

```javascript
class DecentralizedMarketplace {
  constructor(apiURL, wallet) {
    this.apiURL = apiURL
    this.wallet = wallet
    this.address = wallet.address
  }

  // List product
  async listProduct(product) {
    const productId = Date.now()

    const operations = [
      {
        type: 'SET',
        key: `product:${productId}:name`,
        value: btoa(product.name)
      },
      {
        type: 'SET',
        key: `product:${productId}:price`,
        value: btoa(String(product.price))
      },
      {
        type: 'SET',
        key: `product:${productId}:stock`,
        value: btoa(String(product.stock))
      },
      {
        type: 'SET',
        key: `product:${productId}:seller`,
        value: btoa(this.address)
      },
      {
        type: 'SET',
        key: `product:${productId}:category`,
        value: btoa(product.category)
      },
      {
        type: 'SET',
        key: `product:${productId}:description`,
        value: btoa(product.description)
      },
      {
        type: 'SET',
        key: `product:${productId}:image`,
        value: btoa(product.image)
      },
      // Index by seller
      {
        type: 'SET',
        key: `index:seller:${this.address}:products:${productId}`,
        value: btoa('1')
      },
      // Index by category
      {
        type: 'SET',
        key: `index:category:${product.category}:products:${productId}`,
        value: btoa('1')
      }
    ]

    const tx = await this.createTransaction(operations)
    await this.submitTransaction(tx)

    return productId
  }

  // Get products by category
  async getProductsByCategory(category, limit = 50) {
    const response = await fetch(`${this.apiURL}/state/query/prefix`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        prefix: `index:category:${category}:products:`,
        limit: limit
      })
    })

    const { data } = await response.json()

    const productIds = Object.keys(data.results)
      .map(key => key.split(':').pop())

    return Promise.all(productIds.map(id => this.getProduct(id)))
  }

  // Create order
  async createOrder(productId, quantity) {
    const product = await this.getProduct(productId)
    const stock = parseInt(product.stock)

    if (stock < quantity) {
      throw new Error('Insufficient stock')
    }

    const orderId = Date.now()
    const total = parseFloat(product.price) * quantity
    const newStock = stock - quantity

    const operations = [
      // Create order
      {
        type: 'SET',
        key: `order:${orderId}:buyer`,
        value: btoa(this.address)
      },
      {
        type: 'SET',
        key: `order:${orderId}:product`,
        value: btoa(productId)
      },
      {
        type: 'SET',
        key: `order:${orderId}:quantity`,
        value: btoa(String(quantity))
      },
      {
        type: 'SET',
        key: `order:${orderId}:total`,
        value: btoa(String(total))
      },
      {
        type: 'SET',
        key: `order:${orderId}:status`,
        value: btoa('pending')
      },
      {
        type: 'SET',
        key: `order:${orderId}:timestamp`,
        value: btoa(String(Date.now()))
      },
      // Update stock
      {
        type: 'SET',
        key: `product:${productId}:stock`,
        value: btoa(String(newStock))
      },
      // Index order
      {
        type: 'SET',
        key: `index:buyer:${this.address}:orders:${orderId}`,
        value: btoa('1')
      }
    ]

    const tx = await this.createTransaction(operations)
    await this.submitTransaction(tx)

    return orderId
  }

  // Get user's orders
  async getUserOrders(address) {
    const response = await fetch(`${this.apiURL}/state/query/prefix`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        prefix: `index:buyer:${address}:orders:`,
        limit: 100
      })
    })

    const { data } = await response.json()

    const orderIds = Object.keys(data.results)
      .map(key => key.split(':').pop())

    return Promise.all(orderIds.map(id => this.getOrder(id)))
  }

  // Helper methods
  async getProduct(productId) {
    const response = await fetch(`${this.apiURL}/state/query/prefix`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        prefix: `product:${productId}:`,
        limit: 10
      })
    })

    const { data } = await response.json()

    const product = { id: productId }
    for (const [key, value] of Object.entries(data.results)) {
      const field = key.split(':')[2]
      product[field] = atob(value)
    }

    return product
  }

  async getOrder(orderId) {
    const response = await fetch(`${this.apiURL}/state/query/prefix`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        prefix: `order:${orderId}:`,
        limit: 10
      })
    })

    const { data } = await response.json()

    const order = { id: orderId }
    for (const [key, value] of Object.entries(data.results)) {
      const field = key.split(':')[2]
      order[field] = atob(value)
    }

    return order
  }

  async createTransaction(operations) {
    // Same as previous examples
  }

  async submitTransaction(tx) {
    // Same as previous examples
  }
}

// Usage
const marketplace = new DecentralizedMarketplace(
  'http://localhost:8545/api/v1',
  wallet
)

// List product
const productId = await marketplace.listProduct({
  name: 'Laptop',
  price: 999.99,
  stock: 10,
  category: 'electronics',
  description: 'High-performance laptop',
  image: 'ipfs://Qm...'
})

// Browse category
const electronics = await marketplace.getProductsByCategory('electronics')
console.log(electronics)

// Create order
const orderId = await marketplace.createOrder(productId, 2)
console.log(`Order created: ${orderId}`)

// View orders
const myOrders = await marketplace.getUserOrders(wallet.address)
console.log(myOrders)
```

---

## Next Steps

These examples demonstrate the power and flexibility of Podoru Chain for building decentralized applications. You can adapt these patterns for:

- Gaming platforms
- IoT data storage
- Identity management
- Content distribution
- And much more!

## Further Reading

- [Data Storage Patterns](data-patterns.md) - Key design best practices
- [Querying Data](querying.md) - Advanced query techniques
- [API Reference](../api-reference/README.md) - Complete API documentation
