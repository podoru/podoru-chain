# Chain Endpoints

Endpoints for querying blockchain information and metadata.

## GET /chain/info

Get comprehensive blockchain information including height, latest block, and authorities.

### Request

```http
GET /api/v1/chain/info
```

### Response

```json
{
  "success": true,
  "data": {
    "height": 1234,
    "latest_block_hash": "0xabc123def456...",
    "authorities": [
      "0x3D4b25CBdda1014F74F9C80f040ce1Bb69130CBB",
      "0x4aa37EEc2a26a4e04b7b206f32D6C2C63219F5cd",
      "0x304F73DD4CabF754eF2240fF2bC2446eB7709652"
    ],
    "genesis_hash": "0x000000...",
    "chain_id": "podoru-mainnet"
  }
}
```

### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| height | integer | Current blockchain height (latest block number) |
| latest_block_hash | string | Hash of the most recent block |
| authorities | array | List of authorized block producer addresses |
| genesis_hash | string | Hash of the genesis block |
| chain_id | string | Unique identifier for this blockchain |

### Example

```bash
curl http://localhost:8545/api/v1/chain/info | jq
```

```javascript
const response = await fetch('http://localhost:8545/api/v1/chain/info')
const { data } = await response.json()
console.log(`Blockchain height: ${data.height}`)
```

```python
import requests

response = requests.get('http://localhost:8545/api/v1/chain/info')
data = response.json()['data']
print(f"Blockchain height: {data['height']}")
```

### Use Cases

- Display blockchain statistics
- Monitor blockchain health
- Verify node synchronization
- Check authority set
- Application initialization

### Error Responses

**500 Internal Server Error**:
```json
{
  "success": false,
  "error": "Failed to retrieve blockchain info"
}
```

## Related Endpoints

- [GET /block/latest](blocks.md#get-blocklatest) - Get latest block details
- [GET /node/info](node.md#get-nodeinfo) - Get node-specific information
- [GET /node/peers](node.md#get-nodepeers) - Get peer information
