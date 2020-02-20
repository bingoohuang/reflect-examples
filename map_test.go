// nolint gomnd
package goreflect

import (
	"fmt"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapKeys(t *testing.T) {
	m := map[string]string{"k1": "v1", "k2": "v2"}
	keys := MapKeys(m)
	sort.Strings(keys)
	assert.Equal(t, []string{"k1", "k2"}, keys)
}

func TestMapKeysSorted(t *testing.T) {
	m := map[int]int{1: 11, 2: 22, 3: 33}
	keys := MapKeysSortedX(m).([]int)
	assert.Equal(t, []int{1, 2, 3}, keys)
}

func TestMapGetOr(t *testing.T) {
	a := assert.New(t)

	m := map[int]int{1: 11}

	a.Equal(11, MapGetOr(m, 1, 100))
	a.Equal(22, MapGetOr(m, 2, 22))

	m2 := map[string]string{"1": "11"}

	a.Equal("11", MapGetOr(m2, "1", "100"))
	a.Equal("22", MapGetOr(m2, "2", "22"))
}

func TestWalkMapInt(t *testing.T) {
	a := assert.New(t)

	m := map[int]int{1: 11, 2: 22, 3: 33}
	ks := ""
	vs := ""

	WalkMap(m, func(k, v int) {
		ks += fmt.Sprintf("%d", k)
		vs += fmt.Sprintf("%d", v)
	})
	a.Equal("123", ks)
	a.Equal("112233", vs)
}

func TestWalkMapString(t *testing.T) {
	a := assert.New(t)

	m := map[string]int{"1": 11, "2": 22, "3": 33}
	ks := ""
	vs := ""

	WalkMap(m, func(k string, v int) {
		ks += k
		vs += fmt.Sprintf("%v", v)
	})
	a.Equal("123", ks)
	a.Equal("112233", vs)
}

func TestWalkMapFloat64(t *testing.T) {
	a := assert.New(t)

	m := map[float64]int{1.1: 11, 2.2: 22, 3.3: 33}
	ks := ""
	vs := ""

	WalkMap(m, func(k float64, v int) {
		ks += fmt.Sprintf("%.1f", k)
		vs += fmt.Sprintf("%v", v)
	})
	a.Equal("1.12.23.3", ks)
	a.Equal("112233", vs)
}

func TestWalkMapOther(t *testing.T) {
	a := assert.New(t)

	m := map[float32]int{1.1: 11, 2.2: 22, 3.3: 33}
	ks := ""
	vs := ""

	WalkMap(m, func(k float32, v int) {
		ks += fmt.Sprintf("%.1f", k)
		vs += fmt.Sprintf("%v", v)
	})
	a.Equal("1.12.23.3", ks)
	a.Equal("112233", vs)
}
