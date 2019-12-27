package dark

import (
	"fmt"
	"reflect"
	"unsafe"
)

// deepCopier keeps track of the pointers we discovered so far.
type deepCopier struct {
	pointers map[unsafe.Pointer]interface{}
}

func (c *deepCopier) deepCopy(i interface{}) interface{} {
	if i == nil {
		return nil
	}

	typ := reflect.TypeOf(i)
	value := reflect.ValueOf(i)

	// if the thing is nil from the start just leave, IsNil only works on
	// certain types though.
	switch typ.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface,
		reflect.Map, reflect.Ptr, reflect.Slice:
		if value.IsNil() {
			return value.Interface()
		}
	}

	switch typ.Kind() {
	case reflect.Bool:
		return value.Bool()

	case reflect.Int: // ints
		return int(value.Int())
	case reflect.Int8:
		return int8(value.Int())
	case reflect.Int16:
		return int16(value.Int())
	case reflect.Int32:
		return int32(value.Int())
	case reflect.Int64:
		return int64(value.Int())

	case reflect.Uint: // uints
		return uint(value.Uint())
	case reflect.Uint8:
		return uint8(value.Uint())
	case reflect.Uint16:
		return uint16(value.Uint())
	case reflect.Uint32:
		return uint32(value.Uint())
	case reflect.Uint64:
		return uint64(value.Uint())

	case reflect.Float32: // floats
		return float32(value.Float())
	case reflect.Float64:
		return float64(value.Float())

	case reflect.Complex64: // complexs
		return complex64(value.Complex())
	case reflect.Complex128:
		return complex128(value.Complex())

	case reflect.Uintptr:
		// void* can't be recursively copied as the info of the type they point
		// to is not accessible via reflection
		return uintptr(value.Uint())
	case reflect.UnsafePointer:
		// same as reflect.Uintptr
		return unsafe.Pointer(value.Pointer())

	case reflect.String:
		// ensure a proper copy of the strings content
		src := []byte(value.String())
		return string(src)

	case reflect.Func:
		// func cannot be copied per se, that means closures will remain closure.
		return value.Interface()

	case reflect.Array:
		a := reflect.Indirect(reflect.New(typ))
		for i := 0; i < value.Len(); i++ {
			a.Index(i).Set(reflect.ValueOf(c.deepCopy(value.Index(i).Interface())))
		}
		return a.Interface()
	case reflect.Slice:
		// verify if we already have copied that slice
		old, ok := c.pointers[unsafe.Pointer(value.Pointer())]
		if ok {
			return old
		}

		s := reflect.MakeSlice(typ, value.Len(), value.Cap())

		// store the copy in the map for future references
		c.pointers[unsafe.Pointer(value.Pointer())] = s.Interface()

		for i := 0; i < value.Len(); i++ {
			s.Index(i).Set(reflect.ValueOf(c.deepCopy(value.Index(i).Interface())))
		}

		return s.Interface()

	case reflect.Chan:
		// verify if we already have copied that channel
		old, ok := c.pointers[unsafe.Pointer(value.Pointer())]
		if ok {
			return old
		}

		ch := reflect.MakeChan(typ, value.Len())

		// store the copy in the map for future references
		c.pointers[unsafe.Pointer(value.Pointer())] = ch.Interface()

		return ch.Interface()

	case reflect.Map:
		// verify if we already have copied that map
		old, ok := c.pointers[unsafe.Pointer(value.Pointer())]
		if ok {
			return old
		}

		m := reflect.MakeMap(typ)

		// store the copy in the map for future references
		c.pointers[unsafe.Pointer(value.Pointer())] = m.Interface()

		// copy the data between maps
		for _, key := range value.MapKeys() {
			m.SetMapIndex(key, reflect.ValueOf(c.deepCopy(value.MapIndex(key).Interface())))
		}

		return m.Interface()

	case reflect.Struct:
		s := reflect.Indirect(reflect.New(typ))
		for i := 0; i < value.NumField(); i++ {
			Sudo(s.Field(i)).Set(reflect.ValueOf(c.deepCopy(Sudo(value.Field(i)).Interface())))
		}
		return s.Interface()
	case reflect.Ptr:
		// verify if we already have copied that pointer
		pointto, ok := c.pointers[unsafe.Pointer(value.Pointer())]
		if ok {
			return pointto
		}

		// in this case we indirect before allocating so we get a reflect.Value
		// already all set to receive data.
		ptr := reflect.New(reflect.Indirect(value).Type())

		interf := ptr.Interface()
		// store the copy in the map for future references
		c.pointers[unsafe.Pointer(value.Pointer())] = interf

		cp := reflect.ValueOf(c.deepCopy(reflect.Indirect(value).Interface()))
		reflect.Indirect(ptr).Set(cp)

		return interf
	}
	panic(fmt.Sprintf("Problem copying a %s of kind %s, please file in a bug report on github with the type you tried to copy", typ, typ.Kind()))
}

// DeepCopy does a deep copy of the given data. It doesn't recursively copy
// function pointers, uintptrs or unsafe.Pointers as the data it points to is
// unaccessible via reflection but it will copy the pointer value. It does
// handle circular structures. It does not copy buffered channel content as
// there are no reflect.ChanHeader or similar provided by stdlib. It also isn't
// particularely efficient on string copy, by making a complete copy of every
// string it comes across we could effectively be allocating way more memory
// then needed. If anyone comes across this problem ping @hydroflame and he will
// implement something to reuse memory when possible.
func DeepCopy(i interface{}) interface{} {
	c := deepCopier{
		pointers: make(map[unsafe.Pointer]interface{}),
	}
	return c.deepCopy(i)
}
