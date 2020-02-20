// nolint gomnd
package dark

import (
	"fmt"
	"testing"
	"unsafe"
)

func TestStandardAllocation(t *testing.T) {
	a := NewStandardAllocator()
	x := (*int)(unsafe.Pointer(a.Malloc(unsafe.Sizeof(int(0)))))
	if x == nil {
		t.Error("allocated memory is nil")
		return
	}

	*x = 5

	if *x != 5 {
		t.Error("assigning to memory failed")
		return
	}
}

func ExampleAllocator() {
	a := NewStandardAllocator()
	x := (*int)(unsafe.Pointer(a.Malloc(unsafe.Sizeof(int(0)))))
	*x = 5
	fmt.Println(*x)
	// Output: 5
}
