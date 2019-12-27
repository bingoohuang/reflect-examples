package goreflect

import (
	"fmt"
	"reflect"
	"sort"
)

// MapKeys 返回Map的key切片
func MapKeys(m interface{}) []string {
	return MapKeysX(m).([]string)
}

// MapKeysSorted 返回Map排序后的key切片
func MapKeysSorted(m interface{}) []string {
	keys := MapKeys(m)
	sort.Strings(keys)
	return keys
}

// MapKeysX 返回Map的key切片
func MapKeysX(m interface{}) interface{} {
	v := reflect.ValueOf(m)
	if v.Kind() != reflect.Map {
		return nil
	}

	keyType := v.Type().Key()
	ks := reflect.MakeSlice(reflect.SliceOf(keyType), v.Len(), v.Len())
	for i, key := range v.MapKeys() {
		ks.Index(i).Set(key)
	}

	return ks.Interface()
}

// MapKeysSortedX 返回Map排序后的key切片
func MapKeysSortedX(m interface{}) interface{} {
	mv := reflect.ValueOf(m)
	if mv.Kind() != reflect.Map {
		return nil
	}

	mapLen := mv.Len()
	keyType := mv.Type().Key()
	keyKind := keyType.Kind()
	var keyMap map[interface{}]string
	switch keyKind {
	case reflect.String:
	case reflect.Int:
	case reflect.Float64:
	default:
		keyMap = make(map[interface{}]string, mapLen)
	}

	ks := reflect.MakeSlice(reflect.SliceOf(keyType), mapLen, mapLen)
	i := 0
	for _, k := range mv.MapKeys() {
		if keyMap != nil {
			keyMap[k.Interface()] = fmt.Sprintf("%v", k.Interface())
		}

		ks.Index(i).Set(k)
		i++
	}

	ksi := ks.Interface()

	if keyMap != nil {
		sort.Slice(ksi, func(i, j int) bool {
			ki := keyMap[ks.Index(i).Interface()]
			kj := keyMap[ks.Index(j).Interface()]
			return ki < kj
		})
	} else {
		switch keyKind {
		case reflect.String:
			sort.Strings(ksi.([]string))
		case reflect.Int:
			sort.Ints(ksi.([]int))
		case reflect.Float64:
			sort.Float64s(ksi.([]float64))
		}
	}

	return ksi
}

// MapValues 返回Map的value切片
func MapValues(m interface{}) []string {
	return MapValuesX(m).([]string)
}

// MapValuesX 返回Map的value切片
func MapValuesX(m interface{}) interface{} {
	v := reflect.ValueOf(m)
	if v.Kind() != reflect.Map {
		return nil
	}

	typ := v.Type().Elem()
	sl := reflect.MakeSlice(reflect.SliceOf(typ), v.Len(), v.Len())
	for i, key := range v.MapKeys() {
		sl.Index(i).Set(v.MapIndex(key))
	}

	return sl.Interface()
}

// MapGetOr get value from m by key or returns the defaultValue
func MapGetOr(m, k, defaultValue interface{}) interface{} {
	mv := reflect.ValueOf(m)
	if mv.Kind() != reflect.Map {
		return nil
	}

	v := mv.MapIndex(reflect.ValueOf(k))
	if v.IsValid() {
		return v.Interface()
	}

	return defaultValue
}

// WalkMap iterates the map by iterFunc.
func WalkMap(m interface{}, iterFunc interface{}) {
	mv := reflect.ValueOf(m)
	if mv.Kind() != reflect.Map {
		return
	}

	mapLen := mv.Len()
	keyType := mv.Type().Key()
	keyKind := keyType.Kind()
	var keyMap map[interface{}]string
	switch keyKind {
	case reflect.String:
	case reflect.Int:
	case reflect.Float64:
	default:
		keyMap = make(map[interface{}]string, mapLen)
	}

	ks := reflect.MakeSlice(reflect.SliceOf(keyType), mapLen, mapLen)
	for i, k := range mv.MapKeys() {
		if keyMap != nil {
			keyMap[k.Interface()] = fmt.Sprintf("%v", k.Interface())
		}

		ks.Index(i).Set(k)
	}

	ksi := ks.Interface()

	if keyMap != nil {
		sort.Slice(ksi, func(i, j int) bool {
			ki := keyMap[ks.Index(i).Interface()]
			kj := keyMap[ks.Index(j).Interface()]
			return ki < kj
		})
	} else {
		switch keyKind {
		case reflect.String:
			sort.Strings(ksi.([]string))
		case reflect.Int:
			sort.Ints(ksi.([]int))
		case reflect.Float64:
			sort.Float64s(ksi.([]float64))
		}
	}

	funcValue := reflect.ValueOf(iterFunc)

	for j := 0; j < mapLen; j++ {
		k := ks.Index(j)
		v := mv.MapIndex(k)
		funcValue.Call([]reflect.Value{k, v})
	}
}
