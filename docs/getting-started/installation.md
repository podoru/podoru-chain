# Installation

This guide covers installing Podoru Chain and its dependencies on various platforms.

## System Requirements

### Minimum Requirements
- CPU: 2 cores
- RAM: 2 GB
- Disk: 10 GB free space
- OS: Linux, macOS, or Windows (WSL2)

### Recommended Requirements
- CPU: 4+ cores
- RAM: 4+ GB
- Disk: 50+ GB SSD
- OS: Ubuntu 20.04+ or macOS 12+

## Installing Prerequisites

### Go Programming Language

Podoru Chain requires Go 1.24 or higher.

#### Linux (Ubuntu/Debian)

```bash
# Download Go
wget https://go.dev/dl/go1.24.0.linux-amd64.tar.gz

# Remove old installation (if exists)
sudo rm -rf /usr/local/go

# Extract to /usr/local
sudo tar -C /usr/local -xzf go1.24.0.linux-amd64.tar.gz

# Add to PATH
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Verify installation
go version
```

#### macOS

```bash
# Using Homebrew
brew install go@1.24

# Or download from https://go.dev/dl/
```

#### Windows (WSL2)

Follow the Linux instructions within WSL2.

### Docker and Docker Compose

Docker is required for multi-node deployments.

#### Linux (Ubuntu/Debian)

```bash
# Update package index
sudo apt-get update

# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Add user to docker group
sudo usermod -aG docker $USER

# Log out and back in, then verify
docker --version

# Install Docker Compose
sudo apt-get install docker-compose-plugin

# Verify
docker compose version
```

#### macOS

```bash
# Download Docker Desktop from:
# https://www.docker.com/products/docker-desktop

# Or using Homebrew
brew install --cask docker

# Start Docker Desktop
open /Applications/Docker.app

# Verify
docker --version
docker compose version
```

### Dialog (for Setup Wizard)

The interactive setup wizard requires the `dialog` package.

#### Linux (Ubuntu/Debian)

```bash
sudo apt-get install dialog
```

#### macOS

```bash
brew install dialog
```

### Make Build Tool

Most Unix systems have Make pre-installed.

#### Linux (Ubuntu/Debian)

```bash
sudo apt-get install build-essential
```

#### macOS

```bash
xcode-select --install
```

### Git Version Control

#### Linux (Ubuntu/Debian)

```bash
sudo apt-get install git
```

#### macOS

```bash
brew install git
# Or: xcode-select --install
```

## Installing Podoru Chain

### Option 1: From Source (Recommended)

```bash
# Clone the repository
git clone https://github.com/podoru/podoru-chain.git
cd podoru-chain

# Install Go dependencies
make deps

# Build binaries
make build

# Verify build
./bin/podoru-node --help
./bin/keygen --help
```

This creates two binaries in the `bin/` directory:
- `podoru-node` - The main blockchain node
- `keygen` - Cryptographic key generation utility

### Option 2: Using Docker

```bash
# Clone the repository
git clone https://github.com/podoru/podoru-chain.git
cd podoru-chain

# Build Docker image
make docker

# Verify
docker images | grep podoru-chain
```

### Option 3: Download Pre-built Binaries

Pre-built binaries are available from the [GitHub Releases](https://github.com/podoru/podoru-chain/releases) page.

```bash
# Download latest release (Linux AMD64 example)
wget https://github.com/podoru/podoru-chain/releases/download/v1.0.0/podoru-chain-linux-amd64.tar.gz

# Extract
tar -xzf podoru-chain-linux-amd64.tar.gz

# Move to /usr/local/bin (optional)
sudo mv podoru-node keygen /usr/local/bin/

# Verify
podoru-node --help
```

## Verifying Installation

### Check Go Installation

```bash
go version
# Expected: go version go1.24.0 linux/amd64
```

### Check Docker Installation

```bash
docker --version
docker compose version

# Test Docker
docker run hello-world
```

### Check Podoru Chain Binaries

```bash
# If built from source
./bin/podoru-node --help
./bin/keygen --help

# If installed to PATH
podoru-node --help
keygen --help
```

Expected output:
```
╔═══════════════════════════════════════╗
║                                       ║
║        PODORU CHAIN v1.0.0            ║
║   Decentralized Blockchain Platform   ║
║                                       ║
╚═══════════════════════════════════════╝
```

## Build Options

Podoru Chain provides several build targets:

```bash
# Build everything
make build

# Build only the node
make build-node

# Build only keygen
make build-keygen

# Build for specific platform
GOOS=linux GOARCH=amd64 make build
GOOS=darwin GOARCH=arm64 make build

# Build Docker image
make docker

# Clean build artifacts
make clean
```

## Directory Structure

After installation, your directory should look like:

```
podoru-chain/
├── bin/                    # Compiled binaries
│   ├── podoru-node
│   └── keygen
├── cmd/                    # Source code for binaries
├── config/                 # Configuration examples
├── docker/                 # Docker setup files
├── internal/               # Core implementation
├── Makefile               # Build automation
└── README.md
```

## Updating Podoru Chain

To update to the latest version:

```bash
# Pull latest changes
git pull origin main

# Rebuild
make clean
make build

# Or rebuild Docker image
make docker
```

## Uninstalling

To remove Podoru Chain:

```bash
# Remove binaries
rm -rf bin/

# Remove build cache
make clean

# Remove Docker images
docker rmi podoru-chain:latest

# Remove data (CAUTION: This deletes blockchain data)
rm -rf docker/data/
```

## Next Steps

Now that you have Podoru Chain installed:

1. [Run the Setup Wizard](setup-wizard.md) for automated configuration
2. Or follow the [Manual Setup Guide](manual-setup.md) for custom configuration
3. Return to [Quick Start](README.md) to launch your network

## Troubleshooting Installation

### Go Version Too Old

If you see "requires Go 1.24 or higher":

```bash
# Check current version
go version

# Update Go using the installation steps above
```

### Docker Permission Denied

If you see "permission denied" when running Docker:

```bash
# Add user to docker group
sudo usermod -aG docker $USER

# Log out and back in
# Or run: newgrp docker
```

### Build Fails with Missing Dependencies

```bash
# Clear Go cache and retry
go clean -modcache
make deps
make build
```

### Port Already in Use

If ports 8545 or 9000 are already in use:

```bash
# Find process using the port
sudo lsof -i :8545

# Kill the process or use different ports in configuration
```

For more help, see the [Troubleshooting Guide](../troubleshooting/README.md).
