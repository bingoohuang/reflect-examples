package examples

import (
	"fmt"
	"reflect"
	"testing"
)

type TestReadStructTagsUser struct {
	Email  string `mcl:"email"`
	Name   string `mcl:"name"`
	Age    int    `mcl:"age"`
	Github string `mcl:"github" default:"a8m"`
}

func TestReadStructTags(x *testing.T) {
	var u interface{} = TestReadStructTagsUser{}
	// TypeOf returns the reflection Type that represents the dynamic type of u.
	t := reflect.TypeOf(u)
	// Kind returns the specific kind of this type.
	if t.Kind() != reflect.Struct {
		return
	}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		fmt.Println(f.Tag.Get("mcl"), f.Tag.Get("default"))
	}
}
