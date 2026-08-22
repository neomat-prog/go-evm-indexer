package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

const (
	DefaultListenAddr = ":8080"
	DefaultTokenAddr  = "0xA0b86991c6218b36c1d19D4a2E9EB0cE3606eB48" // USDC
	DefaultInterval   = 12 * time.Second
	DefaultLookback   = 100
	DefaultChunk      = 10
)

type ConfigOpts struct {
	ListenAddr string
	DBURL      string
	ETHRPCURL  string

	Token    common.Address // ERC-20 contract whose Transfer logs are indexed
	Interval time.Duration  // wait between head polls once caught up
	Lookback uint64         // blocks behind head to start at, first run only
	Chunk    uint64         // blocks per eth_getLogs call; providers cap this
}

func Load() (ConfigOpts, error) {
	cfg := ConfigOpts{
		ListenAddr: envStr("LISTEN_ADDR", DefaultListenAddr),
		DBURL:      envStr("DB_URL", ""),
		ETHRPCURL:  envStr("ETH_RPC_URL", ""),
		Interval:   envDur("INTERVAL", DefaultInterval),
		Lookback:   envUint("LOOKBACK", DefaultLookback),
		Chunk:      envUint("ETH_CHUNK_SIZE", DefaultChunk),
	}

	if cfg.DBURL == "" {
		return cfg, fmt.Errorf("config: DB_URL is required")
	}
	if cfg.ETHRPCURL == "" {
		return cfg, fmt.Errorf("config: ETH_RPC_URL is required")
	}

	// HexToAddress silently zero-pads garbage, so reject it here instead.
	token := envStr("TOKEN_ADDR", DefaultTokenAddr)
	if !common.IsHexAddress(token) {
		return cfg, fmt.Errorf("config: TOKEN_ADDR %q is not a hex address", token)
	}
	cfg.Token = common.HexToAddress(token)

	return cfg, nil
}

func envStr(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

func envUint(key string, def uint64) uint64 {
	v, err := strconv.ParseUint(os.Getenv(key), 10, 64)
	if err != nil {
		return def
	}
	return v
}

func envDur(key string, def time.Duration) time.Duration {
	v, err := time.ParseDuration(os.Getenv(key))
	if err != nil {
		return def
	}
	return v
}
