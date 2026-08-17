package eth

import (
	"context"
	"fmt"
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

func parseTransfer(log types.Log) (Transfer, error) {
	if len(log.Topics) != 3 {
		return Transfer{}, fmt.Errorf("invalid Transfer log: expected 3 got %d", len(log.Topics))
	}
	if len(log.Data) != 32 {
		return Transfer{}, fmt.Errorf("invalid Transfer log data length: %d", len(log.Data))
	}
	return Transfer{
		From:   common.BytesToAddress(log.Topics[1].Bytes()[12:]),
		To:     common.BytesToAddress(log.Topics[2].Bytes()[12:]),
		Value:  new(big.Int).SetBytes(log.Data),
		Block:  log.BlockNumber,
		TxHash: log.TxHash,
	}, nil
}

var transferTopic = crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)"))

func FetchTransfers(ctx context.Context, token common.Address, client *ethclient.Client, from, to *big.Int) ([]Transfer, error) {

	q := ethereum.FilterQuery{
		FromBlock: from,
		ToBlock:   to,
		Addresses: []common.Address{token},
		Topics:    [][]common.Hash{{transferTopic}},
	}

	log, err := client.FilterLogs(ctx, q)
	if err != nil {
		return nil, err
	}

	transfers := make([]Transfer, 0, len(log))

	for _, log := range log {
		transfer, err := parseTransfer(log)
		if err != nil {
			return nil, err
		}

		transfers = append(transfers, transfer)
	}

	return transfers, nil

}
