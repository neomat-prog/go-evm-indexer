package db

const Schema = `
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

  CREATE TABLE IF NOT EXISTS indexer_state (
        chain_id           BIGINT NOT NULL,
        token              TEXT   NOT NULL,
        last_indexed_block BIGINT NOT NULL,
        PRIMARY KEY (chain_id, token)
  );
  `

const InsertTransfer = `
      INSERT INTO transfers (
              chain_id, token, block_number, block_hash,
              tx_hash, tx_index, log_index,
              from_addr, to_addr, value
      ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
      ON CONFLICT (chain_id, tx_hash, log_index) DO NOTHING
`

// The WHERE clause keeps the cursor monotonic: a retried or out-of-order chunk
// can never rewind it.
const UpsertCursor = `
      INSERT INTO indexer_state (chain_id, token, last_indexed_block)
      VALUES ($1,$2,$3)
      ON CONFLICT (chain_id, token) DO UPDATE
              SET last_indexed_block = EXCLUDED.last_indexed_block
      WHERE indexer_state.last_indexed_block < EXCLUDED.last_indexed_block
`
