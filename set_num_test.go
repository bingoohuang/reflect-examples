package reflectexamples

import (
	"fmt"
	"reflect"
	"testing"
)

const n = 255

func TestSetNum(x *testing.T) {
	var (
		a int8
		b int16
		c uint
		d float32
		e string
	)
	fmt.Println(fillNum(&a), a)
	fmt.Println(fillNum(&b), b)
	fmt.Println(fillNum(&c), c)
	fmt.Println(fillNum(&d), c)
	fmt.Println(fillNum(&e), e)
}

func fillNum(i interface{}) error {
	v := reflect.ValueOf(i)
	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("non-pointer %v", v.Type())
	}
	// get the value that the pointer v points to.
	v = v.Elem()
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if v.OverflowInt(n) {
			return fmt.Errorf("can't assign value due to %s-overflow", v.Kind())
		}
		v.SetInt(n)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		if v.OverflowUint(n) {
			return fmt.Errorf("can't assign value due to %s-overflow", v.Kind())
		}
		v.SetUint(n)
	case reflect.Float32, reflect.Float64:
		if v.OverflowFloat(n) {
			return fmt.Errorf("can't assign value due to %s-overflow", v.Kind())
		}
		v.SetFloat(n)
	default:
		return fmt.Errorf("can't assign value to a non-number type")
	}
	return nil
}
