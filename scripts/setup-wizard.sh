#!/bin/bash
# Podoru Chain Setup Wizard
# Interactive blockchain network configuration and deployment tool

set -e

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Source helper libraries
source "$SCRIPT_DIR/lib/utils.sh"
source "$SCRIPT_DIR/lib/validators.sh"
source "$SCRIPT_DIR/lib/config-generator.sh"
source "$SCRIPT_DIR/lib/docker-compose-generator.sh"

# Change to project root
cd "$PROJECT_ROOT"

# Configuration variables (with defaults)
NUM_PRODUCERS=3
NUM_FULLNODES=1
CHAIN_NAME="Podoru Chain"
CHAIN_DESCRIPTION="Decentralized blockchain for storing any data"
BLOCK_TIME="5s"
API_PORT_START=8545
P2P_PORT_START=9000
SKIP_DOCKER_BUILD=false
AUTO_START_NETWORK=false

# Temp file for storing addresses
ADDRESSES_FILE="/tmp/wizard_addresses.$$"

# Dialog dimensions
DIALOG_HEIGHT=20
DIALOG_WIDTH=70

#==============================================================================
# UI Functions
#==============================================================================

show_welcome() {
    dialog --title "Podoru Chain Setup Wizard" --msgbox "\
Welcome to the Podoru Chain Setup Wizard!

This wizard will help you configure and deploy a multi-node blockchain network.

The wizard will:
  • Configure the number of producer and full nodes
  • Set up chain metadata (name, description, block time)
  • Generate cryptographic keys for all producers
  • Create configuration files for all nodes
  • Build Docker image (optional)
  • Start the blockchain network (optional)

Press OK to continue..." 18 70
}

show_final_summary() {
    local num_producers="$1"
    local num_fullnodes="$2"
    local api_port_start="$3"
    local chain_name="$4"
    local block_time="$5"

    local summary="╔═══════════════════════════════════════════════════════════════╗\n"
    summary+="║           Podoru Chain Setup Complete!                       ║\n"
    summary+="╚═══════════════════════════════════════════════════════════════╝\n\n"

    summary+="📊 Network Configuration:\n"
    summary+="   • Producers: ${num_producers}\n"
    summary+="   • Full Nodes: ${num_fullnodes}\n"
    summary+="   • Chain Name: ${chain_name}\n"
    summary+="   • Block Time: ${block_time}\n\n"

    summary+="🌐 Node Endpoints:\n"

    for ((i=1; i<=num_producers; i++)); do
        local api_port=$((api_port_start + i - 1))
        summary+="   Producer ${i}: http://localhost:${api_port}\n"
    done

    if [ "$num_fullnodes" -gt 0 ]; then
        for ((i=1; i<=num_fullnodes; i++)); do
            local api_port=$((api_port_start + num_producers + i - 1))
            summary+="   Full Node ${i}: http://localhost:${api_port}\n"
        done
    fi

    summary+="\n🧪 Quick Test Commands:\n"
    summary+="   curl http://localhost:${api_port_start}/api/v1/chain/info | jq\n"
    summary+="   curl http://localhost:${api_port_start}/api/v1/state/chain:name | jq\n\n"

    summary+="📋 Management Commands:\n"
    summary+="   make docker-compose-up    # Start network\n"
    summary+="   make docker-compose-logs  # View logs\n"
    summary+="   make docker-compose-down  # Stop network\n"

    dialog --title "Setup Summary" --msgbox "$summary" 30 70
}

#==============================================================================
# Configuration Collection Functions
#==============================================================================

collect_node_configuration() {
    while true; do
        exec 3>&1
        local result=$(dialog --title "Node Configuration" \
            --form "Configure the number of nodes in your blockchain network:" \
            12 70 2 \
            "Number of Producers (1-10):" 1 1 "$NUM_PRODUCERS" 1 32 10 10 \
            "Number of Full Nodes (0-5):" 2 1 "$NUM_FULLNODES" 2 32 10 10 \
            2>&1 1>&3)
        exec 3>&-

        if [ $? -ne 0 ]; then
            return 1  # User cancelled
        fi

        # Parse results
        NUM_PRODUCERS=$(echo "$result" | sed -n '1p')
        NUM_FULLNODES=$(echo "$result" | sed -n '2p')

        # Validate
        local error_msg=""
        if ! validate_producer_count "$NUM_PRODUCERS" 2>/dev/null; then
            error_msg+="Invalid producer count: must be 1-10\n"
        fi
        if ! validate_fullnode_count "$NUM_FULLNODES" 2>/dev/null; then
            error_msg+="Invalid fullnode count: must be 0-5\n"
        fi

        if [ -n "$error_msg" ]; then
            dialog --title "Validation Error" --msgbox "$error_msg" 10 60
            continue
        fi

        break
    done

    return 0
}

