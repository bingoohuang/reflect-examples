package dark

import (
	"fmt"
	"testing"
)

func TestString(t *testing.T) {
	b := []byte("foo")
	s := String(b)
	b[0], b[1], b[2] = 'b', 'a', 'r'

	if s != "bar" {
		t.Errorf("String conversion failed, expected %q, got \"bar\"", s)
		return
	}
}

func TestByteSlice(t *testing.T) {
	b0 := []byte("foo")
	s := String(b0)
	b := ByteSlice(s)
	b[0], b[1], b[2] = 'b', 'a', 'r'

	if string(b) != "bar" {
		t.Errorf("String conversion failed, expected %q, got \"bar\"", s)
		return
	}
}

func ExampleString() {
	b := []byte("Hello world")
	s := String(b)
	fmt.Println(s)
	// Output: Hello world
}

func ExampleByteSlice_read() {
	s := "Hello world"
	b := ByteSlice(s)
	fmt.Println(string(b))
	// Output: Hello world
}

func ExampleByteSlice_write() {
	b0 := []byte("foo") // here we allocate a block of writeable memory
	s := String(b0)     // convert it to a valid string
	b := ByteSlice(s)   // convert back to demonstrate ByteSlice
	b[0], b[1], b[2] = 'b', 'a', 'r'
	fmt.Println(string(b))
	// Output: bar
}

func ExampleByteSlice_segfault() {
	s := "Hello world" // allocates a read-only block.
	b := ByteSlice(s)
	_ = b
	// the next line segfaults because the original memory was allocated
	// read-only. segfaults are not recoverable.
	// b[0] = 'a'
}
