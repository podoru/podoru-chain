#!/bin/bash
# Generate join information for new nodes to connect to your Podoru Chain network

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$PROJECT_ROOT"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}╔═══════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║                                                       ║${NC}"
echo -e "${BLUE}║     Podoru Chain Network Join Information            ║${NC}"
echo -e "${BLUE}║                                                       ║${NC}"
echo -e "${BLUE}╚═══════════════════════════════════════════════════════╝${NC}"
echo ""

# Check if network is set up
if [ ! -f "docker/data/genesis.json" ]; then
    echo -e "${RED}Error: No genesis.json found. Run 'make setup-wizard' first.${NC}"
    exit 1
fi

# Get network information
GENESIS_FILE="docker/data/genesis.json"
CHAIN_NAME=$(jq -r '.initial_state["chain:name"]' "$GENESIS_FILE" 2>/dev/null || echo "Unknown")
CHAIN_DESC=$(jq -r '.initial_state["chain:description"]' "$GENESIS_FILE" 2>/dev/null || echo "Unknown")
BLOCK_TIME=$(grep "block_time:" docker/data/producer1/config.yaml 2>/dev/null | awk '{print $2}' || echo "5s")
AUTHORITIES=$(jq -r '.authorities[]' "$GENESIS_FILE" 2>/dev/null)
GENESIS_HASH=$(sha256sum "$GENESIS_FILE" | awk '{print $1}')

# Get IP addresses
echo -e "${YELLOW}Detecting network addresses...${NC}"
echo ""

# Public IP
PUBLIC_IP=$(curl -s -4 ifconfig.me 2>/dev/null || echo "Unable to detect")

# Local IP (first non-loopback)
LOCAL_IP=$(ip addr show | grep "inet " | grep -v 127.0.0.1 | head -1 | awk '{print $2}' | cut -d'/' -f1 || echo "Unable to detect")

# Docker network IP (if running)
DOCKER_IP=$(docker inspect podoru-producer1 2>/dev/null | jq -r '.[0].NetworkSettings.Networks[].IPAddress' || echo "Not running")

# Get port configuration
API_PORT=$(grep -r "\"8545:" docker/docker-compose.yml 2>/dev/null | head -1 | grep -oP '\d+(?=:8545)' || echo "8545")
P2P_PORT=$(grep -r "\"9000:" docker/docker-compose.yml 2>/dev/null | head -1 | grep -oP '\d+(?=:9000)' || echo "9000")

# Display information
echo -e "${GREEN}═══════════════════════════════════════════════════════${NC}"
echo -e "${GREEN}Network Information${NC}"
echo -e "${GREEN}═══════════════════════════════════════════════════════${NC}"
echo ""
echo "Network Name: $CHAIN_NAME"
echo "Description:  $CHAIN_DESC"
echo "Block Time:   $BLOCK_TIME"
echo ""

echo -e "${GREEN}Network Addresses:${NC}"
echo ""
echo "Public IP:    $PUBLIC_IP"
echo "Local IP:     $LOCAL_IP"
echo "Docker IP:    $DOCKER_IP"
echo ""

echo -e "${GREEN}Ports:${NC}"
echo ""
echo "API Port:     $API_PORT"
echo "P2P Port:     $P2P_PORT"
echo ""

echo -e "${GREEN}Authorities:${NC}"
echo ""
echo "$AUTHORITIES" | while read -r addr; do
    echo "  - $addr"
done
echo ""

echo -e "${GREEN}Genesis File:${NC}"
echo ""
echo "SHA256: $GENESIS_HASH"
echo "Location: $GENESIS_FILE"
echo ""

echo -e "${GREEN}═══════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}Join Instructions${NC}"
echo -e "${GREEN}═══════════════════════════════════════════════════════${NC}"
echo ""

# Determine best scenario
echo -e "${YELLOW}Choose connection scenario:${NC}"
echo ""
echo "1. Same Machine    - Node running on this computer"
echo "2. Local Network   - Node running on same LAN"
echo "3. Internet        - Node running over internet"
echo ""
read -p "Select scenario (1-3): " SCENARIO

