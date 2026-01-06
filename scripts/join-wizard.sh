#!/bin/bash
# Podoru Chain Join Wizard
# Helps users join an existing Podoru Chain network as a full node

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
    log_info "Checking prerequisites..."

    local missing=()

    # Check dialog
    if ! command -v dialog &> /dev/null; then
        missing+=("dialog")
    fi

    # Check Docker
    if ! is_docker_running; then
        log_error "Docker is not running"
        echo ""
        echo "Please start Docker before running this wizard."
        exit 1
    fi

    # Check docker-compose
    if ! has_docker_compose; then
        missing+=("docker-compose")
    fi

    # Check tar
    if ! command -v tar &> /dev/null; then
        missing+=("tar")
    fi

    # Check jq
    if ! command -v jq &> /dev/null; then
        missing+=("jq")
    fi

    if [ ${#missing[@]} -gt 0 ]; then
        log_error "Missing required tools: ${missing[*]}"
        echo ""
        echo "Please install missing tools:"
        for tool in "${missing[@]}"; do
            echo "  - $tool"
        done
        exit 1
    fi

    log_success "Prerequisites check passed"
}

get_tarball_path() {
    while true; do
        exec 3>&1
        local tarball=$(dialog --title "Join-Info Tarball" \
            --inputbox "Enter the path to the join-info tarball:

Suggested location:
~/podoru-chain/podoru-chain-join-info.tar.gz

Or provide full path to the tarball file:" \
            15 70 \
            2>&1 1>&3)
        exec 3>&-

        if [ $? -ne 0 ]; then
            log_info "Setup cancelled by user"
            exit 0
        fi

        # Expand tilde
        tarball="${tarball/#\~/$HOME}"

        # Validate file exists
        if [ ! -f "$tarball" ]; then
            dialog --title "Error" --msgbox "File not found: $tarball

Please check the path and try again." 10 60
            continue
        fi

        # Validate file is readable
        if [ ! -r "$tarball" ]; then
            dialog --title "Error" --msgbox "Cannot read file: $tarball

Please check permissions and try again." 10 60
            continue
        fi

        echo "$tarball"
        return 0
    done
}

check_setup_directory() {
    if [ -d "$SETUP_DIR" ]; then
        dialog --title "Directory Exists" --yesno \
            "The setup directory already exists:
$SETUP_DIR

Do you want to overwrite it?

Yes: Remove existing directory and continue
No: Exit wizard" \
            12 60

        case $? in
            0)  # Yes - remove
                log_info "Removing existing setup directory..."
                rm -rf "$SETUP_DIR"
                ;;
            *)  # No or ESC
                log_info "Setup cancelled by user"
                exit 0
                ;;
        esac
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
        dialog --title "Invalid Tarball" --msgbox \
            "The tarball is missing required files:

$(printf '%s\n' "${missing[@]}")

Please get a valid join-info tarball from the network operator." \
            15 60
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

    # Show network information
    dialog --title "Network Information" --yesno \
        "Please verify this is the correct network:

Chain Name: $chain_name
Description: $chain_desc
Authorities: $authorities_count
Genesis Hash: ${genesis_hash:0:16}...

Is this the correct network?

Yes: Continue with setup
No: Exit wizard" \
        16 70

    case $? in
        0)  # Yes
            return 0
            ;;
        *)  # No or ESC
            log_info "Genesis validation cancelled by user"
            return 1
            ;;
    esac
}

check_port_conflicts() {
    local api_port=$(grep -oP '"\K\d+(?=:8545")' "$SETUP_DIR/docker-compose.yml" | head -1 || echo "8545")
    local p2p_port=$(grep -oP '"\K\d+(?=:9000")' "$SETUP_DIR/docker-compose.yml" | head -1 || echo "9000")

    local conflicts=()
    local new_api_port=$api_port
    local new_p2p_port=$p2p_port

    # Check API port
    if is_port_in_use "$api_port"; then
        conflicts+=("API port $api_port is in use")

        # Suggest alternative
        for port in {8546..8550}; do
            if ! is_port_in_use "$port"; then
                new_api_port=$port
                break
            fi
        done
    fi

    # Check P2P port
    if is_port_in_use "$p2p_port"; then
        conflicts+=("P2P port $p2p_port is in use")

        # Suggest alternative
        for port in {9001..9010}; do
            if ! is_port_in_use "$port"; then
                new_p2p_port=$port
                break
            fi
        done
    fi

    # If conflicts, ask user for resolution
    if [ ${#conflicts[@]} -gt 0 ]; then
        exec 3>&1
        local result=$(dialog --title "Port Conflicts" \
            --form "Port conflicts detected. Please choose alternative ports:

Conflicts:
$(printf '%s\n' "${conflicts[@]}")
" \
            15 70 2 \
            "API Port:" 1 1 "$new_api_port" 1 15 10 0 \
            "P2P Port:" 2 1 "$new_p2p_port" 2 15 10 0 \
            2>&1 1>&3)
        exec 3>&-

        if [ $? -ne 0 ]; then
            return 1
        fi

        # Parse results
        new_api_port=$(echo "$result" | sed -n '1p')
        new_p2p_port=$(echo "$result" | sed -n '2p')

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

    dialog --title "Review Configuration" --yesno \
        "Please review your configuration:

Network Name: $chain_name
Bootstrap Peers: $bootstrap_peers
API Port: $api_port
P2P Port: $p2p_port
Setup Directory: $SETUP_DIR

Proceed with setup?

Yes: Start the node
No: Cancel setup" \
        18 70

    return $?
}

start_node() {
    local api_port=$(grep -oP '"\K\d+(?=:8545")' "$SETUP_DIR/docker-compose.yml" | head -1 || echo "8545")

    log_info "Starting node..."

    # Start docker-compose
    cd "$SETUP_DIR"

    if ! docker-compose up -d 2>&1 | tee /tmp/join-wizard-start.log; then
        log_error "Failed to start node"
        dialog --title "Startup Failed" --msgbox \
            "Failed to start the node.

Check logs at: /tmp/join-wizard-start.log

You can try manually:
  cd $SETUP_DIR
  docker-compose up -d" \
            15 60
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

    dialog --title "Setup Complete" --msgbox \
        "╔═══════════════════════════════════════╗
║                                       ║
║           Success!                    ║
║                                       ║
║   Your node is now running            ║
║                                       ║
╚═══════════════════════════════════════╝

Node Status: Running
Setup Directory: $SETUP_DIR

Useful Commands:
• View logs:
  cd $SETUP_DIR && docker-compose logs -f

• Check node info:
  curl http://localhost:$api_port/api/v1/node/info

• Check chain info:
  curl http://localhost:$api_port/api/v1/chain/info

• Stop node:
  cd $SETUP_DIR && docker-compose down

Documentation:
  docs/joining-network.md" \
        28 65

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
    echo "API endpoint: http://localhost:$api_port"
    echo ""
    echo -e "${YELLOW}Next steps:${NC}"
    echo ""
    echo "  • View logs:  cd $SETUP_DIR && docker-compose logs -f"
    echo "  • Check node: curl http://localhost:$api_port/api/v1/node/info"
    echo "  • Stop node:  cd $SETUP_DIR && docker-compose down"
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
    log_info "Getting tarball location..."
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
