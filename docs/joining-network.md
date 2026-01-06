# Joining an Existing Podoru Chain Network

This guide explains how others can join your Podoru Chain network as full nodes.

## Overview

Podoru Chain uses a **Proof of Authority (PoA)** consensus mechanism where:
- **Producer nodes** are pre-configured authorities that create blocks (fixed at genesis)
- **Full nodes** can join anytime to sync the blockchain and submit transactions
- New nodes cannot become producers without network-wide genesis block update

## Requirements to Join

To join an existing Podoru Chain network, you need:

1. **Genesis file** (`genesis.json`) - Must match exactly with the network
2. **Bootstrap peers** - IP addresses of at least one existing node
3. **Podoru Chain node binary** or Docker image
4. **Network connectivity** - Reachable P2P port (default: 9000)

## Network Scenarios

### Scenario 1: Joining on the Same Machine

Join a local network running on the same computer:

**Configuration:**
```yaml
node_type: full

# Network configuration
p2p_port: 9003  # Use different port
p2p_bind_addr: "0.0.0.0"
bootstrap_peers:
  - "127.0.0.1:9000"  # Producer1
  - "127.0.0.1:9001"  # Producer2 (if exists)
max_peers: 50

# API configuration
api_enabled: true
api_port: 8548  # Use different port
api_bind_addr: "0.0.0.0"

# Storage configuration
data_dir: "/data"

# Consensus configuration (copy from network)
authorities:
  - "0xa050E943a40b07e6e7B0B423F51c2E8536059689"
block_time: 5s

# Genesis configuration
genesis_path: "/data/genesis.json"
```

### Scenario 2: Joining on the Same Local Network (LAN)

Join from another computer on the same network:

**Configuration:**
```yaml
node_type: full

# Network configuration
p2p_port: 9000
p2p_bind_addr: "0.0.0.0"
bootstrap_peers:
  - "192.168.1.100:9000"  # Replace with host's local IP
  - "192.168.1.100:9001"
max_peers: 50

# API configuration
api_enabled: true
api_port: 8545
api_bind_addr: "0.0.0.0"

# Storage configuration
data_dir: "/data"

# Consensus configuration (copy from network)
authorities:
  - "0xa050E943a40b07e6e7B0B423F51c2E8536059689"
block_time: 5s

# Genesis configuration
genesis_path: "/data/genesis.json"
```

**Host Setup Required:**
1. Find your local IP: `ip addr show` or `ifconfig`
2. Ensure firewall allows P2P port (9000)
3. On Ubuntu/Debian: `sudo ufw allow 9000/tcp`

### Scenario 3: Joining Over the Internet

Join from anywhere over the internet:

**Configuration:**
```yaml
node_type: full

# Network configuration
p2p_port: 9000
p2p_bind_addr: "0.0.0.0"
bootstrap_peers:
  - "podoru-node.example.com:9000"  # Public domain/IP
  - "203.0.113.10:9000"  # Or direct IP
max_peers: 50

# API configuration
api_enabled: true
api_port: 8545
api_bind_addr: "0.0.0.0"

# Storage configuration
data_dir: "/data"

# Consensus configuration (copy from network)
authorities:
  - "0xa050E943a40b07e6e7B0B423F51c2E8536059689"
block_time: 5s

# Genesis configuration
genesis_path: "/data/genesis.json"
```

**Host Setup Required:**
1. Configure port forwarding on router (external 9000 → internal 9000)
2. Use static IP or dynamic DNS service
3. Ensure firewall allows incoming connections
4. Consider using a VPN for security

## Step-by-Step: Joining the Network

### Method 1: Using Docker (Recommended)

1. **Get Required Files**

   Request these from the network operator:
   - `genesis.json` - Must be identical to network's genesis
   - Bootstrap peer addresses (IP:port)
   - Authorities list

2. **Create Configuration Directory**
   ```bash
   mkdir -p podoru-fullnode/data
   cd podoru-fullnode
   ```

3. **Create `config.yaml`**

   Use one of the configurations above based on your scenario.

4. **Copy Genesis File**
   ```bash
   # Place the genesis.json you received
   cp /path/to/genesis.json data/genesis.json
   ```

5. **Create `docker-compose.yml`**
   ```yaml
   version: '3.8'

   services:
     fullnode:
       image: podoru-chain:latest
       container_name: my-podoru-fullnode
       ports:
         - "8545:8545"  # API port
         - "9000:9000"  # P2P port
       volumes:
         - ./data:/data
       environment:
         - NODE_TYPE=full
       restart: unless-stopped
   ```

6. **Start the Node**
   ```bash
   docker-compose up -d
   ```

7. **Verify Connection**
   ```bash
   # Check logs
   docker-compose logs -f

   # Check node status
   curl http://localhost:8545/api/v1/node/info

   # Check blockchain height
   curl http://localhost:8545/api/v1/chain/info
   ```

### Method 2: Using Binary

1. **Build or Download Binary**
   ```bash
   # Build from source
   git clone https://github.com/podoru/podoru-chain
   cd podoru-chain
   make build
   ```

