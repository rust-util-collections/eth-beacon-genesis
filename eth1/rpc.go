package eth1

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type JSONData struct {
	Jsonrpc string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  json.RawMessage `json:"result"`
}

func GetBlockFromRPC(ctx context.Context, host string) (*types.Block, error) {
	client, err := ethclient.Dial(host)
	if err != nil {
		return nil, fmt.Errorf("failed to create the ETH client %s", err)
	}

	// Get the latest block
	blockNumberUint64, err := client.BlockNumber(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get the block number %s", err)
	}

	blockNumberBigint := new(big.Int).SetUint64(blockNumberUint64)

	resultBlock, err := client.BlockByNumber(ctx, blockNumberBigint)
	if err != nil {
		return nil, fmt.Errorf("failed to get the ETH block %s", err)
	}

	return resultBlock, nil
}
