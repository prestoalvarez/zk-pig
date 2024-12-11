package tag

// Key represents the key part in key-value pairs. It's a string. The
// allowed character set in the key depends on the use of the key.
type Key string

// Bool creates a new tag with a boolean value.
func (k Key) Bool(b bool) *Tag {
	return &Tag{Key: k, Value: BoolValue(b)}
}

// Int64 creates a new tag with a 64-bit signed integral value.
func (k Key) Int64(i int64) *Tag {
	return &Tag{Key: k, Value: Int64Value(i)}
}

// Float64 creates a new tag with a 64-bit floating point value.
func (k Key) Float64(f float64) *Tag {
	return &Tag{Key: k, Value: Float64Value(f)}
}

// String creates a new tag with a string value.
func (k Key) String(s string) *Tag {
	return &Tag{Key: k, Value: StringValue(s)}
}

// Object creates a new tag with a generic object value.
func (k Key) Object(o interface{}) *Tag {
	return &Tag{Key: k, Value: ObjectValue(o)}
}
