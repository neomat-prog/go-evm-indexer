package db

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neomat-prog/go-evm-indexer/internal/eth"
)

const schema = `
  CREATE TABLE IF NOT EXISTS transfers (
        chain_id     BIGINT        NOT NULL,
        token        TEXT          NOT NULL,
        block_number BIGINT        NOT NULL,
        block_hash   TEXT          NOT NULL,
        tx_hash      TEXT          NOT NULL,
        tx_index     INT           NOT NULL,
        log_index    INT           NOT NULL,
        from_addr    TEXT          NOT NULL,
        to_addr      TEXT          NOT NULL,
        value        NUMERIC(78,0) NOT NULL,
        PRIMARY KEY (chain_id, tx_hash, log_index)
  );
  CREATE INDEX IF NOT EXISTS transfers_block_idx ON transfers (chain_id, 
  token, block_number);
  `

const insertTransfer = `
      INSERT INTO transfers (
              chain_id, token, block_number, block_hash,
              tx_hash, tx_index, log_index,
              from_addr, to_addr, value
      ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
      ON CONFLICT (chain_id, tx_hash, log_index) DO NOTHING
`

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
	if _, err := pool.Exec(ctx, schema); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	return nil
}

func SaveTransfers(ctx context.Context, pool *pgxpool.Pool, transfers []eth.Transfer) error {
	if len(transfers) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	for _, t := range transfers {
		batch.Queue(insertTransfer,
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

	if err := pool.SendBatch(ctx, batch).Close(); err != nil {
		return fmt.Errorf("save transfers: %w", err)
	}
	return nil
}
