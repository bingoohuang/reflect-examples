package defaults

import (
	"encoding/json"
	"errors"
	"reflect"
	"strconv"
	"time"

	"github.com/bingoohuang/gor"
)

// ErrInvalidType is the error for non-struct pointer
var ErrInvalidType = errors.New("not a struct pointer")

// Option is the options for Validate.
type Option struct {
	TagName string
}

// OptionFn is the function prototype to apply option
type OptionFn func(*Option)

// TagName defines the tag name for validate.
func TagName(tagName string) OptionFn { return func(o *Option) { o.TagName = tagName } }

// Set initializes members in a struct referenced by a pointer.
// Maps and slices are initialized by `make` and other primitive types are set with default values.
// `ptr` should be a struct pointer
func Set(ptr interface{}, optionFns ...OptionFn) error {
	if reflect.TypeOf(ptr).Kind() != reflect.Ptr {
		return ErrInvalidType
	}

	option := createOption(optionFns)

	v := reflect.ValueOf(ptr).Elem()
	t := v.Type()

	if t.Kind() != reflect.Struct {
		return ErrInvalidType
	}

	for i := 0; i < t.NumField(); i++ {
		defaultVal := t.Field(i).Tag.Get(option.TagName)
		if defaultVal == "-" {
			continue
		}

		if err := setField(v.Field(i), defaultVal); err != nil {
			return err
		}
	}

	return nil
}

func createOption(optionFns []OptionFn) *Option {
	option := &Option{}

	for _, fn := range optionFns {
		fn(option)
	}

	if option.TagName == "" {
		option.TagName = "default"
	}

	return option
}

func setField(field reflect.Value, v string) error {
	if !field.CanSet() {
		return nil
	}

	if !shouldInitializeField(field, v) {
		return nil
	}

	if hasInitialValue(field) {
		err := setZeroField(field, v)
		if err != nil {
			return err
		}
	}

	switch field.Kind() {
	case reflect.Ptr:
		if err := setField(field.Elem(), v); err != nil {
			return err
		}

		callSetter(field.Interface())
	case reflect.Struct:
		ref := reflect.New(field.Type())
		ref.Elem().Set(field)

		if err := Set(ref.Interface()); err != nil {
			return err
		}

		callSetter(ref.Interface())
		field.Set(ref.Elem())
	case reflect.Slice:
		for j := 0; j < field.Len(); j++ {
			if err := setField(field.Index(j), v); err != nil {
				return err
			}
		}
	}

	return nil
}

func setZeroField(field reflect.Value, v string) error {
	m := map[reflect.Kind]converterFn{
		reflect.Bool:    convertBool,
		reflect.Int:     convertInt,
		reflect.Int8:    convertInt8,
		reflect.Int16:   convertInt16,
		reflect.Int32:   convertInt32,
		reflect.Int64:   convertInt64,
		reflect.Uint:    convertUInt,
		reflect.Uint8:   convertUInt8,
		reflect.Uint16:  convertUInt16,
		reflect.Uint32:  convertUInt32,
		reflect.Uint64:  convertUInt64,
		reflect.Uintptr: convertUintptr,
		reflect.Float32: convertFloat32,
		reflect.Float64: convertFloat64,
		reflect.String:  convertString,
		reflect.Slice:   convertSlice,
		reflect.Map:     convertMap,
		reflect.Struct:  convertStruct,
		reflect.Ptr:     convertPtr,
	}

	f, ok := m[field.Kind()]
	if !ok {
		return nil
	}

	val, err := f(field.Type(), v)

	if err == nil {
		field.Set(val)
		return nil
	}

	if werr, ok := err.(*wrapError); ok {
		return werr.error
	}

	return nil
}

// Setter is an interface for setting default values
type Setter interface {
	SetDefaults()
}

func callSetter(v interface{}) {
	if ds, ok := v.(Setter); ok {
		ds.SetDefaults()
	}
}

type converterFn func(t reflect.Type, v string) (reflect.Value, error)

func convertBool(t reflect.Type, v string) (reflect.Value, error) {
	val, err := strconv.ParseBool(v)
	if err != nil {
		return reflect.Value{}, err
	}

	return reflect.ValueOf(val).Convert(t), nil
}

func convertInt(t reflect.Type, v string) (reflect.Value, error) {
	val, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return reflect.Value{}, err
	}

	return reflect.ValueOf(int(val)).Convert(t), nil
}

