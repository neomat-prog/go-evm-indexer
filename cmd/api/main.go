package main

import (
	"log"

	"github.com/neomat-prog/go-evm-indexer/internal/indexer"
)

func main() {

	// TODO(neomat-prog): implement go functions in order to run concurrently go-eth client and go-server

	if err := indexer.RunIndexing(); err != nil {
		log.Fatal(err)
	}
}
