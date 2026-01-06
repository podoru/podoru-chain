#!/usr/bin/env node

/**
 * Simple Podoru Chain Application Example
 *
 * This demonstrates how to build apps on Podoru Chain
 * Run: node app.js
 */

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

    // Try to get a user profile (will be empty initially)
    await getUserProfile('alice');

    console.log('\n✅ Demo complete!');
    console.log('\n💡 Next steps:');
    console.log('   1. Submit a transaction to store data');
    console.log('   2. Query it back using the examples above');
    console.log('   3. Build your own app on Podoru Chain!');
    console.log('\n   See docs/APP_DEVELOPMENT.md for more examples');

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
  getUserProfile
};