echo ""
echo -e "${GREEN}═══════════════════════════════════════════════════════${NC}"

case $SCENARIO in
    1)
        echo -e "${BLUE}Scenario 1: Same Machine${NC}"
        echo -e "${GREEN}═══════════════════════════════════════════════════════${NC}"
        echo ""
        echo "Bootstrap Peers:"
        echo "  - \"127.0.0.1:$P2P_PORT\""
        echo ""
        echo "Note: Use different ports for API and P2P"
        echo "  Suggested API Port: 8548"
        echo "  Suggested P2P Port: 9003"
        ;;
    2)
        echo -e "${BLUE}Scenario 2: Local Network (LAN)${NC}"
        echo -e "${GREEN}═══════════════════════════════════════════════════════${NC}"
        echo ""
        echo "Bootstrap Peers:"
        echo "  - \"$LOCAL_IP:$P2P_PORT\""
        echo ""
        echo "Required: Open firewall port $P2P_PORT on host"
        echo "  Ubuntu/Debian: sudo ufw allow $P2P_PORT/tcp"
        echo "  Fedora/RHEL:   sudo firewall-cmd --add-port=$P2P_PORT/tcp --permanent"
        ;;
    3)
        echo -e "${BLUE}Scenario 3: Internet${NC}"
        echo -e "${GREEN}═══════════════════════════════════════════════════════${NC}"
        echo ""
        echo "Bootstrap Peers:"
        if [ "$PUBLIC_IP" != "Unable to detect" ]; then
            echo "  - \"$PUBLIC_IP:$P2P_PORT\""
        else
            echo "  - \"YOUR_PUBLIC_IP:$P2P_PORT\""
        fi
        echo ""
        echo "Required Setup:"
        echo "  1. Configure port forwarding on router"
        echo "     External $P2P_PORT → Internal $P2P_PORT"
        echo "  2. Open firewall port $P2P_PORT"
        echo "  3. Consider using dynamic DNS if IP changes"
        ;;
    *)
        echo -e "${RED}Invalid selection${NC}"
        exit 1
        ;;
esac

echo ""
echo -e "${GREEN}═══════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}Configuration Files${NC}"
echo -e "${GREEN}═══════════════════════════════════════════════════════${NC}"
echo ""

# Generate sample config.yaml
OUTPUT_DIR="network-join-info"
mkdir -p "$OUTPUT_DIR"

# Copy genesis
cp "$GENESIS_FILE" "$OUTPUT_DIR/genesis.json"

# Determine bootstrap peer
case $SCENARIO in
    1) BOOTSTRAP_PEER="127.0.0.1:$P2P_PORT" ;;
    2) BOOTSTRAP_PEER="$LOCAL_IP:$P2P_PORT" ;;
    3) BOOTSTRAP_PEER="${PUBLIC_IP:-YOUR_PUBLIC_IP}:$P2P_PORT" ;;
esac

# Generate config.yaml
cat > "$OUTPUT_DIR/config.yaml" <<EOF
node_type: full

# Network configuration
p2p_port: 9000
p2p_bind_addr: "0.0.0.0"
bootstrap_peers:
  - "$BOOTSTRAP_PEER"
max_peers: 50

# API configuration
api_enabled: true
api_port: 8545
api_bind_addr: "0.0.0.0"

# Storage configuration
data_dir: "/data"

# Consensus configuration
authorities:
EOF

echo "$AUTHORITIES" | while read -r addr; do
    echo "  - \"$addr\"" >> "$OUTPUT_DIR/config.yaml"
done

cat >> "$OUTPUT_DIR/config.yaml" <<EOF
block_time: $BLOCK_TIME

# Genesis configuration
genesis_path: "/data/genesis.json"
EOF

# Generate docker-compose.yml
cat > "$OUTPUT_DIR/docker-compose.yml" <<EOF
version: '3.8'

services:
  fullnode:
    image: podoru-chain:latest
    container_name: podoru-fullnode
    hostname: fullnode
    ports:
      - "8545:8545"  # API port
      - "9000:9000"  # P2P port
    volumes:
      - ./:/data
    environment:
      - NODE_TYPE=full
    restart: unless-stopped
