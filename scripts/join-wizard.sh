#!/bin/bash
# Podoru Chain Join Wizard
# Helps users join an existing Podoru Chain network as a full node
# Simple terminal version (no dialog required)

set -e

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Source helper libraries
source "$SCRIPT_DIR/lib/utils.sh"

# Configuration
SETUP_DIR="./podoru-fullnode"
CREATED_SETUP_DIR=false

#==============================================================================
# Helper Functions
#==============================================================================

show_welcome() {
    clear
    echo ""
    echo -e "${BLUE}╔═══════════════════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║                                                       ║${NC}"
    echo -e "${BLUE}║           Podoru Chain Join Wizard                    ║${NC}"
    echo -e "${BLUE}║                                                       ║${NC}"
    echo -e "${BLUE}║      Join an existing Podoru network                  ║${NC}"
    echo -e "${BLUE}║      as a full node                                   ║${NC}"
    echo -e "${BLUE}║                                                       ║${NC}"
    echo -e "${BLUE}╚═══════════════════════════════════════════════════════╝${NC}"
    echo ""
    echo "This wizard will help you:"
    echo "  • Extract and validate a join-info tarball"
    echo "  • Configure your full node"
    echo "  • Check for port conflicts"
    echo "  • Start your node automatically"
    echo ""
}

