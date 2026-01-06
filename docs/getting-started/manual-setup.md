# Manual Setup

This guide walks through manually configuring a Podoru Chain network without using the automated setup wizard.

## Overview

Manual setup gives you complete control over:
- Node configuration
- Network topology
- Genesis block contents
- Key management
- Deployment strategy

## Prerequisites

- [Podoru Chain installed](installation.md)
- Understanding of [Podoru Chain architecture](../architecture/README.md)
- Basic knowledge of YAML and JSON configuration

## Setup Steps

### Step 1: Generate Cryptographic Keys

Producer nodes require cryptographic keys for signing blocks.

```bash
# Create keys directory
mkdir -p keys

# Generate keys for each producer
./bin/keygen -output keys/producer1.key
./bin/keygen -output keys/producer2.key
./bin/keygen -output keys/producer3.key
```

Each command outputs:
```
Private key saved to: keys/producer1.key
Address: 0x3D4b25CBdda1014F74F9C80f040ce1Bb69130CBB
Public Key: 04a1b2c3d4...
```

**Important**: Save these addresses - you'll need them for the genesis file and node configurations.

### Step 2: Create Genesis Block

Create a `genesis.json` file defining the initial blockchain state:

```json
{
  "timestamp": 1704556800,
  "authorities": [
    "0x3D4b25CBdda1014F74F9C80f040ce1Bb69130CBB",
    "0x4aa37EEc2a26a4e04b7b206f32D6C2C63219F5cd",
    "0x304F73DD4CabF754eF2240fF2bC2446eB7709652"
  ],
  "initial_state": {
    "chain:name": "My Custom Chain",
    "chain:version": "1.0.0",
    "chain:description": "A custom blockchain for my application",
    "system:initialized": "true"
  }
}
```

**Field Explanations**:

- `timestamp`: Unix timestamp for genesis block (use current time or fixed value)
- `authorities`: Array of producer addresses (from Step 1)
- `initial_state`: Key-value pairs stored in genesis block

**Important**: All nodes must use the identical genesis file for the network to function.

### Step 3: Create Node Configurations

Create a YAML configuration file for each node.

#### Producer Node Configuration

Create `config/producer1.yaml`:

```yaml
# Node type: "producer" or "full"
node_type: producer

# Address and private key (from Step 1)
address: "0x3D4b25CBdda1014F74F9C80f040ce1Bb69130CBB"
private_key: "/path/to/keys/producer1.key"

# P2P Network Settings
p2p_port: 9000
p2p_bind_addr: "0.0.0.0"
bootstrap_peers:
  - "192.168.1.10:9001"  # Producer 2
  - "192.168.1.11:9002"  # Producer 3
max_peers: 50

# REST API Settings
api_enabled: true
api_port: 8545
api_bind_addr: "0.0.0.0"

# Storage Settings
data_dir: "./data/producer1"

# Consensus Settings
authorities:
  - "0x3D4b25CBdda1014F74F9C80f040ce1Bb69130CBB"
  - "0x4aa37EEc2a26a4e04b7b206f32D6C2C63219F5cd"
  - "0x304F73DD4CabF754eF2240fF2bC2446eB7709652"
block_time: 5s

# Genesis Settings
genesis_path: "./genesis.json"
```

Repeat for `producer2.yaml` and `producer3.yaml`, changing:
- `address` and `private_key`
- `p2p_port` (9001, 9002)
- `api_port` (8546, 8547)
- `data_dir` (data/producer2, data/producer3)
- `bootstrap_peers` (other nodes)

#### Full Node Configuration

Create `config/fullnode1.yaml`:

