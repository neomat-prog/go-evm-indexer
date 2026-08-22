package config

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestLoad(t *testing.T) {
	t.Setenv("DB_URL", "postgres://localhost/x")
	t.Setenv("ETH_RPC_URL", "http://localhost:8545")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() = %v", err)
	}
	if cfg.Chunk != DefaultChunk || cfg.Interval != DefaultInterval || cfg.Lookback != DefaultLookback {
		t.Errorf("optional settings not defaulted: %+v", cfg)
	}
	if cfg.Token != common.HexToAddress(DefaultTokenAddr) {
		t.Errorf("Token = %s, want %s", cfg.Token, DefaultTokenAddr)
	}

	t.Setenv("ETH_CHUNK_SIZE", "2000")
	if cfg, _ := Load(); cfg.Chunk != 2000 {
		t.Errorf("ETH_CHUNK_SIZE ignored: Chunk = %d", cfg.Chunk)
	}

	// A typo would otherwise zero-pad into a contract with no logs.
	t.Setenv("TOKEN_ADDR", "0xnothex")
	if _, err := Load(); err == nil {
		t.Error("Load() accepted a non-hex TOKEN_ADDR")
	}
}

func TestLoadRequiresURLs(t *testing.T) {
	t.Setenv("DB_URL", "")
	t.Setenv("ETH_RPC_URL", "http://localhost:8545")
	if _, err := Load(); err == nil {
		t.Error("Load() accepted an empty DB_URL")
	}

	t.Setenv("DB_URL", "postgres://localhost/x")
	t.Setenv("ETH_RPC_URL", "")
	if _, err := Load(); err == nil {
		t.Error("Load() accepted an empty ETH_RPC_URL")
	}
}
