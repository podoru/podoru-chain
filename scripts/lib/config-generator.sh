#!/bin/bash
# Configuration file generator for Podoru Chain Setup Wizard

# Source utils for logging
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/utils.sh"

# Generate bootstrap peers for a producer node
# Args: node_num, total_producers
generate_producer_bootstrap_peers() {
    local node_num="$1"
    local total_producers="$2"
    local indent="  "

    for ((i=1; i<=total_producers; i++)); do
        if [ $i -ne $node_num ]; then
            local ip_suffix=$((9 + i))
            echo "${indent}- \"172.20.0.${ip_suffix}:9000\""
        fi
    done
}

# Generate bootstrap peers for a fullnode
# Args: total_producers
generate_fullnode_bootstrap_peers() {
    local total_producers="$1"
    local indent="  "

    for ((i=1; i<=total_producers; i++)); do
        local ip_suffix=$((9 + i))
        echo "${indent}- \"172.20.0.${ip_suffix}:9000\""
    done
}

# Generate authorities list for YAML format
# Args: addresses_array (passed as string with newlines)
generate_authorities_yaml() {
    local addresses="$1"
    local indent="  "

    echo "$addresses" | while IFS= read -r addr; do
        if [ -n "$addr" ]; then
            echo "${indent}- \"${addr}\""
        fi
    done
}

# Generate authorities list for JSON format
# Args: addresses_array (passed as string with newlines)
generate_authorities_json() {
    local addresses="$1"
    local result="["
    local first=true

    echo "$addresses" | while IFS= read -r addr; do
        if [ -n "$addr" ]; then
            if [ "$first" = true ]; then
                result="${result}\n    \"${addr}\""
                first=false
            else
                result="${result},\n    \"${addr}\""
            fi
        fi
    done

    result="${result}\n  ]"
    echo -e "$result"
}

# Generate producer configuration file
# Args: node_num, address, bootstrap_peers, authorities, block_time
generate_producer_config() {
    local node_num="$1"
    local address="$2"
    local bootstrap_peers="$3"
    local authorities="$4"
    local block_time="$5"

    local config_file="docker/data/producer${node_num}/config.yaml"

    # Create directory
    ensure_directory "docker/data/producer${node_num}"

    # Generate config using heredoc
    cat > "$config_file" <<EOF
node_type: producer
address: "${address}"
private_key: "/data/keys/producer${node_num}.key"

# Network configuration
p2p_port: 9000
p2p_bind_addr: "0.0.0.0"
bootstrap_peers:
${bootstrap_peers}
max_peers: 50

# API configuration
api_enabled: true
api_port: 8545
api_bind_addr: "0.0.0.0"

# Storage configuration
data_dir: "/data"

# Consensus configuration
authorities:
${authorities}
block_time: ${block_time}

# Genesis configuration
genesis_path: "/data/genesis.json"
EOF

    if [ $? -eq 0 ]; then
        log_success "Generated config for producer${node_num}"
        return 0
    else
        log_error "Failed to generate config for producer${node_num}"
        return 1
    fi
}

# Generate fullnode configuration file
# Args: node_num, bootstrap_peers, authorities, block_time
generate_fullnode_config() {
    local node_num="$1"
    local bootstrap_peers="$2"
    local authorities="$3"
    local block_time="$4"

    local config_file="docker/data/fullnode${node_num}/config.yaml"

    # Create directory
    ensure_directory "docker/data/fullnode${node_num}"

    # Generate config using heredoc
    cat > "$config_file" <<EOF
node_type: full

# Network configuration
p2p_port: 9000
p2p_bind_addr: "0.0.0.0"
bootstrap_peers:
${bootstrap_peers}
max_peers: 50

# API configuration
api_enabled: true
api_port: 8545
api_bind_addr: "0.0.0.0"

# Storage configuration
data_dir: "/data"

# Consensus configuration
authorities:
${authorities}
block_time: ${block_time}

# Genesis configuration
genesis_path: "/data/genesis.json"
EOF

    if [ $? -eq 0 ]; then
        log_success "Generated config for fullnode${node_num}"
        return 0
    else
        log_error "Failed to generate config for fullnode${node_num}"
        return 1
    fi
}

# Generate genesis.json file
# Args: authorities, chain_name, chain_description, chain_version, timestamp
generate_genesis_json() {
    local authorities="$1"
    local chain_name="$2"
    local chain_description="$3"
    local chain_version="${4:-1.0.0}"
    local timestamp="${5:-$(get_genesis_timestamp)}"

    local genesis_file="docker/data/genesis.json"
    local authorities_json=$(generate_authorities_json "$authorities")

    # Sort initial_state keys alphabetically for deterministic genesis hash
    # Order: chain:description, chain:name, chain:version, system:initialized

    cat > "$genesis_file" <<EOF
{
  "timestamp": ${timestamp},
  "authorities": ${authorities_json},
  "initial_state": {
    "chain:description": "${chain_description}",
    "chain:name": "${chain_name}",
    "chain:version": "${chain_version}",
    "system:initialized": "true"
  }
}
EOF

    if [ $? -eq 0 ]; then
        log_success "Generated genesis.json"

        # Copy to all node directories
        for dir in docker/data/producer* docker/data/fullnode*; do
            if [ -d "$dir" ]; then
                cp "$genesis_file" "$dir/genesis.json"
                log_info "Copied genesis.json to $dir"
            fi
        done

        return 0
    else
        log_error "Failed to generate genesis.json"
        return 1
    fi
}

# Generate all configuration files
# Args: addresses_array, num_producers, num_fullnodes, chain_name, chain_description, block_time
generate_all_configs() {
    local addresses="$1"
    local num_producers="$2"
    local num_fullnodes="$3"
    local chain_name="$4"
    local chain_description="$5"
    local block_time="$6"

    log_info "Generating configuration files..."

    # Convert addresses string to array
    local -a addr_array
    while IFS= read -r line; do
        if [ -n "$line" ]; then
            addr_array+=("$line")
        fi
    done <<< "$addresses"

    # Generate authorities list (same for all nodes)
    local authorities=$(generate_authorities_yaml "$addresses")

    # Generate producer configs
    for ((i=1; i<=num_producers; i++)); do
        local address="${addr_array[$((i-1))]}"
        local bootstrap_peers=$(generate_producer_bootstrap_peers "$i" "$num_producers")

        if ! generate_producer_config "$i" "$address" "$bootstrap_peers" "$authorities" "$block_time"; then
            return 1
        fi
    done

    # Generate fullnode configs
    if [ "$num_fullnodes" -gt 0 ]; then
        local fullnode_bootstrap_peers=$(generate_fullnode_bootstrap_peers "$num_producers")

        for ((i=1; i<=num_fullnodes; i++)); do
            if ! generate_fullnode_config "$i" "$fullnode_bootstrap_peers" "$authorities" "$block_time"; then
                return 1
            fi
        done
    fi

    # Generate genesis.json
    if ! generate_genesis_json "$addresses" "$chain_name" "$chain_description" "1.0.0" ""; then
        return 1
    fi

    log_success "All configuration files generated successfully"
    return 0
}

# Export functions
export -f generate_producer_bootstrap_peers
export -f generate_fullnode_bootstrap_peers
export -f generate_authorities_yaml
export -f generate_authorities_json
export -f generate_producer_config
export -f generate_fullnode_config
export -f generate_genesis_json
export -f generate_all_configs
