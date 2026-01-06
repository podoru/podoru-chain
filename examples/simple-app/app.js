#!/usr/bin/env node

/**
 * Simple Podoru Chain Application Example
 *
 * This demonstrates how to build apps on Podoru Chain
 * Run: npm install && node app.js
 */

const crypto = require('crypto');
const { ethers } = require('ethers');
const API_URL = 'http://localhost:8545';

// Simple HTTP client
async function request(endpoint, options = {}) {
  const url = `${API_URL}${endpoint}`;
  const response = await fetch(url, options);
  return await response.json();
}

// Get a single key
async function get(key) {
  const result = await request(`/api/v1/state/${key}`);
  if (result.success && result.data) {
    return Buffer.from(result.data.value, 'base64').toString('utf-8');
  }
  return null;
}

// Get multiple keys at once
async function getBatch(keys) {
  const result = await request('/api/v1/state/batch', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ keys })
  });

  if (result.success && result.data) {
    const decoded = {};
    for (const [key, value] of Object.entries(result.data)) {
      if (value) {
        decoded[key] = Buffer.from(value, 'base64').toString('utf-8');
      }
    }
    return decoded;
  }
  return {};
}

// Query by prefix
async function queryPrefix(prefix, limit = 100) {
  const result = await request('/api/v1/state/query/prefix', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ prefix, limit })
  });

  if (result.success && result.data && result.data.results) {
    const decoded = {};
    for (const [key, value] of Object.entries(result.data.results)) {
      decoded[key] = Buffer.from(value, 'base64').toString('utf-8');
    }
    return decoded;
  }
  return {};
}

// Get chain info
async function getChainInfo() {
  const result = await request('/api/v1/chain/info');
  return result.success ? result.data : null;
}

// ============================================
// Transaction Creation Functions
// ============================================

// Create a test wallet (for demo purposes only!)
// In production, load this from a secure location
let testWallet = null;

function getTestWallet() {
  if (!testWallet) {
    // Generate a random wallet for testing
    // WARNING: This is for demo only - keys are not persisted!
    testWallet = ethers.Wallet.createRandom();
    console.log(`\n🔑 Using test wallet: ${testWallet.address}`);
    console.log(`   Private key: ${testWallet.privateKey}`);
    console.log(`   (This is a temporary test wallet)\n`);
  }
  return testWallet;
}

// Create and sign a transaction
async function createTransaction(operations) {
  const wallet = getTestWallet();

  const from = wallet.address;
  const timestamp = Math.floor(Date.now() / 1000);
  const nonce = Date.now(); // Simple nonce

  // Build transaction data
  const data = {
    operations: operations.map(op => ({
      type: op.type || 'SET',
      key: op.key,
      value: op.value ? Buffer.from(op.value).toString('base64') : null
    }))
  };

  // Create transaction hash (this must match the Go implementation)
  // Go marshals in struct field order: From, Timestamp, Data, Nonce
  const hashData = {
    from,
    timestamp,
    data,
    nonce
  };
  const hashString = JSON.stringify(hashData);
  const hash = crypto.createHash('sha256').update(hashString).digest();

  // For debugging - show what we're hashing
  if (process.env.DEBUG) {
    console.log('Hash input:', hashString);
    console.log('Hash output:', hash.toString('hex'));
  }

  // Sign the hash using ethers (produces recoverable ECDSA signature)
  const hashHex = '0x' + hash.toString('hex');
  const signature = await wallet.signingKey.sign(hashHex);

  // Convert signature to the format go-ethereum expects: r + s + v (65 bytes)
  // v must be the recovery ID (0 or 1), not 27/28
  const r = Buffer.from(signature.r.slice(2), 'hex');
  const s = Buffer.from(signature.s.slice(2), 'hex');
  const v = Buffer.from([signature.v - 27]); // Normalize v to 0 or 1

  const sigBytes = Array.from(Buffer.concat([r, s, v]));

  // Create transaction object
  const tx = {
    id: Array.from(hash),
    from,
    timestamp,
    nonce,
    data,
    signature: sigBytes
  };

  return tx;
}

// Submit a transaction to the blockchain
async function submitTransaction(operations) {
  const tx = await createTransaction(operations);

  const result = await request('/api/v1/transaction', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ transaction: tx })
  });

  if (result.success) {
    return result.data;
  } else {
    throw new Error(result.error || 'Transaction failed');
  }
}

// Helper function to set a single key-value pair
async function set(key, value) {
  return await submitTransaction([
    { type: 'SET', key, value }
  ]);
}

// Helper function to set multiple key-value pairs
async function setBatch(kvPairs) {
  const operations = Object.entries(kvPairs).map(([key, value]) => ({
    type: 'SET',
    key,
    value
  }));
  return await submitTransaction(operations);
}

// Helper function to delete a key
async function del(key) {
  return await submitTransaction([
    { type: 'DELETE', key, value: null }
  ]);
}

