package main

import (
	"context"
	"log"
	"os"

	"github.com/neomat-prog/go-evm-indexer/internal/db"
)

// Reference Transfer struct{}

// {
// From:
// 	0x2B3e7Adb2827c8183c1c6A733e8013c7D236F2f8
// To:
//  0x654de8A1b6F2B2f8De8C32b7b5B50d401D3fc897
// Amount: exactly 444.91
// 	444910000
// Block number:
// 	25782690
// 	Transaction hash: 0x133f759486001acf716ad8dbdc89573a8f414eb04b15c0357260c0c0621135b1
// }

func main() {

	ctx := context.Background()

	pool, err := db.NewPostgres(
		ctx,
		os.Getenv("DB_URL"),
	)

	if err != nil {
		log.Fatal(err)
	}

	defer pool.Close()

	if err := db.Migrate(ctx, pool); err != nil {
		log.Fatal(err)
	}

	// Unlock only after we have chain readers

	// listenAddr := flag.String("listen", ":3000", "HTTP listen address")

	// if err := api.NewServer(*listenAddr).Run(); err != nil {
	// 	log.Fatal(err)
	// }

	// ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	// defer cancel()

	// client, err := eth.NewClient(ctx, eth.ConfigOpts{
	// 	RPCServerAddr: os.Getenv("ETH_RPC_URL"),
	// })
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer client.Close()

	// log.Println("connected")

	// usdc := common.HexToAddress("0xA0b86991c6218b36c1d19D4a2E9EB0cE3606eB48")

	// head, err := client.BlockNumber(ctx)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// from := new(big.Int).SetUint64(head - 9)
	// to := new(big.Int).SetUint64(head)

	// logs, err := eth.FetchTransfers(ctx, usdc, client, from, to)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// if len(logs) == 0 {
	// 	log.Fatalf("no logs in blocks %s-%s", from, to)
	// }

	// for _, t := range logs {
	// 	log.Printf("block %d  %s -> %s  %s USDC  %s", t.BlockNumber, t.From, t.To, usdc6(t.Value), t.TxHash)
	// }
}

// usdc6 formats a raw USDC amount (6 decimals) as a decimal string.
// func usdc6(v *big.Int) string {
// 	return new(big.Rat).SetFrac(v, big.NewInt(1e6)).FloatString(6)
// }
