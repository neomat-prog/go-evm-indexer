package server

import (
	"errors"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"
)

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