EOF

# Generate README
cat > "$OUTPUT_DIR/README.md" <<EOF
# Join $CHAIN_NAME Network

## Quick Start

1. Copy these files to your node directory:
   - \`config.yaml\`
   - \`genesis.json\`
   - \`docker-compose.yml\`

2. Start the node:
   \`\`\`bash
   docker-compose up -d
   \`\`\`

3. Check logs:
   \`\`\`bash
   docker-compose logs -f
   \`\`\`

4. Verify connection:
   \`\`\`bash
   curl http://localhost:8545/api/v1/node/info
   curl http://localhost:8545/api/v1/chain/info
   \`\`\`

## Network Details

- **Network Name**: $CHAIN_NAME
- **Description**: $CHAIN_DESC
- **Block Time**: $BLOCK_TIME
- **Bootstrap Peer**: $BOOTSTRAP_PEER

## Genesis Verification

SHA256 checksum of genesis.json:
\`\`\`
$GENESIS_HASH
\`\`\`

Verify: \`sha256sum genesis.json\`

## Authorities

EOF

echo "$AUTHORITIES" | while read -r addr; do
    echo "- \`$addr\`" >> "$OUTPUT_DIR/README.md"
done

cat >> "$OUTPUT_DIR/README.md" <<EOF

## Need Help?

See full documentation at:
https://github.com/podoru/podoru-chain/docs/joining-network.md

## Troubleshooting

Cannot connect to bootstrap peer?
- Check firewall settings on host
- Verify host node is running
- Test connectivity: \`telnet $BOOTSTRAP_PEER\`

Genesis block mismatch?
- Verify genesis.json SHA256 matches: $GENESIS_HASH
- Delete data directory and restart

Not syncing blocks?
- Check peer connections: \`curl localhost:8545/api/v1/node/info\`
- Verify bootstrap peers are reachable
- Check logs for errors
EOF

echo -e "${GREEN}Files generated in: $OUTPUT_DIR/${NC}"
echo ""
echo "Contents:"
echo "  - config.yaml          (node configuration)"
echo "  - genesis.json         (genesis block)"
echo "  - docker-compose.yml   (Docker setup)"
echo "  - README.md            (instructions)"
echo ""
echo -e "${YELLOW}Creating tarball for easy sharing...${NC}"
echo ""

# Always create tarball at consistent location
TARBALL_DIR="$HOME/podoru-chain"
TARBALL="$TARBALL_DIR/podoru-chain-join-info.tar.gz"

# Ensure directory exists
mkdir -p "$TARBALL_DIR"

# Create tarball
tar -czf "$TARBALL" -C "$OUTPUT_DIR" .

echo ""
echo -e "${GREEN}═══════════════════════════════════════════════════════${NC}"
echo -e "${GREEN}Tarball Created Successfully!${NC}"
echo -e "${GREEN}═══════════════════════════════════════════════════════${NC}"
echo ""
echo -e "${BLUE}Location:${NC} $TARBALL"
echo ""
echo -e "${GREEN}Share this file with users who want to join your network.${NC}"
echo ""
echo -e "${YELLOW}For recipients:${NC}"
echo ""
echo "  1. Receive the tarball file"
echo "  2. Run: ${BLUE}make join-wizard${NC}"
echo "  3. Follow the interactive prompts"
echo ""
echo "Or manually extract and configure:"
echo "  tar -xzf podoru-chain-join-info.tar.gz"
echo "  cd <extract-directory>"
echo "  docker-compose up -d"
echo ""

echo -e "${GREEN}═══════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}Additional Resources${NC}"
echo -e "${GREEN}═══════════════════════════════════════════════════════${NC}"
echo ""
echo "Documentation: docs/joining-network.md"
echo "API Reference: http://localhost:$API_PORT/api/v1/"
echo "Node Info:     curl http://localhost:$API_PORT/api/v1/node/info"
echo ""
echo -e "${GREEN}Done!${NC}"
