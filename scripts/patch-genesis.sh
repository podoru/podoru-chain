#!/bin/bash
# Patch existing genesis.json with token and gas configuration
# Usage: ./scripts/patch-genesis.sh [genesis_file] [output_file]

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Default values
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Default token config
TOKEN_NAME="Podoru"
TOKEN_SYMBOL="PDR"
TOKEN_DECIMALS=18
INITIAL_SUPPLY="100000000000000000000000000" # 100M PDR in wei

# Default gas config
BASE_FEE="1000"
PER_BYTE_FEE="10"

usage() {
    echo "Usage: $0 [OPTIONS] [genesis_file] [output_file]"
    echo ""
    echo "Patch existing genesis.json with token and gas configuration."
    echo ""
    echo "Arguments:"
    echo "  genesis_file    Path to existing genesis.json (default: config/genesis.json)"
    echo "  output_file     Path to output file (default: overwrites genesis_file)"
    echo ""
    echo "Options:"
    echo "  -h, --help           Show this help message"
    echo "  --dry-run            Show what would be changed without modifying files"
    echo "  --no-balances        Skip adding initial balances"
    echo "  --supply AMOUNT      Initial supply in wei (default: $INITIAL_SUPPLY)"
    echo "  --base-fee FEE       Base fee in wei (default: $BASE_FEE)"
    echo "  --per-byte-fee FEE   Per-byte fee in wei (default: $PER_BYTE_FEE)"
    echo ""
    echo "Examples:"
    echo "  $0                                    # Patch config/genesis.json in place"
    echo "  $0 config/genesis.json                # Patch specific file in place"
    echo "  $0 config/genesis.json config/new.json # Create new patched file"
    echo "  $0 --dry-run                          # Preview changes"
    exit 1
}

# Parse options
DRY_RUN=false
NO_BALANCES=false

while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            usage
            ;;
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        --no-balances)
            NO_BALANCES=true
            shift
            ;;
        --supply)
            INITIAL_SUPPLY="$2"
            shift 2
            ;;
        --base-fee)
            BASE_FEE="$2"
            shift 2
            ;;
        --per-byte-fee)
            PER_BYTE_FEE="$2"
            shift 2
            ;;
        *)
            break
            ;;
    esac
done

# Set input and output files
GENESIS_FILE="${1:-$PROJECT_ROOT/config/genesis.json}"
OUTPUT_FILE="${2:-$GENESIS_FILE}"

# Check if genesis file exists
if [[ ! -f "$GENESIS_FILE" ]]; then
    echo -e "${RED}Error: Genesis file not found: $GENESIS_FILE${NC}"
    exit 1
fi

# Check if jq is installed
if ! command -v jq &> /dev/null; then
    echo -e "${RED}Error: jq is required but not installed.${NC}"
    echo "Install with: sudo apt-get install jq (Ubuntu) or brew install jq (macOS)"
    exit 1
fi

echo -e "${GREEN}Patching genesis file: $GENESIS_FILE${NC}"

# Read existing genesis
GENESIS=$(cat "$GENESIS_FILE")

# Check if already patched
if echo "$GENESIS" | jq -e '.token_config' &> /dev/null; then
    echo -e "${YELLOW}Warning: Genesis file already has token_config. Overwriting...${NC}"
fi

# Get authorities from genesis
AUTHORITIES=$(echo "$GENESIS" | jq -r '.authorities[]')
AUTH_COUNT=$(echo "$AUTHORITIES" | wc -l)

if [[ $AUTH_COUNT -eq 0 ]]; then
    echo -e "${RED}Error: No authorities found in genesis file${NC}"
    exit 1
fi

echo -e "Found ${GREEN}$AUTH_COUNT${NC} authorities"

# Calculate balance per authority (distribute evenly)
# Using bc for big number division
BALANCE_PER_AUTH=$(echo "scale=0; $INITIAL_SUPPLY / $AUTH_COUNT" | bc)
REMAINDER=$(echo "scale=0; $INITIAL_SUPPLY % $AUTH_COUNT" | bc)

echo -e "Distributing ${GREEN}$INITIAL_SUPPLY${NC} wei among authorities"
echo -e "  Each authority gets: ${GREEN}$BALANCE_PER_AUTH${NC} wei"
if [[ "$REMAINDER" != "0" ]]; then
    echo -e "  Remainder (${GREEN}$REMAINDER${NC} wei) goes to first authority"
fi

# Build initial_balances object
BALANCE_JSON="{"
FIRST=true
COUNTER=0
while IFS= read -r auth; do
    if [[ -z "$auth" ]]; then
        continue
    fi

    # Normalize address to lowercase
    auth=$(echo "$auth" | tr '[:upper:]' '[:lower:]')

    # First authority gets remainder
    if [[ $FIRST == true && "$REMAINDER" != "0" ]]; then
        AUTH_BALANCE=$(echo "$BALANCE_PER_AUTH + $REMAINDER" | bc)
        FIRST=false
    else
        AUTH_BALANCE=$BALANCE_PER_AUTH
        FIRST=false
    fi

    if [[ $COUNTER -gt 0 ]]; then
        BALANCE_JSON+=","
    fi
    BALANCE_JSON+="\"$auth\":\"$AUTH_BALANCE\""
    COUNTER=$((COUNTER + 1))
done <<< "$AUTHORITIES"
BALANCE_JSON+="}"

# Build the patched genesis
PATCHED_GENESIS=$(echo "$GENESIS" | jq --argjson balances "$BALANCE_JSON" \
    --arg token_name "$TOKEN_NAME" \
    --arg token_symbol "$TOKEN_SYMBOL" \
    --argjson token_decimals "$TOKEN_DECIMALS" \
    --arg initial_supply "$INITIAL_SUPPLY" \
    --arg base_fee "$BASE_FEE" \
    --arg per_byte_fee "$PER_BYTE_FEE" \
    --argjson no_balances "$NO_BALANCES" \
    '. + {
        token_config: {
            name: $token_name,
            symbol: $token_symbol,
            decimals: $token_decimals,
            initial_supply: $initial_supply
        },
        gas_config: {
            base_fee: $base_fee,
            per_byte_fee: $per_byte_fee
        }
    } + (if $no_balances then {} else {initial_balances: $balances} end)')

if [[ $DRY_RUN == true ]]; then
    echo ""
    echo -e "${YELLOW}=== DRY RUN - Changes preview ===${NC}"
    echo ""
    echo "$PATCHED_GENESIS" | jq .
    echo ""
    echo -e "${YELLOW}=== End of preview ===${NC}"
    echo ""
    echo "To apply changes, run without --dry-run"
else
    # Write output
    echo "$PATCHED_GENESIS" | jq . > "$OUTPUT_FILE"
    echo ""
    echo -e "${GREEN}Successfully patched genesis file!${NC}"
    echo -e "Output written to: ${GREEN}$OUTPUT_FILE${NC}"
    echo ""
    echo "Token Configuration:"
    echo "  Name: $TOKEN_NAME"
    echo "  Symbol: $TOKEN_SYMBOL"
    echo "  Decimals: $TOKEN_DECIMALS"
    echo "  Initial Supply: $INITIAL_SUPPLY wei"
    echo ""
    echo "Gas Configuration:"
    echo "  Base Fee: $BASE_FEE wei"
    echo "  Per-Byte Fee: $PER_BYTE_FEE wei"
    echo ""
    if [[ $NO_BALANCES == false ]]; then
        echo "Initial Balances distributed to $AUTH_COUNT authorities"
    fi
fi
