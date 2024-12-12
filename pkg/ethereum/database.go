package ethereum

import (
	"context"
	"fmt"

	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethdb"
)

type FetchHeader func(ctx context.Context, hash gethcommon.Hash) (*gethtypes.Header, error)

// FillDBWithAncestors fills an ethdb.Database with the headers of the ancestors of a block
// The ancestors headers are fetched from a distinct source (e.g. a remote node, an external database...)
// If an ancestor header is already present in the database, it is not fetched.
func FillDBWithAncestors(ctx context.Context, db ethdb.Database, fetcher FetchHeader, block *gethtypes.Block, count int) ([]*gethtypes.Header, error) {
	parentHash := block.ParentHash()
	parentNumber := block.NumberU64() - 1
	ancestors := make([]*gethtypes.Header, 0)
	var err error
	for i := 0; i < count; i++ {
		ancestor := rawdb.ReadHeader(db, parentHash, parentNumber)
		if ancestor == nil {
			ancestor, err = fetcher(ctx, parentHash)
			if err != nil {
				return nil, fmt.Errorf("failed to fetch ancestor header number=%d hash=%d: %w", parentNumber, parentHash, err)
			}

			if ancestor == nil {
				break
			}

			ancestors = append(ancestors, ancestor)
			rawdb.WriteHeader(db, ancestor)
		}

		parentHash = ancestor.ParentHash
		parentNumber--
	}
	return ancestors, nil
}

// WriteCodes fills an ethdb.Database with the provided bytecodes
func WriteCodes(db ethdb.Database, codes ...[]byte) {
	var (
		hasher = crypto.NewKeccakState()
		hash   = make([]byte, 32)
	)

	//nolint:errcheck // Can't fail
	for _, code := range codes {
		hasher.Reset()
		hasher.Write(code)
		hasher.Read(hash)

		rawdb.WriteCode(db, gethcommon.BytesToHash(hash), code)
	}
}

// WriteHeaders fills an ethdb.Database with the provided headers
func WriteHeaders(db ethdb.Database, headers ...*gethtypes.Header) {
	for _, header := range headers {
		rawdb.WriteHeader(db, header)
	}
}

// WriteNodesToHashDB fills an ethdb.Database with the provided nodes
func WriteNodesToHashDB(db ethdb.Database, nodes ...[]byte) {
	var (
		hasher = crypto.NewKeccakState()
		hash   = make([]byte, 32)
	)
	//nolint:errcheck // Can't fail
	for _, node := range nodes {
		hasher.Reset()
		hasher.Write(node)
		hasher.Read(hash)

		rawdb.WriteLegacyTrieNode(db, gethcommon.BytesToHash(hash), node)
	}
}
