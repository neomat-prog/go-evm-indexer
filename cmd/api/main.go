package main

import (
	"flag"
	"log"

	"github.com/neomat-prog/go-evm-indexer/internal/api"
)

func main() {
	listenAddr := flag.String("listen", ":3000", "HTTP listen address")
	flag.Parse()

	if err := api.NewServer(*listenAddr).Run(); err != nil {
		log.Fatal(err)
	}
}
