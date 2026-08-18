package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
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
