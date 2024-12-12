package trie

import (
	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/ethereum/go-ethereum/trie/trienode"
)

// NoOpTrie is a gethstate.Trie that does nothing.
type NoOpTrie struct {
	isVerkle bool
}

func NewNoOpTrie(isVerkle bool) *NoOpTrie {
	return &NoOpTrie{isVerkle: isVerkle}
}

func (t *NoOpTrie) GetKey([]byte) []byte { return nil }
func (t *NoOpTrie) GetAccount(_ gethcommon.Address) (*types.StateAccount, error) {
	return nil, nil
}
func (t *NoOpTrie) GetStorage(_ gethcommon.Address, _ []byte) ([]byte, error) { return nil, nil }
func (t *NoOpTrie) UpdateAccount(_ gethcommon.Address, _ *types.StateAccount, _ int) error {
	return nil
}
func (t *NoOpTrie) UpdateStorage(_ gethcommon.Address, _, _ []byte) error { return nil }
func (t *NoOpTrie) DeleteAccount(_ gethcommon.Address) error              { return nil }
func (t *NoOpTrie) DeleteStorage(_ gethcommon.Address, _ []byte) error    { return nil }
func (t *NoOpTrie) UpdateContractCode(_ gethcommon.Address, _ gethcommon.Hash, _ []byte) error {
	return nil
}
func (t *NoOpTrie) Hash() gethcommon.Hash { return gethcommon.Hash{} }
func (t *NoOpTrie) Commit(_ bool) (gethcommon.Hash, *trienode.NodeSet) {
	return gethcommon.Hash{}, nil
}
func (t *NoOpTrie) Witness() map[string]struct{}                     { return nil }
func (t *NoOpTrie) NodeIterator(_ []byte) (trie.NodeIterator, error) { return nil, nil }
func (t *NoOpTrie) Prove(_ []byte, _ ethdb.KeyValueWriter) error     { return nil }
func (t *NoOpTrie) IsVerkle() bool                                   { return t.isVerkle }
