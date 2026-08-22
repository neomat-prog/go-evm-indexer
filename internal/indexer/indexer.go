package indexer

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neomat-prog/go-evm-indexer/internal/config"
	"github.com/neomat-prog/go-evm-indexer/internal/db"
	"github.com/neomat-prog/go-evm-indexer/internal/eth"
)

// Blocks behind head to stop at. Reorg safety, not a tunable.
const confirmations = 12

type Indexer struct {
	pool   *pgxpool.Pool
	client *ethclient.Client
	cfg    config.ConfigOpts
}

func New(pool *pgxpool.Pool, client *ethclient.Client, cfg config.ConfigOpts) *Indexer {
	// Zero Interval panics NewTicker; zero Chunk loops forever.
	if cfg.Interval <= 0 {
		cfg.Interval = config.DefaultInterval
	}
	if cfg.Chunk == 0 {
		cfg.Chunk = config.DefaultChunk
	}
	return &Indexer{pool: pool, client: client, cfg: cfg}
}

// Run blocks until ctx is cancelled or an unrecoverable error occurs.
func (ix *Indexer) Run(ctx context.Context) error {
	chainID, err := ix.client.ChainID(ctx)
	if err != nil {
		return fmt.Errorf("chain id: %w", err)
	}

	head, err := ix.client.BlockNumber(ctx)
	if err != nil {
		return fmt.Errorf("block number: %w", err)
	}

	cursor, err := db.LastIndexedBlock(ctx, ix.pool, chainID, ix.cfg.Token,
		saturatingSub(head, ix.cfg.Lookback))
	if err != nil {
		return err
	}

	log.Printf("indexer: chain=%s token=%s resuming after block %d (head %d)",
		chainID, ix.cfg.Token.Hex(), cursor, head)

	ticker := time.NewTicker(ix.cfg.Interval)
	defer ticker.Stop()

	for {
		if cursor, err = ix.catchUp(ctx, chainID, cursor); err != nil {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

// catchUp scans chunk by chunk up to the safe head, returning the last block
// scanned. Drains the whole backlog at RPC speed, not one chunk per interval.
func (ix *Indexer) catchUp(ctx context.Context, chainID *big.Int, cursor uint64) (uint64, error) {
	head, err := ix.client.BlockNumber(ctx)
	if err != nil {
		return cursor, fmt.Errorf("block number: %w", err)
	}
	safe := saturatingSub(head, confirmations)

	for {
		from, to, ok := nextRange(cursor, safe, ix.cfg.Chunk)
		if !ok {
			return cursor, nil
		}

		transfers, err := eth.FetchTransfers(ctx, ix.client, chainID, ix.cfg.Token,
			new(big.Int).SetUint64(from), new(big.Int).SetUint64(to))
		if err != nil {
			return cursor, err
		}

		if err := db.SaveChunk(ctx, ix.pool, chainID, ix.cfg.Token, to, transfers); err != nil {
			return cursor, err
		}

		log.Printf("indexer: blocks %d-%d (head %d): %d transfers",
			from, to, head, len(transfers))
		cursor = to

		// Notice cancellation mid-backfill, not just at the ticker.
		if err := ctx.Err(); err != nil {
			return cursor, err
		}
	}
}

// nextRange picks the next window, or ok == false once cursor reaches safe.
func nextRange(cursor, safe, chunk uint64) (from, to uint64, ok bool) {
	if cursor >= safe {
		return 0, 0, false
	}
	return cursor + 1, min(cursor+chunk, safe), true
}

// saturatingSub avoids the wraparound of a-b when head < offset (fresh devnet).
func saturatingSub(a, b uint64) uint64 {
	if a < b {
		return 0
	}
	return a - b
}
