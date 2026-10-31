package server

import (
	"net/http"

	"github.com/neomat-prog/go-evm-indexer/internal/db"
)

func (s *Server) handleLatestTransfers(w http.ResponseWriter, r *http.Request) error {
	transfers, err := db.LatestTransferBlocks(r.Context(), s.pool, s.chainID, s.token, 10)
	if err != nil {
		return err
	}

	return writeJSON(w, http.StatusOK, transfers)
}

func (s *Server) handleIndexerState(w http.ResponseWriter, r *http.Request) error {
	block, err := db.LastIndexedBlock(r.Context(), s.pool, s.chainID, s.token, 0)
	if err != nil {
		return err
	}
	if block == 0 {
		return writeJSON(w, http.StatusOK, map[string]any{"last_indexed_block": nil})
	}
	return writeJSON(w, http.StatusOK, map[string]any{"last_indexed_block": block})
}
