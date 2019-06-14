package reflect_examples

import (
	"reflect"
	"testing"

	"github.com/mitchellh/reflectwalk"
	"github.com/stretchr/testify/assert"
)

type WalkMe struct {
	Name string
}

type Walker struct {
	FieldNameValues string
}

func (Walker) Struct(reflect.Value) error {
	return nil
}

func (w *Walker) StructField(f reflect.StructField, v reflect.Value) error {
	w.FieldNameValues += f.Name + ":" + v.String() + ","

	return nil
}

func TestReflectWalk(t *testing.T) {
	var walker Walker

	walkMe := WalkMe{Name: "bingoo"}

	err := reflectwalk.Walk(walkMe, &walker)
	assert.Nil(t, err)

	assert.Equal(t, "Name:bingoo,", walker.FieldNameValues)
}
