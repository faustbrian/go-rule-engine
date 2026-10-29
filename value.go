package ruleengine

import "time"

// Kind is the exact runtime type of a Value.
type Kind uint8

const (
	// KindMissing represents an absent path rather than a supplied value.
	KindMissing Kind = iota
	// KindNull represents an explicitly supplied null.
	KindNull
	// KindBool represents a boolean.
	KindBool
	// KindInt represents a signed 64-bit integer.
	KindInt
	// KindFloat represents a finite 64-bit floating-point number.
	KindFloat
	// KindString represents valid UTF-8 text.
	KindString
	// KindTime represents an instant with no monotonic clock reading.
	// Core ordering uses UnixNano and is limited to its representable range.
	KindTime
	// KindDuration represents a time duration.
	KindDuration
	// KindList represents an ordered immutable list of values.
	KindList
	// KindDecimal represents an exact base-10 number that retains its original
	// RFC 8259 lexical text. It is appended after KindList so earlier kind
	// numbers are unchanged.
	KindDecimal
)

// maxKind is the largest kind accepted by contexts, compilation and signatures.
const maxKind = KindDecimal

// Value is a typed fact or operand value; its zero value is Missing. Constructors
// do not validate service budgets. Context construction and compilation validate
// supplied values. A Value is immutable: List copies the caller's slice and
// ListValue and Interface return copies, so list storage, including nested
// lists, is never reachable by callers and is shared rather than recopied.
type Value struct {
	kind Kind
	data any
}

// Missing returns the absent-path sentinel.
func Missing() Value { return Value{kind: KindMissing} }

// Null returns an explicit null value.
func Null() Value { return Value{kind: KindNull} }

// Bool returns a boolean value.
func Bool(value bool) Value { return Value{kind: KindBool, data: value} }

// Int returns a signed integer value.
func Int(value int64) Value { return Value{kind: KindInt, data: value} }

// Float returns a floating-point value; contexts reject non-finite values.
func Float(value float64) Value { return Value{kind: KindFloat, data: value} }

// String returns a string value.
func String(value string) Value { return Value{kind: KindString, data: value} }

// Time returns a time value with its monotonic reading removed.
func Time(value time.Time) Value { return Value{kind: KindTime, data: value.Round(0)} }

// Duration returns a duration value.
func Duration(value time.Duration) Value { return Value{kind: KindDuration, data: value} }

// List copies its input slice and returns a collection value. Elements are
// immutable Values, so nested lists share their storage instead of being
// copied again.
func List(values ...Value) Value {
	items := make([]Value, len(values))
	copy(items, values)
	return Value{kind: KindList, data: items}
}

// Kind returns the exact value kind.
func (v Value) Kind() Kind { return v.kind }

// Interface returns the scalar value or a copied []Value for a list. A
// decimal returns its original lexical text as a string.
func (v Value) Interface() any {
	if v.kind == KindList {
		values, _ := v.ListValue()
		return values
	}
	if text, ok := v.DecimalText(); ok {
		return text
	}
	return v.data
}

// BoolValue returns the boolean and whether the kind matches.
func (v Value) BoolValue() (bool, bool) {
	value, ok := v.data.(bool)
	return value, ok && v.kind == KindBool
}

// IntValue returns the integer and whether the kind matches.
func (v Value) IntValue() (int64, bool) {
	value, ok := v.data.(int64)
	return value, ok && v.kind == KindInt
}

// FloatValue returns the float and whether the kind matches.
func (v Value) FloatValue() (float64, bool) {
	value, ok := v.data.(float64)
	return value, ok && v.kind == KindFloat
}

// StringValue returns the string and whether the kind matches.
func (v Value) StringValue() (string, bool) {
	value, ok := v.data.(string)
	return value, ok && v.kind == KindString
}

// TimeValue returns the time and whether the kind matches.
func (v Value) TimeValue() (time.Time, bool) {
	value, ok := v.data.(time.Time)
	return value, ok && v.kind == KindTime
}

// DurationValue returns the duration and whether the kind matches.
func (v Value) DurationValue() (time.Duration, bool) {
	value, ok := v.data.(time.Duration)
	return value, ok && v.kind == KindDuration
}

// ListValue returns a copy of the elements and whether the kind matches.
// Mutating the returned slice does not change v; nested list elements are
// immutable Values and need no copy.
func (v Value) ListValue() ([]Value, bool) {
	values, ok := v.listItems()
	if !ok {
		return nil, false
	}
	return append(make([]Value, 0, len(values)), values...), true
}

// Equal reports whether v and other have the same kind and value. For two
// operands of the same non-Missing kind it matches OpEqual: times compare as
// instants, decimals by exact numeric value regardless of spelling, and lists
// element by element. Unlike OpEqual, values of different kinds are simply
// unequal rather than a type mismatch, and Missing equals Missing where
// OpEqual reports false. A NaN float, which contexts reject, equals nothing.
// Equal does not allocate.
func (v Value) Equal(other Value) bool { return valuesEqual(v, other) }

// listItems returns the shared immutable list storage without copying. Callers
// inside the package must not modify or expose the returned slice.
func (v Value) listItems() ([]Value, bool) {
	values, ok := v.data.([]Value)
	return values, ok && v.kind == KindList
}
