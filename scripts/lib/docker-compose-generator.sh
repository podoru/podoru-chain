#!/bin/bash
# Docker Compose generator for Podoru Chain Setup Wizard

# Source utils for logging
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/utils.sh"

# Generate producer service definition
# Args: node_num, ip_suffix, api_port, p2p_port, total_producers
generate_producer_service() {
    local node_num="$1"
    local ip_suffix="$2"
    local api_port="$3"
    local p2p_port="$4"
    local total_producers="$5"

    cat <<EOF
  producer${node_num}:
    image: podoru-chain:latest
    container_name: podoru-producer${node_num}
    hostname: producer${node_num}
    ports:
      - "${api_port}:8545"
      - "${p2p_port}:9000"
    volumes:
      - ./data/producer${node_num}:/data
    environment:
      - NODE_TYPE=producer
    networks:
      podoru:
        ipv4_address: 172.20.0.${ip_suffix}
    restart: unless-stopped
EOF

    # Add dependencies (all producers except the first depend on producer1)
    if [ $node_num -gt 1 ]; then
        cat <<EOF
    depends_on:
      - producer1
EOF
    fi

    echo ""
}

# Generate fullnode service definition
# Args: node_num, ip_suffix, api_port, p2p_port
generate_fullnode_service() {
    local node_num="$1"
    local ip_suffix="$2"
    local api_port="$3"
    local p2p_port="$4"

    cat <<EOF
  fullnode${node_num}:
    image: podoru-chain:latest
    container_name: podoru-fullnode${node_num}
    hostname: fullnode${node_num}
    ports:
      - "${api_port}:8545"
      - "${p2p_port}:9000"
    volumes:
      - ./data/fullnode${node_num}:/data
    environment:
      - NODE_TYPE=full
    networks:
      podoru:
        ipv4_address: 172.20.0.${ip_suffix}
    restart: unless-stopped
    depends_on:
      - producer1
EOF

    echo ""
}

# Generate network configuration
generate_network_config() {
    cat <<EOF
networks:
  podoru:
    driver: bridge
    ipam:
      config:
        - subnet: 172.20.0.0/16
EOF
}

# Generate complete docker-compose.yml
# Args: num_producers, num_fullnodes, api_port_start, p2p_port_start
generate_docker_compose() {
    local num_producers="$1"
    local num_fullnodes="$2"
    local api_port_start="$3"
    local p2p_port_start="$4"

    local compose_file="docker/docker-compose.yml"

    log_info "Generating docker-compose.yml..."

    # Start with header
    cat > "$compose_file" <<'EOF'
version: '3.8'

services:
EOF

    # Generate producer services
    for ((i=1; i<=num_producers; i++)); do
        local ip_suffix=$((9 + i))  # 172.20.0.10, .11, .12, ...
        local api_port=$((api_port_start + i - 1))
        local p2p_port=$((p2p_port_start + i - 1))

        generate_producer_service "$i" "$ip_suffix" "$api_port" "$p2p_port" "$num_producers" >> "$compose_file"
    done

    # Generate fullnode services
    for ((i=1; i<=num_fullnodes; i++)); do
        local ip_suffix=$((99 + i))  # 172.20.0.100, .101, .102, ...
        local api_port=$((api_port_start + num_producers + i - 1))
        local p2p_port=$((p2p_port_start + num_producers + i - 1))

        generate_fullnode_service "$i" "$ip_suffix" "$api_port" "$p2p_port" >> "$compose_file"
    done

    # Add network configuration
    generate_network_config >> "$compose_file"

    if [ $? -eq 0 ]; then
        log_success "Generated docker-compose.yml with $num_producers producers and $num_fullnodes fullnodes"
        return 0
    else
        log_error "Failed to generate docker-compose.yml"
        return 1
    fi
}

# Display docker compose configuration summary
# Args: num_producers, num_fullnodes, api_port_start, p2p_port_start
display_docker_compose_summary() {
    local num_producers="$1"
    local num_fullnodes="$2"
    local api_port_start="$3"
    local p2p_port_start="$4"

    echo ""
    echo "Docker Compose Configuration:"
    echo "=============================="
    echo ""
    echo "Producer Nodes:"

    for ((i=1; i<=num_producers; i++)); do
        local ip_suffix=$((9 + i))
        local api_port=$((api_port_start + i - 1))
        local p2p_port=$((p2p_port_start + i - 1))

        echo "  producer${i}:"
        echo "    Container: podoru-producer${i}"
        echo "    IP: 172.20.0.${ip_suffix}"
        echo "    API Port: ${api_port}:8545"
        echo "    P2P Port: ${p2p_port}:9000"
        echo ""
    done

    if [ "$num_fullnodes" -gt 0 ]; then
        echo "Full Nodes:"

        for ((i=1; i<=num_fullnodes; i++)); do
            local ip_suffix=$((99 + i))
            local api_port=$((api_port_start + num_producers + i - 1))
            local p2p_port=$((p2p_port_start + num_producers + i - 1))

            echo "  fullnode${i}:"
            echo "    Container: podoru-fullnode${i}"
            echo "    IP: 172.20.0.${ip_suffix}"
            echo "    API Port: ${api_port}:8545"
            echo "    P2P Port: ${p2p_port}:9000"
            echo ""
        done
    fi

    echo "Network: podoru (172.20.0.0/16)"
    echo ""
}

# Export functions
export -f generate_producer_service
export -f generate_fullnode_service
export -f generate_network_config
export -f generate_docker_compose
export -f display_docker_compose_summary
