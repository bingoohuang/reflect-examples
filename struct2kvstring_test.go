package reflectexamples

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

type TestStruct2KvStringUser struct {
	Email   string `kv:"email,omitempty"`
	Name    string `kv:"name,omitempty"`
	Github  string `kv:"github,omitempty"`
	private string
}

func TestStruct2KvString(x *testing.T) {
	var (
		u = TestStruct2KvStringUser{Name: "Ariel", Github: "a8m"}
		v = struct {
			A, B, C string
		}{
			"foo",
			"bar",
			"baz",
		}
		w = &TestStruct2KvStringUser{}
	)
	fmt.Println(struct2KvString(u))
	fmt.Println(struct2KvString(v))
	fmt.Println(struct2KvString(w))
}

// this example supports only structs, and assume their
// fields are type string.
func struct2KvString(i interface{}) (string, error) {
	v := reflect.ValueOf(i)
	t := v.Type()
	if t.Kind() != reflect.Struct {
		return "", fmt.Errorf("type %s is not supported", t.Kind())
	}
	var s []string
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		// skip unexported fields. from godoc:
		// PkgPath is the package path that qualifies a lower case (unexported)
		// field name. It is empty for upper case (exported) field names.
		if f.PkgPath != "" {
			continue
		}
		fv := v.Field(i)
		key, omit := readTag(f)
		// skip empty values when "omitempty" set.
		if omit && fv.String() == "" {
			continue
		}
		s = append(s, fmt.Sprintf("%s=%s", key, fv.String()))
	}
	return strings.Join(s, ","), nil
}

func readTag(f reflect.StructField) (string, bool) {
	val, ok := f.Tag.Lookup("kv")
	if !ok {
		return f.Name, false
	}
	opts := strings.Split(val, ",")
	omit := false
	if len(opts) == 2 {
		omit = opts[1] == "omitempty"
	}
	return opts[0], omit
}
