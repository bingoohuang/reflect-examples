package reflect_examples

import (
	"fmt"
	"reflect"
	"testing"
)

type TestTypeImplInterfaceMarshaler interface {
	MarshalKV() (string, error)
}

type TestTypeImplInterfaceUser struct {
	Email   string `kv:"email,omitempty"`
	Name    string `kv:"name,omitempty"`
	Github  string `kv:"github,omitempty"`
	private string
}

func (u TestTypeImplInterfaceUser) MarshalKV() (string, error) {
	return fmt.Sprintf("name=%s,email=%s,github=%s", u.Name, u.Email, u.Github), nil
}

func TestTypeImplInterface(x *testing.T) {
	fmt.Println(encode(TestTypeImplInterfaceUser{"boring", "Ariel", "a8m", ""}))
	fmt.Println(encode(&TestTypeImplInterfaceUser{Github: "posener", Name: "Eyal", Email: "boring"}))
}

var marshalerType = reflect.TypeOf(new(TestTypeImplInterfaceMarshaler)).Elem()

func encode(i interface{}) (string, error) {
	t := reflect.TypeOf(i)
	if !t.Implements(marshalerType) {
		return "", fmt.Errorf("encode only supports structs that implement the Marshaler interface")
	}
	m, _ := reflect.ValueOf(i).Interface().(TestTypeImplInterfaceMarshaler)
	return m.MarshalKV()
}
