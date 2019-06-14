package reflectexamples

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tkrajina/go-reflector/reflector"
)

type Address struct {
	Street string `tag:"be" tag2:"1,2,3"`
	Number int    `tag:"bi"`
}

type Person struct {
	Name string `tag:"bu"`
	Address
}

func (p Person) Hi(name string) string {
	return fmt.Sprintf("Hi %s my name is %s", name, p.Name)
}

func TestReflector(t *testing.T) {
	p := Person{}
	obj := reflector.New(p)

	fmt.Println(obj.Field("Name").IsValid())

	val, err := obj.Field("Name").Get()
	fmt.Println(val, err)

	p2 := Person{}
	obj2 := reflector.New(&p2)
	err = obj2.Field("Name").Set("Something")
	assert.Nil(t, err)

	jsonTag, _ := obj.Field("Name").Tag("tag")
	assert.Equal(t, "bu", jsonTag)
}