```yaml
# Full nodes don't need address/private_key
node_type: full

# P2P Network Settings
p2p_port: 9003
p2p_bind_addr: "0.0.0.0"
bootstrap_peers:
  - "192.168.1.10:9000"  # Producer 1
  - "192.168.1.11:9001"  # Producer 2
  - "192.168.1.12:9002"  # Producer 3
max_peers: 50

# REST API Settings
api_enabled: true
api_port: 8548
api_bind_addr: "0.0.0.0"

# Storage Settings
data_dir: "./data/fullnode1"

# Consensus Settings (same authorities as producers)
authorities:
  - "0x3D4b25CBdda1014F74F9C80f040ce1Bb69130CBB"
  - "0x4aa37EEc2a26a4e04b7b206f32D6C2C63219F5cd"
  - "0x304F73DD4CabF754eF2240fF2bC2446eB7709652"
block_time: 5s

# Genesis Settings
genesis_path: "./genesis.json"
```

### Step 4: Create Data Directories

```bash
# Create data directories
mkdir -p data/producer1
mkdir -p data/producer2
mkdir -p data/producer3
mkdir -p data/fullnode1

# Copy genesis file to each
cp genesis.json data/producer1/
cp genesis.json data/producer2/
cp genesis.json data/producer3/
cp genesis.json data/fullnode1/
```

### Step 5: Start the Nodes

Start each node in a separate terminal or as a background service.

#### Terminal 1 - Producer 1
```bash
./bin/podoru-node -config config/producer1.yaml
```

#### Terminal 2 - Producer 2
```bash
./bin/podoru-node -config config/producer2.yaml
```

#### Terminal 3 - Producer 3
```bash
./bin/podoru-node -config config/producer3.yaml
```

#### Terminal 4 - Full Node
```bash
./bin/podoru-node -config config/fullnode1.yaml
```

### Step 6: Verify the Network

Once all nodes are running:

```bash
# Check chain info
curl http://localhost:8545/api/v1/chain/info | jq

# Check connected peers
curl http://localhost:8545/api/v1/node/peers | jq

# Query genesis data
curl http://localhost:8545/api/v1/state/chain:name | jq
```

Expected output for chain info:
```json
{
  "success": true,
  "data": {
    "height": 10,
    "latest_block_hash": "0xabc123...",
    "authorities": [
      "0x3D4b25CBdda1014F74F9C80f040ce1Bb69130CBB",
      "0x4aa37EEc2a26a4e04b7b206f32D6C2C63219F5cd",
      "0x304F73DD4CabF754eF2240fF2bC2446eB7709652"
    ]
  }
}
```

## Docker Deployment

For production deployments, use Docker for better isolation and management.

### Create Dockerfile

The repository includes a Dockerfile, but you can customize it:

```dockerfile
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY . .
RUN go build -o podoru-node ./cmd/node

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/podoru-node .
EXPOSE 8545 9000
CMD ["./podoru-node", "-config", "/data/config.yaml"]
```

### Build Docker Image

```bash
make docker
# Or manually:
docker build -t podoru-chain:latest .
```

### Create Docker Compose File

Create `docker-compose.yml`:

```yaml
version: '3.8'

services:
  producer1:
    image: podoru-chain:latest
    container_name: podoru-producer1
    ports:
      - "8545:8545"
      - "9000:9000"
    volumes:
      - ./docker/data/producer1:/data
      - ./config/producer1.yaml:/data/config.yaml
      - ./keys/producer1.key:/data/keys/producer1.key
    networks:
      podoru_network:
        ipv4_address: 172.20.0.10

  producer2:
    image: podoru-chain:latest
    container_name: podoru-producer2
    ports:
      - "8546:8545"
      - "9001:9000"
    volumes:
      - ./docker/data/producer2:/data
      - ./config/producer2.yaml:/data/config.yaml
      - ./keys/producer2.key:/data/keys/producer2.key
    networks:
      podoru_network:
        ipv4_address: 172.20.0.11

  producer3:
    image: podoru-chain:latest
    container_name: podoru-producer3
    ports:
      - "8547:8545"
      - "9002:9000"
    volumes:
      - ./docker/data/producer3:/data
      - ./config/producer3.yaml:/data/config.yaml
      - ./keys/producer3.key:/data/keys/producer3.key
    networks:
      podoru_network:
        ipv4_address: 172.20.0.12

  fullnode1:
    image: podoru-chain:latest
    container_name: podoru-fullnode1
    ports:
      - "8548:8545"
      - "9003:9000"
    volumes:
      - ./docker/data/fullnode1:/data
      - ./config/fullnode1.yaml:/data/config.yaml
    networks:
      podoru_network:
        ipv4_address: 172.20.0.13

networks:
  podoru_network:
    driver: bridge
    ipam:
      config:
        - subnet: 172.20.0.0/16
```

