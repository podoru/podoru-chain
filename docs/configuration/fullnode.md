# Full Node Configuration

Configuration reference for Podoru Chain full nodes.

## Overview

Full nodes validate blocks and maintain blockchain state but don't produce blocks. They:
- Validate all blocks
- Maintain complete blockchain state
- Serve API requests
- Participate in P2P network
- Don't require private keys

## Configuration Template

```yaml
# Node Type
node_type: full

# P2P Network Configuration
p2p_port: 9000
p2p_bind_addr: "0.0.0.0"
bootstrap_peers:
  - "172.20.0.10:9000"
  - "172.20.0.11:9000"
  - "172.20.0.12:9000"
max_peers: 50

# API Configuration
api_enabled: true
api_port: 8545
api_bind_addr: "0.0.0.0"

# Storage Configuration
data_dir: "/data"

# Consensus Configuration
authorities:
  - "0x3D4b25CBdda1014F74F9C80f040ce1Bb69130CBB"
  - "0x4aa37EEc2a26a4e04b7b206f32D6C2C63219F5cd"
  - "0x304F73DD4CabF754eF2240fF2bC2446eB7709652"
block_time: 5s

# Genesis Configuration
genesis_path: "/data/genesis.json"
```

## Required Parameters

All parameters are same as producer nodes except:

**Not Required**:
- `address` - Full nodes don't have identity
- `private_key` - No signing capability

**Must Set**:
```yaml
node_type: full
```

See [Producer Configuration](producer.md) for detailed parameter descriptions.

## Use Cases

### Additional API Endpoints

Add full nodes to distribute API load:

```yaml
# Load Balancer
# ├── Producer 1 (API: 8545)
# ├── Producer 2 (API: 8546)
# ├── Full Node 1 (API: 8548) ← Additional capacity
# └── Full Node 2 (API: 8549) ← Additional capacity
```

### Geographic Distribution

Deploy full nodes in different regions:

```yaml
# Asia Full Node
p2p_port: 9000
bootstrap_peers:
  - "us-producer1.example.com:9000"
  - "eu-producer1.example.com:9000"

# Europe Full Node
p2p_port: 9000
bootstrap_peers:
  - "us-producer1.example.com:9000"
  - "asia-fullnode1.example.com:9000"
```

### Backup/Redundancy

Ensure network availability:

```yaml
# Even if 1-2 producers fail,
# full nodes maintain network access
```

## Complete Examples

### Basic Full Node

```yaml
node_type: full

p2p_port: 9003
p2p_bind_addr: "0.0.0.0"
bootstrap_peers:
  - "producer1:9000"
  - "producer2:9001"
  - "producer3:9002"
max_peers: 50

api_enabled: true
api_port: 8548
api_bind_addr: "0.0.0.0"

data_dir: "/data"

authorities:
  - "0x3D4b25CBdda1014F74F9C80f040ce1Bb69130CBB"
  - "0x4aa37EEc2a26a4e04b7b206f32D6C2C63219F5cd"
  - "0x304F73DD4CabF754eF2240fF2bC2446eB7709652"

block_time: 5s
genesis_path: "/data/genesis.json"
```

### API-Only Full Node

Dedicated for serving API requests:

```yaml
node_type: full

p2p_port: 9010
p2p_bind_addr: "0.0.0.0"
bootstrap_peers:
  - "internal-producer1:9000"
  - "internal-producer2:9001"
max_peers: 20

api_enabled: true
api_port: 8545
api_bind_addr: "0.0.0.0"  # Public API

data_dir: "/var/lib/podoru"

authorities:
  - "0x3D4b25CBdda1014F74F9C80f040ce1Bb69130CBB"
  - "0x4aa37EEc2a26a4e04b7b206f32D6C2C63219F5cd"
  - "0x304F73DD4CabF754eF2240fF2bC2446eB7709652"

block_time: 5s
genesis_path: "/etc/podoru/genesis.json"
```

### Archive Node

Store complete blockchain history:

```yaml
node_type: full

p2p_port: 9020
p2p_bind_addr: "0.0.0.0"
bootstrap_peers:
  - "producer1.example.com:9000"
max_peers: 50

api_enabled: true
api_port: 8545
api_bind_addr: "127.0.0.1"

data_dir: "/mnt/archive/podoru"  # Large disk

authorities:
  - "0x3D4b25CBdda1014F74F9C80f040ce1Bb69130CBB"
  - "0x4aa37EEc2a26a4e04b7b206f32D6C2C63219F5cd"
  - "0x304F73DD4CabF754eF2240fF2bC2446eB7709652"

block_time: 5s
genesis_path: "/etc/podoru/genesis.json"
```

## Differences from Producer Nodes

| Feature | Producer Node | Full Node |
|---------|--------------|-----------|
| Block Production | Yes | No |
| Block Validation | Yes | Yes |
| Requires Keys | Yes | No |
| In Authorities | Yes | No |
| API Server | Yes | Yes |
| P2P Network | Yes | Yes |
| State Storage | Yes | Yes |

## Best Practices

### Security

**API Security**:
```yaml
# Use reverse proxy for public access
api_bind_addr: "127.0.0.1"
# Configure nginx/caddy to proxy requests
```

**Firewall**:
```bash
# Allow P2P
sudo ufw allow 9000/tcp

# Allow API (if public)
sudo ufw allow 8545/tcp
```

### Performance

**Resource Allocation**:
- CPU: 2-4 cores
- RAM: 4-8 GB
- Disk: 50-100 GB SSD

**Tuning**:
```yaml
# Increase max_peers for better connectivity
max_peers: 100

# Use SSD for data_dir
data_dir: "/mnt/ssd/podoru"
```

### Monitoring

Monitor full node health:

```bash
# Health check
curl http://localhost:8545/api/v1/node/health

# Peer count
curl http://localhost:8545/api/v1/node/peers | jq '.data.peer_count'

# Sync status
curl http://localhost:8545/api/v1/chain/info | jq '.data.height'
```

## Troubleshooting

### Not Syncing

Check bootstrap peers connectivity:

```bash
# Test connectivity
nc -zv producer1 9000
nc -zv producer2 9001

# Check logs
docker logs fullnode1 | grep -i "peer"
```

### No API Response

Verify API configuration:

```bash
# Check if API is enabled
grep api_enabled config/fullnode.yaml

# Check port
netstat -an | grep 8545

# Test locally
curl http://localhost:8545/api/v1/node/info
```

## Further Reading

- [Producer Configuration](producer.md)
- [Genesis Configuration](genesis.md)
- [Manual Setup](../getting-started/manual-setup.md)
- [CLI Reference](../cli-reference/node.md)
