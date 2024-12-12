// Copyright 2024 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package state

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	gethstate "github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/ethereum/go-ethereum/trie/utils"
	"github.com/ethereum/go-ethereum/triedb"
	"github.com/ethereum/go-ethereum/triedb/database"
	"github.com/holiman/uint256"
	"github.com/kkrt-labs/kakarot-controller/pkg/ethereum/rpc"
)

// stateReader wraps a database state reader.
type stateReader struct {
	reader database.StateReader
	buff   crypto.KeccakState
}

// newStateReader constructs a state reader with on the given state root.
func newStateReader(reader database.StateReader) *stateReader {
	return &stateReader{
		reader: reader,
		buff:   crypto.NewKeccakState(),
	}
}

// Account implements Reader, retrieving the account specified by the address.
//
// An error will be returned if the associated snapshot is already stale or
// the requested account is not yet covered by the snapshot.
//
// The returned account might be nil if it's not existent.
func (r *stateReader) Account(addr common.Address) (*types.StateAccount, error) {
	account, err := r.reader.Account(crypto.HashData(r.buff, addr.Bytes()))
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, nil
	}
	acct := &types.StateAccount{
		Nonce:    account.Nonce,
		Balance:  account.Balance,
		CodeHash: account.CodeHash,
		Root:     common.BytesToHash(account.Root),
	}
	if len(acct.CodeHash) == 0 {
		acct.CodeHash = types.EmptyCodeHash.Bytes()
	}
	if acct.Root == (common.Hash{}) {
		acct.Root = types.EmptyRootHash
	}
	return acct, nil
}

// Storage implements Reader, retrieving the storage slot specified by the
// address and slot key.
//
// An error will be returned if the associated snapshot is already stale or
// the requested storage slot is not yet covered by the snapshot.
//
// The returned storage slot might be empty if it's not existent.
func (r *stateReader) Storage(addr common.Address, key common.Hash) (common.Hash, error) {
	addrHash := crypto.HashData(r.buff, addr.Bytes())
	slotHash := crypto.HashData(r.buff, key.Bytes())
	ret, err := r.reader.Storage(addrHash, slotHash)
	if err != nil {
		return common.Hash{}, err
	}
	if len(ret) == 0 {
		return common.Hash{}, nil
	}
	// Perform the rlp-decode as the slot value is RLP-encoded in the state
	// snapshot.
	_, content, _, err := rlp.Split(ret)
	if err != nil {
		return common.Hash{}, err
	}
	var value common.Hash
	value.SetBytes(content)
	return value, nil
}

// Copy implements Reader, returning a deep-copied snap reader.
func (r *stateReader) Copy() gethstate.Reader {
	return &stateReader{
		reader: r.reader,
		buff:   crypto.NewKeccakState(),
	}
}

// trieReader implements the Reader interface, providing functions to access
// state from the referenced trie.
type trieReader struct {
	root     common.Hash                       // State root which uniquely represent a state
	db       *triedb.Database                  // Database for loading trie
	buff     crypto.KeccakState                // Buffer for keccak256 hashing
	mainTrie gethstate.Trie                    // Main trie, resolved in constructor
	subRoots map[common.Address]common.Hash    // Set of storage roots, cached when the account is resolved
	subTries map[common.Address]gethstate.Trie // Group of storage tries, cached when it's resolved
}

// trieReader constructs a trie reader of the specific state. An error will be
// returned if the associated trie specified by root is not existent.
func newTrieReader(root common.Hash, db *triedb.Database, cache *utils.PointCache) (*trieReader, error) {
	var (
		tr  gethstate.Trie
		err error
	)
	if !db.IsVerkle() {
		tr, err = trie.NewStateTrie(trie.StateTrieID(root), db)
	} else {
		tr, err = trie.NewVerkleTrie(root, db, cache)
	}
	if err != nil {
		return nil, err
	}
	return &trieReader{
		root:     root,
		db:       db,
		buff:     crypto.NewKeccakState(),
		mainTrie: tr,
		subRoots: make(map[common.Address]common.Hash),
		subTries: make(map[common.Address]gethstate.Trie),
	}, nil
}

// Account implements Reader, retrieving the account specified by the address.
//
// An error will be returned if the trie state is corrupted. An nil account
// will be returned if it's not existent in the trie.
func (r *trieReader) Account(addr common.Address) (*types.StateAccount, error) {
	account, err := r.mainTrie.GetAccount(addr)
	if err != nil {
		return nil, err
	}
	if account == nil {
		r.subRoots[addr] = types.EmptyRootHash
	} else {
		r.subRoots[addr] = account.Root
	}
	return account, nil
}

// Storage implements Reader, retrieving the storage slot specified by the
// address and slot key.
//
// An error will be returned if the trie state is corrupted. An empty storage
// slot will be returned if it's not existent in the trie.
func (r *trieReader) Storage(addr common.Address, key common.Hash) (common.Hash, error) {
	var (
		tr    gethstate.Trie
		found bool
		value common.Hash
	)
	if r.db.IsVerkle() {
		tr = r.mainTrie
	} else {
		tr, found = r.subTries[addr]
		if !found {
			root, ok := r.subRoots[addr]

			// The storage slot is accessed without account caching. It's unexpected
			// behavior but try to resolve the account first anyway.
			if !ok {
				_, err := r.Account(addr)
				if err != nil {
					return common.Hash{}, err
				}
				root = r.subRoots[addr]
			}
			var err error
			tr, err = trie.NewStateTrie(trie.StorageTrieID(r.root, crypto.HashData(r.buff, addr.Bytes()), root), r.db)
			if err != nil {
				return common.Hash{}, err
			}
			r.subTries[addr] = tr
		}
	}
	ret, err := tr.GetStorage(addr, key.Bytes())
	if err != nil {
		return common.Hash{}, err
	}
	value.SetBytes(ret)
	return value, nil
}

