package dark

import (
	"unsafe"
)

// The cost of copying memory from RAM to CPU is quite high. This strategy here
// guarantees that small enough data can be layed out continuously in memory.
// That way your data will be fetch in the minimal amount memory access. The
// example only packs 2 slice but this can easilly be extended to more then 2
// slice or even to other pointers to any structs or any combination of these.
func ExampleMemoryPacking() {
	mem0, mem1 := []int{1, 2, 3}, []float64{4, 5, 6, 7}
	sizeElem0 := unsafe.Sizeof(mem0[0]) // sizeof(int)
	sizeElem1 := unsafe.Sizeof(mem1[0]) // sizeof(float64)

	// I need to find out what are the rules for being guaranteed that early
	// bailing guarantees that continuous blocks stay continuous.
	byteAftermem0 := uintptr(unsafe.Pointer(&mem0[0])) + sizeElem0*uintptr(len(mem0))
	beginingOfmem1 := uintptr(unsafe.Pointer(&mem1[0]))
	if byteAftermem0 == beginingOfmem1 {
		// return
	}

	// find the total memory we'll need as well as the memory used by every
	// block
	totalMem0 := uintptr(len(mem0)) * sizeElem0
	totalMem1 := uintptr(len(mem1)) * sizeElem1

	totalMem := totalMem0 + totalMem1

	// Allocate a continuous block for all the memory we'll need
	raw := make([]byte, totalMem)

	// next we create the new headers and point the memory to the raw backing
	// array.
	header0 := SliceHeader{
		Data: unsafe.Pointer(&raw[0]), // the first available byte
		Len:  len(mem0),
		Cap:  len(mem0),
	}

	header1 := SliceHeader{
		// the first free byte after the memory used by mem0
		Data: unsafe.Pointer(&raw[totalMem0]),
		Len:  len(mem1),
		Cap:  len(mem1),
	}

	// convert back to actual slices
	s0 := *(*[]int)(unsafe.Pointer(&header0))
	s1 := *(*[]float64)(unsafe.Pointer(&header1))

	// Copy the old data to the new memory
	copy(s0, mem0)
	copy(s1, mem1)

	// Here we verify that they are indeed continuous, you wouldn't normally
	// have that last part in your code
	byteAfters0 := uintptr(unsafe.Pointer(&s0[0])) + sizeElem0*uintptr(len(s0))
	beginingOfs1 := uintptr(unsafe.Pointer(&s1[0]))
	// fmt.Printf("%X %X\n", byteAfters0, beginingOfs1)
	if byteAfters0 != beginingOfs1 {
		panic("memory not continuous")
	}

	// Then verify that the two bit of memory don't overlap
	// First assign data to everything
	s0[0], s0[1], s0[2] = 9, 9, 9
	s1[0], s1[1], s1[2], s1[3] = 8, 8, 8, 8
	// then verify that nothing overwrote over anything else.
	if s0[0] != 9 || s0[1] != 9 || s0[2] != 9 ||
		s1[0] != 8 || s1[1] != 8 || s1[2] != 8 || s1[3] != 8 {
		panic("memory overlap")
	}
	// there should be no output

	// Output:
}
