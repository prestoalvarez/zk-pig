package evm

import (
	"context"
	"math/big"

	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/tracing"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	gethvm "github.com/ethereum/go-ethereum/core/vm"
	"github.com/kkrt-labs/kakarot-controller/pkg/log"
	"github.com/kkrt-labs/kakarot-controller/pkg/tag"
	"github.com/sirupsen/logrus"
)

// ExecutorWithTags is an executor decorator that adds tags relative to a block execution to the context
// It attaches tags: chain.id, block.number, block.hash
// It also adds the component tag if provided
// If no namespaces are provided (recommended), it attaches the tags to the default namespace
func ExecutorWithTags(component string, namespaces ...string) ExecutorDecorator {
	return func(executor Executor) Executor {
		return ExecutorFunc(func(ctx context.Context, params *ExecParams) (*core.ProcessResult, error) {
			if component != "" {
				ctx = tag.WithComponent(ctx, component)
			}

			tags := []*tag.Tag{
				tag.Key("chain.id").String(params.Chain.Config().ChainID.String()),
				tag.Key("block.number").Int64(params.Block.Number().Int64()),
				tag.Key("block.hash").String(params.Block.Hash().Hex()),
			}

			if len(namespaces) == 0 {
				namespaces = []string{tag.DefaultNamespace}
			}

			for _, ns := range namespaces {
				ctx = tag.WithNamespaceTags(ctx, ns, tags...)
			}

			return executor.Execute(ctx, params)
		})
	}
}

// ExecutorWithLog is an executor decorator that logs block execution
// If namespaces are provided, it loads tags from the provided namespaces
// By default (recommended) it logs tags from the default namespace
func ExecutorWithLog(namespaces ...string) ExecutorDecorator {
	return func(executor Executor) Executor {
		return ExecutorFunc(func(ctx context.Context, params *ExecParams) (*core.ProcessResult, error) {
			logger := log.LoggerWithFieldsFromNamespaceContext(ctx, namespaces...)

			// Set tracing logger
			params.VMConfig.Tracer = NewLoggerTracer(logger).Hooks()

			logger.Debug("Executing block...")
			res, err := executor.Execute(log.WithLogger(ctx, logger), params)
			if err != nil {
				logger.WithError(err).Error("Failed to execute block")
			} else {
				logger.WithField("gasUsed", res.GasUsed).Debug("Executed block")
			}

			return res, err
		})
	}
}

// LoggerTracer is an EVM tracer that logs EVM execution
type LoggerTracer struct {
	logger      logrus.FieldLogger
	blockLogger logrus.FieldLogger
	txLogger    logrus.FieldLogger
}

// NewLoggerTracer creates a new logger tracer
func NewLoggerTracer(logger logrus.FieldLogger) *LoggerTracer {
	return &LoggerTracer{logger: logger}
}

// OnBlockStart logs block execution start
func (t *LoggerTracer) OnBlockStart(event tracing.BlockEvent) {
	t.blockLogger = t.logger.WithFields(logrus.Fields{
		"block.number": event.Block.Number(),
		"block.hash":   event.Block.Hash().Hex(),
	})
	t.blockLogger.Debug("Start executing block")
}

// OnBlockEnd logs block execution end
func (t *LoggerTracer) OnBlockEnd(err error) {
	if err != nil {
		t.blockLogger.WithError(err).Error("failed to execute block")
	} else {
		t.blockLogger.Info("Executed block")
	}
	t.blockLogger = nil
}

// OnTxStart logs transaction execution start
func (t *LoggerTracer) OnTxStart(vm *tracing.VMContext, tx *gethtypes.Transaction, from gethcommon.Address) {
	t.txLogger = t.blockLogger.WithFields(logrus.Fields{
		"tx.type": "transaction",
		"tx.hash": tx.Hash().Hex(),
		"tx.from": from.Hex(),
	})
	t.txLogger.WithField("vm.blocknumber", vm.BlockNumber.String()).Debug("Start executing transaction")
}

