<img
  src="assets/gopher.png"
  alt="Gopher"
  width="400"
  align="right"
/>

<h1>evm-indexer</h1>

<p>A small Go service that indexes ERC-20 <code>Transfer</code> events from an
EVM chain into Postgres and serves them over HTTP.</p>

The indexer scans blocks in chunks, stays 12 blocks behind head for reorg safety, and saves each chunk with its cursor in one transaction, so restarts resume where they stopped.

## Requirements

* Go 1.27+
* Postgres
* An Ethereum JSON-RPC URL (Alchemy, Infura, or your own node)

## Setup

```sh
cp .env.example .env   # set ETH_RPC_URL and DB_URL
make run
```

Tables are created on startup.

## Configuration

| Variable | Default | Description |
|---|---|---|
| `ETH_RPC_URL` | required | JSON-RPC endpoint |
| `DB_URL` | required | Postgres connection string |
| `TOKEN_ADDR` | USDC | ERC-20 contract to index |
| `ETH_CHUNK_SIZE` | 10 | Blocks per `eth_getLogs` call |
| `LOOKBACK` | 100 | Blocks behind head to start from on first run |
| `INTERVAL` | 12s | Wait between head polls once caught up |
| `LISTEN_ADDR` | :8080 | HTTP listen address |

## API

| Endpoint | Returns |
|---|---|
| `GET /healthz` | `{"status":"ok"}` |
| `GET /index` | Last indexed block, or `null` |
| `GET /transfers` | 10 latest transfers |
| `GET /logs?token=0x...` | First 10 blocks with transfers for a token (defaults to `TOKEN_ADDR`) |

## Tests

```sh
make test
```

Postgres tests run only when `DB_URL` is set.
