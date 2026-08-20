package main

import (
	"log"

	"github.com/neomat-prog/go-evm-indexer/internal/indexer"
)

func main() {
	if err := indexer.RunIndexing(); err != nil {
		log.Fatal(err)
	}
}