// OnTxEnd logs transaction execution end
func (t *LoggerTracer) OnTxEnd(receipt *gethtypes.Receipt, err error) {
	if err != nil {
		t.txLogger.WithError(err).Error("failed to execute transaction")
	} else {
		t.txLogger.WithFields(logrus.Fields{
			"receipt.txHash":          receipt.TxHash.Hex(),
			"receipt.status":          receipt.Status,
			"receipt.gasUsed":         receipt.GasUsed,
			"receipt.postState":       hexutil.Encode(receipt.PostState),
			"receipt.contractAddress": receipt.ContractAddress.Hex(),
		}).Debug("Executed transaction")
	}
	t.txLogger = nil
}

// OnSystemCallStart logs system call execution start
func (t *LoggerTracer) OnSystemCallStart() {
	t.txLogger = t.blockLogger.WithFields(logrus.Fields{
		"tx.type": "system",
	})
	t.txLogger.Debug("Execute system call")
}

// OnSystemCallEnd logs system call execution end
func (t *LoggerTracer) OnSystemCallEnd() {
	t.txLogger.Debug("System call executed")
	t.txLogger = nil
}

// OnEnter logs EVM message execution start
func (t *LoggerTracer) OnEnter(depth int, typ byte, from, to gethcommon.Address, input []byte, gas uint64, value *big.Int) {
	if value == nil {
		value = new(big.Int)
	}
	t.txLogger.WithFields(logrus.Fields{
		"msg.type":  gethvm.OpCode(typ).String(),
		"msg.depth": depth,
		"msg.from":  from.Hex(),
		"msg.to":    to.Hex(),
		"msg.input": hexutil.Encode(input),
		"msg.gas":   gas,
		"msg.value": hexutil.EncodeBig(value),
	}).Debug("Execute EVM message...")
}

// OnExit logs EVM message execution end
func (t *LoggerTracer) OnExit(depth int, output []byte, gasUsed uint64, err error, reverted bool) {
	logger := t.txLogger.WithFields(logrus.Fields{
		"msg.depth":    depth,
		"msg.output":   hexutil.Encode(output),
		"msg.gasUsed":  gasUsed,
		"msg.reverted": reverted,
	})
	if err != nil {
		logger.WithError(err).Error("Failed to process EVM message")
	} else {
		logger.Debug("EVM message executed")
	}
}

// OnOpcode logs opcode execution
func (t *LoggerTracer) OnOpcode(pc uint64, op byte, gas, cost uint64, _ tracing.OpContext, _ []byte, depth int, err error) {
	if err != nil {
		t.txLogger.WithFields(logrus.Fields{
			"pc":    pc,
			"op":    gethvm.OpCode(op).String(),
			"gas":   gas,
			"cost":  cost,
			"depth": depth,
		}).WithError(err).Error("Cannot execute opcode")
	}
}

// OnFault logs opcode execution fault
func (t *LoggerTracer) OnFault(pc uint64, op byte, gas, cost uint64, _ tracing.OpContext, depth int, err error) {
	t.txLogger.WithFields(logrus.Fields{
		"pc":    pc,
		"op":    gethvm.OpCode(op).String(),
		"gas":   gas,
		"cost":  cost,
		"depth": depth,
	}).WithError(err).Error("Failed to execute opcode")
}

// Hooks returns the logger tracer hooks
func (t *LoggerTracer) Hooks() *tracing.Hooks {
	return &tracing.Hooks{
		OnBlockStart:      t.OnBlockStart,
		OnBlockEnd:        t.OnBlockEnd,
		OnTxStart:         t.OnTxStart,
		OnTxEnd:           t.OnTxEnd,
		OnEnter:           t.OnEnter,
		OnExit:            t.OnExit,
		OnOpcode:          t.OnOpcode,
		OnFault:           t.OnFault,
		OnSystemCallStart: t.OnSystemCallStart,
		OnSystemCallEnd:   t.OnSystemCallEnd,
	}
}
