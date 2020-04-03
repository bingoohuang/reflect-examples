package gor

import (
	"reflect"
)

// IsEmpty returns if the object is considered as empty or not.
func IsEmpty(value interface{}) bool {
	if v, ok := value.(reflect.Value); ok {
		return IsEmptyValue(v)
	}

	return IsEmptyValue(reflect.ValueOf(value))
}

// IsEmptyValue returns if the object is considered as empty or not.
func IsEmptyValue(value reflect.Value) bool {
	switch value.Kind() {
	case reflect.String:
		return value.Len() == 0
	case reflect.Bool:
		return !value.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return value.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return value.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return value.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return value.IsNil()
	}

	return reflect.DeepEqual(value.Interface(), reflect.Zero(value.Type()).Interface())
}

// IndirectAll returns the value that v points to.
// If v is a nil pointer, Indirect returns a zero Value.
// If v is not a pointer, Indirect returns v.
func IndirectAll(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Ptr && v.Elem().IsValid() {
		v = v.Elem()
	}

	return v
}

// 参考 https://github.com/uber-go/dig/blob/master/types.go
// nolint gochecknoglobals
var (
	// ErrType defines the error's type
	ErrType = reflect.TypeOf((*error)(nil)).Elem()
)

// ImplType tells src whether it implements target type.
func ImplType(src, target reflect.Type) bool {
	if src == target || src.Kind() == reflect.Ptr && src.Elem() == target {
		return true
	}

	if target.Kind() != reflect.Interface {
		return false
	}

	return reflect.PtrTo(src).Implements(target)
}

// IsError tells t whether it is error type exactly.
func IsError(t reflect.Type) bool { return t == ErrType }

// AsError tells t whether it implements error type exactly.
func AsError(t reflect.Type) bool { return ImplType(t, ErrType) }

// V returns the variadic arguments to slice.
func V(v ...interface{}) []interface{} {
	return v
}

// V0 returns the one of variadic arguments at index 0.
func V0(v ...interface{}) interface{} {
	return v[0]
}