// Example: Query user profile
async function getUserProfile(username) {
  console.log(`\n📋 Getting profile for: ${username}`);

  // Get all user data using prefix query
  const userData = await queryPrefix(`user:${username}:`);

  if (Object.keys(userData).length === 0) {
    console.log(`❌ User '${username}' not found`);
    return null;
  }

  console.log(`✅ Found user data:`);
  for (const [key, value] of Object.entries(userData)) {
    const field = key.split(':').pop();
    console.log(`   ${field}: ${value}`);
  }

  return userData;
}

// Example: Get chain data
async function showChainData() {
  console.log('\n🔗 Blockchain Information:');
  console.log('='.repeat(50));

  const chainInfo = await getChainInfo();
  if (chainInfo) {
    console.log(`Height: ${chainInfo.height}`);
    console.log(`Current Hash: ${chainInfo.current_hash}`);
    console.log(`Genesis Hash: ${chainInfo.genesis_hash}`);
    console.log(`Authorities: ${chainInfo.authorities.length}`);
    chainInfo.authorities.forEach((auth, i) => {
      console.log(`  ${i + 1}. ${auth}`);
    });
  }
}

// Example: Query initial state
async function showGenesisData() {
  console.log('\n📦 Genesis State Data:');
  console.log('='.repeat(50));

  const chainData = await queryPrefix('chain:');

  for (const [key, value] of Object.entries(chainData)) {
    const field = key.replace('chain:', '');
    console.log(`${field}: ${value}`);
  }
}

// Example: Batch query
async function batchQueryExample() {
  console.log('\n🔄 Batch Query Example:');
  console.log('='.repeat(50));

  const keys = ['chain:name', 'chain:version', 'chain:description'];
  console.log(`Querying keys: ${keys.join(', ')}`);

  const results = await getBatch(keys);

  for (const [key, value] of Object.entries(results)) {
    console.log(`${key}: ${value}`);
  }
}

// Example: Create a user profile
async function createUserProfile(username, profileData) {
  console.log(`\n✍️  Creating profile for: ${username}`);

  // Convert profile data to key-value pairs with user prefix
  const kvPairs = {};
  for (const [field, value] of Object.entries(profileData)) {
    kvPairs[`user:${username}:${field}`] = value;
  }

  try {
    const result = await setBatch(kvPairs);
    console.log(`✅ Profile created! Transaction: ${result.transaction_hash}`);
    console.log(`   Status: ${result.status}`);
    return result;
  } catch (error) {
    console.log(`❌ Failed to create profile: ${error.message}`);
    throw error;
  }
}

// Example: Update a single field
async function updateUserField(username, field, value) {
  console.log(`\n📝 Updating ${username}'s ${field}...`);

  try {
    const result = await set(`user:${username}:${field}`, value);
    console.log(`✅ Updated! Transaction: ${result.transaction_hash}`);
    return result;
  } catch (error) {
    console.log(`❌ Failed to update: ${error.message}`);
    throw error;
  }
}

// Main demo
async function main() {
  console.log('🚀 Podoru Chain - Simple App Example');
  console.log('='.repeat(50));

  try {
    // Show blockchain info
    await showChainData();

    // Show genesis data
    await showGenesisData();

    // Batch query example
    await batchQueryExample();

    // === TRANSACTION EXAMPLES ===
    console.log('\n📝 Transaction Examples:');
    console.log('='.repeat(50));

    // Example 1: Create a user profile
    await createUserProfile('alice', {
      name: 'Alice Johnson',
      email: 'alice@example.com',
      bio: 'Blockchain enthusiast'
    });

    // Wait a moment for the block to be produced
    console.log('\n⏳ Waiting for block to be produced...');
    await new Promise(resolve => setTimeout(resolve, 5000));

    // Query the profile we just created
    await getUserProfile('alice');

    // Example 2: Update a single field
    await updateUserField('alice', 'bio', 'Podoru Chain developer');

    // Wait again
    await new Promise(resolve => setTimeout(resolve, 5000));

    // Query again to see the update
    await getUserProfile('alice');

    console.log('\n✅ Demo complete!');
    console.log('\n💡 What you learned:');
    console.log('   ✓ Query blockchain data (GET)');
    console.log('   ✓ Submit transactions (POST)');
    console.log('   ✓ Create user profiles');
    console.log('   ✓ Update individual fields');
    console.log('\n📖 See docs/APP_DEVELOPMENT.md for more examples');

  } catch (error) {
    console.error('\n❌ Error:', error.message);
    console.error('\nMake sure Podoru Chain is running on', API_URL);
    console.error('Start it with: make docker-compose-up');
  }
}

// Run if called directly
if (require.main === module) {
  main();
}

// Export for use as a module
module.exports = {
  get,
  getBatch,
  queryPrefix,
  getChainInfo,
  getUserProfile,
  set,
  setBatch,
  del,
  submitTransaction,
  createUserProfile,
  updateUserField
};
