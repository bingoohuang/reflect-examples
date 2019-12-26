package dark

import (
    "unsafe"
)

// Allocator represents a memory allocator, this can be used to manage your
// memory very carefully when standard allocation is not good enough.
type Allocator interface {
    // Malloc allocates n bytes via this allocator and returns a pointer to
    // at least size valid bytes of memory.
    Malloc(size uintptr) unsafe.Pointer
    // Free frees the memory this pointer points to.
    Free(unsafe.Pointer)
}

// stdAllocator uses go standard allocation algorithm to allocate and free
// memory.
type stdAllocator struct{}

// NewStandardAllocator returns an allocator based on Go standard allocation
// algorithm.
func NewStandardAllocator() Allocator {
    return stdAllocator{}
}

// Guarantee that stdAllocator implements Allocator at compile time.
var _ Allocator = stdAllocator{}

// Malloc allocates n bytes via this allocator and returns a pointer to at least
// n valid bytes of memory.
func (a stdAllocator) Malloc(size uintptr) unsafe.Pointer {
    return unsafe.Pointer(&make([]byte, size)[0])
}

// Free frees the memory this pointer points to.
func (a stdAllocator) Free(p unsafe.Pointer) {}
