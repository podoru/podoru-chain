# Simple Podoru Chain App Example

This is a basic example showing how to query data from Podoru Chain.

## Prerequisites

1. Node.js 18+ installed
2. Podoru Chain running on `localhost:8545`

## Run the Example

```bash
# Start Podoru Chain
cd ../..
make docker-compose-up

# Wait for blockchain to start, then run the example
cd examples/simple-app
node app.js
```

## What It Does

This example demonstrates:

1. **Chain Info Query** - Get blockchain metadata
2. **Single Key Query** - Get one value
3. **Batch Query** - Get multiple keys at once
4. **Prefix Query** - Get all keys matching a pattern

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

See [docs/APP_DEVELOPMENT.md](../../docs/APP_DEVELOPMENT.md) for complete guide.
