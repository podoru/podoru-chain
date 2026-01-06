# Troubleshooting

Common issues and solutions for Podoru Chain.

## Installation Issues

### Go Version Too Old

**Error**: `requires Go 1.24 or higher`

**Solution**:
```bash
# Check version
go version

# Update Go
wget https://go.dev/dl/go1.24.0.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.24.0.linux-amd64.tar.gz
```

### Build Fails

**Error**: `cannot find package` or `undefined: ...`

**Solution**:
```bash
# Clean and rebuild
go clean -modcache
make deps
make build
```

### Docker Permission Denied

**Error**: `permission denied while trying to connect to Docker daemon`

**Solution**:
```bash
# Add user to docker group
sudo usermod -aG docker $USER

# Log out and back in, or run:
newgrp docker

# Verify
docker ps
```

## Node Startup Issues

### Config File Not Found

**Error**: `Config file path is required` or `Failed to load configuration`

**Solution**:
```bash
# Verify file exists
ls -l config/producer1.yaml

# Use correct path
./bin/podoru-node -config config/producer1.yaml
```

### Invalid Configuration

**Error**: `Failed to load configuration: yaml: ...`

**Solution**:
```bash
# Validate YAML syntax
yamllint config/producer1.yaml

# Check for common mistakes:
# - Missing colons
# - Incorrect indentation
# - Invalid values
```

### Private Key Not Found

**Error**: `Failed to load private key`

**Solution**:
```bash
# Check file exists
ls -l /path/to/private.key

# Verify permissions (should be 600)
chmod 600 /path/to/private.key

# Verify path in config
grep private_key config/producer1.yaml
```

### Port Already in Use

**Error**: `bind: address already in use`

**Solution**:
```bash
# Find process using port
sudo lsof -i :8545
sudo lsof -i :9000

# Kill process
kill <PID>

# Or use different port
# Edit config file and change port
```

### Genesis File Not Found

**Error**: `Failed to load genesis file`

**Solution**:
```bash
# Verify file exists
ls -l genesis.json

# Check path in config
grep genesis_path config/*.yaml

# Ensure all nodes have same genesis
sha256sum genesis.json
sha256sum docker/data/*/genesis.json
```

## Network Issues

### No Peers Connecting

**Symptom**: `peer_count: 0`

**Diagnosis**:
```bash
# Check peer count
curl http://localhost:8545/api/v1/node/peers | jq '.data.peer_count'

# Check bootstrap peers in config
grep bootstrap_peers config/*.yaml
```

**Solutions**:
```bash
# 1. Verify network connectivity
nc -zv peer-host 9000

# 2. Check firewall
sudo ufw status
sudo ufw allow 9000/tcp

# 3. Verify bootstrap_peers are running
curl http://peer-host:8545/api/v1/node/health

# 4. Check Docker network (if using Docker)
docker network inspect podoru_network
```

### Nodes Not Syncing

**Symptom**: Different heights across nodes

**Diagnosis**:
```bash
# Check heights on all nodes
for port in 8545 8546 8547; do
  echo "Port $port:"
  curl -s http://localhost:$port/api/v1/chain/info | jq '.data.height'
done
```

**Solutions**:
```bash
# 1. Verify same genesis file
sha256sum docker/data/*/genesis.json

# 2. Check all nodes have same authorities
grep -A 5 authorities config/*.yaml

# 3. Restart lagging nodes
docker restart podoru-producer1

# 4. Check for errors
docker logs podoru-producer1 | grep -i error
```

### Network Partition

**Symptom**: Some nodes have different blockchain

**Solution**:
```bash
# 1. Stop all nodes
make docker-compose-down

# 2. Verify all nodes have same genesis
for node in producer1 producer2 producer3; do
  sha256sum docker/data/$node/genesis.json
done

# 3. Clear data on out-of-sync nodes
rm -rf docker/data/problem-node/badger

# 4. Restart all nodes
make docker-compose-up
```

## Block Production Issues

### Blocks Not Being Produced

**Symptom**: Blockchain height not increasing

**Diagnosis**:
```bash
# Check if any producers are running
curl http://localhost:8545/api/v1/node/info | jq '.data.node_type'

# Check latest block time
curl http://localhost:8545/api/v1/block/latest | jq '.data.timestamp'
```

**Solutions**:
```bash
# 1. Verify producer node is running
docker ps | grep producer

# 2. Check producer address in authorities
curl http://localhost:8545/api/v1/chain/info | jq '.data.authorities'

# 3. Verify private key is correct
# Address from config should match address from keygen

# 4. Check logs for errors
docker logs podoru-producer1 | grep -i "block\|producer"
```

### Invalid Block Signature

**Error**: `Invalid block signature` in logs

**Solution**:
```bash
# 1. Verify address matches private key
./bin/keygen -output test.key
# Compare address with config

# 2. Ensure address is in authorities list
grep -A 5 authorities config/*.yaml

# 3. Check all nodes have same authorities
# All authority lists must be identical
```

## API Issues

### API Not Responding

**Symptom**: `Connection refused` when accessing API

**Diagnosis**:
```bash
# Check if API is enabled
grep api_enabled config/*.yaml

# Check API port
grep api_port config/*.yaml

# Test locally
curl http://localhost:8545/api/v1/node/health
```

