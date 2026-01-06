# keygen

Cryptographic key generation utility for Podoru Chain producer nodes.

## Synopsis

```bash
keygen [-output <file>] [-address]
```

## Description

`keygen` generates ECDSA keypairs using the secp256k1 curve (Ethereum-compatible). Producer nodes require a keypair for signing blocks.

## Options

### -output

Path to save the private key file.

```bash
keygen -output mynode.key
```

If not specified, the private key is printed to stdout (not recommended for production).

### -address

Show the derived Ethereum-compatible address (default: true).

```bash
keygen -address=false  # Don't show address
```

## Examples

### Generate and Save Key

```bash
./bin/keygen -output keys/producer1.key
```

**Output**:
```
Private key saved to: keys/producer1.key
Address: 0x3D4b25CBdda1014F74F9C80f040ce1Bb69130CBB
Public Key: 04a1b2c3d4e5f6...
```

### Generate Multiple Keys

```bash
./bin/keygen -output keys/producer1.key
./bin/keygen -output keys/producer2.key
./bin/keygen -output keys/producer3.key
```

### Print Key to Console

```bash
./bin/keygen
```

**Output**:
```
Private Key: 1234567890abcdef...
Address: 0x3D4b25CBdda1014F74F9C80f040ce1Bb69130CBB
Public Key: 04a1b2c3d4e5f6...
```

**Warning**: Don't use console output for production keys.

## Generated Files

### Key File Format

The generated private key file contains the raw private key bytes.

**File Structure**:
- Binary format
- 32 bytes (256 bits)
- No encryption (encrypt separately if needed)

**Permissions**:
```bash
-rw------- 1 user user 32 Jan  6 10:00 producer1.key
```

Always set restrictive permissions:

```bash
chmod 600 keys/producer1.key
```

## Address Derivation

The address is derived from the public key:

1. Generate ECDSA keypair (secp256k1)
2. Get public key (uncompressed, 65 bytes)
3. Hash public key with Keccak256
4. Take last 20 bytes
5. Prefix with "0x"

**Result**: Ethereum-compatible address like `0x3D4b25CBdda1014F74F9C80f040ce1Bb69130CBB`

## Security Considerations

### Key Storage

**Good Practices**:
- Store keys in encrypted filesystem
- Use restrictive file permissions (600)
- Keep backups in secure location
- Never commit keys to version control

**Bad Practices**:
- Storing keys in plaintext
- Sharing keys via email/chat
- Using same key for multiple nodes
- Committing to Git

### Key Backup

Always backup your keys:

```bash
# Encrypted backup
tar -czf keys-backup.tar.gz keys/
gpg -c keys-backup.tar.gz
rm keys-backup.tar.gz

# Store keys-backup.tar.gz.gpg securely
```

### Key Recovery

If you lose your private key:
- You cannot recover it
- The node cannot produce blocks
- You must generate a new key
- Update network configuration with new address

## Use Cases

### Setting Up Producer Nodes

```bash
# Generate keys for 3 producers
./bin/keygen -output docker/data/producer1/keys/producer1.key
./bin/keygen -output docker/data/producer2/keys/producer2.key
./bin/keygen -output docker/data/producer3/keys/producer3.key

# Note the addresses
# Add addresses to genesis.json authorities list
# Configure each producer with their respective key
```

### Testing

```bash
# Generate temporary key for testing
./bin/keygen -output /tmp/test.key
# Use for development, discard after
```

### Key Rotation

If you need to change keys:

1. Generate new key
2. Update node configuration
3. Coordinate with network to update authorities list
4. Securely destroy old key

## Integration with Node

Use the generated key in your node configuration:

```yaml
node_type: producer
address: "0x3D4b25CBdda1014F74F9C80f040ce1Bb69130CBB"  # From keygen output
private_key: "/path/to/keys/producer1.key"              # Path to generated key
```

## Troubleshooting

### Permission Denied

```bash
$ ./bin/keygen -output keys/producer1.key
Error creating directory: permission denied
```

**Solution**: Create directory first or use writable path:

```bash
mkdir -p keys
chmod 700 keys
./bin/keygen -output keys/producer1.key
```

### Key File Too Large/Small

The key file should be exactly 32 bytes.

```bash
$ ls -l keys/producer1.key
-rw------- 1 user user 32 Jan 6 10:00 keys/producer1.key
```

If different, regenerate:

```bash
rm keys/producer1.key
./bin/keygen -output keys/producer1.key
```

## Advanced Usage

### Programmatic Generation

Generate keys in scripts:

```bash
#!/bin/bash

for i in {1..5}; do
  echo "Generating key for producer${i}..."
  ./bin/keygen -output "keys/producer${i}.key"
done
```

### Extract Address from Existing Key

Currently not supported. Use `keygen` output during generation.

## Further Reading

- [Configuration Guide](../configuration/producer.md) - Using keys in config
- [Manual Setup](../getting-started/manual-setup.md) - Setting up nodes
- [Security](../architecture/README.md#security-model) - Security overview
