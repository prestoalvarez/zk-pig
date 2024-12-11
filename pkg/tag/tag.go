package tag

// A tag is a key-value pair that can be attached to a context
// It is used to carry metadata about a given context execution along the context chain
// It is primarily used for rich instrumentation of the application: logging, metrics, tracing
type Tag struct {
	Key     Key
	Value   Value
	chained *bool
}

// Chained flags the tag as chained.
func (t *Tag) Chained(b bool) *Tag {
	t.chained = &b
	return t
}

// Copy returns a copy of the tag.
func (t *Tag) Copy() *Tag {
	return &Tag{
		Key:     t.Key,
		Value:   t.Value,
		chained: t.chained,
	}
}

// BoolTag creates a new tag with a boolean value.
func BoolTag(key string, value bool) *Tag {
	return Key(key).Bool(value)
}

// Int64Tag creates a new tag with a 64-bit signed integral value.
func Int64Tag(key string, value int64) *Tag {
	return Key(key).Int64(value)
}

// Float64Tag creates a new tag with a 64-bit floating point value.
func Float64Tag(key string, value float64) *Tag {
	return Key(key).Float64(value)
}

// StringTag creates a new tag with a string value.
func StringTag(key, value string) *Tag {
	return Key(key).String(value)
}

// ObjectTag creates a new tag with a generic object value.
func ObjectTag(key string, value interface{}) *Tag {
	return Key(key).Object(value)
}
