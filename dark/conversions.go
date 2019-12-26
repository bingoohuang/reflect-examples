package dark

import (
    "unsafe"
)

// String transforms a slice of byte into a string without doing the actual copy
// of the data.
func String(b []byte) string {
    return *(*string)(unsafe.Pointer(&b))
}

// ByteSlice converts a strings into the equivalent byte slice without doing the
// actual copy of the data. The slice returned by this function may be read-only.
// See examples for more details.
func ByteSlice(s string) []byte {
    sh := *(*StringHeader)(unsafe.Pointer(&s))
    bh := SliceHeader{
        Data: sh.Data,
        Len:  sh.Len,
        Cap:  sh.Len,
    }
    return *(*[]byte)(unsafe.Pointer(&bh))
}
