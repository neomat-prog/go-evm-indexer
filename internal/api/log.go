package api

import (
	"net/http"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/jackc/pgx/v5"
	"github.com/neomat-prog/go-evm-indexer/internal/db"
)

type Log struct {
	Token string
}

const firstBlocksLimit = 10

func (s *Server) handleTokenLog(w http.ResponseWriter, r *http.Request) error {
	token := s.token
	if q := r.URL.Query().Get("token"); q != "" {
		if !common.IsHexAddress(q) {
			return apiError{Err: "token is not a hex address", Status: http.StatusBadRequest}
		}
		token = common.HexToAddress(q)
	}

	rows, err := s.pool.Query(r.Context(), db.FetchFirstBlocks, s.chainID.Int64(), strings.ToLower(token.Hex()), firstBlocksLimit)
	if err != nil {
		return err
	}

	blocks, err := pgx.CollectRows(rows, pgx.RowTo[int64])
	if err != nil {
		return err
	}

	return writeJSON(w, http.StatusOK, map[string]any{
		"token":  strings.ToLower(token.Hex()),
		"blocks": blocks,
	})

}
