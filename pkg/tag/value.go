package tag

// Type describes the type of the Value holds.
type Type int

const (
	// INVALID is used for a Value with no value set.
	INVALID Type = iota
	// BOOL is a boolean Type Value.
	BOOL
	// INT64 is a 64-bit signed integral Type Value.
	INT64
	// FLOAT64 is a 64-bit floating point Type Value.
	FLOAT64
	// STRING is a string Type Value.
	STRING
	// OBJECT is a generic object Type Value.
	OBJECT
)

// Value represents a tag value.
type Value struct {
	Type      Type
	Interface interface{}
}

// InvalidValue returns a Value with no value set.
func InvalidValue() Value {
	return Value{Type: INVALID}
}

// BoolValue returns a Value with a boolean value set.
func BoolValue(v bool) Value {
	return Value{Type: BOOL, Interface: v}
}

// Int64Value returns a Value with a 64-bit signed integral value set.
func Int64Value(v int64) Value {
	return Value{Type: INT64, Interface: v}
}

// Float64Value returns a Value with a 64-bit floating point value set.
func Float64Value(v float64) Value {
	return Value{Type: FLOAT64, Interface: v}
}

// StringValue returns a Value with a string value set.
func StringValue(v string) Value {
	return Value{Type: STRING, Interface: v}
}

// ObjectValue returns a Value with a generic object value set.
func ObjectValue(v interface{}) Value {
	return Value{Type: OBJECT, Interface: v}
}
