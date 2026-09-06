package eth

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/neomat-prog/go-evm-indexer/internal/config"
)

func NewClient(ctx context.Context, cfg config.ConfigOpts) (*ethclient.Client, *big.Int, error) {
	if cfg.ETHRPCURL == "" {
		return nil, nil, errors.New("eth: ETH_RPC_URL is required")
	}

	client, err := ethclient.DialContext(ctx, cfg.ETHRPCURL)
	if err != nil {
		return nil, nil, fmt.Errorf("dial: %w", err)
	}
	chainID, err := client.ChainID(ctx)
	if err != nil {
		client.Close()
		return nil, nil, fmt.Errorf("eth_chainId: %w", err)
	}

	return client, chainID, nil
}