func convertInt8(t reflect.Type, v string) (reflect.Value, error) {
	val, err := strconv.ParseInt(v, 10, 8)
	if err != nil {
		return reflect.Value{}, err
	}

	return reflect.ValueOf(int8(val)).Convert(t), nil
}
func convertInt16(t reflect.Type, v string) (reflect.Value, error) {
	val, err := strconv.ParseInt(v, 10, 16)
	if err != nil {
		return reflect.Value{}, err
	}

	return reflect.ValueOf(int16(val)).Convert(t), nil
}
func convertInt32(t reflect.Type, v string) (reflect.Value, error) {
	val, err := strconv.ParseInt(v, 10, 32)
	if err != nil {
		return reflect.Value{}, err
	}

	return reflect.ValueOf(int32(val)).Convert(t), nil
}

func convertInt64(t reflect.Type, v string) (reflect.Value, error) {
	d, err := time.ParseDuration(v)
	if err == nil {
		return reflect.ValueOf(d).Convert(t), nil
	}

	val, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return reflect.Value{}, err
	}

	return reflect.ValueOf(val).Convert(t), nil
}

func convertUInt(t reflect.Type, v string) (reflect.Value, error) {
	val, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return reflect.Value{}, err
	}

	return reflect.ValueOf(uint(val)).Convert(t), nil
}

func convertUInt8(t reflect.Type, v string) (reflect.Value, error) {
	val, err := strconv.ParseUint(v, 10, 8)
	if err != nil {
		return reflect.Value{}, err
	}

	return reflect.ValueOf(uint8(val)).Convert(t), nil
}

func convertUInt16(t reflect.Type, v string) (reflect.Value, error) {
	val, err := strconv.ParseUint(v, 10, 16)
	if err != nil {
		return reflect.Value{}, err
	}

	return reflect.ValueOf(uint16(val)).Convert(t), nil
}

func convertUInt32(t reflect.Type, v string) (reflect.Value, error) {
	val, err := strconv.ParseUint(v, 10, 32)
	if err != nil {
		return reflect.Value{}, err
	}

	return reflect.ValueOf(uint32(val)).Convert(t), nil
}

func convertUInt64(t reflect.Type, v string) (reflect.Value, error) {
	val, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return reflect.Value{}, err
	}

	return reflect.ValueOf(val).Convert(t), nil
}

func convertUintptr(t reflect.Type, v string) (reflect.Value, error) {
	val, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return reflect.Value{}, err
	}

	return reflect.ValueOf(uintptr(val)).Convert(t), nil
}

func convertFloat32(t reflect.Type, v string) (reflect.Value, error) {
	val, err := strconv.ParseFloat(v, 32)
	if err != nil {
		return reflect.Value{}, err
	}

	return reflect.ValueOf(float32(val)).Convert(t), nil
}

func convertFloat64(t reflect.Type, v string) (reflect.Value, error) {
	val, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return reflect.Value{}, err
	}

	return reflect.ValueOf(val).Convert(t), nil
}

type wrapError struct {
	error
}

func convertString(t reflect.Type, v string) (reflect.Value, error) {
	return reflect.ValueOf(v).Convert(t), nil
}

func convertSlice(t reflect.Type, v string) (reflect.Value, error) {
	ref := reflect.New(t)
	ref.Elem().Set(reflect.MakeSlice(t, 0, 0))

	if v != "" && v != "[]" {
		if err := json.Unmarshal([]byte(v), ref.Interface()); err != nil {
			return reflect.Value{}, &wrapError{err}
		}
	}

	return ref.Elem().Convert(t), nil
}

func convertMap(t reflect.Type, v string) (reflect.Value, error) {
	ref := reflect.New(t)
	ref.Elem().Set(reflect.MakeMap(t))

	if v != "" && v != "{}" {
		if err := json.Unmarshal([]byte(v), ref.Interface()); err != nil {
			return reflect.Value{}, &wrapError{err}
		}
	}

	return ref.Elem().Convert(t), nil
}

func convertPtr(t reflect.Type, v string) (reflect.Value, error) {
	return reflect.New(t.Elem()), nil
}

func convertStruct(t reflect.Type, v string) (reflect.Value, error) {
	ref := reflect.New(t)

	if v != "" && v != "{}" {
		if err := json.Unmarshal([]byte(v), ref.Interface()); err != nil {
			return reflect.Value{}, &wrapError{err}
		}
	}

	return ref.Elem(), nil
}

func hasInitialValue(field reflect.Value) bool {
	return gor.IsEmptyValue(field)
}

func shouldInitializeField(field reflect.Value, defaultVal string) bool {
	switch field.Kind() {
	case reflect.Struct:
		return true
	case reflect.Slice:
		return field.Len() > 0 || defaultVal != ""
	}

	return defaultVal != ""
}

// CanUpdate returns true when the given value is an initial value of its type
func CanUpdate(v interface{}) bool {
	return hasInitialValue(reflect.ValueOf(v))
}