// Copy implements Reader, returning a deep-copied trie reader.
func (r *trieReader) Copy() gethstate.Reader {
	tries := make(map[common.Address]gethstate.Trie)
	for addr, tr := range r.subTries {
		tries[addr] = mustCopyTrie(tr)
	}
	return &trieReader{
		root:     r.root,
		db:       r.db,
		buff:     crypto.NewKeccakState(),
		mainTrie: mustCopyTrie(r.mainTrie),
		subRoots: maps.Clone(r.subRoots),
		subTries: tries,
	}
}

// multiReader is the aggregation of a list of Reader interface, providing state
// access by leveraging all readers. The checking priority is determined by the
// position in the reader list.
type multiReader struct {
	readers []gethstate.Reader // List of readers, sorted by checking priority
}

// newMultiReader constructs a multiReader instance with the given readers. The
// priority among readers is assumed to be sorted. Note, it must contain at least
// one reader for constructing a multiReader.
func newMultiReader(readers ...gethstate.Reader) (*multiReader, error) {
	if len(readers) == 0 {
		return nil, errors.New("empty reader set")
	}
	return &multiReader{
		readers: readers,
	}, nil
}

// Account implementing Reader interface, retrieving the account associated with
// a particular address.
//
// - Returns a nil account if it does not exist
// - Returns an error only if an unexpected issue occurs
// - The returned account is safe to modify after the call
func (r *multiReader) Account(addr common.Address) (*types.StateAccount, error) {
	var errs []error
	for _, reader := range r.readers {
		acct, err := reader.Account(addr)
		if err == nil {
			return acct, nil
		}
		errs = append(errs, err)
	}
	return nil, errors.Join(errs...)
}

// Storage implementing Reader interface, retrieving the storage slot associated
// with a particular account address and slot key.
//
// - Returns an empty slot if it does not exist
// - Returns an error only if an unexpected issue occurs
// - The returned storage slot is safe to modify after the call
func (r *multiReader) Storage(addr common.Address, slot common.Hash) (common.Hash, error) {
	var errs []error
	for _, reader := range r.readers {
		slot, err := reader.Storage(addr, slot)
		if err == nil {
			return slot, nil
		}
		errs = append(errs, err)
	}
	return common.Hash{}, errors.Join(errs...)
}

// Copy implementing Reader interface, returning a deep-copied state reader.
func (r *multiReader) Copy() gethstate.Reader {
	var readers []gethstate.Reader
	for _, reader := range r.readers {
		readers = append(readers, reader.Copy())
	}
	return &multiReader{readers: readers}
}

// --- Below this line is the code that was not present in the original goethereum file ---

// rpcReader is a state reader that retrieves state information from a remote node (typically an archive node).
// it provides the ability to read account and storage information from a remote node
// it is useful when the local node does not have a full state trie
//
// Note:
//   - rpcReader needs a blockNumber and corresponding state root to retrieve the state information
//   - in case a block is not final and a re-org happens, the rpcReader may lead to inconsistent state information. If using
//     rpcReader on a non-finalized blocks, it is recommended to control the validity of the computed state information, with the finalized block
//     (in particular verify that the state root is the same as the one in the finalized block header)
type rpcReader struct {
	remote rpc.Client // Remote client to retrieve state information from remote node

	blockNumber *big.Int    // Block number to retrieve state information
	root        common.Hash // State root corresponding to the block number (it is assumed that the state root for the given block does not change (i.e. no re-org))
}

// Account implementing Reader interface, retrieving the account associated with
// a particular address.
//
// - Returns a nil account if account is not available in the remote node
// - Returns an error only if remote node returns an error
// - The returned account is safe to modify after the call
func (r *rpcReader) Account(addr common.Address) (*gethtypes.StateAccount, error) {
	account, err := r.remote.GetProof(context.Background(), addr, nil, r.blockNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get proof for address %s and block %v: %v", addr.Hex(), r.blockNumber, err)
	}

	if account == nil {
		return nil, nil
	}

	balance, hasOverflowed := uint256.FromBig(account.Balance)
	if hasOverflowed {
		return nil, fmt.Errorf("failed to convert balance %v to uint256", account.Balance)
	}

	return &gethtypes.StateAccount{
		Nonce:    account.Nonce,
		Balance:  balance,
		Root:     account.StorageHash,
		CodeHash: account.CodeHash.Bytes(),
	}, nil
}

// Storage implementing Reader interface, retrieving the storage slot associated
// with a particular account address and slot key.
//
// - Returns an empty slot if it does not exist
// - Returns an error only if an unexpected issue occurs
// - The returned storage slot is safe to modify after the call
func (r *rpcReader) Storage(addr common.Address, slot common.Hash) (common.Hash, error) {
	value, err := r.remote.StorageAt(context.Background(), addr, slot, r.blockNumber)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to get storage slot for address %s and slot %s and block %v: %v", addr.Hex(), slot.Hex(), r.blockNumber, err)
	}
	return common.BytesToHash(value), nil
}

// Copy implementing Reader interface, returning a deep-copied state reader.
func (r *rpcReader) Copy() gethstate.Reader {
	return &rpcReader{
		blockNumber: r.blockNumber,
		remote:      r.remote,
		root:        r.root,
	}
}
