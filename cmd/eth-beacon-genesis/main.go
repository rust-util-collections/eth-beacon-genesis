package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/attestantio/go-eth2-client/http"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/urfave/cli/v3"

	"github.com/ethpandaops/eth-beacon-genesis/config"
	"github.com/ethpandaops/eth-beacon-genesis/eth1"
	"github.com/ethpandaops/eth-beacon-genesis/generator"
	"github.com/ethpandaops/eth-beacon-genesis/utils"
	"github.com/ethpandaops/eth-beacon-genesis/validators"
)

var (
	eth1ConfigFlag = &cli.StringFlag{
		Name:     "eth1-config",
		Usage:    "Path to execution genesis config (genesis.json)",
		Required: true,
	}
	configFlag = &cli.StringFlag{
		Name:     "config",
		Usage:    "Path to consensus genesis config (config.yaml)",
		Required: true,
	}
	mnemonicsFileFlag = &cli.StringFlag{
		Name:  "mnemonics",
		Usage: "Path to the file containing the mnemonics for genesis validators",
	}
	validatorsFileFlag = &cli.StringFlag{
		Name:  "additional-validators",
		Usage: "Path to the file with a list of additional genesis validators validators",
	}
	shadowForkBlockFlag = &cli.StringFlag{
		Name:  "shadow-fork-block",
		Usage: "Path to the file with a execution block to create a shadow fork from",
	}
	shadowForkRPCFlag = &cli.StringFlag{
		Name:  "shadow-fork-rpc",
		Usage: "Execution RPC URL to fetch the block to create a shadow fork from",
	}
	stateOutputFlag = &cli.StringFlag{
		Name:  "state-output",
		Usage: "Path to the file to write the genesis state to in SSZ format",
	}
	jsonOutputFlag = &cli.StringFlag{
		Name:  "json-output",
		Usage: "Path to the file to write the genesis state to in JSON format",
	}

	quietFlag = &cli.BoolFlag{
		Name:    "quiet",
		Aliases: []string{"q"},
		Usage:   "Suppress output",
	}

	app = &cli.Command{
		Name:  "eth-beacon-genesis",
		Usage: "Generate the Ethereum 2.0 Beacon Chain genesis state",
		Commands: []*cli.Command{
			{
				Name:  "devnet",
				Usage: "Generate a devnet genesis state",
				Flags: []cli.Flag{
					eth1ConfigFlag, configFlag, mnemonicsFileFlag, validatorsFileFlag,
					shadowForkBlockFlag, shadowForkRPCFlag, stateOutputFlag, jsonOutputFlag,
					quietFlag,
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return runDevnet(ctx, cmd)
				},
				UsageText: "eth-beacon-genesis devnet [options]",
			},
			{
				Name:  "version",
				Usage: "Print the version of the application",
				Flags: []cli.Flag{},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					fmt.Printf("eth-beacon-genesis version %s\n", utils.GetBuildVersion())
					return nil
				},
			},
		},
		DefaultCommand: "devnet",
	}
)

func main() {
	if err := app.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

func runDevnet(ctx context.Context, cmd *cli.Command) error {
	eth1Config := cmd.String(eth1ConfigFlag.Name)
	eth2Config := cmd.String(configFlag.Name)
	mnemonicsFile := cmd.String(mnemonicsFileFlag.Name)
	validatorsFile := cmd.String(validatorsFileFlag.Name)
	shadowForkBlock := cmd.String(shadowForkBlockFlag.Name)
	shadowForkRPC := cmd.String(shadowForkRPCFlag.Name)
	stateOutputFile := cmd.String(stateOutputFlag.Name)
	jsonOutputFile := cmd.String(jsonOutputFlag.Name)
	quiet := cmd.Bool(quietFlag.Name)

	elGenesis, err := eth1.LoadEth1GenesisConfig(eth1Config)
	if err != nil {
		return fmt.Errorf("failed to load execution genesis: %w", err)
	}

	clConfig, err := config.LoadConfig(eth2Config)
	if err != nil {
		return fmt.Errorf("failed to load consensus config: %w", err)
	}

	var clValidators []*validators.Validator
	if mnemonicsFile != "" {
		vals, err := validators.GenerateValidatorsByMnemonic(mnemonicsFile, quiet)
		if err != nil {
			return fmt.Errorf("failed to load validators from mnemonics file: %w", err)
		}

		if len(vals) > 0 {
			clValidators = vals
		}
	}

	if validatorsFile != "" {
		vals, err := validators.LoadValidatorsFromFile(validatorsFile)
		if err != nil {
			return fmt.Errorf("failed to load validators from file: %w", err)
		}

		if len(vals) > 0 {
			clValidators = append(clValidators, vals...)
		}
	}

	if len(clValidators) == 0 {
		return fmt.Errorf("no validators found")
	}

	builder := generator.NewGenesisBuilder(elGenesis, clConfig)
	builder.AddValidators(clValidators)

	if shadowForkBlock != "" || shadowForkRPC != "" {
		var gensisBlock *types.Block

		if shadowForkBlock != "" {
			blockBytes, err := os.ReadFile(shadowForkBlock)
			if err != nil {
				return fmt.Errorf("failed to read shadow fork block: %w", err)
			}

			// Unmarshal the JSON into a types.Block object
			var resultData eth1.JSONData
			err = json.Unmarshal(blockBytes, &resultData)
			if err != nil {
				return fmt.Errorf("failed to unmarshal JSON: %w", err)
			}

			// Set the eth1Block value for use later
			block, err := eth1.ParseEthBlock(resultData.Result)
			if err != nil {
				return fmt.Errorf("failed to parse eth1 block: %w", err)
			}

			gensisBlock = block
		} else {
			block, err := eth1.GetBlockFromRPC(ctx, shadowForkRPC)
			if err != nil {
				return fmt.Errorf("failed to get shadow fork block: %w", err)
			}
			gensisBlock = block
		}

		builder.SetShadowForkBlock(gensisBlock)
	}
	genesisState, err := builder.BuildState()
	if err != nil {
		return fmt.Errorf("failed to build genesis: %w", err)
	}

	if stateOutputFile != "" {
		sszData, err := builder.Serialize(genesisState, http.ContentTypeSSZ)
		if err != nil {
			return fmt.Errorf("failed to serialize genesis state: %w", err)
		}

		os.WriteFile(stateOutputFile, sszData, 0644)
	}

	if jsonOutputFile != "" {
		jsonData, err := builder.Serialize(genesisState, http.ContentTypeJSON)
		if err != nil {
			return fmt.Errorf("failed to serialize genesis state: %w", err)
		}

		os.WriteFile(jsonOutputFile, jsonData, 0644)
	}

	if stateOutputFile == "" && jsonOutputFile == "" {
		jsonData, err := builder.Serialize(genesisState, http.ContentTypeJSON)
		if err != nil {
			return fmt.Errorf("failed to serialize genesis state: %w", err)
		}

		fmt.Println(string(jsonData))
	}

	return nil
}
