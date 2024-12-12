package ethereum

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/core/rawdb"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	rpcmock "github.com/kkrt-labs/kakarot-controller/pkg/ethereum/rpc/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestFillDBWithAncestors(t *testing.T) {
	// Generate 50 fake ancestor headers
	count := 50
	blockNumber := big.NewInt(260)
	fakeAncestors := make([]*gethtypes.Header, count)
	ancestorHash := crypto.Keccak256Hash([]byte("first-ancestor"))
	for i := (count - 1); i >= 0; i-- {
		fakeAncestors[i] = &gethtypes.Header{
			Number:     big.NewInt(int64(blockNumber.Uint64() - uint64(i) - 1)),
			Root:       crypto.Keccak256Hash([]byte{byte(i)}),
			ParentHash: ancestorHash,
		}
		ancestorHash = fakeAncestors[i].Hash()
	}

	// Generate block with the latest ancestor as parent
	block := gethtypes.NewBlockWithHeader(&gethtypes.Header{
		Number:     blockNumber,
		Root:       crypto.Keccak256Hash([]byte{byte(count)}),
		ParentHash: fakeAncestors[0].Hash(),
	})

	// Prepare a mock RPC client with the fake ancestors
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	remote := rpcmock.NewMockClient(ctrl)
	for i := 0; i < count; i++ {
		remote.EXPECT().
			HeaderByHash(gomock.Any(), fakeAncestors[i].Hash()).
			Return(fakeAncestors[i], nil)
	}

	// Execute hook
	db := rawdb.NewMemoryDatabase()
	ancestors, err := FillDBWithAncestors(context.Background(), db, remote.HeaderByHash, block, count)
	require.NoError(t, err)

	// Verify that the ancestors were loaded to chain correctly
	parentHash := block.ParentHash()
	parentNumber := block.NumberU64() - 1
	for i := 0; i < count; i++ {
		ancestor := rawdb.ReadHeader(db, parentHash, parentNumber)
		require.NotNil(t, ancestor, "Expected ancestor %v to be non nil", i)
		assert.Equal(t, fakeAncestors[i].Root, ancestor.Root, "Expected ancestor %v root to be correct", i)
		parentHash = ancestor.ParentHash
		parentNumber--
	}

	// Verify that the return ancestors are correct
	assert.Len(t, ancestors, count, "Expected 256 ancestors to be loaded")
	for i, ancestor := range ancestors {
		assert.Equal(t, fakeAncestors[i], ancestor, "Expected ancestor %v to be correct", i)
	}
}

func TestFillDBWithBytecode(t *testing.T) {
	db := rawdb.NewMemoryDatabase()
	codes := [][]byte{
		[]byte("code1"),
		[]byte("code2"),
	}

	// Fill the database with the bytecodes
	WriteCodes(db, codes...)

	code1 := rawdb.ReadCode(db, crypto.Keccak256Hash(codes[0]))
	assert.Equal(t, codes[0], code1, "Expected code1 to be correct")
	code2 := rawdb.ReadCode(db, crypto.Keccak256Hash(codes[1]))
	assert.Equal(t, codes[1], code2, "Expected code2 to be correct")
}
