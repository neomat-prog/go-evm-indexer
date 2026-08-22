package db

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neomat-prog/go-evm-indexer/internal/eth"
)

func NewPostgres(ctx context.Context, databaseUrl string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, databaseUrl)
	if err != nil {
		return nil, fmt.Errorf("create postgres pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	fmt.Println("bluetooth device has connected ah succesfully :)")
	return pool, nil
}

func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	if _, err := pool.Exec(ctx, Schema); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	return nil
}

// LastIndexedBlock returns the highest block scanned for this (chain, token),
// or fallback if the pair has never been indexed. Note this is the last block
// *scanned*, not the last block that happened to contain a Transfer.
func LastIndexedBlock(ctx context.Context, pool *pgxpool.Pool, chainID *big.Int, token common.Address, fallback uint64) (uint64, error) {
	var block int64

	err := pool.QueryRow(ctx,
		`SELECT last_indexed_block FROM indexer_state WHERE chain_id = $1 AND token = $2`,
		chainID.Int64(), strings.ToLower(token.Hex()),
	).Scan(&block)

	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return fallback, nil
	case err != nil:
		return 0, fmt.Errorf("last indexed block: %w", err)
	}

	return uint64(block), nil
}

// SaveChunk writes the transfers found in blocks up to toBlock and advances the
// cursor to toBlock, both in one transaction.
//
// The atomicity is the point: if the cursor were committed ahead of the rows, a
// crash in between would leave the cursor claiming blocks are indexed that are
// not, and nothing would ever go back for them. Committing together means a
// crash can only ever cost a re-scan, which insertTransfer's ON CONFLICT
// absorbs.
func SaveChunk(ctx context.Context, pool *pgxpool.Pool, chainID *big.Int, token common.Address, toBlock uint64, transfers []eth.Transfer) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer tx.Rollback(ctx) // no-op once Commit has succeeded

	if len(transfers) > 0 {
		batch := &pgx.Batch{}
		for _, t := range transfers {
			batch.Queue(InsertTransfer,
				t.ChainID.Int64(),
				strings.ToLower(t.Token.Hex()),
				int64(t.BlockNumber),
				t.BlockHash.Hex(),
				t.TxHash.Hex(),
				int32(t.TxIndex),
				int32(t.LogIndex),
				strings.ToLower(t.From.Hex()),
				strings.ToLower(t.To.Hex()),
				pgtype.Numeric{Int: t.Value, Exp: 0, Valid: true},
			).Exec(func(pgconn.CommandTag) error {
				return nil
			})
		}

		if err := tx.SendBatch(ctx, batch).Close(); err != nil {
			return fmt.Errorf("save transfers: %w", err)
		}
	}

	// Runs even when the chunk held no transfers: an empty range is still a
	// scanned range, and the cursor has to move or the loop refetches it forever.
	if _, err := tx.Exec(ctx, UpsertCursor,
		chainID.Int64(), strings.ToLower(token.Hex()), int64(toBlock),
	); err != nil {
		return fmt.Errorf("advance cursor: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit chunk: %w", err)
	}
	return nil
}
