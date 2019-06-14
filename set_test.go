package reflectexamples

import (
	"fmt"
	"testing"
	"unsafe"

	"github.com/modern-go/reflect2"
)

func TestSet(t *testing.T) {
	valType := reflect2.TypeOf(1)
	i := 1
	j := 10
	valType.Set(&i, &j)
	// i will be 10
	fmt.Println(i)
}

func TestUnpoinerSet(t *testing.T) {
	valType := reflect2.TypeOf(1)
	i := 1
	j := 10
	valType.UnsafeSet(unsafe.Pointer(&i), unsafe.Pointer(&j))
	// i will be 10
	fmt.Println(i)
}
