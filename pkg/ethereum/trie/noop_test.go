package trie

import (
	"testing"

	gethstate "github.com/ethereum/go-ethereum/core/state"
	"github.com/stretchr/testify/assert"
)

func TestNoOpTrieImplementsInterface(t *testing.T) {
	assert.Implements(t, (*gethstate.Trie)(nil), new(NoOpTrie))
}
