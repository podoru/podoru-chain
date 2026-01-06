# Podoru Chain Explorer

A real-time blockchain explorer for Podoru Chain built with TanStack Start and React.

## Features

- **Real-time Activity Feed** - WebSocket-powered live updates for new blocks and transactions
- **Block Explorer** - Browse recent blocks with detailed information
- **Transaction Browser** - View pending transactions in the mempool
- **State Browser** - Query and explore the key-value state
- **Network Dashboard** - Monitor connected peers, authorities, and node health

## Quick Start

### Option 1: Using Docker (Recommended)

Build and run the full stack (blockchain + explorer):

```bash
# From the root of podoru-chain repository
make stack-up
```

The explorer will be available at **http://localhost:3000**

View logs:
```bash
make stack-logs
```

Stop the stack:
```bash
make stack-down
```

### Option 2: Development Mode

Run the explorer in development mode with hot reloading:

```bash
# From the root of podoru-chain repository
make explorer-dev

# Or from the explorer directory
cd explorer
npm install
npm run dev
```

The explorer will be available at **http://localhost:3000**

## Configuration

The explorer connects to a Podoru Chain node via environment variables:

- `VITE_API_URL` - REST API endpoint (default: `http://localhost:8545`)
- `VITE_WS_URL` - WebSocket endpoint (default: `ws://localhost:8545/api/v1/ws`)

### Docker Environment

In Docker Compose, these are set automatically:
- API: `http://producer1:8545`
- WebSocket: `ws://producer1:8545/api/v1/ws`

### Local Development

Create a `.env` file in the explorer directory:

```env
VITE_API_URL=http://localhost:8545
VITE_WS_URL=ws://localhost:8545/api/v1/ws
```

## Pages

### Dashboard (`/`)
- Real-time activity feed with live blocks and transactions
- Current chain statistics (height, authorities)
- WebSocket connection status

### Blocks (`/blocks`)
- List of recent blocks
- Block height, hash, producer, timestamp
- Transaction count per block

### Transactions (`/transactions`)
- Pending transactions in mempool
- Transaction operations (SET/DELETE)
- Sender addresses and nonces

### State Browser (`/state`)
- Query key-value state by prefix
- View all state entries
- JSON-formatted values

### Network (`/network`)
- Node information and type
- Connected peers
- Block authorities list

## Building

### Production Build

```bash
# From root
make explorer-build

# Or from explorer directory
cd explorer
npm run build
```

### Docker Image

```bash
# From root
make explorer-docker

# Or manually
docker build -t podoru-explorer:latest -f explorer/Dockerfile explorer/
```

## Technology Stack

- **TanStack Start** - Full-stack React framework with SSR
- **TanStack Router** - Type-safe file-based routing
- **React 18** - UI framework
- **WebSocket API** - Real-time updates
- **Vinxi** - Build tool
- **Vite** - Fast development server

## Architecture

```
┌─────────────────┐
│  Explorer UI    │  (React + TanStack Start)
└────────┬────────┘
         │
    WebSocket + REST
         │
┌────────▼────────┐
│  Podoru Node    │  (Go + WebSocket Server)
│  REST API       │
└─────────────────┘
```

## API Integration

The explorer uses two communication methods:

1. **REST API** - For querying historical data
   - Chain info, blocks, transactions
   - State queries
   - Node and peer information

2. **WebSocket** - For real-time updates
   - New block events
   - New transaction events
   - Chain updates
   - Mempool changes

## Development

### File Structure

```
explorer/
├── app/
│   ├── routes/          # Page routes
│   │   ├── __root.tsx   # Root layout
│   │   ├── index.tsx    # Dashboard
│   │   ├── blocks.tsx   # Blocks page
│   │   ├── transactions.tsx
│   │   ├── state.tsx
│   │   └── network.tsx
│   ├── hooks/
│   │   └── useWebSocket.ts  # WebSocket hook
│   ├── lib/
│   │   ├── api.ts       # REST API client
│   │   └── utils.ts     # Utility functions
│   ├── router.tsx       # Router config
│   └── ssr.tsx         # SSR entry
├── Dockerfile
├── package.json
└── tsconfig.json
```

### Adding New Features

1. **New Page** - Create file in `app/routes/`
2. **API Method** - Add to `app/lib/api.ts`
3. **WebSocket Event** - Update `app/hooks/useWebSocket.ts`

## Troubleshooting

### Explorer can't connect to node

Check that:
1. Node is running on the correct port (8545)
2. CORS is enabled (automatically configured)
3. Environment variables are set correctly

### WebSocket shows disconnected

Verify:
1. WebSocket endpoint is accessible
2. Network allows WebSocket connections
3. Check browser console for errors

### No data showing

Ensure:
1. Blockchain is producing blocks
2. API endpoint is correct
3. Check network tab in browser DevTools

## License

MIT