collect_chain_metadata() {
    while true; do
        exec 3>&1
        local result=$(dialog --title "Chain Metadata" \
            --form "Configure your blockchain metadata:" \
            14 70 3 \
            "Chain Name:" 1 1 "$CHAIN_NAME" 1 20 45 50 \
            "Description:" 2 1 "$CHAIN_DESCRIPTION" 2 20 45 200 \
            "Block Time (e.g., 5s):" 3 1 "$BLOCK_TIME" 3 25 10 10 \
            2>&1 1>&3)
        exec 3>&-

        if [ $? -ne 0 ]; then
            return 1  # User cancelled
        fi

        # Parse results
        CHAIN_NAME=$(echo "$result" | sed -n '1p')
        CHAIN_DESCRIPTION=$(echo "$result" | sed -n '2p')
        BLOCK_TIME=$(echo "$result" | sed -n '3p')

        # Validate
        local error_msg=""
        if ! validate_chain_name "$CHAIN_NAME" 2>/dev/null; then
            error_msg+="Invalid chain name\n"
        fi
        if ! validate_block_time "$BLOCK_TIME" 2>/dev/null; then
            error_msg+="Invalid block time (format: Ns, e.g., 5s)\n"
        fi

        if [ -n "$error_msg" ]; then
            dialog --title "Validation Error" --msgbox "$error_msg" 10 60
            continue
        fi

        break
    done

    return 0
}

collect_network_ports() {
    while true; do
        exec 3>&1
        local result=$(dialog --title "Network Ports" \
            --form "Configure starting port numbers for the network:" \
            12 70 2 \
            "API Port Start:" 1 1 "$API_PORT_START" 1 20 10 10 \
            "P2P Port Start:" 2 1 "$P2P_PORT_START" 2 20 10 10 \
            2>&1 1>&3)
        exec 3>&-

        if [ $? -ne 0 ]; then
            return 1  # User cancelled
        fi

        # Parse results
        API_PORT_START=$(echo "$result" | sed -n '1p')
        P2P_PORT_START=$(echo "$result" | sed -n '2p')

        # Validate
        local error_msg=""
        if ! validate_api_port "$API_PORT_START" 2>/dev/null; then
            error_msg+="Invalid API port: must be 1024-65535\n"
        fi
        if ! validate_p2p_port "$P2P_PORT_START" 2>/dev/null; then
            error_msg+="Invalid P2P port: must be 1024-65535\n"
        fi
        if ! validate_total_ports "$NUM_PRODUCERS" "$NUM_FULLNODES" "$API_PORT_START" "$P2P_PORT_START" 2>/dev/null; then
            error_msg+="Port ranges overlap or exceed limits\n"
        fi

        if [ -n "$error_msg" ]; then
            dialog --title "Validation Error" --msgbox "$error_msg" 10 60
            continue
        fi

        break
    done

    return 0
}

collect_advanced_options() {
    exec 3>&1
    local result=$(dialog --title "Advanced Options" \
        --checklist "Select advanced options:" \
        12 70 2 \
        1 "Skip Docker build (use existing image)" off \
        2 "Auto-start network after setup" off \
        2>&1 1>&3)
    exec 3>&-

    if [ $? -ne 0 ]; then
        return 1  # User cancelled
    fi

    # Parse results
    SKIP_DOCKER_BUILD=false
    AUTO_START_NETWORK=false

    echo "$result" | grep -q "1" && SKIP_DOCKER_BUILD=true
    echo "$result" | grep -q "2" && AUTO_START_NETWORK=true

    return 0
}

