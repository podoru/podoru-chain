# Quick Start Guide

Get your Podoru Chain network up and running in minutes using our interactive setup wizard.

## Prerequisites

Before you begin, make sure you have the following installed:

- **Go 1.24 or higher** - [Download Go](https://golang.org/dl/)
- **Docker and Docker Compose** - [Install Docker](https://docs.docker.com/get-docker/)
- **dialog** (for interactive setup wizard)
  - Ubuntu/Debian: `sudo apt-get install dialog`
  - macOS: `brew install dialog`
  - Fedora/RHEL: `sudo dnf install dialog`
- **Make** - Usually pre-installed on Unix systems
- **Git** - [Install Git](https://git-scm.com/downloads)

## Quick Start (Recommended)

The fastest way to get started is using our interactive setup wizard:

### 1. Clone the Repository

```bash
git clone https://github.com/podoru/podoru-chain.git
cd podoru-chain
```

### 2. Run the Setup Wizard

```bash
make setup-wizard
```

The wizard will guide you through:
- Configuring the number of producer and full nodes
- Setting chain metadata (name, description, block time)
- Configuring network ports
- Building Docker images
- Starting the network

### 3. Verify Your Network

Once the wizard completes, test your blockchain:

```bash
# Check blockchain info
curl http://localhost:8545/api/v1/chain/info | jq

# Query genesis data
curl http://localhost:8545/api/v1/state/chain:name | jq
```

Expected output:
```json
{
  "success": true,
  "data": {
    "key": "chain:name",
    "value": "Podoru Chain"
  }
}
```

### 4. Start Building

Your blockchain is now running! You can:
- Submit transactions via the REST API
- Query blockchain data
- Build decentralized applications

See the [Development Guide](../development/README.md) for examples.

## Management Commands

After setup, use these commands to manage your network:

```bash
# Start the network
make docker-compose-up

# Stop the network
make docker-compose-down

# View logs from all nodes
make docker-compose-logs

# View logs from a specific node
docker logs podoru-producer1 -f

# Restart a single node
docker restart podoru-producer1
```

## Default Network Configuration

The wizard creates the following default setup:

| Node | Type | API Port | P2P Port |
|------|------|----------|----------|
| Producer 1 | Producer | 8545 | 9000 |
| Producer 2 | Producer | 8546 | 9001 |
| Producer 3 | Producer | 8547 | 9002 |
| Full Node | Full | 8548 | 9003 |

You can customize these settings in the wizard.

## Next Steps

- **Learn More**: Check out the [Architecture Overview](../architecture/README.md)
- **Build Apps**: Read the [Development Guide](../development/README.md)
- **API Reference**: Explore the [REST API](../api-reference/README.md)
- **Advanced Setup**: See [Manual Setup](manual-setup.md) for custom configurations

## Alternative Setup Methods

If you prefer more control or want to understand the setup process:

- [Installation](installation.md) - Manual installation steps
- [Setup Wizard](setup-wizard.md) - Detailed wizard documentation
- [Manual Setup](manual-setup.md) - Complete manual configuration guide

## Troubleshooting

If you encounter issues during setup:

1. Ensure all prerequisites are installed
2. Check Docker is running: `docker ps`
3. Verify ports are available: `netstat -an | grep 8545`
4. See [Troubleshooting Guide](../troubleshooting/README.md)

## Getting Help

- Check the [Troubleshooting Guide](../troubleshooting/README.md)
- Search [GitHub Issues](https://github.com/podoru/podoru-chain/issues)
- Open a new issue if you need help
