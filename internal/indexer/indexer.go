// Package indexer walks the chain forward in bounded chunks, writing ERC-20
// Transfer logs to Postgres and recording how far it has scanned.
//
// The cursor is an in-memory number: "every block at or below this is fully in
// the database". It is seeded once at startup from indexer_state and advanced
// in lockstep with the rows it describes, so a restart resumes exactly where
// the previous process stopped instead of skipping whatever was mined while it
// was down.
package indexer

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neomat-prog/go-evm-indexer/internal/db"
	"github.com/neomat-prog/go-evm-indexer/internal/eth"
)

const (
	// hardcoded default chunk value, choose any other default chunk you need
	defaultChunk = 10

	// confirmations keeps the window this far behind head, since blocks nearer
	// the tip can still be reorged out from under us.
	//
	// ponytail: one constant for one chain. Move it into Config when a second
	// chain with a different finality assumption shows up.
	confirmations = 12

	defaultInterval = 12 * time.Second
	defaultLookback = 100
)

// Config holds the tunables for an Indexer. Every field except Token has a
// usable zero value.
type Config struct {
	// Token is the ERC-20 contract whose Transfer logs are indexed.
	Token common.Address

	// Interval is how long to wait after catching up before polling head again.
	Interval time.Duration

	// Lookback is how far behind head to begin when this token has never been
	// indexed. Ignored once a cursor exists.
	Lookback uint64

	// Chunk is the number of blocks per eth_getLogs call. Providers cap this
	// and reject anything wider, so it is a property of the RPC plan, not a
	// performance dial.
	Chunk uint64
}

func (c *Config) setDefaults() {
	if c.Interval <= 0 {
		c.Interval = defaultInterval
	}
	if c.Lookback == 0 {
		c.Lookback = defaultLookback
	}
	if c.Chunk == 0 {
		c.Chunk = defaultChunk
	}
}

type Indexer struct {
	pool   *pgxpool.Pool
	client *ethclient.Client
	cfg    Config
}

func New(pool *pgxpool.Pool, client *ethclient.Client, cfg Config) *Indexer {
	cfg.setDefaults()
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

// catchUp scans chunk by chunk from cursor to the current safe head, returning
// the last block scanned. It drains the whole backlog without waiting on the
// ticker, so a long outage is caught up at RPC speed rather than one chunk per
// interval.
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

		// A backfill can run for a long time; notice cancellation between
		// chunks rather than only at the ticker.
		if err := ctx.Err(); err != nil {
			return cursor, err
		}
	}
}

// nextRange picks the next window to fetch, or reports ok == false when the
// cursor has reached safe. from excludes cursor and to includes safe, so
// consecutive calls tile the chain with no gap and no overlap.
//
// The window spans at most chunk blocks inclusive of both ends, which is how
// providers count the limit they advertise.
func nextRange(cursor, safe, chunk uint64) (from, to uint64, ok bool) {
	if cursor >= safe {
		return 0, 0, false
	}
	return cursor + 1, min(cursor+chunk, safe), true
}

// saturatingSub avoids the wraparound that plain a-b gives on a chain whose
// head is still lower than the offset, e.g. a fresh devnet.
func saturatingSub(a, b uint64) uint64 {
	if a < b {
		return 0
	}
	return a - b
}
