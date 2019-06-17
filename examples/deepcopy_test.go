package examples

import (
	"reflect"
	"testing"

	"github.com/darkgopher/dark"
)

func TestDeepCopy(t *testing.T) {
	s := "Hello world!"
	r := struct {
		s0 string
		s1 string
	}{
		s0: s,
		s1: s[:5],
	}
	u := dark.DeepCopy(r)
	if !reflect.DeepEqual(u, r) {
		t.Fatalf("not equal got %v, want %v", u, r)
	}
}
