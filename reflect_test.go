package gor

import (
	"fmt"
	"image"
	"reflect"
	"testing"
)

type t int // A type definition

// Some methods on the type
func (r t) Twice() t       { return r * 2 }
func (r t) Half() t        { return r / 2 }
func (r t) Less(r2 t) bool { return r < r2 }
func (r t) privateMethod() {} // nolint unused

type FooService interface {
	Foo1(x int) int
	Foo2(x string) string
}

func TestListMethods(te *testing.T) {
	report(t(0))
	report(image.Point{})
	report((*FooService)(nil))
	report(struct{ FooService }{})
}

func report(x interface{}) {
	v := reflect.ValueOf(x)
	t := reflect.TypeOf(x) // or v.Type()

	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	n := t.NumMethod()

	fmt.Printf("Type %v has %d exported methods:\n", t, n)

	const format = "%-6s %-46s %s\n"

	fmt.Printf(format, "Name", "Methods expression", "Methods value")

	for i := 0; i < n; i++ {
		if t.Kind() == reflect.Interface {
			fmt.Printf(format, t.Method(i).Name, "(N/A)", t.Method(i).Type)
		} else {
			fmt.Printf(format, t.Method(i).Name, t.Method(i).Type, v.Method(i).Type())
		}
	}

	fmt.Println()
}
