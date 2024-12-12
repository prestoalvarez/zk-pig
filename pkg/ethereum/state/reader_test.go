package state

import (
	"testing"

	gethstate "github.com/ethereum/go-ethereum/core/state"
	"github.com/stretchr/testify/assert"
)

func TestReadersImplementInterface(t *testing.T) {
	assert.Implements(t, (*gethstate.Reader)(nil), new(rpcReader))
}
