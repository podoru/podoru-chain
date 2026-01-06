# CLI Reference

Command-line tools for running and managing Podoru Chain nodes.

## Available Commands

Podoru Chain provides two command-line tools:

### podoru-node

The main blockchain node executable.

- **Purpose**: Run a blockchain node (producer or full node)
- **Location**: `bin/podoru-node`
- **Documentation**: [podoru-node Reference](node.md)

### keygen

Cryptographic key generation utility.

- **Purpose**: Generate keypairs for producer nodes
- **Location**: `bin/keygen`
- **Documentation**: [keygen Reference](keygen.md)

## Installation

### From Source

```bash
git clone https://github.com/podoru/podoru-chain.git
cd podoru-chain
make build
```

Binaries will be in the `bin/` directory.

### Pre-built Binaries

Download from [GitHub Releases](https://github.com/podoru/podoru-chain/releases).

## Quick Reference

### Running a Node

```bash
./bin/podoru-node -config path/to/config.yaml
```

### Generating Keys

```bash
./bin/keygen -output mynode.key
```

## Common Tasks

### Start Producer Node

```bash
./bin/podoru-node -config config/producer1.yaml
```

### Start Full Node

```bash
./bin/podoru-node -config config/fullnode.yaml
```

### Generate Producer Keys

```bash
./bin/keygen -output keys/producer1.key
```

### View Help

```bash
./bin/podoru-node --help
./bin/keygen --help
```

## Environment Variables

Currently, Podoru Chain does not use environment variables for configuration. All configuration is done via YAML files.

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Error (configuration, startup, runtime) |
| 130 | Interrupted (Ctrl+C) |

## Further Reading

- [podoru-node Command](node.md) - Node executable reference
- [keygen Command](keygen.md) - Key generation reference
- [Configuration Guide](../configuration/README.md) - Configuration files
- [Getting Started](../getting-started/README.md) - Setup guide
