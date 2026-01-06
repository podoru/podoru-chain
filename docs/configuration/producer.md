# Producer Node Configuration

Configuration reference for Podoru Chain producer nodes.

## Overview

Producer nodes are authorized to create new blocks. They require:
- Cryptographic keypair for signing blocks
- Address listed in genesis authorities
- Complete node configuration

## Configuration Template

```yaml
# Node Type
node_type: producer

# Producer Identity
address: "0x3D4b25CBdda1014F74F9C80f040ce1Bb69130CBB"
private_key: "/data/keys/producer1.key"

# P2P Network Configuration
p2p_port: 9000
p2p_bind_addr: "0.0.0.0"
bootstrap_peers:
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

### node_type

**Type**: String
**Required**: Yes
**Value**: `"producer"`

```yaml
node_type: producer
```

### address

**Type**: String
**Required**: Yes (producer only)
**Format**: Ethereum-compatible address (0x-prefixed, 42 characters)

```yaml
address: "0x3D4b25CBdda1014F74F9C80f040ce1Bb69130CBB"
```

**Must Match**:
- Address derived from private_key
- Address listed in genesis authorities
- Address in other nodes' authorities list

### private_key

**Type**: String
**Required**: Yes (producer only)
**Value**: Path to private key file

```yaml
private_key: "/data/keys/producer1.key"
```

**Security**:
- Use absolute paths
- Set file permissions to 600
- Never commit to version control
- Back up securely

### p2p_port

**Type**: Integer
**Required**: Yes
**Range**: 1-65535
**Default**: 9000

```yaml
p2p_port: 9000
```

**Recommendations**:
- Use 9000-9999 range
- Each node needs unique port
- Allow through firewall

### p2p_bind_addr

**Type**: String
**Required**: Yes
**Default**: `"0.0.0.0"`

```yaml
p2p_bind_addr: "0.0.0.0"  # Bind to all interfaces
p2p_bind_addr: "127.0.0.1"  # Localhost only
p2p_bind_addr: "192.168.1.10"  # Specific interface
```

### bootstrap_peers

**Type**: Array of strings
**Required**: Yes
**Format**: `"host:port"`

```yaml
bootstrap_peers:
  - "192.168.1.11:9000"
  - "192.168.1.12:9000"
  - "node3.example.com:9000"
```

**Tips**:
- List other producer nodes
- Use static IPs or DNS names
- Don't include self
- At least 1 peer recommended

### api_enabled

**Type**: Boolean
**Required**: Yes

```yaml
api_enabled: true  # Enable REST API
api_enabled: false  # Disable REST API
```

### api_port

**Type**: Integer
**Required**: If api_enabled is true
**Range**: 1-65535
**Default**: 8545

```yaml
api_port: 8545
```

### api_bind_addr

**Type**: String
**Required**: If api_enabled is true
**Default**: `"0.0.0.0"`

```yaml
api_bind_addr: "0.0.0.0"  # Public
api_bind_addr: "127.0.0.1"  # Localhost only
```

**Security**:
- Use `127.0.0.1` with reverse proxy for production
- Use `0.0.0.0` only if firewall configured

### data_dir

**Type**: String
**Required**: Yes

```yaml
data_dir: "/data"
data_dir: "./data/producer1"
data_dir: "/var/lib/podoru"
```

**Recommendations**:
- Use absolute paths in production
- Ensure sufficient disk space
- Use SSD for better performance
- Regular backups

### authorities

**Type**: Array of strings
**Required**: Yes
**Format**: Ethereum addresses

```yaml
authorities:
  - "0x3D4b25CBdda1014F74F9C80f040ce1Bb69130CBB"
  - "0x4aa37EEc2a26a4e04b7b206f32D6C2C63219F5cd"
  - "0x304F73DD4CabF754eF2240fF2bC2446eB7709652"
