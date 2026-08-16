package eth

import (
	"context"
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/ethclient"
)

func NewClient(ctx context.Context, cfg ConfigOpts) (*ethclient.Client, error) {
	if cfg.RPCServerAddr == "" {
		return nil, errors.New("eth: RPCServerAddr is required")
	}

	client, err := ethclient.DialContext(ctx, cfg.RPCServerAddr)
	if err != nil {
		return nil, fmt.Errorf("dial: %w", err)
	}

	if _, err := client.ChainID(ctx); err != nil {
		client.Close()
		return nil, fmt.Errorf("eth_chainId: %w", err)
	}

	return client, nil
}
