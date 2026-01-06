# Podoru Chain

A fully decentralized blockchain platform built in Go for storing arbitrary data as key-value pairs, powered by Proof of Authority (PoA) consensus.

## Features

- **Decentralized Architecture**: Fully distributed P2P network
- **Flexible Data Storage**: Store any data as key-value pairs
- **Proof of Authority Consensus**: Fast and efficient block production with round-robin selection
- **Two Node Types**:
  - **Full Nodes**: Validate all blocks and maintain complete blockchain state
  - **Producer Nodes**: Create new blocks when authorized
- **BadgerDB Storage**: High-performance embedded key-value database
- **REST API**: Query blockchain data and submit transactions
- **Docker Ready**: Multi-node setup with Docker Compose

## Architecture

- **Consensus**: Proof of Authority (PoA) with configurable block producers
- **Database**: BadgerDB for persistent storage
- **Networking**: TCP-based P2P with gossip protocol
- **API**: RESTful HTTP interface

## Quick Start with Setup Wizard (Recommended)

The easiest way to set up your Podoru Chain network is using the interactive setup wizard:

### Prerequisites

- Go 1.24 or higher
- Docker and Docker Compose
- dialog (for interactive UI)
  - Ubuntu/Debian: `sudo apt-get install dialog`
  - macOS: `brew install dialog`
  - Fedora/RHEL: `sudo dnf install dialog`

### Setup

1. Clone the repository:
```bash
git clone https://github.com/podoru/podoru-chain.git
cd podoru-chain
```

2. Run the setup wizard:
```bash
make setup-wizard
```

3. Follow the interactive prompts to configure:
   - Number of producer nodes (1-10)
   - Number of full nodes (0-5)
   - Chain metadata (name, description, block time)
   - Network ports (API and P2P starting ports)
   - Advanced options (skip Docker build, auto-start network)

4. The wizard will automatically:
   - Generate cryptographic keys for all producers
   - Create configuration files for all nodes
   - Generate a synchronized genesis block
   - Build the Docker image
   - Optionally start the network

5. Once complete, test your blockchain:
```bash
# Check blockchain info
curl http://localhost:8545/api/v1/chain/info | jq

# Query genesis data
curl http://localhost:8545/api/v1/state/chain:name | jq
```

### Management Commands

```bash
make docker-compose-up      # Start the network
make docker-compose-down    # Stop the network
make docker-compose-logs    # View logs from all nodes
```

## Manual Setup

For advanced users who prefer manual configuration:

### Prerequisites

- Go 1.24 or higher
- Docker and Docker Compose (for multi-node setup)
- Make (optional, for convenience commands)

### Installation

1. Clone the repository:
```bash
git clone https://github.com/podoru/podoru-chain.git
cd podoru-chain
```

2. Install dependencies:
```bash
make deps
```

3. Build the binaries:
```bash
make build
```

This creates two binaries in `bin/`:
- `podoru-node` - The main blockchain node
- `keygen` - Key generation utility

## Running a Multi-Node Network

### Using Docker Compose (Recommended)

1. Setup Docker data directories and generate keys:
```bash
make setup-docker
make keygen-all
```

2. Start the network (3 producers + 1 full node):
```bash
make docker-compose-up
```

3. View logs:
```bash
make docker-compose-logs
```

4. Stop the network:
```bash
make docker-compose-down
```

The network will have the following nodes:
- Producer 1: API on port 8545, P2P on port 9000
- Producer 2: API on port 8546, P2P on port 9001
- Producer 3: API on port 8547, P2P on port 9002
- Full Node: API on port 8548, P2P on port 9003

### Running a Single Node Locally

1. Generate a key pair:
```bash
./bin/keygen -output mynode.key
```

2. Create a configuration file (see `config/producer1.yaml` for example)

3. Run the node:
```bash
make run CONFIG=path/to/config.yaml
```

Or directly:
```bash
./bin/podoru-node -config path/to/config.yaml
```

## Configuration

### Node Configuration

Create a YAML configuration file with the following structure:

```yaml
node_type: producer  # or "full"
address: "0xYourAddress"
private_key: "/path/to/private/key"

p2p_port: 9000
p2p_bind_addr: "0.0.0.0"
bootstrap_peers:
  - "peer1:9000"
  - "peer2:9000"

api_enabled: true
api_port: 8545
api_bind_addr: "0.0.0.0"

data_dir: "./data"

authorities:
  - "0xAuthority1Address"
  - "0xAuthority2Address"
  - "0xAuthority3Address"

block_time: 5s
genesis_path: "./genesis.json"
```

### Genesis Configuration

Define the initial blockchain state in `genesis.json`:

```json
{
  "timestamp": 1704556800,
  "authorities": [
    "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb1",
    "0x8626f6940E2eb28930eFb4CeF49B2d1F2C9C1199",
    "0xdD2FD4581271e230360230F9337D5c0430Bf44C0"
  ],
  "initial_state": {
    "chain:name": "Podoru Chain",
    "chain:version": "1.0.0"
  }
}
```