**Solutions**:
```bash
# 1. Verify API is enabled
api_enabled: true

# 2. Check bind address
api_bind_addr: "0.0.0.0"  # Not 127.0.0.1 for external access

# 3. Check firewall
sudo ufw allow 8545/tcp

# 4. Verify node is running
docker ps | grep podoru
```

### 404 Not Found

**Error**: API returns 404 for valid endpoints

**Solution**:
```bash
# Verify endpoint path (note /api/v1 prefix)
curl http://localhost:8545/api/v1/chain/info  # ✓ Correct
curl http://localhost:8545/chain/info         # ✗ Wrong
```

### Slow API Responses

**Symptom**: API requests take too long

**Solutions**:
```bash
# 1. Check disk I/O (use SSD)
iostat -x 1

# 2. Monitor CPU/RAM
top

# 3. Reduce query size
# Use limit on prefix queries
curl -X POST http://localhost:8545/api/v1/state/query/prefix \
  -d '{"prefix": "user:", "limit": 100}'

# 4. Implement caching in application
```

## Database Issues

### Database Corruption

**Error**: `Failed to open database` or `corrupted`

**Solution**:
```bash
# 1. Stop node
docker stop podoru-producer1

# 2. Try repair (if available)
# Or restore from backup

# 3. If repair fails, delete and resync
rm -rf docker/data/producer1/badger
docker start podoru-producer1
# Node will resync from peers
```

### Disk Space Full

**Error**: `no space left on device`

**Solution**:
```bash
# Check disk usage
df -h

# Check blockchain size
du -sh docker/data/*/badger

# Solutions:
# 1. Add more disk space
# 2. Enable pruning (future feature)
# 3. Use different disk
```

### Slow Database Performance

**Symptom**: Slow block processing

**Solutions**:
```bash
# 1. Use SSD for data_dir
# 2. Increase RAM
# 3. Run garbage collection
curl -X POST http://localhost:8545/api/v1/admin/gc  # If available
```

## Transaction Issues

### Transaction Not Confirmed

**Symptom**: Transaction submitted but not in blockchain

**Diagnosis**:
```bash
# Check mempool
curl http://localhost:8545/api/v1/mempool | jq '.data.count'

# Check transaction hash
curl http://localhost:8545/api/v1/transaction/<hash>
```

**Solutions**:
```bash
# 1. Verify transaction signature is valid
# 2. Check nonce is correct
# 3. Wait for next block (5 seconds default)
# 4. Check if producers are running
```

### Invalid Signature

**Error**: `Invalid transaction signature`

**Solution**:
```javascript
// Verify signing process
// 1. Build transaction object correctly
// 2. Calculate hash properly
// 3. Sign with correct private key
// 4. Include signature in transaction
```

### Invalid Nonce

**Error**: `Invalid nonce: expected X, got Y`

**Solution**:
```javascript
// Track nonce properly
let nonce = await getNonce(address)

// For each transaction:
tx.nonce = nonce
await submitTx(tx)
nonce++  // Increment for next transaction
```

## Docker Issues

### Container Won't Start

**Solution**:
```bash
# Check logs
docker logs podoru-producer1

# Check configuration
docker inspect podoru-producer1

# Verify volumes are mounted
docker inspect podoru-producer1 | jq '.[0].Mounts'

# Restart container
docker restart podoru-producer1
```

### Network Issues in Docker

**Solution**:
```bash
# Check Docker network
docker network inspect podoru_network

# Verify all containers are on same network
docker network ls

# Recreate network if needed
docker network rm podoru_network
docker network create podoru_network
```

## Performance Issues

### High CPU Usage

**Diagnosis**:
```bash
# Check CPU usage
top
htop

# Check which process
ps aux | grep podoru
```

**Solutions**:
```bash
# 1. Increase block_time to reduce processing
block_time: 10s

# 2. Limit peer connections
max_peers: 20

# 3. Disable API if not needed
api_enabled: false
```

### High Memory Usage

**Diagnosis**:
```bash
# Check memory
free -h

# Check process memory
ps aux | grep podoru | awk '{print $6}'
```

**Solutions**:
```bash
# 1. Reduce state cache (code change required)
# 2. Add more RAM
# 3. Monitor for memory leaks
```

## Getting Help

If your issue isn't listed here:

1. **Check logs**:
   ```bash
   docker logs podoru-producer1
   journalctl -u podoru-node
   ```

2. **Search GitHub Issues**:
   [github.com/podoru/podoru-chain/issues](https://github.com/podoru/podoru-chain/issues)

3. **Open New Issue**:
   Include:
   - Error messages
   - Configuration files (remove private keys!)
   - Steps to reproduce
   - System information

4. **Community Support**:
   - GitHub Discussions
   - Discord (if available)

## Debug Mode

Enable verbose logging (if available):

```yaml
# In config
log_level: debug
```

Or via environment:

```bash
LOG_LEVEL=debug ./bin/podoru-node -config config.yaml
```

## Common Commands Reference

```bash
# Check node status
curl http://localhost:8545/api/v1/node/health

# Check blockchain height
curl http://localhost:8545/api/v1/chain/info | jq '.data.height'

# Check peers
curl http://localhost:8545/api/v1/node/peers | jq '.data.peer_count'

# View logs
docker logs podoru-producer1 -f

# Restart node
docker restart podoru-producer1

# Check disk space
df -h

# Check ports
netstat -an | grep -E "8545|9000"
```
