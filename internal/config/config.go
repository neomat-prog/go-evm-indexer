package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	ListenAddr   string
	DBURL        string
	ETHRPCURL    string
	TokenAddr    string
	BlockWindow  uint64
	PollInterval time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		ListenAddr:   envStr("LISTEN_ADDR", ":3000"),
		DBURL:        envStr("DB_URL", ""),
		ETHRPCURL:    envStr("ETH_RPC_URL", ""),
		TokenAddr:    envStr("TOKEN_ADDR", "0xA0b86991c6218b36c1d19D4a2E9EB0cE3606eB48"),
		BlockWindow:  envUint("BLOCK_WINDOW", 10),
		PollInterval: envDur("POLL_INTERVAL", 12*time.Second),
	}

	if cfg.DBURL == "" {
		return cfg, fmt.Errorf("config: DB_URL is required")
	}
	if cfg.ETHRPCURL == "" {
		return cfg, fmt.Errorf("config: ETH_RPC_URL is required")
	}
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
