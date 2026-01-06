#!/bin/bash
# Validation functions for Podoru Chain Setup Wizard

# Source utils for logging (assuming it's already loaded or in same directory)
LIB_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
[ -f "$LIB_DIR/utils.sh" ] && source "$LIB_DIR/utils.sh" || true

# Validate node count
validate_node_count() {
    local count="$1"
    local min="$2"
    local max="$3"
    local type="$4"

    # Check if it's a number
    if ! [[ "$count" =~ ^[0-9]+$ ]]; then
        echo "ERROR: $type count must be a number"
        return 1
    fi

    # Check range
    if [ "$count" -lt "$min" ] || [ "$count" -gt "$max" ]; then
        echo "ERROR: $type count must be between $min and $max"
        return 1
    fi

    return 0
}

# Validate producer count (1-10)
validate_producer_count() {
    local count="$1"
    validate_node_count "$count" 1 10 "Producer"
    return $?
}

# Validate fullnode count (0-5)
validate_fullnode_count() {
    local count="$1"
    validate_node_count "$count" 0 5 "Full node"
    return $?
}

# Validate port number
validate_port() {
    local port="$1"
    local port_name="${2:-Port}"

    # Check if it's a number
    if ! [[ "$port" =~ ^[0-9]+$ ]]; then
        echo "ERROR: $port_name must be a number"
        return 1
    fi

    # Check range (avoid privileged ports < 1024)
    if [ "$port" -lt 1024 ] || [ "$port" -gt 65535 ]; then
        echo "ERROR: $port_name must be between 1024 and 65535"
        return 1
    fi

    return 0
}

# Validate API port
validate_api_port() {
    local port="$1"
    validate_port "$port" "API port"
    return $?
}

# Validate P2P port
validate_p2p_port() {
    local port="$1"
    validate_port "$port" "P2P port"
    return $?
}

