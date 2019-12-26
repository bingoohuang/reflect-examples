package examples

import (
    "fmt"
    "reflect"
    "testing"
)

type A struct{}

func (A) Hello() { fmt.Println("World") }

func TestFunctionCall(x *testing.T) {
    // ValueOf returns a new Value, which is the reflection interface to a Go value.
    v := reflect.ValueOf(A{})
    m := v.MethodByName("Hello")
    if m.Kind() != reflect.Func {
        return
    }
    m.Call(nil)
}