### Start Docker Network

```bash
docker-compose up -d

# View logs
docker-compose logs -f

# Stop network
docker-compose down
```

## Advanced Configuration

### Custom Block Time

Adjust consensus speed:

```yaml
block_time: 3s  # Faster blocks (more overhead)
block_time: 10s # Slower blocks (more efficient)
```

### Multiple Networks

Run separate networks on the same machine:

```yaml
# Network 1
api_port: 8545
p2p_port: 9000

# Network 2 (different ports)
api_port: 9545
p2p_port: 10000
```

### External Nodes

Connect nodes across different machines:

```yaml
# Node on Machine A (192.168.1.10)
p2p_bind_addr: "0.0.0.0"
p2p_port: 9000

# Node on Machine B
bootstrap_peers:
  - "192.168.1.10:9000"  # Connect to Machine A
```

### Production Security

For production deployments:

```yaml
# Bind API only to localhost (use reverse proxy)
api_bind_addr: "127.0.0.1"

# Limit max peers
max_peers: 20

# Use firewall rules for P2P port
# Only allow known peer IPs
```

### Key Security

Protect private keys:

```bash
# Encrypt key file
openssl enc -aes-256-cbc -in producer1.key -out producer1.key.enc

# Set strict permissions
chmod 600 keys/*.key

# Use environment variables for paths
export PRODUCER1_KEY=/secure/path/producer1.key
```

## Running as System Service

### Systemd Service (Linux)

Create `/etc/systemd/system/podoru-producer1.service`:

```ini
[Unit]
Description=Podoru Chain Producer 1
After=network.target

[Service]
Type=simple
User=podoru
Group=podoru
WorkingDirectory=/opt/podoru-chain
ExecStart=/opt/podoru-chain/bin/podoru-node -config /opt/podoru-chain/config/producer1.yaml
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
```

Enable and start:

```bash
sudo systemctl enable podoru-producer1
sudo systemctl start podoru-producer1
sudo systemctl status podoru-producer1
```

## Monitoring

### Health Checks

```bash
# Node health endpoint
curl http://localhost:8545/api/v1/node/health

# Chain height
curl http://localhost:8545/api/v1/chain/info | jq '.data.height'
```

### Log Management

```bash
# Redirect logs to file
./bin/podoru-node -config config.yaml > logs/node.log 2>&1 &

# Rotate logs with logrotate
/var/log/podoru/*.log {
    daily
    rotate 7
    compress
    delaycompress
}
```

## Backup and Recovery

### Backup Blockchain Data

```bash
# Stop node
sudo systemctl stop podoru-producer1

# Backup data directory
tar -czf backup-$(date +%Y%m%d).tar.gz data/

# Restart node
sudo systemctl start podoru-producer1
```

### Restore from Backup

```bash
# Stop node
sudo systemctl stop podoru-producer1

# Restore data
tar -xzf backup-20240106.tar.gz

# Restart node
sudo systemctl start podoru-producer1
```

## Troubleshooting

See the [Troubleshooting Guide](../troubleshooting/README.md) for common issues.

## Next Steps

- [Configuration Reference](../configuration/README.md) - Detailed configuration options
- [CLI Reference](../cli-reference/README.md) - Command-line tools
- [Development Guide](../development/README.md) - Build applications
- [API Reference](../api-reference/README.md) - REST API documentation
