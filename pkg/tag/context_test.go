package tag

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContextWithNamespaceTags(t *testing.T) {
	ctx := WithNamespaceTags(
		context.Background(),
		"test",
		Key("key1").String("value1"),
		Key("key2").String("value2"),
	)

	set := FromNamespaceContext(ctx, "test")
	assert.Equal(t, 2, len(set))
}

func TestContextWithTags(t *testing.T) {
	ctx := WithTags(
		context.Background(),
		Key("key1").String("value1"),
		Key("key2").String("value2"),
	)

	set := FromContext(ctx)
	assert.Equal(t, 2, len(set))
}

func TestContextWithComponent(t *testing.T) {
	ctx := WithComponent(context.Background(), "component1")

	set := FromContext(ctx)
	assert.Equal(t, 1, len(set))
	assert.Equal(t, "component", string(set[0].Key))
	assert.Equal(t, "component1", set[0].Value.Interface.(string))

	ctx = WithComponent(ctx, "component2")
	set = FromContext(ctx)
	assert.Equal(t, 1, len(set))
	assert.Equal(t, "component", string(set[0].Key))
	assert.Equal(t, "component1.component2", set[0].Value.Interface.(string))
}
