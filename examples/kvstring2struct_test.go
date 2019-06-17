package examples

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

type TestKvString2StructUser struct {
	Name    string
	Github  string
	private string
}

func TestKvString2Struct(x *testing.T) {
	var (
		v0 TestKvString2StructUser
		v1 *TestKvString2StructUser
		v2 = new(TestKvString2StructUser)
		v3 struct{ Name string }
		s  = "Name=Ariel,Github=a8m"
	)
	fmt.Println(kvString2Struct(s, &v0), v0) // pass
	fmt.Println(kvString2Struct(s, v1), v1)  // fail
	fmt.Println(kvString2Struct(s, v2), v2)  // pass
	fmt.Println(kvString2Struct(s, v3), v3)  // fail
	fmt.Println(kvString2Struct(s, &v3), v3) // pass
}

func kvString2Struct(s string, i interface{}) error {
	v := reflect.ValueOf(i)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return fmt.Errorf("kvString2Struct requires non-nil pointer")
	}
	// get the value that the pointer v points to.
	v = v.Elem()
	// assume that the input is valid.
	for _, kv := range strings.Split(s, ",") {
		s := strings.Split(kv, "=")
		f := v.FieldByName(s[0])
		// make sure that this field is defined, and can be changed.
		if !f.IsValid() || !f.CanSet() {
			continue
		}
		// assume all the fields are type string.
		f.SetString(s[1])
	}
	return nil
}
