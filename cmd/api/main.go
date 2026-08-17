package main

import (
	"context"
	"log"
	"math/big"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/neomat-prog/go-evm-indexer/internal/eth"
)

func main() {

	// Unlock only after we have chain readers

	// listenAddr := flag.String("listen", ":3000", "HTTP listen address")

	// if err := api.NewServer(*listenAddr).Run(); err != nil {
	// 	log.Fatal(err)
	// }

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := eth.NewClient(ctx, eth.ConfigOpts{
		RPCServerAddr: os.Getenv("ETH_RPC_URL"),
	})
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	log.Println("connected")

	usdc := common.HexToAddress("0xA0b86991c6218b36c1d19D4a2E9EB0cE3606eB48")

	head, err := client.BlockNumber(ctx)
	if err != nil {
		log.Fatal(err)
	}
	from := new(big.Int).SetUint64(head - 10)
	to := new(big.Int).SetUint64(head)

	logs, err := eth.FetchTransfers(ctx, usdc, client, from, to)
	if err != nil {
		log.Fatal(err)
	}

	if len(logs) == 0 {
		log.Fatalf("no logs in blocks %s-%s", from, to)
	}

	// TODO(neomat-prog): implement todo transfer print after parse
}
