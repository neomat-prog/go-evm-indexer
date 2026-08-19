package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"os"
	"os/signal"
	"syscall"

	"github.com/ethereum/go-ethereum/common"
	"github.com/neomat-prog/go-evm-indexer/internal/db"
	"github.com/neomat-prog/go-evm-indexer/internal/eth"
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

	head, err := client.BlockNumber(ctx)
	if err != nil {
		return fmt.Errorf("block number: %w", err)
	}

	from := new(big.Int).SetUint64(head - 9)
	to := new(big.Int).SetUint64(head)

	transfers, err := eth.FetchTransfers(ctx, usdc, client, from, to)
	if err != nil {
		return err
	}

	if err := db.SaveTransfers(ctx, pool, transfers); err != nil {
		return err
	}

	log.Printf("saved %d transfers from blocks %s-%s", len(transfers), from, to)
	return nil
}