review_configuration() {
    local total_nodes=$((NUM_PRODUCERS + NUM_FULLNODES))
    local max_api_port=$((API_PORT_START + total_nodes - 1))
    local max_p2p_port=$((P2P_PORT_START + total_nodes - 1))

    local review="Please review your configuration:\n\n"
    review+="Nodes:\n"
    review+="  • Producers: ${NUM_PRODUCERS}\n"
    review+="  • Full Nodes: ${NUM_FULLNODES}\n"
    review+="  • Total: ${total_nodes}\n\n"

    review+="Chain:\n"
    review+="  • Name: ${CHAIN_NAME}\n"
    review+="  • Description: ${CHAIN_DESCRIPTION}\n"
    review+="  • Block Time: ${BLOCK_TIME}\n\n"

    review+="Ports:\n"
    review+="  • API: ${API_PORT_START}-${max_api_port}\n"
    review+="  • P2P: ${P2P_PORT_START}-${max_p2p_port}\n\n"

    review+="Options:\n"
    review+="  • Skip Docker build: $SKIP_DOCKER_BUILD\n"
    review+="  • Auto-start network: $AUTO_START_NETWORK\n\n"

    review+="Proceed with setup?"

    dialog --title "Review Configuration" --yesno "$review" 24 70
    return $?
}

#==============================================================================
# Setup Execution Functions
#==============================================================================

check_and_handle_existing_setup() {
    if check_existing_setup; then
        dialog --title "Existing Setup Detected" --yesno \
            "Found existing blockchain data directories.\n\nDo you want to clean them and start fresh?\n\nYes: Remove existing data and reconfigure\nNo: Exit wizard" \
            12 70

        case $? in
            0)  # Yes - clean
                log_info "Cleaning existing setup..."
                rm -rf docker/data/producer* docker/data/fullnode* 2>/dev/null
                rm -f docker/docker-compose.yml 2>/dev/null
                log_success "Existing setup cleaned"
                return 0
                ;;
            1)  # No - exit
                log_info "Setup cancelled by user"
                return 1
                ;;
            255)  # ESC
                log_info "Setup cancelled by user"
                return 1
                ;;
        esac
    fi

    return 0
}

generate_keys_with_progress() {
    log_info "Generating cryptographic keys for ${NUM_PRODUCERS} producers..."

    # Clear addresses file
    > "$ADDRESSES_FILE"

    (
        for ((i=1; i<=NUM_PRODUCERS; i++)); do
            local percent=$((i * 100 / NUM_PRODUCERS))

            echo "XXX"
            echo "$percent"
            echo "Generating keys for producer${i}..."
            echo "This may take a few seconds."
            echo "XXX"

            # Create keys directory
            ensure_directory "docker/data/producer${i}/keys"

            # Generate key
            local key_output=$(go run ./cmd/tools/keygen/main.go \
                -output "docker/data/producer${i}/keys/producer${i}.key" 2>&1)

            # Extract address
            local address=$(extract_address "$key_output")

            if [ -z "$address" ]; then
                log_error "Failed to extract address for producer${i}"
                echo "100"
                exit 1
            fi

            # Save address
            echo "$address" >> "$ADDRESSES_FILE"

            sleep 0.5  # Small delay for visual feedback
        done

        echo "XXX"
        echo "100"
        echo "All keys generated successfully!"
        echo "XXX"
        sleep 1

    ) | dialog --title "Generating Cryptographic Keys" --gauge "Initializing..." 10 70 0

    if [ ${PIPESTATUS[0]} -ne 0 ]; then
        log_error "Key generation failed"
        return 1
    fi

    log_success "Generated keys for all producers"
    return 0
}

build_docker_image_with_progress() {
    if [ "$SKIP_DOCKER_BUILD" = true ]; then
        log_info "Skipping Docker build (using existing image)"
        return 0
    fi

    log_info "Building Docker image..."

    (
        echo "10" ; sleep 1
        echo "XXX" ; echo "30" ; echo "Building Docker image..." ; echo "This may take several minutes." ; echo "XXX"

        # Build image (redirect output to log file)
        if docker build -t podoru-chain:latest -f docker/Dockerfile . > /tmp/wizard_docker_build.$$ 2>&1; then
            echo "XXX" ; echo "100" ; echo "Docker image built successfully!" ; echo "XXX" ; sleep 1
            exit 0
        else
            echo "XXX" ; echo "100" ; echo "Docker build failed! Check logs." ; echo "XXX" ; sleep 2
            exit 1
        fi

    ) | dialog --title "Building Docker Image" --gauge "Preparing build context..." 10 70 0

    local build_result=${PIPESTATUS[0]}

    if [ $build_result -ne 0 ]; then
        log_error "Docker build failed"
        dialog --title "Build Error" --msgbox "Docker build failed. Check /tmp/wizard_docker_build.$$ for details." 10 60
        return 1
    fi

    log_success "Docker image built successfully"
    rm -f /tmp/wizard_docker_build.$$
    return 0
}

