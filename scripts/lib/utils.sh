#!/bin/bash
# Utility functions for Podoru Chain Setup Wizard

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1" >&2
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1" >&2
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1" >&2
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
}

# Safe file write - atomic operation using temp file
safe_write_config() {
    local target="$1"
    local content="$2"

    local temp_file="${target}.tmp.$$"

    # Write to temp file
    if echo "$content" > "$temp_file" 2>/dev/null; then
        # Move temp to target (atomic)
        if mv "$temp_file" "$target" 2>/dev/null; then
            return 0
        else
            log_error "Failed to move temp file to $target"
            rm -f "$temp_file"
            return 1
        fi
    else
        log_error "Failed to write to temp file $temp_file"
        rm -f "$temp_file"
        return 1
    fi
}

# Cleanup on error - remove partial setup
cleanup_on_error() {
    local exit_code=$?

    if [ $exit_code -ne 0 ]; then
        log_error "Setup failed with exit code $exit_code"

        if [ "${WIZARD_CLEANUP_ON_ERROR:-1}" = "1" ]; then
            log_warning "Rolling back partial setup..."

            # Remove generated data directories
            rm -rf docker/data/producer* docker/data/fullnode* 2>/dev/null

            # Remove temp docker-compose file
            rm -f docker/docker-compose.yml.tmp 2>/dev/null
            rm -f docker/docker-compose.yml.new 2>/dev/null

            # Remove temp files
            rm -f /tmp/wizard_*.$$  2>/dev/null

            log_info "Partial setup has been rolled back"
        fi
    fi

    # Clean up temp files on normal exit too
    rm -f /tmp/wizard_*.$$  2>/dev/null
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check if port is in use
is_port_in_use() {
    local port="$1"

    if command_exists lsof; then
        lsof -Pi ":$port" -sTCP:LISTEN -t >/dev/null 2>&1
        return $?
    elif command_exists netstat; then
        netstat -tuln 2>/dev/null | grep -q ":$port "
        return $?
    elif command_exists ss; then
        ss -tuln 2>/dev/null | grep -q ":$port "
        return $?
    else
        # Can't check, assume available
        return 1
    fi
}

# Generate timestamp for genesis
get_genesis_timestamp() {
    date +%s
}

# Extract address from keygen output
extract_address() {
    local keygen_output="$1"
    echo "$keygen_output" | grep -oP 'Address:\s*\K0x[a-fA-F0-9]{40}' | head -1
}

# Create directory if it doesn't exist
ensure_directory() {
    local dir="$1"

    if [ ! -d "$dir" ]; then
        mkdir -p "$dir" 2>/dev/null
        if [ $? -ne 0 ]; then
            log_error "Failed to create directory: $dir"
            return 1
        fi
    fi

    return 0
}

# Check if Docker is running
is_docker_running() {
    if ! command_exists docker; then
        return 1
    fi

    docker info >/dev/null 2>&1
    return $?
}

# Check if docker-compose is available
has_docker_compose() {
    if command_exists docker-compose; then
        return 0
    elif docker compose version >/dev/null 2>&1; then
        return 0
    else
        return 1
    fi
}

# Run docker-compose command (handles both docker-compose and docker compose)
run_docker_compose() {
    if command_exists docker-compose; then
        docker-compose "$@"
    else
        docker compose "$@"
    fi
}

# Confirm action with user
confirm() {
    local message="$1"
    local response

    while true; do
        read -p "$message (y/n): " response
        case "$response" in
            [Yy]* ) return 0;;
            [Nn]* ) return 1;;
            * ) echo "Please answer yes (y) or no (n).";;
        esac
    done
}

# Display a progress spinner
spinner() {
    local pid=$1
    local delay=0.1
    local spinstr='|/-\'

    while kill -0 "$pid" 2>/dev/null; do
        local temp=${spinstr#?}
        printf " [%c]  " "$spinstr"
        spinstr=$temp${spinstr%"$temp"}
        sleep $delay
        printf "\b\b\b\b\b\b"
    done
    printf "    \b\b\b\b"
}

# Check all prerequisites
check_prerequisites() {
    local missing=()

    # Check for dialog
    if ! command_exists dialog; then
        missing+=("dialog")
    fi

    # Check for Go
    if ! command_exists go; then
        missing+=("go")
    fi

    # Check for Docker
    if ! command_exists docker; then
        missing+=("docker")
    fi

    # Check for jq (optional but recommended)
    if ! command_exists jq; then
        log_warning "jq is not installed (optional for testing)"
    fi

    if [ ${#missing[@]} -gt 0 ]; then
        log_error "Missing required dependencies: ${missing[*]}"
        log_info "Please install them and try again"
        return 1
    fi

    # Check if Docker is running
    if ! is_docker_running; then
        log_error "Docker is not running. Please start Docker and try again"
        return 1
    fi

    return 0
}

# Export functions for use in other scripts
export -f log_info
export -f log_success
export -f log_warning
export -f log_error
export -f safe_write_config
export -f cleanup_on_error
export -f command_exists
export -f is_port_in_use
export -f get_genesis_timestamp
export -f extract_address
export -f ensure_directory
export -f is_docker_running
export -f has_docker_compose
export -f run_docker_compose
export -f confirm
export -f spinner
export -f check_prerequisites
