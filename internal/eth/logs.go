package eth

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Transfer struct {
	From   common.Address
	To     common.Address
	Value  *big.Int
	Block  uint64
	TxHash common.Hash
}

// func parseTransfer(types.Log) (Transfer, error) {}

var transferTopic = crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)"))

func FetchTransfers(ctx context.Context, token common.Address, client *ethclient.Client, from, to *big.Int) ([]types.Log, error) {

	q := ethereum.FilterQuery{
		FromBlock: from,
		ToBlock:   to,
		Addresses: []common.Address{token},
		Topics:    [][]common.Hash{{transferTopic}},
	}

	return client.FilterLogs(ctx, q)

}
