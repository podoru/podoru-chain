.PHONY: all build test clean clean-wizard docker docker-compose-up docker-compose-down keygen deps run setup-wizard join-info join-wizard update-node explorer-build explorer-dev explorer-docker stack-up stack-down stack-logs patch-genesis patch-all-genesis

# Build the node binary
build:
	@echo "Building Podoru Chain node..."
	@go build -o bin/podoru-node ./cmd/node
	@echo "Building keygen tool..."
	@go build -o bin/keygen ./cmd/tools/keygen
	@echo "Build complete!"

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -rf docker/data/
	@rm -f coverage.out coverage.html
	@echo "Clean complete!"

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy
	@echo "Dependencies installed!"

# Build Docker image
docker:
	@echo "Building Docker image..."
	@docker build -t podoru-chain:latest -f docker/Dockerfile .
	@echo "Docker image built!"

# Start multi-node setup with Docker Compose
docker-compose-up:
	@echo "Starting Podoru Chain network..."
	@cd docker && docker-compose up -d
	@echo "Network started! View logs with: make docker-compose-logs"

# Stop Docker Compose
docker-compose-down:
	@echo "Stopping Podoru Chain network..."
	@cd docker && docker-compose down
	@echo "Network stopped!"

# View Docker Compose logs
docker-compose-logs:
	@cd docker && docker-compose logs -f

# Generate a new key pair
keygen:
	@echo "Generating new key pair..."
	@go run ./cmd/tools/keygen/main.go

# Generate keys for all producers
keygen-all:
	@echo "Generating keys for all producers..."
	@mkdir -p docker/data/producer1/keys
	@mkdir -p docker/data/producer2/keys
	@mkdir -p docker/data/producer3/keys
	@mkdir -p docker/data/fullnode1
	@go run ./cmd/tools/keygen/main.go -output docker/data/producer1/keys/producer1.key
	@go run ./cmd/tools/keygen/main.go -output docker/data/producer2/keys/producer2.key
	@go run ./cmd/tools/keygen/main.go -output docker/data/producer3/keys/producer3.key
	@echo "Keys generated!"

# Setup Docker data directories
setup-docker:
	@echo "Setting up Docker data directories..."
	@mkdir -p docker/data/producer1/keys
	@mkdir -p docker/data/producer2/keys
	@mkdir -p docker/data/producer3/keys
	@mkdir -p docker/data/fullnode1
	@cp config/producer1.yaml docker/data/producer1/config.yaml
	@cp config/producer2.yaml docker/data/producer2/config.yaml
	@cp config/producer3.yaml docker/data/producer3/config.yaml
	@cp config/fullnode.yaml docker/data/fullnode1/config.yaml
	@cp config/genesis.json docker/data/producer1/genesis.json
	@cp config/genesis.json docker/data/producer2/genesis.json
	@cp config/genesis.json docker/data/producer3/genesis.json
	@cp config/genesis.json docker/data/fullnode1/genesis.json
	@echo "Docker setup complete!"

# Run a single node locally
run:
	@if [ -z "$(CONFIG)" ]; then \
		echo "Error: CONFIG variable is required. Usage: make run CONFIG=config/producer1.yaml"; \
		exit 1; \
	fi
	@echo "Running node with config: $(CONFIG)"
	@go run ./cmd/node/main.go -config $(CONFIG)

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "Format complete!"

# Run linter
lint:
	@echo "Running linter..."
	@golangci-lint run || echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"

# Run interactive setup wizard
setup-wizard:
	@echo "Starting Podoru Chain Setup Wizard..."
	@bash scripts/setup-wizard.sh

# Clean wizard-generated data (requires sudo for Docker-created files)
clean-wizard:
	@echo "Cleaning wizard-generated data..."
	@cd docker && docker-compose down 2>/dev/null || true
	@sudo rm -rf docker/data/producer* docker/data/fullnode* docker/data/genesis.json 2>/dev/null || true
	@rm -f docker/docker-compose.yml 2>/dev/null || true
	@echo "Wizard data cleaned! Run 'make setup-wizard' to start fresh."

# Generate join information for others to connect to your network
join-info:
	@echo "Generating network join information..."
	@bash scripts/generate-join-info.sh

# Join an existing network using a join-info tarball
join-wizard:
	@echo "Starting Podoru Chain Join Wizard..."
	@bash scripts/join-wizard.sh

# Update existing node with latest code from git
update-node:
	@echo "Updating Podoru Chain node..."
	@bash scripts/update-node.sh

