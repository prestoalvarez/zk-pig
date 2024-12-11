package jsonrpc

import (
	"encoding/json"
	"math/big"
	"os"
	"testing"

	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRPCHeader(t *testing.T) {
	header := &gethtypes.Header{
		ParentHash:  gethcommon.Hash{},
		UncleHash:   gethcommon.HexToHash("0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347"),
		ReceiptHash: gethcommon.HexToHash("0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421"),
		TxHash:      gethcommon.HexToHash("0x661a9febcfa8f1890af549b874faf9fa274aede26ef489d9db0b25daa569450e"),
		Number:      big.NewInt(100),
		GasLimit:    1000,
		GasUsed:     10000,
		Time:        1000,
		Difficulty:  big.NewInt(100),
		Extra:       hexutil.MustDecode("0xabcd"),
		Bloom:       gethtypes.Bloom{},
		Coinbase:    gethcommon.HexToAddress("0x5E0c840B55d49377070e131ed93B69393218C444"),
		MixDigest:   gethcommon.Hash{},
		Nonce:       gethtypes.EncodeNonce(25),
		Root:        gethcommon.Hash{},
	}

	rpcHeader := new(Header).FromHeader(header)
	assert.Equal(t, header.Hash(), rpcHeader.Hash)

	b, err := json.Marshal(rpcHeader)
	require.NoError(t, err)

	rpcHeader2 := new(Header)
	err = json.Unmarshal(b, rpcHeader2)
	require.NoError(t, err)

	header2 := rpcHeader2.Header()
	assert.Equal(t, header.Hash(), header2.Hash())
}

func TestRPCBlock(t *testing.T) {
	f, err := os.Open("testdata/block.json")
	require.NoError(t, err)
	defer f.Close()

	rpcBlock := new(Block)
	err = json.NewDecoder(f).Decode(rpcBlock)
	require.NoError(t, err)

	block := rpcBlock.Block()
	assert.Equal(t, rpcBlock.Header.Hash, block.Hash())

	b, err := json.Marshal(rpcBlock)
	require.NoError(t, err)

	rpcBlock2 := new(Block)
	err = json.Unmarshal(b, rpcBlock2)
	require.NoError(t, err)

	block2 := rpcBlock2.Block()
	assert.Equal(t, block.Hash(), block2.Hash())
}
