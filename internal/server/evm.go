package server

import (
	"errors"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/neomat-prog/go-evm-indexer/internal/db"
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
	rows, err := s.pool.Query(r.Context(), db.FetchLatestTransferBlock, s.chainID.Int64(), strings.ToLower(s.token.Hex()), 10)
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
	err := s.pool.QueryRow(r.Context(), db.FetchIndexState, s.chainID.Int64(), strings.ToLower(s.token.Hex())).Scan(&block)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return writeJSON(w, http.StatusOK, map[string]any{"last_indexed_block": nil})
	case err != nil:
		return err
	}
	return writeJSON(w, http.StatusOK, map[string]any{"last_indexed_block": block})
}
