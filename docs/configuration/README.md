# Configuration

Guide to configuring Podoru Chain nodes and networks.

## Configuration Files

Podoru Chain uses two main configuration files:

### 1. Node Configuration (YAML)

Defines node-specific settings like ports, keys, and network peers.

- **Format**: YAML
- **Location**: `config/*.yaml`
- **Required**: Yes
- **Per Node**: Each node needs its own config file

**Types**:
- [Producer Node Configuration](producer.md)
- [Full Node Configuration](fullnode.md)

### 2. Genesis Configuration (JSON)

Defines the initial blockchain state and authorities.

- **Format**: JSON
- **Location**: `genesis.json` or custom path
- **Required**: Yes
- **Network-wide**: All nodes must use the identical genesis file

**Documentation**: [Genesis Block Configuration](genesis.md)

## Quick Start

### Using Setup Wizard

The easiest way to configure your network:

```bash
make setup-wizard
```

The wizard automatically generates all configuration files.

### Manual Configuration

For custom setups, create configuration files manually:

1. [Generate keys](../cli-reference/keygen.md) for producers
2. [Create genesis file](genesis.md)
3. [Configure producer nodes](producer.md)
4. [Configure full nodes](fullnode.md)

## Configuration Overview

### Producer Node

```yaml
node_type: producer
address: "0xYourAddress"
private_key: "/path/to/key"
p2p_port: 9000
api_enabled: true
api_port: 8545
data_dir: "./data"
authorities: [...]
block_time: 5s
genesis_path: "./genesis.json"
```

[Full Producer Configuration](producer.md)

### Full Node

```yaml
node_type: full
p2p_port: 9000
api_enabled: true
api_port: 8545
data_dir: "./data"
authorities: [...]
block_time: 5s
genesis_path: "./genesis.json"
```

[Full Node Configuration](fullnode.md)

### Genesis Block

```json
{
  "timestamp": 1704556800,
  "authorities": [
    "0xAddress1",
    "0xAddress2",
    "0xAddress3"
  ],
  "initial_state": {
    "chain:name": "My Chain",
    "chain:version": "1.0.0"
  }
}
```

[Genesis Configuration](genesis.md)

## Configuration Parameters

### Common Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| node_type | string | Yes | "producer" or "full" |
| p2p_port | integer | Yes | P2P network port |
| p2p_bind_addr | string | Yes | P2P bind address |
| bootstrap_peers | array | Yes | Initial peer addresses |
| api_enabled | boolean | Yes | Enable REST API |
| api_port | integer | If API enabled | API server port |
| data_dir | string | Yes | Data directory path |
| authorities | array | Yes | Block producer addresses |
| block_time | duration | Yes | Time between blocks |
| genesis_path | string | Yes | Genesis file path |

### Producer-Only Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| address | string | Yes | Producer's address |
| private_key | string | Yes | Path to private key |

## Environment-Specific Configurations

### Development

```yaml
# Fast blocks, local peers
block_time: 3s
p2p_port: 9000
bootstrap_peers:
  - "localhost:9001"
  - "localhost:9002"
```

### Production

```yaml
# Standard blocks, distributed peers
block_time: 5s
p2p_port: 9000
bootstrap_peers:
  - "node1.example.com:9000"
  - "node2.example.com:9000"
  - "node3.example.com:9000"
```

### Testing

```yaml
# Very fast blocks, minimal setup
block_time: 1s
p2p_port: 19000
bootstrap_peers: []
```

## Best Practices

### Security

- **Never commit private keys to version control**
- Set restrictive permissions on key files (600)
- Use different keys for each producer
- Store keys encrypted at rest
- Back up keys securely

### Network

- Use 3+ producers for fault tolerance
- Distribute nodes geographically
- Use static IPs for bootstrap peers
- Configure firewall rules for P2P ports
- Monitor peer connectivity

### Storage

- Use SSD for `data_dir`
- Ensure sufficient disk space
- Regular backups of blockchain data
- Monitor disk usage

### Performance

- Adjust `block_time` based on needs
- Tune `max_peers` for network size
- Use appropriate hardware
- Monitor resource usage

## Configuration Examples

### 3-Producer Network

```yaml
# Producer 1
node_type: producer
address: "0xProducer1Address"
private_key: "/keys/producer1.key"
p2p_port: 9000
bootstrap_peers:
  - "producer2:9001"
  - "producer3:9002"

# Producer 2
node_type: producer
address: "0xProducer2Address"
private_key: "/keys/producer2.key"
p2p_port: 9001
bootstrap_peers:
  - "producer1:9000"
  - "producer3:9002"

# Producer 3
node_type: producer
address: "0xProducer3Address"
private_key: "/keys/producer3.key"
p2p_port: 9002
bootstrap_peers:
  - "producer1:9000"
  - "producer2:9001"
```

### Mixed Network (Producers + Full Nodes)

```yaml
# 3 Producers + 2 Full Nodes
# Producer configs: Same as above
# Full Node 1
node_type: full
p2p_port: 9003
bootstrap_peers:
  - "producer1:9000"
  - "producer2:9001"
  - "producer3:9002"

# Full Node 2
node_type: full
p2p_port: 9004
bootstrap_peers:
  - "producer1:9000"
  - "fullnode1:9003"
```

## Validating Configuration

### Before Starting

```bash
# Check YAML syntax
yamllint config/producer1.yaml

# Verify file permissions
ls -l config/*.yaml keys/*.key

# Test configuration
./bin/podoru-node -config config/producer1.yaml --dry-run
```

### After Starting

```bash
# Check node health
curl http://localhost:8545/api/v1/node/health

# Verify peer connections
curl http://localhost:8545/api/v1/node/peers

# Check blockchain sync
curl http://localhost:8545/api/v1/chain/info
```

## Troubleshooting

See [Troubleshooting Guide](../troubleshooting/README.md) for common configuration issues.

## Further Reading

- [Genesis Block Configuration](genesis.md)
- [Producer Node Configuration](producer.md)
- [Full Node Configuration](fullnode.md)
- [CLI Reference](../cli-reference/README.md)
- [Manual Setup Guide](../getting-started/manual-setup.md)
