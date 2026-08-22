package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/ethereum/go-ethereum/common"
	"github.com/neomat-prog/go-evm-indexer/internal/api"
	"github.com/neomat-prog/go-evm-indexer/internal/db"
	"github.com/neomat-prog/go-evm-indexer/internal/eth"
	"github.com/neomat-prog/go-evm-indexer/internal/indexer"
	"golang.org/x/sync/errgroup"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := db.NewPostgres(ctx, os.Getenv("DB_URL"))
	if err != nil {
		return err
	}
	defer pool.Close()

	if err := db.Migrate(ctx, pool); err != nil {
		return err
	}

	client, err := eth.NewClient(ctx, eth.ConfigOpts{
		RPCServerAddr: os.Getenv("ETH_RPC_URL"),
	})
	if err != nil {
		return err
	}
	defer client.Close()

	usdc := common.HexToAddress("0xA0b86991c6218b36c1d19D4a2E9EB0cE3606eB48")

	listenAddr := os.Getenv("LISTEN_ADDR")
	if listenAddr == "" {
		listenAddr = ":8080"
	}

	srv := api.NewServer(listenAddr)
	ix := indexer.New(pool, client, indexer.Config{
		Token: usdc,
		Chunk: envUint("ETH_CHUNK_SIZE"),
	})

	// Both loops block forever, so each needs its own goroutine. errgroup ties
	// them together: whichever fails first cancels gctx, and the other unwinds
	// through its own ctx.Done() path.
	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() error { return srv.Run(gctx) })
	g.Go(func() error { return ix.Run(gctx) })

	// A clean Ctrl-C cancels ctx, which surfaces as context.Canceled rather
	// than a real failure.
	if err := g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		return err
	}

	log.Println("shutdown complete")
	return nil
}

// envUint reads an optional numeric setting, returning 0 when unset or
// unparseable so the caller's own default applies.
func envUint(key string) uint64 {
	n, err := strconv.ParseUint(os.Getenv(key), 10, 64)
	if err != nil {
		return 0
	}
	return n
}
