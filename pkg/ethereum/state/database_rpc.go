package state

import (
	"context"
	"fmt"
	"math/big"

	gethcommon "github.com/ethereum/go-ethereum/common"
	gethstate "github.com/ethereum/go-ethereum/core/state"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/trie/utils"
	"github.com/kkrt-labs/kakarot-controller/pkg/ethereum/rpc"
	"github.com/kkrt-labs/kakarot-controller/pkg/ethereum/trie"
)

// RPCDatabase is a gethstate.Database that reads the state from a remote RPC node.
type RPCDatabase struct {
	gethstate.Database

	remote                 rpc.Client
	stateRootToBlockNumber map[gethcommon.Hash]*big.Int
	currentBlockNumber     *big.Int

	pointCache *utils.PointCache
}

// NewRPCDatabase creates a new state database that reads the state from a remote RPC node.
func NewRPCDatabase(db gethstate.Database, remote rpc.Client) *RPCDatabase {
	return &RPCDatabase{
		Database:               db,
		remote:                 remote,
		stateRootToBlockNumber: make(map[gethcommon.Hash]*big.Int),
	}
}

// MarkBlock records a mapping from state root to the corresponding block number.
// This is necessary as the underlying RPC node expects parameters to be block numbers and not a state root.
func (db *RPCDatabase) MarkBlock(header *gethtypes.Header) {
	db.stateRootToBlockNumber[header.Root] = header.Number
	db.currentBlockNumber = header.Number
}

func (db *RPCDatabase) getBlockNumber(stateRoot gethcommon.Hash) (*big.Int, error) {
	if blockNumber, ok := db.stateRootToBlockNumber[stateRoot]; ok {
		return blockNumber, nil
	}
	return nil, fmt.Errorf("missing block for state root %s", stateRoot.Hex())
}

// Reader implements the gethstate.Database interface.
func (db *RPCDatabase) Reader(root gethcommon.Hash) (gethstate.Reader, error) {
	blockNumber, err := db.getBlockNumber(root)
	if err != nil {
		return nil, err
	}

	// This is the reader that reads from the remote node.
	return &rpcReader{
		remote:      db.remote,
		blockNumber: blockNumber,
		root:        root,
	}, nil
}

// OpenTrie implements the gethstate.Database interface.
func (db *RPCDatabase) OpenTrie(root gethcommon.Hash) (gethstate.Trie, error) {
	if tr, err := db.Database.OpenTrie(root); err == nil {
		return tr, nil
	}
	// We return a no-op trie to avoid some errors on block execution.
	// But it should be treated as suched and not used for any state access.
	return trie.NewNoOpTrie(false), nil
}

// OpenStorageTrie implements the gethstate.Database interface.
func (db *RPCDatabase) OpenStorageTrie(stateRoot gethcommon.Hash, address gethcommon.Address, root gethcommon.Hash, tr gethstate.Trie) (gethstate.Trie, error) {
	if tr, err := db.Database.OpenStorageTrie(stateRoot, address, root, tr); err == nil {
		return tr, nil
	}

	// We return a no-op trie to avoid some errors on block execution.
	// But it should be treated as suched and not used for any state access.
	return trie.NewNoOpTrie(false), nil
}

// ContractCode implements the gethstate.Database interface.
func (db *RPCDatabase) ContractCode(addr gethcommon.Address, _ gethcommon.Hash) ([]byte, error) {
	code, err := db.remote.CodeAt(context.Background(), addr, db.currentBlockNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get code for address %s: %v", addr.Hex(), err)
	}

	return code, nil
}

// ContractCodeSize implements the gethstate.Database interface.
func (db *RPCDatabase) ContractCodeSize(addr gethcommon.Address, codeHash gethcommon.Hash) (int, error) {
	code, err := db.ContractCode(addr, codeHash)
	return len(code), err
}