# Build explorer
explorer-build:
	@echo "Building explorer..."
	@cd explorer && npm install && npm run build
	@echo "Explorer built!"

# Run explorer in dev mode
explorer-dev:
	@echo "Starting explorer in development mode..."
	@cd explorer && npm run dev

# Build explorer Docker image
explorer-docker:
	@echo "Building explorer Docker image..."
	@docker build -t podoru-explorer:latest -f explorer/Dockerfile explorer/
	@echo "Explorer Docker image built!"

# Start full stack (nodes + explorer)
stack-up: docker
	@echo "Starting full stack (nodes + explorer)..."
	@cd docker && docker-compose up -d
	@echo "Stack started!"
	@echo ""
	@echo "Access points:"
	@echo "  Explorer:     http://localhost:3000"
	@echo "  Producer API: http://localhost:8545"
	@echo "  Fullnode API: http://localhost:8546"
	@echo ""
	@echo "View logs with: make stack-logs"

# Stop full stack
stack-down:
	@echo "Stopping full stack..."
	@cd docker && docker-compose down
	@echo "Stack stopped!"

# View stack logs
stack-logs:
	@cd docker && docker-compose logs -f

# Patch genesis.json with token and gas configuration
patch-genesis:
	@echo "Patching genesis.json with token and gas configuration..."
	@if [ -z "$(GENESIS)" ]; then \
		bash scripts/patch-genesis.sh; \
	else \
		bash scripts/patch-genesis.sh "$(GENESIS)" "$(OUTPUT)"; \
	fi

# Patch all genesis files in docker/data with token and gas configuration
patch-all-genesis:
	@echo "Patching all genesis files in docker/data..."
	@for f in docker/data/*/genesis.json; do \
		if [ -f "$$f" ]; then \
			echo "Patching $$f..."; \
			bash scripts/patch-genesis.sh "$$f" "$$f"; \
		fi; \
	done
	@echo "All genesis files patched!"

# Preview genesis patch without applying
patch-genesis-dry-run:
	@echo "Previewing genesis patch..."
	@if [ -z "$(GENESIS)" ]; then \
		bash scripts/patch-genesis.sh --dry-run; \
	else \
		bash scripts/patch-genesis.sh --dry-run "$(GENESIS)"; \
	fi

# Show help
help:
	@echo "Podoru Chain Makefile Commands:"
	@echo ""
	@echo "Quick Start:"
	@echo "  make setup-wizard      - Run interactive setup wizard (recommended)"
	@echo "  make stack-up          - Start full stack (nodes + explorer)"
	@echo "  make stack-down        - Stop full stack"
	@echo "  make stack-logs        - View all logs"
	@echo ""
	@echo "Explorer:"
	@echo "  make explorer-build    - Build explorer (npm install & build)"
	@echo "  make explorer-dev      - Run explorer in development mode"
	@echo "  make explorer-docker   - Build explorer Docker image"
	@echo ""
	@echo "Node Operations:"
	@echo "  make clean-wizard      - Clean wizard-generated data and start fresh"
	@echo "  make join-info         - Generate info for others to join your network"
	@echo "  make join-wizard       - Join an existing network with a tarball"
	@echo "  make update-node       - Pull latest code and update running node"
	@echo "  make build             - Build the node and keygen binaries"
	@echo "  make test              - Run tests"
	@echo "  make test-coverage     - Run tests with coverage report"
	@echo "  make clean             - Clean build artifacts"
	@echo "  make deps              - Install dependencies"
	@echo ""
	@echo "Docker:"
	@echo "  make docker            - Build Docker image"
	@echo "  make setup-docker      - Setup Docker data directories"
	@echo "  make keygen-all        - Generate keys for all producers"
	@echo "  make docker-compose-up - Start multi-node network"
	@echo "  make docker-compose-down - Stop multi-node network"
	@echo "  make docker-compose-logs - View network logs"
	@echo ""
	@echo "Genesis Migration:"
	@echo "  make patch-genesis         - Patch config/genesis.json with token/gas config"
	@echo "  make patch-genesis GENESIS=<path> OUTPUT=<path> - Patch specific genesis file"
	@echo "  make patch-all-genesis     - Patch all genesis files in docker/data"
	@echo "  make patch-genesis-dry-run - Preview genesis patch without applying"
	@echo ""
	@echo "Development:"
	@echo "  make run CONFIG=<path> - Run a single node locally"
	@echo "  make fmt               - Format code"
	@echo "  make lint              - Run linter"
	@echo "  make help              - Show this help message"
	@echo ""

# Default target
all: deps build

