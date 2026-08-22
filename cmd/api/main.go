package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/neomat-prog/go-evm-indexer/internal/api"
	"github.com/neomat-prog/go-evm-indexer/internal/config"
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

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	pool, err := db.NewPostgres(ctx, cfg.DBURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	if err := db.Migrate(ctx, pool); err != nil {
		return err
	}

	client, err := eth.NewClient(ctx, eth.ConfigOpts{RPCServerAddr: cfg.ETHRPCURL})
	if err != nil {
		return err
	}
	defer client.Close()

	srv := api.NewServer(cfg.ListenAddr)
	idx := indexer.New(pool, client, cfg)

	// Both loops block forever; whichever fails first cancels gctx.
	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() error { return srv.Run(gctx) })
	g.Go(func() error { return idx.Run(gctx) })

	// Ctrl-C surfaces as context.Canceled, not a real failure.
	if err := g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		return err
	}

	log.Println("shutdown complete")
	return nil
}
