package reflectexamples

import (
	"fmt"
	"reflect"
	"testing"
)

func Add(a, b int) int { return a + b }

func TestFunctionCallArgs(x *testing.T) {
	v := reflect.ValueOf(Add)
	if v.Kind() != reflect.Func {
		return
	}
	t := v.Type()
	argv := make([]reflect.Value, t.NumIn())
	for i := range argv {
		// validate the type of parameter "i".
		if t.In(i).Kind() != reflect.Int {
			return
		}
		argv[i] = reflect.ValueOf(i)
	}
	// note that, len(result) == t.NumOut()
	result := v.Call(argv)
	if len(result) != 1 || result[0].Kind() != reflect.Int {
		return
	}
	fmt.Println(result[0].Int())
}
