package eth1

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/core"
)

func LoadEth1GenesisConfig(configPath string) (*core.Genesis, error) {
	eth1ConfData, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read eth1 config file: %v", err)
	}
	var eth1Genesis core.Genesis
	if err := json.NewDecoder(bytes.NewReader(eth1ConfData)).Decode(&eth1Genesis); err != nil {
		return nil, fmt.Errorf("failed to decode eth1 config file: %v", err)
	}
	return &eth1Genesis, nil
}
