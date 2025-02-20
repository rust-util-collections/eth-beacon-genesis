package utils

import (
	"github.com/ethpandaops/eth-beacon-genesis/config"
	dynssz "github.com/pk910/dynamic-ssz"
)

func GetDynSSZ(config *config.Config) *dynssz.DynSsz {
	spec := config.GetSpecs()
	dynSsz := dynssz.NewDynSsz(spec)
	return dynSsz
}
