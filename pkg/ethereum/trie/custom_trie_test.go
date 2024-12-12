package trie

import (
	"testing"

	gethstate "github.com/ethereum/go-ethereum/core/state"
	"github.com/stretchr/testify/assert"
)

func TestStateTrieImplementsTrieInterface(t *testing.T) {
	assert.Implements(t, (*gethstate.Trie)(nil), new(StateTrie))
}