2. **Create Data Directory**
   ```bash
   mkdir -p ~/podoru-data
   cd ~/podoru-data
   ```

3. **Create Configuration**

   Create `config.yaml` and `genesis.json` as described above.

4. **Run Node**
   ```bash
   /path/to/podoru-node --config config.yaml
   ```

## Getting Network Information

As the network operator, you can share this information:

### Generate Join Instructions

```bash
# Get your public IP (if sharing over internet)
curl -4 ifconfig.me

# Or local IP (if sharing on LAN)
ip addr show | grep "inet " | grep -v 127.0.0.1

# Share these files:
# 1. Genesis file
cat docker/data/genesis.json

# 2. Bootstrap peers list
echo "Bootstrap Peers:"
echo "  - \"YOUR_IP:9000\""
echo "  - \"YOUR_IP:9001\""

# 3. Authorities list
cat docker/data/genesis.json | jq '.authorities'
```

## Security Considerations

### For Network Operators

1. **Firewall Rules**: Only expose necessary ports
2. **Rate Limiting**: Protect against DoS attacks
3. **Monitoring**: Track peer connections and resource usage
4. **Genesis Security**: Keep genesis file authentic (consider checksums)

### For Joining Nodes

1. **Verify Genesis**: Ensure genesis.json is from trusted source
2. **Secure RPC**: Don't expose API port publicly unless necessary
3. **Resource Limits**: Monitor disk space and bandwidth
4. **Updates**: Keep node software updated

## Troubleshooting

### Cannot Connect to Bootstrap Peers

**Symptoms:**
```
Failed to connect to bootstrap peer: connection refused
```

**Solutions:**
1. Check firewall on host machine
2. Verify IP address and port are correct
3. Ensure host node is running
4. Test connectivity: `telnet HOST_IP 9000`

### Genesis Block Mismatch

**Symptoms:**
```
Genesis block hash mismatch
```

**Solutions:**
1. Get fresh `genesis.json` from network operator
2. Delete blockchain data directory
3. Restart node with correct genesis

### Not Syncing Blocks

**Symptoms:**
```
Node stuck at height 0 or low block number
```

**Solutions:**
1. Check peer connections: `curl localhost:8545/api/v1/node/info`
2. Verify bootstrap peers are reachable
3. Check node logs for errors
4. Ensure authorities list matches network

## Network Information Template

Share this information with users who want to join:

```
=== Podoru Chain Network Join Information ===

Network Name: Podoru Chain
Chain ID: (from genesis.json)

Bootstrap Peers:
  - "YOUR_IP:9000"
  - "YOUR_IP:9001"

Genesis Block:
  Download: [URL or share file directly]
  SHA256: [checksum of genesis.json]

Authorities:
  - "0xa050E943a40b07e6e7B0B423F51c2E8536059689"

Block Time: 5s

Requirements:
  - Docker or Podoru Chain binary
  - Open port 9000 for P2P
  - At least 1GB disk space
  - Stable internet connection

Documentation:
  https://github.com/podoru/podoru-chain/docs/joining-network.md
```

## Example: Complete Setup Script

Here's a script to help users join your network:

```bash
#!/bin/bash
# join-podoru-network.sh

BOOTSTRAP_PEER="192.168.1.100:9000"  # Replace with your IP
DATA_DIR="./podoru-node-data"

echo "Joining Podoru Chain Network..."

# Create directory
mkdir -p "$DATA_DIR"
cd "$DATA_DIR"

# Download genesis (replace with your method)
echo "Downloading genesis.json..."
curl -o genesis.json http://your-server.com/genesis.json

# Create config
cat > config.yaml <<EOF
node_type: full

p2p_port: 9000
p2p_bind_addr: "0.0.0.0"
bootstrap_peers:
  - "$BOOTSTRAP_PEER"
max_peers: 50

api_enabled: true
api_port: 8545
api_bind_addr: "0.0.0.0"

data_dir: "/data"

authorities:
  - "0xa050E943a40b07e6e7B0B423F51c2E8536059689"
block_time: 5s

genesis_path: "/data/genesis.json"
EOF

# Create docker-compose
cat > docker-compose.yml <<EOF
version: '3.8'

services:
  fullnode:
    image: podoru-chain:latest
    container_name: podoru-fullnode
    ports:
      - "8545:8545"
      - "9000:9000"
    volumes:
      - ./:/data
    environment:
      - NODE_TYPE=full
    restart: unless-stopped
EOF

echo "Starting node..."
docker-compose up -d

echo "Node started! Checking status..."
sleep 5
curl http://localhost:8545/api/v1/node/info

echo ""
echo "View logs: docker-compose logs -f"
echo "Stop node: docker-compose down"
```

## Advanced: Adding More Producer Nodes

To add more producer nodes, you need to:

1. **Generate new network with updated authorities**
2. **All nodes must restart with new genesis**
3. **Coordinate upgrade with all participants**

This requires network-wide consensus and is beyond the scope of joining as a full node. Contact the network operators for governance procedures.
