package eth1

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type rpcBlock struct {
	Hash         common.Hash         `json:"hash"`
	Transactions []rpcTransaction    `json:"transactions"`
	UncleHashes  []common.Hash       `json:"uncles"`
	Withdrawals  []*types.Withdrawal `json:"withdrawals,omitempty"`
}

type rpcTransaction struct {
	tx *types.Transaction
	txExtraInfo
}

type txExtraInfo struct {
	BlockNumber *string         `json:"blockNumber,omitempty"`
	BlockHash   *common.Hash    `json:"blockHash,omitempty"`
	From        *common.Address `json:"from,omitempty"`
}

func (tx *rpcTransaction) UnmarshalJSON(msg []byte) error {
	if err := json.Unmarshal(msg, &tx.tx); err != nil {
		return err
	}

	return json.Unmarshal(msg, &tx.txExtraInfo)
}

func ParseEthBlock(blockData json.RawMessage) (*types.Block, error) {
	var resultHeader types.Header
	if err := json.Unmarshal(blockData, &resultHeader); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON header: %w", err)
	}

	var body rpcBlock
	if err := json.Unmarshal(blockData, &body); err != nil {
		return nil, err
	}

	// get transactions
	txs := make([]*types.Transaction, len(body.Transactions))
	for idx := range body.Transactions {
		txs[idx] = body.Transactions[idx].tx
	}

	return types.NewBlockWithHeader(&resultHeader).WithBody(types.Body{
		Transactions: txs,
		Uncles:       nil,
		Withdrawals:  body.Withdrawals,
	}), nil
}

func LoadBlockFromFile(filePath string) (*types.Block, error) {
	var blockBytes []byte

	if strings.HasPrefix(filePath, "http://") || strings.HasPrefix(filePath, "https://") {
		resp, err := http.Get(filePath) //nolint:gosec // This is a valid use case as we want to load the block from a variable URL
		if err != nil {
			return nil, fmt.Errorf("failed to get block from URL: %w", err)
		}

		defer resp.Body.Close()

		blockBytes, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read block from URL: %w", err)
		}
	} else {
		var err error

		blockBytes, err = os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read shadow fork block: %w", err)
		}
	}

	// Unmarshal the JSON into a types.Block object
	var resultData JSONData

	err2 := json.Unmarshal(blockBytes, &resultData)
	if err2 != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err2)
	}

	// Set the eth1Block value for use later
	block, err2 := ParseEthBlock(resultData.Result)
	if err2 != nil {
		return nil, fmt.Errorf("failed to parse eth1 block: %w", err2)
	}

	return block, nil
}
