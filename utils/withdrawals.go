package utils

import (
	"fmt"

	"github.com/attestantio/go-eth2-client/spec/bellatrix"
	"github.com/attestantio/go-eth2-client/spec/capella"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/ethereum/go-ethereum/core/types"
	ssz "github.com/ferranbt/fastssz"

	"github.com/ethpandaops/eth-beacon-genesis/config"
)

func ComputeWithdrawalsRoot(withdrawals types.Withdrawals, config *config.Config) (phase0.Root, error) {
	// Compute the SSZ hash-tree-root of the withdrawals,
	// since that is what we put as withdrawals_root in the CL execution-payload.
	// Not to be confused with the legacy MPT root in the EL block header.
	num := uint64(len(withdrawals))
	max := config.GetUintDefault("MAX_WITHDRAWALS_PER_PAYLOAD", 16)
	if num > max {
		return phase0.Root{}, fmt.Errorf("withdrawals list is too long")
	}

	clWithdrawals := make([]capella.Withdrawal, len(withdrawals))
	for i, withdrawal := range withdrawals {
		clWithdrawals[i] = capella.Withdrawal{
			Index:          capella.WithdrawalIndex(withdrawal.Index),
			ValidatorIndex: phase0.ValidatorIndex(withdrawal.Validator),
			Address:        bellatrix.ExecutionAddress(withdrawal.Address),
			Amount:         phase0.Gwei(withdrawal.Amount),
		}
	}

	withdrawalsRoot, err := HashWithFastSSZHasher(func(hh *ssz.Hasher) error {
		for _, elem := range clWithdrawals {
			if err := elem.HashTreeRootWith(hh); err != nil {
				return err
			}
		}
		hh.MerkleizeWithMixin(0, num, max)
		return nil
	})
	if err != nil {
		return phase0.Root{}, err
	}
	return phase0.Root(withdrawalsRoot), nil
}
