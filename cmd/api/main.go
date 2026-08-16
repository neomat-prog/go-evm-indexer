package main

import (
	"context"
	"log"
	"os"
	"time"

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

}
