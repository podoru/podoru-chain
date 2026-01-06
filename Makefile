.PHONY: all build test clean clean-wizard docker docker-compose-up docker-compose-down keygen deps run setup-wizard join-info join-wizard

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

# Show help
help:
	@echo "Podoru Chain Makefile Commands:"
	@echo ""
	@echo "  make setup-wizard      - Run interactive setup wizard (recommended)"
	@echo "  make clean-wizard      - Clean wizard-generated data and start fresh"
	@echo "  make join-info         - Generate info for others to join your network"
	@echo "  make join-wizard       - Join an existing network with a tarball"
	@echo "  make build             - Build the node and keygen binaries"
	@echo "  make test              - Run tests"
	@echo "  make test-coverage     - Run tests with coverage report"
	@echo "  make clean             - Clean build artifacts"
	@echo "  make deps              - Install dependencies"
	@echo "  make docker            - Build Docker image"
	@echo "  make setup-docker      - Setup Docker data directories"
	@echo "  make keygen-all        - Generate keys for all producers"
	@echo "  make docker-compose-up - Start multi-node network"
	@echo "  make docker-compose-down - Stop multi-node network"
	@echo "  make docker-compose-logs - View network logs"
	@echo "  make run CONFIG=<path> - Run a single node locally"
	@echo "  make fmt               - Format code"
	@echo "  make lint              - Run linter"
	@echo "  make help              - Show this help message"
	@echo ""

# Default target
all: deps build