```

**Must Match**:
- Genesis file authorities
- All other nodes' configurations

### block_time

**Type**: Duration string
**Required**: Yes
**Format**: `"<number>s"`, `"<number>m"`, etc.

```yaml
block_time: 5s   # 5 seconds
block_time: 10s  # 10 seconds
block_time: 1m   # 1 minute
```

**Recommendations**:
- Development: 3-5s
- Production: 5-10s
- High-throughput: 2-3s

**Must Match**: All nodes must use same block_time

### genesis_path

**Type**: String
**Required**: Yes

```yaml
genesis_path: "/data/genesis.json"
genesis_path: "./genesis.json"
```

**Must**:
- Point to valid genesis JSON file
- Be identical across all nodes

## Optional Parameters

### max_peers

**Type**: Integer
**Default**: 50

```yaml
max_peers: 50
```

**Recommendations**:
- Small networks: 10-20
- Large networks: 50-100

## Complete Examples

### Local Development

```yaml
node_type: producer
address: "0x3D4b25CBdda1014F74F9C80f040ce1Bb69130CBB"
private_key: "./keys/producer1.key"

p2p_port: 9000
p2p_bind_addr: "127.0.0.1"
bootstrap_peers:
  - "localhost:9001"
  - "localhost:9002"
max_peers: 10

api_enabled: true
api_port: 8545
api_bind_addr: "127.0.0.1"

data_dir: "./data/producer1"

authorities:
  - "0x3D4b25CBdda1014F74F9C80f040ce1Bb69130CBB"
  - "0x4aa37EEc2a26a4e04b7b206f32D6C2C63219F5cd"
  - "0x304F73DD4CabF754eF2240fF2bC2446eB7709652"

block_time: 3s
genesis_path: "./genesis.json"
```

### Production

```yaml
node_type: producer
address: "0x3D4b25CBdda1014F74F9C80f040ce1Bb69130CBB"
private_key: "/etc/podoru/keys/producer1.key"

p2p_port: 9000
p2p_bind_addr: "0.0.0.0"
bootstrap_peers:
  - "node2.example.com:9000"
  - "node3.example.com:9000"
  - "node4.example.com:9000"
max_peers: 50

api_enabled: true
api_port: 8545
api_bind_addr: "127.0.0.1"  # Use with nginx reverse proxy

data_dir: "/var/lib/podoru/data"

authorities:
  - "0x3D4b25CBdda1014F74F9C80f040ce1Bb69130CBB"
  - "0x4aa37EEc2a26a4e04b7b206f32D6C2C63219F5cd"
  - "0x304F73DD4CabF754eF2240fF2bC2446eB7709652"
  - "0x5bb48c61FfB69A8A2D9f72C8A5A7e2F8cD3c1234"
  - "0x6cc59d72EEc80B9B3E0g83D9B6B8f3G9dE4d2345"

block_time: 5s
genesis_path: "/etc/podoru/genesis.json"
```

## Validation

### Check Configuration

```bash
# Validate YAML syntax
yamllint config/producer1.yaml

# Verify private key exists
ls -l /data/keys/producer1.key

# Check permissions
# Should be: -rw------- (600)

# Verify address matches key
# (Run keygen with same key to verify)
```

### Common Mistakes

**Address Mismatch**:
```yaml
# ❌ Bad: Address doesn't match private_key
address: "0xWrongAddress"
private_key: "/keys/producer1.key"  # Derives different address
```

**Not in Authorities**:
```yaml
# ❌ Bad: Address not in authorities list
address: "0xProducer1"
authorities:
  - "0xProducer2"  # Missing Producer1!
  - "0xProducer3"
```

**Duplicate Ports**:
```yaml
# ❌ Bad: Multiple nodes on same port
# producer1.yaml
p2p_port: 9000

# producer2.yaml (same machine)
p2p_port: 9000  # Conflict!
```

## Further Reading

- [Genesis Configuration](genesis.md)
- [Full Node Configuration](fullnode.md)
- [CLI Reference](../cli-reference/node.md)
- [Manual Setup](../getting-started/manual-setup.md)
