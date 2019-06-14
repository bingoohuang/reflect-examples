package reflect_examples

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
)

func Sum(x1, x2 int, xs ...int) int {
	sum := x1 + x2
	for _, xi := range xs {
		sum += xi
	}
	return sum
}

func TestFunctionCallVarags(x *testing.T) {
	v := reflect.ValueOf(Sum)
	if v.Kind() != reflect.Func {
		return
	}
	t := v.Type()
	argc := t.NumIn()
	if t.IsVariadic() {
		argc += rand.Intn(10)
	}
	argv := make([]reflect.Value, argc)
	for i := range argv {
		argv[i] = reflect.ValueOf(i)
	}
	result := v.Call(argv)
	fmt.Println(result[0].Int()) // assume that t.NumOut() > 0 tested above.
}
