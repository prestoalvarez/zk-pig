package tag

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithTags(t *testing.T) {
	set := EmptySet.WithTags(
		Key("key1").String("value1.0"),
		Key("key2").String("value2.0"),
	)

	require.Equal(t, 2, len(set))

	newSet := set.WithTags(
		Key("key1").String("value1.1"),
		Key("key3").String("value3.0"),
	)

	require.Equal(t, 3, len(newSet))
	assert.Equal(t, "value1.1", newSet[0].Value.Interface)
	assert.Equal(t, "value2.0", newSet[1].Value.Interface)
	assert.Equal(t, "value3.0", newSet[2].Value.Interface)

	require.Len(t, set, 2)
	assert.Equal(t, "value1.0", set[0].Value.Interface)
	assert.Equal(t, "value2.0", set[1].Value.Interface)
}
