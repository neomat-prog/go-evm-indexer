package server

import (
	"errors"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"
)

type transfer struct {
	BlockNumber int64  `json:"block_number"`
	TxHash      string `json:"tx_hash"`
	LogIndex    int    `json:"log_index"`
	From        string `json:"from"`
	To          string `json:"to"`
	Value       string `json:"value"`
}

func (s *Server) handleLatestTransfers(w http.ResponseWriter, r *http.Request) error {
	rows, err := s.pool.Query(r.Context(), `SELECT block_number, tx_hash, log_index, from_addr, to_addr, value::text
  FROM transfers
  WHERE chain_id = $1 AND token = $2
  ORDER BY block_number DESC, log_index DESC
  LIMIT $3`, s.chainID.Int64(), strings.ToLower(s.token.Hex()), 10)
	if err != nil {
		return err
	}

	transfers, err := pgx.CollectRows(rows, pgx.RowToStructByPos[transfer])
	if err != nil {
		return err
	}

	return writeJSON(w, http.StatusOK, transfers)
}

func (s *Server) handleIndexerState(w http.ResponseWriter, r *http.Request) error {
	var block int64
	err := s.pool.QueryRow(r.Context(), `SELECT last_indexed_block FROM indexer_state WHERE chain_id = $1 AND token = $2`, s.chainID.Int64(), strings.ToLower(s.token.Hex())).Scan(&block)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return writeJSON(w, http.StatusOK, map[string]any{"last_indexed_block": nil})
	case err != nil:
		return err
	}
	return writeJSON(w, http.StatusOK, map[string]any{"last_indexed_block": block})
}