# Check if port range is available
validate_port_range() {
    local start_port="$1"
    local count="$2"
    local port_type="${3:-Port}"

    local conflicts=()

    for ((i=0; i<count; i++)); do
        local port=$((start_port + i))
        if is_port_in_use "$port"; then
            conflicts+=("$port")
        fi
    done

    if [ ${#conflicts[@]} -gt 0 ]; then
        echo "WARNING: The following $port_type ports are in use: ${conflicts[*]}"
        echo "This may cause conflicts when starting the network."
        return 1
    fi

    return 0
}

# Validate block time format
validate_block_time() {
    local time="$1"

    # Check format (Ns where N is a number)
    if ! [[ "$time" =~ ^[0-9]+s$ ]]; then
        echo "ERROR: Block time must be in format: Ns (e.g., 5s, 10s)"
        return 1
    fi

    # Extract number
    local seconds="${time%s}"

    # Check reasonable range (1-60 seconds)
    if [ "$seconds" -lt 1 ] || [ "$seconds" -gt 60 ]; then
        echo "ERROR: Block time must be between 1s and 60s"
        return 1
    fi

    return 0
}

# Validate chain name
validate_chain_name() {
    local name="$1"

    # Check if empty
    if [ -z "$name" ]; then
        echo "ERROR: Chain name cannot be empty"
        return 1
    fi

    # Check length (1-50 characters)
    if [ ${#name} -gt 50 ]; then
        echo "ERROR: Chain name must be 50 characters or less"
        return 1
    fi

    # Check for valid characters (alphanumeric, spaces, hyphens, underscores)
    if ! [[ "$name" =~ ^[a-zA-Z0-9\ _-]+$ ]]; then
        echo "ERROR: Chain name can only contain letters, numbers, spaces, hyphens, and underscores"
        return 1
    fi

    return 0
}

# Validate chain description
validate_chain_description() {
    local description="$1"

    # Description can be empty, but if provided, check length
    if [ -n "$description" ] && [ ${#description} -gt 200 ]; then
        echo "ERROR: Chain description must be 200 characters or less"
        return 1
    fi

    return 0
}

# Check for existing setup
check_existing_setup() {
    local data_dir="docker/data"

    if [ ! -d "$data_dir" ]; then
        return 1  # No existing setup
    fi

    # Check if there are any producer or fullnode directories
    local existing_dirs=$(find "$data_dir" -mindepth 1 -maxdepth 1 -type d \
        \( -name "producer*" -o -name "fullnode*" \) 2>/dev/null)

    if [ -z "$existing_dirs" ]; then
        return 1  # No existing setup
    fi

    # Return success (existing setup found)
    return 0
}

# Validate all basic configuration inputs
validate_basic_config() {
    local num_producers="$1"
    local num_fullnodes="$2"
    local chain_name="$3"
    local block_time="$4"
    local api_port="$5"
    local p2p_port="$6"

    local errors=()

    # Validate producer count
    if ! validate_producer_count "$num_producers" 2>/dev/null; then
        errors+=("Invalid producer count: $num_producers")
    fi

    # Validate fullnode count
    if ! validate_fullnode_count "$num_fullnodes" 2>/dev/null; then
        errors+=("Invalid fullnode count: $num_fullnodes")
    fi

    # Validate chain name
    if ! validate_chain_name "$chain_name" 2>/dev/null; then
        errors+=("Invalid chain name")
    fi

    # Validate block time
    if ! validate_block_time "$block_time" 2>/dev/null; then
        errors+=("Invalid block time: $block_time")
    fi

    # Validate API port
    if ! validate_api_port "$api_port" 2>/dev/null; then
        errors+=("Invalid API port: $api_port")
    fi

    # Validate P2P port
    if ! validate_p2p_port "$p2p_port" 2>/dev/null; then
        errors+=("Invalid P2P port: $p2p_port")
    fi

    # Check for port conflicts
    if [ "$api_port" -eq "$p2p_port" ]; then
        errors+=("API and P2P ports cannot be the same")
    fi

    # Report errors
    if [ ${#errors[@]} -gt 0 ]; then
        for error in "${errors[@]}"; do
            log_error "$error"
        done
        return 1
    fi

    return 0
}

# Validate total port usage
validate_total_ports() {
    local num_producers="$1"
    local num_fullnodes="$2"
    local api_port_start="$3"
    local p2p_port_start="$4"

    local total_nodes=$((num_producers + num_fullnodes))

    # Check if API port range exceeds 65535
    local max_api_port=$((api_port_start + total_nodes - 1))
    if [ "$max_api_port" -gt 65535 ]; then
        echo "ERROR: API port range exceeds 65535 (would need ports up to $max_api_port)"
        return 1
    fi

    # Check if P2P port range exceeds 65535
    local max_p2p_port=$((p2p_port_start + total_nodes - 1))
    if [ "$max_p2p_port" -gt 65535 ]; then
        echo "ERROR: P2P port range exceeds 65535 (would need ports up to $max_p2p_port)"
        return 1
    fi

    # Check for overlap between API and P2P port ranges
    if [ "$api_port_start" -le "$max_p2p_port" ] && [ "$p2p_port_start" -le "$max_api_port" ]; then
        # Ranges overlap, check if they actually conflict
        if [ "$api_port_start" -lt "$p2p_port_start" ]; then
            if [ "$max_api_port" -ge "$p2p_port_start" ]; then
                echo "ERROR: API and P2P port ranges overlap"
                return 1
            fi
        else
            if [ "$max_p2p_port" -ge "$api_port_start" ]; then
                echo "ERROR: API and P2P port ranges overlap"
                return 1
            fi
        fi
    fi

    return 0
}

# Export validation functions
export -f validate_node_count
export -f validate_producer_count
export -f validate_fullnode_count
export -f validate_port
export -f validate_api_port
export -f validate_p2p_port
export -f validate_port_range
export -f validate_block_time
export -f validate_chain_name
export -f validate_chain_description
export -f check_existing_setup
export -f validate_basic_config
export -f validate_total_ports
