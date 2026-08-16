package eth

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// currently locked for testing only filtering logs

// var transferTopic = crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)"))

func FetchTransfers(ctx context.Context, client *ethclient.Client, from, to *big.Int) ([]types.Log, error) {
	// q := ethereum.FilterQuery{
	// 	FromBlock: from,
	// 	ToBlock:   to,
	// 	Addresses: []common.Address{common.HexToAddress("0xA0b86991c6218b36c1d19D4a2E3606eB48")},
	// 	Topics:    [][]common.Hash{{transferTopic}},
	// }

	return client.FilterLogs(ctx, ethereum.FilterQuery{
		FromBlock: from,
		ToBlock:   to,
	})
}
