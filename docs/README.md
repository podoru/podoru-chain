# Podoru Chain Documentation

Welcome to the Podoru Chain documentation! Podoru Chain is a fully decentralized blockchain platform built in Go for storing arbitrary data as key-value pairs, powered by Proof of Authority (PoA) consensus.

## Features

- **Decentralized Architecture**: Fully distributed P2P network with TCP-based gossip protocol
- **Flexible Data Storage**: Store any data as key-value pairs with hierarchical naming conventions
- **Proof of Authority Consensus**: Fast and efficient block production with round-robin selection
- **Customizable Nodes**: Support for 1-10 producer nodes and 0-5 full nodes
- **High Performance**: BadgerDB embedded database for fast data access
- **REST API**: Comprehensive HTTP interface for blockchain interaction
- **Docker Ready**: Multi-node setup with automated Docker Compose configuration
- **Interactive Setup**: Wizard-based configuration for easy deployment

## Quick Navigation

### Getting Started
- [Quick Start](getting-started/README.md) - Get up and running in minutes
- [Installation](getting-started/installation.md) - Detailed installation guide
- [Setup Wizard](getting-started/setup-wizard.md) - Interactive configuration tool
- [Manual Setup](getting-started/manual-setup.md) - Advanced manual configuration

### Architecture
- [Overview](architecture/README.md) - System architecture and components
- [Consensus (PoA)](architecture/consensus.md) - Proof of Authority consensus mechanism
- [Storage (BadgerDB)](architecture/storage.md) - Data storage layer
- [P2P Networking](architecture/networking.md) - Peer-to-peer communication
- [Architecture Diagrams](architecture/diagrams.md) - System diagrams and visualizations

### API Reference
- [Overview](api-reference/README.md) - REST API introduction
- [Chain Endpoints](api-reference/chain.md) - Blockchain information
- [Block Endpoints](api-reference/blocks.md) - Block queries
- [Transaction Endpoints](api-reference/transactions.md) - Transaction submission
- [State Endpoints](api-reference/state.md) - Key-value data queries
- [Node Endpoints](api-reference/node.md) - Node information and health

### Development
- [Getting Started](development/README.md) - Build applications on Podoru Chain
- [Data Storage Patterns](development/data-patterns.md) - Best practices for data design
- [Querying Data](development/querying.md) - Advanced query techniques
- [Example Applications](development/examples.md) - Complete app examples

### CLI Reference
- [Overview](cli-reference/README.md) - Command-line tools
- [podoru-node](cli-reference/node.md) - Main blockchain node
- [keygen Tool](cli-reference/keygen.md) - Key generation utility

### Configuration
- [Overview](configuration/README.md) - Configuration guide
- [Genesis Block](configuration/genesis.md) - Initial blockchain state
- [Producer Nodes](configuration/producer.md) - Producer configuration
- [Full Nodes](configuration/fullnode.md) - Full node configuration

### Help
- [Troubleshooting](troubleshooting/README.md) - Common issues and solutions
- [Contributing](contributing/README.md) - Contribution guidelines

## Viewing Documentation

### Browse Markdown Files (Recommended)
The documentation is written in markdown and can be viewed directly:
- Browse files in this `docs/` directory
- Use any markdown viewer or editor
- View on GitHub (if pushed to a repository)

### Alternative Static Site Generators

**Using Docsify** (runtime renderer, no build needed):
```bash
npm install -g docsify-cli
docsify serve docs
# Open http://localhost:3000
```

**Using MkDocs** (Python-based):
```bash
pip install mkdocs
mkdocs serve
# Open http://localhost:8000
```

## About Podoru Chain

Podoru Chain is designed to be a flexible, developer-friendly blockchain platform that allows you to build any type of application by storing data as key-value pairs. Whether you're building a social network, file storage system, or e-commerce platform, Podoru Chain provides the infrastructure you need.

### Key Capabilities

- **Any Data Type**: Store JSON, binary data, files, or any custom format
- **Flexible Queries**: Single key, batch, and prefix-based queries
- **Transaction Support**: Submit multiple operations in a single transaction
- **Multi-Node Network**: Configurable number of producer and full nodes
- **Production Ready**: Docker deployment with health monitoring

## Quick Start

Get your blockchain network running in 3 steps:

```bash
# 1. Run the setup wizard
make setup-wizard

# 2. Follow the prompts to configure your network

# 3. Your blockchain is ready!
curl http://localhost:8545/api/v1/chain/info
```

See the [Quick Start Guide](getting-started/README.md) for detailed instructions.

## Community & Support

- **Issues**: Report bugs and request features
- **Contributing**: See [Contributing Guidelines](contributing/README.md)
- **License**: MIT License

---

Happy building with Podoru Chain! 🚀
