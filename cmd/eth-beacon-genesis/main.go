package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/attestantio/go-eth2-client/http"
	"github.com/urfave/cli/v3"

	"github.com/ethpandaops/eth-beacon-genesis/config"
	"github.com/ethpandaops/eth-beacon-genesis/eth1"
	"github.com/ethpandaops/eth-beacon-genesis/generator"
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
	outputFlag = &cli.StringFlag{
		Name:  "output",
		Usage: "Path to the file to write the genesis state to in SSZ format",
	}
	outputJsonFlag = &cli.StringFlag{
		Name:  "output-json",
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
					outputFlag, outputJsonFlag, quietFlag,
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return runDevnet(ctx, cmd)
				},
				UsageText: "shamir-msg split [options]",
			},
		},
		DefaultCommand: "run",
	}
)

func main() {
	if err := app.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

func runDevnet(_ context.Context, cmd *cli.Command) error {
	eth1Config := cmd.String(eth1ConfigFlag.Name)
	eth2Config := cmd.String(configFlag.Name)
	mnemonicsFile := cmd.String(mnemonicsFileFlag.Name)
	validatorsFile := cmd.String(validatorsFileFlag.Name)
	outputFile := cmd.String(outputFlag.Name)
	outputJson := cmd.String(outputJsonFlag.Name)
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

	genesisState, err := builder.BuildState()
	if err != nil {
		return fmt.Errorf("failed to build genesis: %w", err)
	}

	if outputFile != "" {
		sszData, err := builder.Serialize(genesisState, http.ContentTypeSSZ)
		if err != nil {
			return fmt.Errorf("failed to serialize genesis state: %w", err)
		}

		os.WriteFile(outputFile, sszData, 0644)
	}

	if outputJson != "" {
		jsonData, err := builder.Serialize(genesisState, http.ContentTypeJSON)
		if err != nil {
			return fmt.Errorf("failed to serialize genesis state: %w", err)
		}

		os.WriteFile(outputJson, jsonData, 0644)
	}

	if outputFile == "" && outputJson == "" {
		jsonData, err := builder.Serialize(genesisState, http.ContentTypeJSON)
		if err != nil {
			return fmt.Errorf("failed to serialize genesis state: %w", err)
		}

		fmt.Println(string(jsonData))
	}

	return nil
}
