// Package goreflect ...
// nolint gomnd
package goreflect

import (
	"fmt"
	"reflect"

	"github.com/averagesecurityguy/random"
	funk "github.com/thoas/go-funk"
)

// EnsureSliceLen grows the slice capability
func EnsureSliceLen(v reflect.Value, len int) {
	// Grow slice if necessary
	if len >= v.Cap() {
		cap2 := v.Cap() + v.Cap()/2
		if cap2 < 4 {
			cap2 = 4
		}
		if cap2 < len {
			cap2 = len
		}

		v2 := reflect.MakeSlice(v.Type(), v.Len(), cap2)
		reflect.Copy(v2, v)
		v.Set(v2)
	}
	if len >= v.Len() {
		v.SetLen(len + 1)
	}
}

// GetSliceByPtr 检查一个值v是否是Slice的指针，返回slice切片的reflect值
func GetSliceByPtr(v interface{}) (reflect.Value, error) {
	iv := reflect.ValueOf(v)
	nilValue := reflect.ValueOf(nil)
	if iv.Kind() != reflect.Ptr {
		return nilValue, fmt.Errorf("non-pointer %v", iv.Type())
	}

	// get the value that the pointer v points to.
	ve := iv.Elem()
	if ve.Kind() != reflect.Slice {
		return nilValue, fmt.Errorf("can't fill non-slice value")
	}

	return ve, nil
}

// SliceContains tells if a slice contains a element.
func SliceContains(slice interface{}, elem interface{}) bool {
	arrValue := reflect.ValueOf(slice)
	arrType := arrValue.Type()
	kind := arrType.Kind()

	if kind == reflect.Slice || kind == reflect.Array {
		for i := 0; i < arrValue.Len(); i++ {
			// XXX - panics if slice element points to an unexported struct field
			// see https://golang.org/pkg/reflect/#Value.Interface
			if arrValue.Index(i).Interface() == elem {
				return true
			}
		}
		return false
	}

	panic(fmt.Sprintf("Type %s is not supported by Map", arrType.String()))
}

// RandomIntN returns a random int
func RandomIntN(n uint64) int {
	i, _ := random.Uint64Range(0, n)
	return int(i)
}

// IterateSlice iterates a slice with a function fn.
func IterateSlice(arr interface{}, start int, fn interface{}) (bool, interface{}) {
	if !funk.IsFunction(fn) {
		panic("Second argument must be function")
	}

	arrValue := reflect.ValueOf(arr)
	arrType := arrValue.Type()
	kind := arrType.Kind()

	if kind == reflect.Slice || kind == reflect.Array {
		if start < 0 {
			start = RandomIntN(uint64(arrValue.Len()))
		}
		return iterateSlice(arrValue, start, reflect.ValueOf(fn))
	}

	panic(fmt.Sprintf("Type %s is not supported by Map", arrType.String()))
}

// var ErrorInterface = reflect.TypeOf((*error)(nil)).Elem()

func iterateSlice(arrValue reflect.Value, start int, funcValue reflect.Value) (bool, interface{}) {
	funcType := funcValue.Type()
	numOut := funcType.NumOut()
	numIn := funcType.NumIn()
	if !(numIn == 1 || numIn == 2) || numOut > 2 {
		panic("Iterate function with an array must have 1/2 parameter " +
			"and must return 0/1(bool)/2(bool,error) parameter")
	}

	if numOut >= 1 && funcType.Out(0).Kind() != reflect.Bool {
		panic("Iterate function must return bool when there is 1 parameters")
	}
	if numOut >= 2 && funcType.Out(1).Kind() != reflect.Interface {
		panic("Iterate function must return (bool, error) when there is 2 parameters")
	}

	arrElemType := arrValue.Type().Elem()

	// Checking whether element type is convertible to function's first argument's type.
	elemPos := 0
	if numIn == 2 {
		elemPos = 1
	}
	if !arrElemType.ConvertibleTo(funcType.In(elemPos)) {
		panic("Iterate function's argument is not compatible with type of array.")
	}

	if numIn == 2 && reflect.Int != funcType.In(0).Kind() {
		panic("Iterate function's 1st argument is not int.")
	}

	if numOut == 0 {
		internalIterateSlice0(start, arrValue.Len(), arrValue, numIn, funcValue)
		internalIterateSlice0(0, start, arrValue, numIn, funcValue)
		return false, nil
	}

	if over, inte := internalIterateSlice1(start, arrValue.Len(), arrValue, numIn, numOut, funcValue); over {
		return true, inte
	}
	return internalIterateSlice1(0, start, arrValue, numIn, numOut, funcValue)
}

func internalIterateSlice1(from, to int, arr reflect.Value, numIn, numOut int, f reflect.Value) (bool, interface{}) {
	for i := from; i < to; i++ {
		var values []reflect.Value
		if numIn == 1 {
			values = []reflect.Value{arr.Index(i)}
		} else if numIn == 2 {
			values = []reflect.Value{reflect.ValueOf(i), arr.Index(i)}
		}

		if results := f.Call(values); results[0].Bool() {
			if numOut >= 2 {
				return true, results[1].Interface()
			}
			return true, nil
		}
	}

	return false, nil
}

func internalIterateSlice0(from, to int, arrValue reflect.Value, numIn int, funcValue reflect.Value) {
	for i := from; i < to; i++ {
		var values []reflect.Value
		if numIn == 1 {
			values = []reflect.Value{arrValue.Index(i)}
		} else if numIn == 2 {
			values = []reflect.Value{reflect.ValueOf(i), arrValue.Index(i)}
		}
		_ = funcValue.Call(values)
	}
}
