package rpc

import (
	"context"
	"math/big"

	geth "github.com/ethereum/go-ethereum"
	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient/gethclient"
)

// TODO: the mockgen command does not set the imports as expected in the output file
//go:generate mockgen -source client.go -destination mock/client.go -package mock -imports geth=github.com/ethereum/go-ethereum,gethcommon=github.com/ethereum/go-ethereum/common,gethtypes=github.com/ethereum/go-ethereum/core/types Client

type Client interface {
	geth.ChainReader
	geth.ChainStateReader
	geth.ChainSyncReader
	geth.ContractCaller
	geth.PendingContractCaller
	geth.LogFilterer
	geth.TransactionReader
	geth.TransactionSender
	geth.GasPricer
	geth.GasPricer1559
	geth.FeeHistoryReader
	geth.PendingStateReader
	geth.GasEstimator
	geth.BlockNumberReader
	geth.ChainIDReader

	NetworkID(ctx context.Context) (*big.Int, error)
	GetProof(ctx context.Context, account gethcommon.Address, keys []string, blockNumber *big.Int) (*gethclient.AccountResult, error)
}
