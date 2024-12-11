package evm

import (
	"context"
	"fmt"
	"runtime"

	"github.com/ethereum/go-ethereum/core"
	gethstate "github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/stateless"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	gethparams "github.com/ethereum/go-ethereum/params"
	"github.com/kkrt-labs/kakarot-controller/src"
)

// ExecParams are the parameters for an EVM execution.
type ExecParams struct {
	VMConfig *vm.Config // VM configuration
	Block    *types.Block
	Validate bool // Whether to validate the block
	State    *gethstate.StateDB
	Chain    *core.HeaderChain
	Reporter func(error)
}

// Executor is an interface for executing EVM blocks.
type Executor interface {
	Execute(ctx context.Context, params *ExecParams) (*core.ProcessResult, error)
}

// ExecutorFunc is a function that executes an EVM block.
type ExecutorFunc func(ctx context.Context, params *ExecParams) (*core.ProcessResult, error)

func (f ExecutorFunc) Execute(ctx context.Context, params *ExecParams) (*core.ProcessResult, error) {
	return f(ctx, params)
}

// ExecutorDecorator is a function that decorates an EVM executor.
type ExecutorDecorator func(Executor) Executor

type executor struct{}

// NewExecutor creates a new EVM executor.
func NewExecutor() Executor {
	return &executor{}
}

// Execute executes an EVM block.
// It processes the block on the given state and chain then validates the block if requested.
//
//nolint:gocritic // TODO: uncomment commented code after adding metrics
func (e *executor) Execute(_ context.Context, params *ExecParams) (res *core.ProcessResult, execErr error) {
	if vmCfg := params.VMConfig; vmCfg != nil && vmCfg.Tracer != nil {
		if vmCfg.Tracer.OnBlockStart != nil {
			vmCfg.Tracer.OnBlockStart(tracing.BlockEvent{
				Block: params.Block,
			})
		}
		if vmCfg.Tracer.OnBlockEnd != nil {
			defer func() {
				vmCfg.Tracer.OnBlockEnd(execErr)
			}()
		}
	}

	if params.Chain.Config().IsByzantium(params.Block.Number()) {
		if params.VMConfig.StatelessSelfValidation {
			// Create witness for tracking state accesses
			witness, err := stateless.NewWitness(params.Block.Header(), params.Chain)
			if err != nil {
				execErr = fmt.Errorf("failed to create witness: %v", err)
				return
			}

			params.State.StartPrefetcher("chain", witness)

			defer func() {
				params.State.StopPrefetcher()
			}()
		}
	}

	// Process block on given state
	processor := core.NewStateProcessor(params.Chain.Config(), params.Chain)

	// pstart := time.Now()
	res, execErr = processor.Process(params.Block, params.State, *params.VMConfig)
	if execErr != nil {
		execErr = fmt.Errorf("failed to process block: %v", execErr)
		if params.Reporter != nil {
			params.Reporter(summarizeBadBlockError(params.Chain.Config(), params.Block, res, execErr))
		}
		return
	}
	// ptime := time.Since(pstart)

	// Process block on given state

	if params.Validate {
		// vstart := time.Now()
		validator := core.NewBlockValidator(params.Chain.Config(), nil)
		if execErr = validator.ValidateState(params.Block, params.State, res, false); execErr != nil {
			execErr = fmt.Errorf("failed to validate block: %v", execErr)
			if params.Reporter != nil {
				params.Reporter(summarizeBadBlockError(params.Chain.Config(), params.Block, res, execErr))
			}
			return
		}
		// vtime := time.Since(vstart)
	}

	// TODO: add metrics
	// Update the metrics touched during block processing and validation
	// accountReadTimer.Update(statedb.AccountReads) // Account reads are complete(in processing)
	// storageReadTimer.Update(statedb.StorageReads) // Storage reads are complete(in processing)
	// if statedb.AccountLoaded != 0 {
	// 	accountReadSingleTimer.Update(statedb.AccountReads / time.Duration(statedb.AccountLoaded))
	// }
	// if statedb.StorageLoaded != 0 {
	// 	storageReadSingleTimer.Update(statedb.StorageReads / time.Duration(statedb.StorageLoaded))
	// }
	// accountUpdateTimer.Update(statedb.AccountUpdates)                                 // Account updates are complete(in validation)
	// storageUpdateTimer.Update(statedb.StorageUpdates)                                 // Storage updates are complete(in validation)
	// accountHashTimer.Update(statedb.AccountHashes)                                    // Account hashes are complete(in validation)
	// triehash := statedb.AccountHashes                                                 // The time spent on tries hashing
	// trieUpdate := statedb.AccountUpdates + statedb.StorageUpdates                     // The time spent on tries update
	// blockExecutionTimer.Update(ptime - (statedb.AccountReads + statedb.StorageReads)) // The time spent on EVM processing
	// blockValidationTimer.Update(vtime - (triehash + trieUpdate))                      // The time spent on block validation
	// blockCrossValidationTimer.Update(xvtime)                                          // The time spent on stateless cross validation

	return
}

// summarizeBadBlock generates a human-readable summary of a bad block.
func summarizeBadBlockError(chainCfg *gethparams.ChainConfig, block *types.Block, res *core.ProcessResult, err error) error {
	var receipts types.Receipts
	if res != nil {
		receipts = res.Receipts
	}

	var receiptString string
	for i, receipt := range receipts {
		receiptString += fmt.Sprintf("\n  %d: cumulative: %v gas: %v contract: %v status: %v tx: %v logs: %v bloom: %x state: %x",
			i, receipt.CumulativeGasUsed, receipt.GasUsed, receipt.ContractAddress.Hex(),
			receipt.Status, receipt.TxHash.Hex(), receipt.Logs, receipt.Bloom, receipt.PostState)
	}

	version := src.Version
	vcs := ""
	platform := fmt.Sprintf("%s %s %s %s", version, runtime.Version(), runtime.GOARCH, runtime.GOOS)
	if vcs != "" {
		vcs = fmt.Sprintf("\nVCS: %s", vcs)
	}

	return fmt.Errorf(`########## BAD BLOCK #########
Block: %v (%#x)
Error: %v
Chain config: %#v
Platform: %v%v
Receipts: %v
##############################`, block.Number(), block.Hash(), err, platform, vcs, chainCfg, receiptString)
}
