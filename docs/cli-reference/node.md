# podoru-node

The main Podoru Chain node executable for running producer and full nodes.

## Synopsis

```bash
podoru-node -config <config-file>
```

## Description

`podoru-node` starts a Podoru Chain blockchain node that:
- Validates blocks and maintains blockchain state
- Participates in P2P network
- Serves REST API (if enabled)
- Produces blocks (if configured as producer)

## Options

### -config

**Required**: Path to YAML configuration file.

```bash
podoru-node -config /path/to/config.yaml
```

**Example**:
```bash
./bin/podoru-node -config config/producer1.yaml
```

## Examples

### Run Producer Node

```bash
./bin/podoru-node -config config/producer1.yaml
```

### Run Full Node

```bash
./bin/podoru-node -config config/fullnode.yaml
```

### Run with Custom Config

```bash
./bin/podoru-node -config /etc/podoru/custom.yaml
```

## Configuration File

The configuration file must be in YAML format. See [Configuration Guide](../configuration/README.md) for details.

**Minimum Configuration**:
```yaml
node_type: full
p2p_port: 9000
api_enabled: true
api_port: 8545
data_dir: "./data"
authorities:
  - "0xAuthority1..."
block_time: 5s
genesis_path: "./genesis.json"
```

**Producer Configuration**:
```yaml
node_type: producer
address: "0xYourAddress"
private_key: "/path/to/key"
# ... plus all other fields
```

## Startup Process

1. Load and validate configuration
2. Initialize BadgerDB storage
3. Load or create genesis block
4. Start P2P networking
5. Sync blockchain (if behind)
6. Start API server (if enabled)
7. Start block production (if producer)

## Logging

The node logs to stdout/stderr:

```
2024-01-06 10:00:00 INFO Loading configuration from config/producer1.yaml...
2024-01-06 10:00:01 INFO Creating blockchain node...
2024-01-06 10:00:02 INFO Starting REST API server...
2024-01-06 10:00:03 INFO Podoru Chain node is running
```

### Log Levels

- **INFO**: Normal operations
- **WARN**: Non-critical issues
- **ERROR**: Critical errors
- **FATAL**: Unrecoverable errors (node exits)

## Signals

### SIGINT / SIGTERM

Graceful shutdown:

```bash
# Send SIGINT (Ctrl+C)
^C

# Or send SIGTERM
kill -TERM <pid>
```

The node will:
1. Stop accepting new transactions
2. Stop API server
3. Close P2P connections
4. Flush database to disk
5. Exit

### SIGKILL

Force shutdown (not recommended):

```bash
kill -KILL <pid>
```

May result in data corruption.

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Normal shutdown |
| 1 | Configuration error |
| 1 | Startup error |
| 1 | Runtime error |
| 130 | Interrupted (SIGINT) |

## Running as Service

### Systemd (Linux)

Create `/etc/systemd/system/podoru-node.service`:

```ini
[Unit]
Description=Podoru Chain Node
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
sudo systemctl enable podoru-node
sudo systemctl start podoru-node
sudo systemctl status podoru-node
```

### Docker

```bash
docker run -d \
  --name podoru-node \
  -p 8545:8545 \
  -p 9000:9000 \
  -v /path/to/data:/data \
  -v /path/to/config.yaml:/config.yaml \
  podoru-chain:latest \
  -config /config.yaml
```

## Monitoring

### Check if Running

```bash
ps aux | grep podoru-node
```

### Check Health

```bash
curl http://localhost:8545/api/v1/node/health
```

### View Logs

```bash
# Systemd
journalctl -u podoru-node -f

# Docker
docker logs podoru-node -f
```

## Troubleshooting

### Node Won't Start

1. Check configuration file syntax
2. Verify ports are available
3. Check file permissions
4. Review error logs

### Node Crashes

1. Check disk space
2. Verify BadgerDB integrity
3. Check for network issues
4. Review logs for errors

### Node Not Syncing

1. Check peer connectivity
2. Verify genesis file matches network
3. Check firewall rules
4. Review P2P port configuration

See [Troubleshooting Guide](../troubleshooting/README.md) for more help.

## Performance Tuning

### Resource Requirements

**Minimum**:
- CPU: 2 cores
- RAM: 2 GB
- Disk: 10 GB

**Recommended**:
- CPU: 4+ cores
- RAM: 8+ GB
- Disk: 100+ GB SSD

### Optimization

Adjust BadgerDB settings in code for better performance:
- Increase cache size for more RAM
- Use SSD for data directory
- Tune GC frequency

## Further Reading

- [Configuration Guide](../configuration/README.md) - Configuration options
- [Architecture Overview](../architecture/README.md) - How it works
- [Troubleshooting](../troubleshooting/README.md) - Common issues
