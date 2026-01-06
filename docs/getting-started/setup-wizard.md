# Setup Wizard

The Podoru Chain Setup Wizard provides an interactive, guided experience to configure and launch your blockchain network.

## Overview

The setup wizard automates:
- Network configuration (node counts, ports)
- Chain metadata setup
- Cryptographic key generation
- Genesis block creation
- Docker image building
- Network startup

## Prerequisites

Before running the wizard, ensure you have:

- [Installed all prerequisites](installation.md)
- Cloned the Podoru Chain repository
- `dialog` package installed for the interactive UI

## Running the Wizard

```bash
cd podoru-chain
make setup-wizard
```

## Wizard Steps

### 1. Welcome Screen

The wizard displays a welcome message and overview of what will be configured.

**Action**: Press Enter to continue

### 2. Node Configuration

Configure the number of producer and full nodes:

**Producer Nodes (1-10)**
- These nodes create new blocks
- Recommended: 3+ for production (provides fault tolerance)
- Minimum: 1 for development

**Full Nodes (0-5)**
- These nodes validate blocks but don't produce them
- Optional but recommended for production
- Useful for providing additional API endpoints

**Default**: 3 producers, 1 full node

### 3. Chain Metadata

Configure your blockchain's identity:

**Chain Name**
- Human-readable name for your blockchain
- Default: "Podoru Chain"
- Example: "My Custom Blockchain"

**Chain Description**
- Brief description of your blockchain's purpose
- Default: "Decentralized blockchain for storing any data"
- Stored in genesis block under `chain:description`

**Chain Version**
- Semantic version for your blockchain
- Default: "1.0.0"
- Stored in genesis block under `chain:version`

### 4. Consensus Configuration

**Block Time**
- Time between blocks in seconds
- Default: 5 seconds
- Range: 1-60 seconds
- Lower = faster but more network overhead
- Higher = slower but more efficient

**Recommendations**:
- Development: 3-5 seconds
- Production: 5-10 seconds
- High-throughput: 2-3 seconds

### 5. Network Configuration

Configure network ports for your nodes:

**API Starting Port**
- First node uses this port, subsequent nodes increment by 1
- Default: 8545
- Producer 1: 8545, Producer 2: 8546, etc.

**P2P Starting Port**
- First node uses this port, subsequent nodes increment by 1
- Default: 9000
- Producer 1: 9000, Producer 2: 9001, etc.

**Note**: Ensure these port ranges are available on your system.

### 6. Advanced Options

**Skip Docker Build**
- Skip building the Docker image (use existing image)
- Default: No
- Use "Yes" if image is already built

**Auto-start Network**
- Automatically start the network after setup
- Default: Yes
- If "No", you'll need to run `make docker-compose-up` manually

### 7. Confirmation

Review your configuration before proceeding:

```
Configuration Summary:
- Producer Nodes: 3
- Full Nodes: 1
- Chain Name: Podoru Chain
- Block Time: 5s
- API Ports: 8545-8548
- P2P Ports: 9000-9003
```

**Action**: Confirm to proceed with setup

### 8. Automated Setup

The wizard now performs:

1. **Directory Setup** - Creates `docker/data/` directories
2. **Key Generation** - Generates keypairs for all producers
3. **Genesis Creation** - Creates synchronized genesis block
4. **Docker Compose** - Generates docker-compose.yml
5. **Docker Build** - Builds the Docker image (if not skipped)
6. **Network Start** - Starts all nodes (if auto-start enabled)

### 9. Completion

The wizard displays:
- Success message
- Generated node addresses
- API endpoints for each node
- Management commands

## Generated Files

The wizard creates the following structure:

```
docker/
├── data/
│   ├── producer1/
│   │   ├── genesis.json
│   │   └── keys/
│   │       └── producer1.key
│   ├── producer2/
│   │   ├── genesis.json
│   │   └── keys/
│   │       └── producer2.key
│   ├── producer3/
│   │   ├── genesis.json
│   │   └── keys/
│   │       └── producer3.key
│   └── fullnode1/
│       └── genesis.json
└── docker-compose.yml
```

## Post-Setup Verification

After the wizard completes, verify your network:

### Check Running Containers

```bash
docker ps
```

Expected output:
```
CONTAINER ID   IMAGE              PORTS                    NAMES
abc123...      podoru-chain      0.0.0.0:8545->8545/tcp   podoru-producer1
def456...      podoru-chain      0.0.0.0:8546->8545/tcp   podoru-producer2
ghi789...      podoru-chain      0.0.0.0:8547->8545/tcp   podoru-producer3
jkl012...      podoru-chain      0.0.0.0:8548->8545/tcp   podoru-fullnode1
```

### Test API Endpoints

```bash
# Check chain info
curl http://localhost:8545/api/v1/chain/info | jq

# Query genesis data
curl http://localhost:8545/api/v1/state/chain:name | jq
```

### View Logs

```bash
# All nodes
make docker-compose-logs

# Specific node
docker logs podoru-producer1 -f
```

## Managing Your Network

### Start Network

```bash
make docker-compose-up
```

### Stop Network

```bash
make docker-compose-down
```

### Restart Single Node

```bash
docker restart podoru-producer1
```

### View Node Logs

```bash
docker logs podoru-producer1 -f
```

## Customization After Setup

### Modify Configuration

Edit generated configuration files:

```bash
# For Docker setup
vi docker/docker-compose.yml

# For manual setup
vi config/producer1.yaml
```

### Add More Nodes

To add additional nodes after initial setup:

1. Generate new keys: `./bin/keygen -output docker/data/producer4/keys/producer4.key`
2. Add the address to `authorities` in all genesis files
3. Create configuration for the new node
4. Update docker-compose.yml
5. Restart the network

### Change Block Time

Edit the `block_time` in each node's configuration and restart.

## Re-running the Wizard

To reconfigure your network:

```bash
# Stop and remove existing network
make docker-compose-down

# Remove existing data (CAUTION: This deletes blockchain data)
rm -rf docker/data/

# Run wizard again
make setup-wizard
```

## Troubleshooting

### Wizard Crashes or Freezes

```bash
# Ensure dialog is installed
sudo apt-get install dialog  # Ubuntu/Debian
brew install dialog           # macOS

# Check terminal size (minimum 24x80)
echo $LINES $COLUMNS
```

### Port Conflicts

If you see "port already in use":

```bash
# Find and kill process using the port
sudo lsof -i :8545
kill <PID>

# Or use different ports in wizard configuration
```

### Docker Build Fails

```bash
# Check Docker is running
docker ps

# Try building manually
make docker

# Check Docker logs
docker logs podoru-producer1
```

### Network Won't Start

```bash
# Check Docker Compose file
cat docker/docker-compose.yml

# Try starting manually with verbose output
docker-compose -f docker/docker-compose.yml up
```

## Advanced Usage

### Environment Variables

Override wizard defaults with environment variables:

```bash
# Custom ports
API_START_PORT=8000 P2P_START_PORT=9100 make setup-wizard

# Custom block time
BLOCK_TIME=10 make setup-wizard
```

### Wizard in Non-Interactive Mode

For automated deployments, use the manual setup scripts:

```bash
# See manual setup guide
./scripts/setup.sh --producers=3 --fullnodes=1 --blocktime=5s
```

## Next Steps

- [Manual Setup](manual-setup.md) - Manual configuration guide
- [Configuration Reference](../configuration/README.md) - Detailed configuration options
- [Development Guide](../development/README.md) - Start building applications
- [API Reference](../api-reference/README.md) - Explore the REST API
