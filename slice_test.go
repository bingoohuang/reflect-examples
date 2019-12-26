package goreflect

import (
    "testing"

    "github.com/stretchr/testify/assert"
)

func TestIterateSlice(t *testing.T) {
    a := []string{"0", "1", "2"}
    st := ""

    ok, err := IterateSlice(a, 1, func(s string) { st += s })
    assert.False(t, ok)
    assert.Nil(t, err)
    assert.Equal(t, "120", st)
}

func TestIterateSlice2(t *testing.T) {
    a := []string{"0", "1", "2"}
    st := ""

    ok, err := IterateSlice(a, 2, func(i int, s string) bool { st += s; return i == 0 })
    assert.True(t, ok)
    assert.Nil(t, err)
    assert.Equal(t, "20", st)
}

func TestIterateSlice3(t *testing.T) {
    a := []string{"0", "1", "2"}
    st := ""

    ok, res := IterateSlice(a, 2, func(i int, s string) (bool, interface{}) { st += s; return i == 0, "xxxx" })
    assert.True(t, ok)
    assert.Equal(t, "xxxx", res)
    assert.Equal(t, "20", st)
}

func TestSliceContains(t *testing.T) {
    a := []string{"0", "1", "2"}
    assert.True(t, SliceContains(a, "0"))
    assert.False(t, SliceContains(a, "3"))
}
