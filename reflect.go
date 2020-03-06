package goreflect

import "reflect"

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
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	return v
}