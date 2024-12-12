package blockstore

import (
	"context"

	blockinputs "github.com/kkrt-labs/kakarot-controller/src/blocks/inputs"
)

type BlockStore interface {
	PreflightDataStore
	ProvableInputsStore
}

type PreflightDataStore interface {
	// StorePreflightData stores the preflight data for a block.
	StorePreflightData(ctx context.Context, data *blockinputs.PreflightData) error

	// LoadPreflightData loads the preflight data for a block.
	LoadPreflightData(ctx context.Context, chainID, blockNumber uint64) (*blockinputs.PreflightData, error)
}

type ProvableInputsStore interface {
	// StoreProvableInputs stores the provable inputs for a block.
	StoreProvableInputs(ctx context.Context, data *blockinputs.ProvableInputs) error

	// LoadProvableInputs loads the provable inputs for a block.
	LoadProvableInputs(ctx context.Context, chainID, blockNumber uint64) (*blockinputs.ProvableInputs, error)
}
