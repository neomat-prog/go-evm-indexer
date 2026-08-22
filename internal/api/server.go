package api

import (
	"context"
	"errors"
	"log"
	"math/big"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ChainReader interface {
	BlockNumber(ctx context.Context) (uint64, error)
}

type Server struct {
	listenAddr string
	pool       *pgxpool.Pool
	chainID    *big.Int
	token      common.Address
}

func NewServer(listenAddr string, pool *pgxpool.Pool, chainID *big.Int, token common.Address) *Server {
	return &Server{listenAddr: listenAddr, pool: pool, chainID: chainID, token: token}
}

// Run serves until ctx is cancelled, then drains in-flight requests. Signal
// handling belongs to main, which owns the ctx shared with the indexer.
func (s *Server) Run(ctx context.Context) error {
	srv := &http.Server{
		Addr:              s.listenAddr,
		Handler:           s.routes(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Println("server listening on", s.listenAddr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		log.Println("server shutting down")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return srv.Shutdown(shutdownCtx)
}
