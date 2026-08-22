package eth

import (
	"context"
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/neomat-prog/go-evm-indexer/internal/config"
)

func NewClient(ctx context.Context, cfg config.ConfigOpts) (*ethclient.Client, error) {
	if cfg.ETHRPCURL == "" {
		return nil, errors.New("eth: ETH_RPC_URL is required")
	}

	client, err := ethclient.DialContext(ctx, cfg.ETHRPCURL)
	if err != nil {
		return nil, fmt.Errorf("dial: %w", err)
	}

	if _, err := client.ChainID(ctx); err != nil {
		client.Close()
		return nil, fmt.Errorf("eth_chainId: %w", err)
	}

	return client, nil
}
