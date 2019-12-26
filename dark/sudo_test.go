package dark

import (
    "fmt"
    "reflect"
    "testing"
)

func TestSudo(t *testing.T) {
    var s struct {
        x int
    }
    v := reflect.ValueOf(&s).Elem()

    // Sudo the reflect.Value
    v = Sudo(v.FieldByName("x"))

    if !v.CanSet() {
        t.Error("Sudo did not function properly")
    }
}

func ExampleSudo() {
    var s struct {
        x int
    }
    v := reflect.ValueOf(&s).Elem()

    // Sudo the reflect.Value
    v = Sudo(v.FieldByName("x"))
    v.Set(reflect.ValueOf(int(2)))

    fmt.Println(s.x)
    // Output: 2
}

func ExampleSudo_panic() {
    var s struct {
        x int
    }
    v := reflect.ValueOf(&s).Elem()

    defer func() {
        if r := recover(); r != nil {
            fmt.Println(r)
        }
    }()
    v.Set(reflect.ValueOf(int(2)))

    fmt.Println(s.x)
    // Output: reflect.Set: value of type int is not assignable to type struct { x int }
}