start_network_if_requested() {
    if [ "$AUTO_START_NETWORK" = true ]; then
        log_info "Starting blockchain network..."

        (
            echo "50" ; sleep 1
            echo "XXX" ; echo "75" ; echo "Starting Docker containers..." ; echo "XXX"

            cd docker
            if run_docker_compose up -d > /tmp/wizard_docker_up.$$ 2>&1; then
                echo "XXX" ; echo "100" ; echo "Network started successfully!" ; echo "XXX" ; sleep 1
                exit 0
            else
                echo "XXX" ; echo "100" ; echo "Failed to start network!" ; echo "XXX" ; sleep 2
                exit 1
            fi

        ) | dialog --title "Starting Network" --gauge "Preparing containers..." 10 70 0

        local start_result=${PIPESTATUS[0]}

        if [ $start_result -ne 0 ]; then
            log_error "Failed to start network"
            dialog --title "Start Error" --msgbox "Failed to start network. Check /tmp/wizard_docker_up.$$ for details." 10 60
            return 1
        fi

        log_success "Network started successfully"
        rm -f /tmp/wizard_docker_up.$$

        # Give nodes a moment to initialize
        dialog --infobox "Waiting for nodes to initialize..." 5 50
        sleep 5

        return 0
    fi

    return 0
}

#==============================================================================
# Main Setup Flow
#==============================================================================

main() {
    # Set up error handling
    trap cleanup_on_error EXIT

    # Check prerequisites
    if ! check_prerequisites; then
        echo "Prerequisites check failed. Exiting."
        exit 1
    fi

    # Show welcome screen
    show_welcome

    # Check for existing setup
    if ! check_and_handle_existing_setup; then
        clear
        exit 0
    fi

    # Collect configuration
    if ! collect_node_configuration; then
        clear
        log_info "Setup cancelled"
        exit 0
    fi

    if ! collect_chain_metadata; then
        clear
        log_info "Setup cancelled"
        exit 0
    fi

    if ! collect_network_ports; then
        clear
        log_info "Setup cancelled"
        exit 0
    fi

    if ! collect_advanced_options; then
        clear
        log_info "Setup cancelled"
        exit 0
    fi

    # Review configuration
    if ! review_configuration; then
        clear
        log_info "Setup cancelled"
        exit 0
    fi

    # Execute setup
    clear
    log_info "Starting Podoru Chain setup..."
    echo ""

    # Create directory structure
    log_info "Creating directory structure..."
    ensure_directory "docker/data"

    # Generate keys
    if ! generate_keys_with_progress; then
        log_error "Setup failed during key generation"
        exit 1
    fi

    # Read generated addresses
    local addresses=$(cat "$ADDRESSES_FILE")

    # Generate configuration files
    dialog --infobox "Generating configuration files..." 5 50
    if ! generate_all_configs "$addresses" "$NUM_PRODUCERS" "$NUM_FULLNODES" \
        "$CHAIN_NAME" "$CHAIN_DESCRIPTION" "$BLOCK_TIME"; then
        log_error "Failed to generate configuration files"
        exit 1
    fi

    # Generate docker-compose.yml
    dialog --infobox "Generating docker-compose.yml..." 5 50
    if ! generate_docker_compose "$NUM_PRODUCERS" "$NUM_FULLNODES" \
        "$API_PORT_START" "$P2P_PORT_START"; then
        log_error "Failed to generate docker-compose.yml"
        exit 1
    fi

    # Build Docker image
    if ! build_docker_image_with_progress; then
        log_error "Setup failed during Docker build"
        exit 1
    fi

    # Start network if requested
    if ! start_network_if_requested; then
        log_warning "Setup completed but network failed to start"
    fi

    # Show final summary
    show_final_summary "$NUM_PRODUCERS" "$NUM_FULLNODES" "$API_PORT_START" \
        "$CHAIN_NAME" "$BLOCK_TIME"

    # Cleanup
    rm -f "$ADDRESSES_FILE"

    clear
    log_success "Setup wizard completed successfully!"
    echo ""

    return 0
}

# Run main function
main "$@"
