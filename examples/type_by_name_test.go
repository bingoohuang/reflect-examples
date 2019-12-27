package examples

import (
    "fmt"
    "testing"

    "github.com/modern-go/reflect2"
)

func TestTypeByName(x *testing.T) {
    t := reflect2.TypeByName("time.Time")
    fmt.Printf("type:%+v", t)
}
