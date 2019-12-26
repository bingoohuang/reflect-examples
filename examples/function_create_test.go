package examples

import (
    "fmt"
    "reflect"
    "testing"
)

type AddCreate func(int64, int64) int64

func TestFunctionCreate(x *testing.T) {
    t := reflect.TypeOf(AddCreate(nil))
    mul := reflect.MakeFunc(t, func(args []reflect.Value) []reflect.Value {
        a := args[0].Int()
        b := args[1].Int()
        return []reflect.Value{reflect.ValueOf(a + b)}
    })
    fn, ok := mul.Interface().(AddCreate)
    if !ok {
        return
    }
    fmt.Println(fn(2, 3))
}
