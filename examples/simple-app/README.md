# Simple Podoru Chain App Example

This is a complete example showing how to **read and write** data on Podoru Chain.

## Prerequisites

1. Node.js 18+ installed
2. Podoru Chain running on `localhost:8545`

## Run the Example

```bash
# Start Podoru Chain
cd ../..
make docker-compose-up

# Wait for blockchain to start, then install dependencies and run
cd examples/simple-app
npm install
node app.js
```

## What It Does

This example demonstrates:

### Reading Data
1. **Chain Info Query** - Get blockchain metadata
2. **Single Key Query** - Get one value
3. **Batch Query** - Get multiple keys at once
4. **Prefix Query** - Get all keys matching a pattern

### Writing Data
5. **Create Transactions** - Submit signed transactions to the blockchain
6. **Set Key-Value Pairs** - Store data on-chain
7. **Batch Operations** - Multiple operations in one transaction
8. **User Profiles** - Create and update user data

## Sample Output

```
🚀 Podoru Chain - Simple App Example
==================================================

🔗 Blockchain Information:
==================================================
Height: 5
Current Hash: 0xabcd...
Genesis Hash: 0x1234...
Authorities: 3
  1. 0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb1
  2. 0x8626f6940E2eb28930eFb4CeF49B2d1F2C9C1199
  3. 0xdD2FD4581271e230360230F9337D5c0430Bf44C0

📦 Genesis State Data:
==================================================
name: Podoru Chain
version: 1.0.0
description: Decentralized blockchain for storing any data

✅ Demo complete!
```

## Building Your Own App

Use this as a starting point! Key functions:

### Reading Data

```javascript
// Get single value
const name = await get('chain:name');

// Get multiple values
const data = await getBatch(['chain:name', 'chain:version']);

// Query by prefix (get all user data)
const userData = await queryPrefix('user:alice:');

// Get blockchain info
const info = await getChainInfo();
```

### Writing Data

```javascript
// Set a single key-value pair
await set('myapp:setting', 'enabled');

// Set multiple key-value pairs in one transaction
await setBatch({
  'user:bob:name': 'Bob Smith',
  'user:bob:email': 'bob@example.com'
});

// Delete a key
await del('myapp:temp');

// Create a user profile (helper function)
await createUserProfile('alice', {
  name: 'Alice Johnson',
  email: 'alice@example.com',
  bio: 'Blockchain enthusiast'
});

// Update a single field
await updateUserField('alice', 'bio', 'Podoru Chain developer');
```

### Advanced: Direct Transaction Submission

```javascript
// Submit a custom transaction with multiple operations
await submitTransaction([
  { type: 'SET', key: 'app:counter', value: '42' },
  { type: 'SET', key: 'app:status', value: 'active' },
  { type: 'DELETE', key: 'app:old_data' }
]);
```

## How It Works

1. **Transaction Signing**: The app generates a random Ethereum-compatible wallet for testing
2. **Transaction Format**: Transactions are JSON objects with operations (SET/DELETE)
3. **Signature**: Uses secp256k1 ECDSA signatures (same as Ethereum)
4. **Submission**: POSTs signed transactions to `/api/v1/transaction`
5. **Block Production**: Transactions are included in the next block (3 second intervals)

See [docs/APP_DEVELOPMENT.md](../../docs/APP_DEVELOPMENT.md) for complete guide.