check_prerequisites() {
    echo "Checking prerequisites..."
    echo ""

    local missing=()
    local warnings=()

    # Check Docker
    if command -v docker &> /dev/null; then
        local docker_version=$(docker --version | awk '{print $3}' | sed 's/,//')
        echo "✓ Docker installed: $docker_version"

        if docker ps &> /dev/null; then
            echo "✓ Docker daemon is running"
        else
            warnings+=("Docker daemon is not running")
            echo "⚠ Docker daemon is not running"
        fi
    else
        missing+=("docker")
        echo "✗ Docker is not installed"
    fi

    # Check Docker Compose
    if docker compose version &> /dev/null; then
        local compose_version=$(docker compose version | awk '{print $4}')
        echo "✓ Docker Compose installed: $compose_version"
    elif command -v docker-compose &> /dev/null; then
        local compose_version=$(docker-compose --version | awk '{print $4}')
        echo "✓ Docker Compose (legacy) installed: $compose_version"
    else
        missing+=("docker-compose")
        echo "✗ Docker Compose is not installed"
    fi

    # Check tar
    if command -v tar &> /dev/null; then
        echo "✓ tar installed"
    else
        missing+=("tar")
        echo "✗ tar is not installed"
    fi

    # Check jq
    if command -v jq &> /dev/null; then
        echo "✓ jq installed"
    else
        missing+=("jq")
        echo "✗ jq is not installed"
    fi

    echo ""

    # Show installation instructions if missing
    if [ ${#missing[@]} -gt 0 ]; then
        echo "Missing required tools: ${missing[*]}"
        echo ""
        echo "Installation instructions:"
        echo ""

        for tool in "${missing[@]}"; do
            case $tool in
                docker)
                    echo "  Docker:"
                    echo "    Ubuntu:  sudo apt-get install -y docker.io"
                    echo "    macOS:   Download Docker Desktop from https://docker.com"
                    echo ""
                    ;;
                docker-compose)
                    echo "  Docker Compose:"
                    echo "    Ubuntu:  sudo apt-get install -y docker-compose-v2"
                    echo "    macOS:   Included with Docker Desktop"
                    echo ""
                    ;;
                tar)
                    echo "  tar:"
                    echo "    Ubuntu:  sudo apt-get install -y tar"
                    echo ""
                    ;;
                jq)
                    echo "  jq:"
                    echo "    Ubuntu:  sudo apt-get install -y jq"
                    echo "    macOS:   brew install jq"
                    echo ""
                    ;;
            esac
        done

        exit 1
    fi

    # Show warnings
    if [ ${#warnings[@]} -gt 0 ]; then
        echo "Warnings:"
        for warning in "${warnings[@]}"; do
            echo "  ⚠ $warning"
        done
        echo ""
        read -p "Press Enter to continue or Ctrl+C to exit..."
    fi

    echo "All prerequisites satisfied!"
    echo ""
}

get_tarball_path() {
    echo "Enter the path to the join-info tarball:"
    echo ""
    echo "  Example: ~/podoru-chain-join-info.tar.gz"
    echo "  Example: /path/to/podoru-chain-join-info.tar.gz"
    echo ""

    while true; do
        read -p "Tarball path: " tarball

        # Handle empty input
        if [ -z "$tarball" ]; then
            echo "Error: Please enter a path"
            continue
        fi

        # Expand tilde
        tarball="${tarball/#\~/$HOME}"

        # Validate file exists
        if [ ! -f "$tarball" ]; then
            echo "Error: File not found: $tarball"
            continue
        fi

        # Validate file is readable
        if [ ! -r "$tarball" ]; then
            echo "Error: Cannot read file: $tarball"
            continue
        fi

        echo "$tarball"
        return 0
    done
}

check_setup_directory() {
    if [ -d "$SETUP_DIR" ]; then
        echo ""
        echo "Warning: Setup directory already exists: $SETUP_DIR"
        echo ""
        read -p "Overwrite existing directory? (y/n): " overwrite

        if [ "$overwrite" = "y" ] || [ "$overwrite" = "Y" ]; then
            log_info "Removing existing setup directory..."
            rm -rf "$SETUP_DIR"
        else
            log_info "Setup cancelled by user"
            exit 0
        fi
    fi
}

extract_tarball() {
    local tarball="$1"

    log_info "Extracting tarball..."

    # Create setup directory
    mkdir -p "$SETUP_DIR"
    CREATED_SETUP_DIR=true

    # Extract tarball
    if ! tar -xzf "$tarball" -C "$SETUP_DIR" 2>/dev/null; then
        log_error "Failed to extract tarball"
        return 1
    fi

    # Validate required files
    local missing=()

    if [ ! -f "$SETUP_DIR/genesis.json" ]; then
        missing+=("genesis.json")
    fi

    if [ ! -f "$SETUP_DIR/config.yaml" ]; then
        missing+=("config.yaml")
    fi

    if [ ! -f "$SETUP_DIR/docker-compose.yml" ]; then
        missing+=("docker-compose.yml")
    fi

    if [ ${#missing[@]} -gt 0 ]; then
        log_error "Missing required files in tarball: ${missing[*]}"
        echo ""
        echo "Please get a valid join-info tarball from the network operator."
        return 1
    fi

    log_success "Tarball extracted successfully"
    return 0
}

validate_genesis() {
    local genesis_file="$SETUP_DIR/genesis.json"

    # Read network info
    local chain_name=$(jq -r '.initial_state["chain:name"]' "$genesis_file" 2>/dev/null || echo "Unknown")
    local chain_desc=$(jq -r '.initial_state["chain:description"]' "$genesis_file" 2>/dev/null || echo "Unknown")
    local authorities_count=$(jq -r '.authorities | length' "$genesis_file" 2>/dev/null || echo "0")
    local genesis_hash=$(sha256sum "$genesis_file" | awk '{print $1}')

    echo ""
    echo "Network Information:"
    echo "  Chain Name:   $chain_name"
    echo "  Description:  $chain_desc"
    echo "  Authorities:  $authorities_count"
    echo "  Genesis Hash: ${genesis_hash:0:16}..."
    echo ""

    read -p "Is this the correct network? (y/n): " confirm

    if [ "$confirm" = "y" ] || [ "$confirm" = "Y" ]; then
        return 0
    else
        log_info "Genesis validation cancelled by user"
        return 1
    fi
}

check_port_conflicts() {
    local api_port=$(grep -oP '"\K\d+(?=:8545")' "$SETUP_DIR/docker-compose.yml" | head -1 || echo "8545")
    local p2p_port=$(grep -oP '"\K\d+(?=:9000")' "$SETUP_DIR/docker-compose.yml" | head -1 || echo "9000")

    local conflicts=()
    local new_api_port=$api_port
    local new_p2p_port=$p2p_port

    # Check API port
    if is_port_in_use "$api_port" 2>/dev/null; then
        conflicts+=("API port $api_port is in use")

        # Suggest alternative
        for port in {8546..8550}; do
            if ! is_port_in_use "$port" 2>/dev/null; then
                new_api_port=$port
                break
            fi
        done
    fi

    # Check P2P port
    if is_port_in_use "$p2p_port" 2>/dev/null; then
        conflicts+=("P2P port $p2p_port is in use")

        # Suggest alternative
        for port in {9001..9010}; do
            if ! is_port_in_use "$port" 2>/dev/null; then
                new_p2p_port=$port
                break
            fi
        done
    fi

    # If conflicts, ask user for resolution
    if [ ${#conflicts[@]} -gt 0 ]; then
        echo ""
        echo "Port conflicts detected:"
        for conflict in "${conflicts[@]}"; do
            echo "  ⚠ $conflict"
        done
        echo ""
        echo "Suggested alternatives:"

        read -p "API Port [$new_api_port]: " input_api
        new_api_port=${input_api:-$new_api_port}

        read -p "P2P Port [$new_p2p_port]: " input_p2p
        new_p2p_port=${input_p2p:-$new_p2p_port}

        # Update docker-compose.yml
        sed -i "s/\"[0-9]\+:8545\"/\"$new_api_port:8545\"/" "$SETUP_DIR/docker-compose.yml"
        sed -i "s/\"[0-9]\+:9000\"/\"$new_p2p_port:9000\"/" "$SETUP_DIR/docker-compose.yml"

        log_success "Ports updated: API=$new_api_port, P2P=$new_p2p_port"
    fi

    echo "$new_api_port:$new_p2p_port"
    return 0
}

review_configuration() {
    local api_port=$(grep -oP '"\K\d+(?=:8545")' "$SETUP_DIR/docker-compose.yml" | head -1 || echo "8545")
    local p2p_port=$(grep -oP '"\K\d+(?=:9000")' "$SETUP_DIR/docker-compose.yml" | head -1 || echo "9000")
    local chain_name=$(jq -r '.initial_state["chain:name"]' "$SETUP_DIR/genesis.json" 2>/dev/null || echo "Unknown")
    local bootstrap_peers=$(grep -A 10 "bootstrap_peers:" "$SETUP_DIR/config.yaml" | grep '  - ' | sed 's/.*"\(.*\)".*/\1/' | tr '\n' ', ' | sed 's/,$//')

    echo ""
    echo "Configuration Summary:"
    echo "  Network Name:    $chain_name"
    echo "  Bootstrap Peers: $bootstrap_peers"
    echo "  API Port:        $api_port"
    echo "  P2P Port:        $p2p_port"
    echo "  Setup Directory: $SETUP_DIR"
    echo ""

    read -p "Proceed with setup? (y/n): " confirm

    if [ "$confirm" = "y" ] || [ "$confirm" = "Y" ]; then
        return 0
    else
        return 1
    fi
}

start_node() {
    local api_port=$(grep -oP '"\K\d+(?=:8545")' "$SETUP_DIR/docker-compose.yml" | head -1 || echo "8545")

    log_info "Starting node..."

    # Start docker-compose
    cd "$SETUP_DIR"

    if ! docker-compose up -d 2>&1; then
        log_error "Failed to start node"
        echo ""
        echo "You can try manually:"
        echo "  cd $SETUP_DIR"
        echo "  docker-compose up -d"
        cd - > /dev/null
        return 1
    fi

    cd - > /dev/null

    log_info "Waiting for node to start..."
    sleep 5

    # Check node status
    local node_info=$(curl -s http://localhost:$api_port/api/v1/node/info 2>/dev/null || echo "")

    if [ -z "$node_info" ]; then
        log_warning "Node may still be starting up"
    else
        log_success "Node is running!"
    fi

    return 0
}

show_completion() {
    local api_port=$(grep -oP '"\K\d+(?=:8545")' "$SETUP_DIR/docker-compose.yml" | head -1 || echo "8545")

    echo ""
    echo -e "${GREEN}╔═══════════════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║                                                       ║${NC}"
    echo -e "${GREEN}║           Setup Complete!                             ║${NC}"
    echo -e "${GREEN}║                                                       ║${NC}"
    echo -e "${GREEN}╚═══════════════════════════════════════════════════════╝${NC}"
    echo ""
    echo "Your Podoru Chain full node is now running!"
    echo ""
    echo "Node directory: $SETUP_DIR"
    echo "API endpoint:   http://localhost:$api_port"
    echo ""
    echo -e "${YELLOW}Useful commands:${NC}"
    echo ""
    echo "  • View logs:   cd $SETUP_DIR && docker-compose logs -f"
    echo "  • Check node:  curl http://localhost:$api_port/api/v1/node/info"
    echo "  • Chain info:  curl http://localhost:$api_port/api/v1/chain/info"
    echo "  • Stop node:   cd $SETUP_DIR && docker-compose down"
    echo ""
}

cleanup_on_error() {
    local exit_code=$?

    if [ $exit_code -ne 0 ] && [ "$CREATED_SETUP_DIR" = true ]; then
        echo ""
        log_warning "Setup failed or was cancelled"

        read -p "Do you want to remove the setup directory? (y/n): " CLEANUP

        if [ "$CLEANUP" = "y" ] || [ "$CLEANUP" = "Y" ]; then
            rm -rf "$SETUP_DIR"
            log_info "Setup directory removed"
        else
            log_info "Setup directory kept at: $SETUP_DIR"
        fi
    fi
}

#==============================================================================
# Main Execution
#==============================================================================

main() {
    # Setup error handling
    trap cleanup_on_error EXIT

    # Show welcome screen
    show_welcome

    # Check prerequisites
    check_prerequisites

    # Get tarball path
    TARBALL_PATH=$(get_tarball_path)

    if [ -z "$TARBALL_PATH" ]; then
        log_error "No tarball specified"
        exit 1
    fi

    log_success "Tarball: $TARBALL_PATH"

    # Check if setup directory exists
    check_setup_directory

    # Extract tarball
    if ! extract_tarball "$TARBALL_PATH"; then
        exit 1
    fi

    # Validate genesis
    if ! validate_genesis; then
        log_info "Setup cancelled by user"
        exit 0
    fi

    # Check for port conflicts
    if ! check_port_conflicts; then
        log_info "Setup cancelled by user"
        exit 0
    fi

    # Review configuration
    if ! review_configuration; then
        log_info "Setup cancelled by user"
        exit 0
    fi

    # Start node
    if ! start_node; then
        exit 1
    fi

    # Show completion
    show_completion

    # Disable cleanup on success
    trap - EXIT
}

# Run main function
main
