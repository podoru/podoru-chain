#!/bin/bash
#
# Podoru Chain Node Update Script
# Pulls latest code, rebuilds Docker image, and restarts containers
#

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}"
echo "========================================"
echo "   Podoru Chain Node Update Script"
echo "========================================"
echo -e "${NC}"

# Find project root (where Makefile exists)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
cd "$PROJECT_ROOT"

echo -e "${YELLOW}Working directory: $PROJECT_ROOT${NC}"

# Detect docker-compose location
if [ -f "docker-compose.yml" ]; then
    COMPOSE_DIR="."
    COMPOSE_FILE="docker-compose.yml"
elif [ -f "docker/docker-compose.yml" ]; then
    COMPOSE_DIR="docker"
    COMPOSE_FILE="docker/docker-compose.yml"
else
    echo -e "${RED}Error: No docker-compose.yml found${NC}"
    exit 1
fi

echo -e "${YELLOW}Using compose file: $COMPOSE_FILE${NC}"

# Show current status
echo ""
echo -e "${BLUE}=== Current Status ===${NC}"
echo -n "Git branch: "
git branch --show-current
echo -n "Last commit: "
git log -1 --oneline

# Check if containers are running
RUNNING_CONTAINERS=$(docker-compose -f "$COMPOSE_FILE" ps -q 2>/dev/null | wc -l)
if [ "$RUNNING_CONTAINERS" -gt 0 ]; then
    echo -e "${GREEN}Running containers: $RUNNING_CONTAINERS${NC}"
else
    echo -e "${YELLOW}No containers running${NC}"
fi

# Pull latest code
echo ""
echo -e "${BLUE}=== Pulling Latest Code ===${NC}"
git pull origin master || git pull origin main || {
    echo -e "${RED}Failed to pull from remote. Continuing with local code...${NC}"
}

echo -n "New commit: "
git log -1 --oneline

# Rebuild Docker image
echo ""
echo -e "${BLUE}=== Rebuilding Docker Image ===${NC}"
docker build -f docker/Dockerfile -t podoru-chain:latest . || {
    echo -e "${RED}Failed to build Docker image${NC}"
    exit 1
}
echo -e "${GREEN}Docker image rebuilt successfully${NC}"

# Restart containers if they were running
if [ "$RUNNING_CONTAINERS" -gt 0 ]; then
    echo ""
    echo -e "${BLUE}=== Restarting Containers ===${NC}"

    # Stop containers
    echo "Stopping containers..."
    docker-compose -f "$COMPOSE_FILE" down

    # Start containers
    echo "Starting containers..."
    docker-compose -f "$COMPOSE_FILE" up -d

    echo -e "${GREEN}Containers restarted${NC}"

    # Wait a moment for containers to start
    sleep 3

    # Show running containers
    echo ""
    echo -e "${BLUE}=== Running Containers ===${NC}"
    docker-compose -f "$COMPOSE_FILE" ps
else
    echo ""
    echo -e "${YELLOW}No containers were running. Use 'docker-compose up -d' to start.${NC}"
fi

echo ""
echo -e "${GREEN}========================================"
echo "   Update Complete!"
echo "========================================${NC}"
