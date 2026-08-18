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
	ChainID *big.Int
	Token   common.Address

	From  common.Address
	To    common.Address
	Value *big.Int

	BlockNumber uint64
	BlockHash   common.Hash

	TxHash   common.Hash
	TxIndex  uint
	LogIndex uint
}

func parseTransfer(chainID *big.Int, log types.Log) (Transfer, error) {
	if len(log.Topics) != 3 {
		return Transfer{}, fmt.Errorf(
			"invalid Transfer log: expected 3 topics, got %d",
			len(log.Topics),
		)
	}

	if len(log.Data) != 32 {
		return Transfer{}, fmt.Errorf(
			"invalid Transfer log data length: %d",
			len(log.Data),
		)
	}

	return Transfer{
		ChainID: new(big.Int).Set(chainID),
		Token:   log.Address,

		From:  common.BytesToAddress(log.Topics[1].Bytes()[12:]),
		To:    common.BytesToAddress(log.Topics[2].Bytes()[12:]),
		Value: new(big.Int).SetBytes(log.Data),

		BlockNumber: log.BlockNumber,
		BlockHash:   log.BlockHash,

		TxHash:   log.TxHash,
		TxIndex:  log.TxIndex,
		LogIndex: log.Index,
	}, nil
}

var transferTopic = crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)"))

func FetchTransfers(ctx context.Context, token common.Address, client *ethclient.Client, from, to *big.Int) ([]Transfer, error) {

	chainID, err := client.ChainID(ctx)

	q := ethereum.FilterQuery{
		FromBlock: from,
		ToBlock:   to,
		Addresses: []common.Address{token},
		Topics:    [][]common.Hash{{transferTopic}},
	}

	log, err := client.FilterLogs(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("filter logs: %w", err)
	}

	transfers := make([]Transfer, 0, len(log))

	for _, log := range log {
		transfer, err := parseTransfer(chainID, log)
		if err != nil {
			return nil, fmt.Errorf(
				"parse transfer tx=%s log=%d: %w",
				log.TxHash.Hex(),
				log.Index,
				err,
			)
		}

		transfers = append(transfers, transfer)
	}

	return transfers, nil

}