## REST API

### Endpoints

#### Chain Information
- `GET /api/v1/chain/info` - Get blockchain info (height, hash, authorities)
- `GET /api/v1/block/{hash}` - Get block by hash
- `GET /api/v1/block/height/{height}` - Get block by height
- `GET /api/v1/block/latest` - Get latest block

#### Transactions
- `GET /api/v1/transaction/{hash}` - Get transaction by hash
- `POST /api/v1/transaction` - Submit a new transaction

#### State (Data Storage)
- `GET /api/v1/state/{key}` - Get value for a single key
- `POST /api/v1/state/batch` - **NEW!** Get multiple keys at once
- `POST /api/v1/state/query/prefix` - **NEW!** Query all keys with a prefix

#### Node
- `GET /api/v1/node/info` - Get node information
- `GET /api/v1/node/peers` - Get connected peers
- `GET /api/v1/node/health` - Health check

#### Mempool
- `GET /api/v1/mempool` - Get pending transactions

### Examples

**Get Chain Info:**
```bash
curl http://localhost:8545/api/v1/chain/info
```

**Get State Value:**
```bash
curl http://localhost:8545/api/v1/state/chain:name
```

**Batch Query (Get Multiple Keys):**
```bash
curl -X POST http://localhost:8545/api/v1/state/batch \
  -H "Content-Type: application/json" \
  -d '{"keys": ["chain:name", "chain:version"]}'
```

**Prefix Query (Get All Keys Matching Pattern):**
```bash
curl -X POST http://localhost:8545/api/v1/state/query/prefix \
  -H "Content-Type: application/json" \
  -d '{"prefix": "user:alice:", "limit": 100}'
```

**Submit Transaction:**
```bash
curl -X POST http://localhost:8545/api/v1/transaction \
  -H "Content-Type: application/json" \
  -d '{
    "transaction": {
      "from": "0xYourAddress",
      "timestamp": 1704556800,
      "data": {
        "operations": [
          {
            "type": "SET",
            "key": "mykey",
            "value": "bXl2YWx1ZQ=="
          }
        ]
      },
      "nonce": 0,
      "signature": "0xSignature..."
    }
  }'
```

## Development

### Project Structure

```
podoru-chain/
├── cmd/                     # Command-line applications
│   ├── node/               # Main node executable
│   └── tools/keygen/       # Key generation tool
├── internal/               # Private application code
│   ├── blockchain/         # Core blockchain logic
│   ├── consensus/          # PoA consensus engine
│   ├── crypto/             # Cryptographic operations
│   ├── storage/            # BadgerDB integration
│   ├── network/            # P2P networking
│   ├── api/rest/           # REST API
│   └── node/               # Node orchestration
├── config/                 # Configuration files
├── docker/                 # Docker setup
├── Makefile               # Build automation
└── README.md              # This file
```

### Building

```bash
# Build everything
make build

# Build Docker image
make docker

# Format code
make fmt

# Run tests
make test

# Run tests with coverage
make test-coverage
```

### Testing

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage
```

## How It Works

### Consensus (Proof of Authority)

- Block producers are pre-configured in the genesis file
- Producers take turns creating blocks using round-robin selection
- Each block is signed by its producer
- All nodes validate block signatures and ensure the correct producer created the block

### Data Storage

- Data is stored as key-value pairs in transactions
- Each transaction can contain multiple operations (SET or DELETE)
- State is maintained in-memory and persisted to BadgerDB
- State root is calculated using a Merkle tree

### Block Production

1. Producer checks if it's their turn (based on block height)
2. Collects pending transactions from mempool
3. Creates block with transactions
4. Calculates Merkle roots (transactions and state)
5. Signs block with private key
6. Broadcasts to network

### Synchronization

- Nodes sync blockchain on startup
- Query peers for their heights
- Request missing blocks in batches
- Validate and apply blocks sequentially

## Security Considerations

- Private keys are stored encrypted at rest
- Transactions require valid signatures
- Nonce prevents replay attacks
- Maximum transaction and block sizes prevent DOS
- Only authorized producers can create blocks

## License

MIT License - see LICENSE file for details

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Building Applications

Podoru Chain is designed for **application development**! See the [App Development Guide](docs/APP_DEVELOPMENT.md) for:

- How to structure data with key-value pairs
- Query patterns (single, batch, prefix)
- Example applications (Twitter clone, file storage, user profiles)
- Best practices and design patterns
- JavaScript/TypeScript SDK examples

**Key Features for Apps:**
- ✅ Flexible key-value storage for any data type
- ✅ Fast batch and prefix queries
- ✅ REST API for easy integration
- ✅ No smart contracts needed
- ✅ Fully decentralized data storage

## Support

For issues and questions, please open an issue on GitHub.
