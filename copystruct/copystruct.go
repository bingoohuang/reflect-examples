package copystruct

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"strings"
)

const (
	// optionField is the from field option name for struct tag.
	optionField = "field"
	// optionContext is the context option name for struct tag.
	optionContext = "context"
	// optionSkip is the skip option name for struct tag.
	optionSkip = "skip"
	// optionForce is the skip option name for struct tag.
	optionForce = "force"
	// optionConvert is the convert option name for struct tag.
	optionConvert = "convert"
)

// tagOptions is a map that contains extracted struct tag context.
type tagOptions map[string]string

// OptionFn types the option func type.
type OptionFn func(cs *CopyStruct)

// CopyStruct deep copies a struct to/from a struct.
type CopyStruct struct {
	dst, src interface{}
	ctx      map[string]interface{}
	tagName  string
}

// TagName customizes the tagName (default is copystruct)
func TagName(tagName string) OptionFn {
	return func(cs *CopyStruct) {
		cs.tagName = tagName
	}
}

// Copy sets source or destination.
func Copy(src interface{}, optionFns ...OptionFn) *CopyStruct {
	c := &CopyStruct{src: src, tagName: "copystruct"}

	for _, fn := range optionFns {
		fn(c)
	}

	return c
}

// WithContext injects the given context into the builder instance.
func (dc *CopyStruct) WithContext(ctx map[string]interface{}) *CopyStruct {
	dc.ctx = ctx
	return dc
}

// To sets the destination.
func (dc *CopyStruct) To(dst interface{}) error {
	dc.dst = dst

	return dc.process(dc.dst, dc.src, false)
}

// From sets the given the source as destination and destination as source.
func (dc *CopyStruct) From(src interface{}) error {
	dc.dst = dc.src
	dc.src = src

	return dc.process(dc.dst, dc.src, true)
}

// process handles copy.
func (dc *CopyStruct) process(dst, src interface{}, reversed bool) error {
	dstValue := reflect.Indirect(reflect.ValueOf(dst))
	if !dstValue.CanAddr() {
		return fmt.Errorf("destination %+v is unaddressable", dstValue.Interface())
	}

	srcValue := reflect.Indirect(reflect.ValueOf(src))

	for _, f := range getFieldNames(src) {
		dc.copyFields(srcValue, dstValue, f, dst, reversed)
	}

	for _, m := range getMethodNames(src) {
		if err := dc.copyMethods(dstValue, src, dst, m); err != nil {
			return err
		}
	}

	return nil
}

func (dc *CopyStruct) copyMethods(dstValue reflect.Value, src, dst interface{}, m string) error {
	name, tagOptions := dc.getRelatedField(dst, m)
	if name == "" {
		return nil
	}

	if _, ok := tagOptions[optionSkip]; ok {
		return nil
	}

	method := reflect.ValueOf(src).MethodByName(m)
	if !method.IsValid() {
		return fmt.Errorf("method %s is invalid", m)
	}

	dstFieldType, _ := dstValue.Type().FieldByName(name)
	dstFieldValue := dstValue.FieldByName(name)

	_, force := tagOptions[optionForce]

	args := make([]reflect.Value, 0)

	if _, withContext := tagOptions[optionContext]; withContext {
		args = []reflect.Value{reflect.ValueOf(dc.ctx)}
	}

	resultValue := method.Call(args)[0]
	resultType := resultValue.Type()

	// Value -> Ptr
	if dstFieldValue.Kind() == reflect.Ptr && force {
		ptr := reflect.New(resultType)
		ptr.Elem().Set(resultValue)

		setFieldValue(ptr.Type(), dstFieldType.Type, dstFieldValue, ptr, tagOptions)

		return nil
	}

	// Ptr -> value
	if resultValue.Kind() == reflect.Ptr && force {
		setFieldValue(resultValue.Elem().Type(), dstFieldType.Type, dstFieldValue, resultValue.Elem(), tagOptions)

		return nil
	}

	if resultValue.IsValid() {
		setFieldValue(resultType, dstFieldType.Type, dstFieldValue, resultValue, tagOptions)
	}

	return nil
}

func (dc *CopyStruct) copyFields(srcValue, dstValue reflect.Value, f string, dst interface{}, reversed bool) {
	srcFieldStruct, srcFieldFound := srcValue.Type().FieldByName(f)
	if !srcFieldFound {
		return
	}

	srcFieldValue := srcValue.FieldByName(f)
	srcFieldType := srcFieldStruct.Type
	srcFieldName := srcFieldStruct.Name

	dstFieldName, tagOptions := dc.parseDstFieldName(srcFieldName, reversed, srcFieldStruct, dst)
	if _, ok := tagOptions[optionSkip]; ok {
		return
	}

	dstStructField, dstFieldFound := dstValue.Type().FieldByName(dstFieldName)
	if !dstFieldFound {
		return
	}

	dstFieldValue := dstValue.FieldByName(dstFieldName)

	// Force option for empty interfaces and nullable types
	_, force := tagOptions[optionForce]

	dstKind := dstFieldValue.Kind()
	if isNullableType(srcFieldType) {
		if dstKind == reflect.Ptr && force { // Valuer -> ptr
			processNullableTypeValuer2Ptr(srcFieldValue, dstFieldValue, dstStructField, tagOptions)
		} else { // Valuer -> value
			processNullableTypeValuer2Value(srcFieldValue, dstFieldValue, dstStructField, force, tagOptions)
		}

		return
	}

	if dstKind == reflect.Interface {
		if force {
			dstFieldValue.Set(srcFieldValue)
		}

		return
	}

	// Ptr -> Value
	if srcFieldType.Kind() == reflect.Ptr && !srcFieldValue.IsNil() && dstKind != reflect.Ptr {
		srcFieldValue = reflect.Indirect(srcFieldValue)
		srcFieldType = srcFieldValue.Type()
	}

	setFieldValue(srcFieldType, dstStructField.Type, dstFieldValue, srcFieldValue, tagOptions)
}

func setFieldValue(srcFieldType reflect.Type, dstFieldType reflect.Type,
	dstFieldValue, srcFieldValue reflect.Value, tagOptions tagOptions) bool {
	if srcFieldType.AssignableTo(dstFieldType) {
		dstFieldValue.Set(srcFieldValue)
		return true
	}

	if _, ok := tagOptions[optionConvert]; ok && srcFieldType.ConvertibleTo(dstFieldType) {
		dstFieldValue.Set(srcFieldValue.Convert(dstFieldType))
		return true
	}

	return false
}

func processNullableTypeValuer2Ptr(srcFieldVal, dstFieldVal reflect.Value,
	dstStructField reflect.StructField, tagOptions tagOptions) {
	// We have same nullable type on both sides
	if setFieldValue(srcFieldVal.Type(), dstStructField.Type, dstFieldVal, srcFieldVal, tagOptions) {
		return
	}

	v, _ := srcFieldVal.Interface().(driver.Valuer).Value()
	if v == nil {
		return
	}

	valueType := reflect.TypeOf(v)

	ptr := reflect.New(valueType)
	ptr.Elem().Set(reflect.ValueOf(v))

	setFieldValue(valueType, dstStructField.Type.Elem(), dstFieldVal, ptr, tagOptions)
}

func processNullableTypeValuer2Value(srcFieldValue, dstFieldValue reflect.Value,
	dstFieldType reflect.StructField, force bool, tagOptions tagOptions) {
	// We have same nullable type on both sides
	if setFieldValue(srcFieldValue.Type(), dstFieldType.Type, dstFieldValue, srcFieldValue, tagOptions) {
		return
	}

	if !force {
		return
	}

	v, _ := srcFieldValue.Interface().(driver.Valuer).Value()
	if v == nil {
		return
	}

	rv := reflect.ValueOf(v)

	setFieldValue(rv.Type(), dstFieldType.Type, dstFieldValue, rv, tagOptions)
}

func (dc *CopyStruct) parseDstFieldName(srcFieldName string, reversed bool,
	srcStructField reflect.StructField, dst interface{}) (string, tagOptions) {
	dstFieldName := srcFieldName

	if reversed {
		tagOptions := parseTagOptions(srcStructField.Tag.Get(dc.tagName))
		if v, ok := tagOptions[optionField]; ok && v != "" {
			dstFieldName = v
		}

		return dstFieldName, tagOptions
	}

	if name, opts := dc.getRelatedField(dst, srcFieldName); name != "" {
		return name, opts
	}

	return dstFieldName, tagOptions{}
}

// parseTagOptions parses deepcopier tag field and returns context.
// nolint:gomnd
func parseTagOptions(value string) tagOptions {
	options := tagOptions{}

	for _, opt := range strings.Split(value, ";") {
		o := strings.SplitN(opt, ":", 2)

		switch len(o) { // nolint:gomnd
		case 1: // copystruct:"keyword; without; value;"
			options[o[0]] = ""
		case 2: // copystruct:"key:value; anotherkey:anothervalue"
			options[strings.TrimSpace(o[0])] = strings.TrimSpace(o[1])
		}
	}

	return options
}

// getRelatedField returns first matching field.
func (dc *CopyStruct) getRelatedField(instance interface{}, name string) (string, tagOptions) {
	value := reflect.Indirect(reflect.ValueOf(instance))
	fieldName := ""
	tagOptions := tagOptions{}

	for i := 0; i < value.NumField(); i++ {
		vField := value.Field(i)
		tField := value.Type().Field(i)
		tagOptions := parseTagOptions(tField.Tag.Get(dc.tagName))

		if tField.Type.Kind() == reflect.Struct && tField.Anonymous {
			if n, o := dc.getRelatedField(vField.Interface(), name); n != "" {
				return n, o
			}
		}

		if v, ok := tagOptions[optionField]; ok && v == name {
			return tField.Name, tagOptions
		}

		if tField.Name == name {
			return tField.Name, tagOptions
		}
	}

	return fieldName, tagOptions
}

// getMethodNames returns instance's method names.
func getMethodNames(instance interface{}) []string {
	t := reflect.TypeOf(instance)
	methods := make([]string, t.NumMethod())

	for i := 0; i < t.NumMethod(); i++ {
		methods[i] = t.Method(i).Name
	}

	return methods
}

// getFieldNames returns instance's field names.
func getFieldNames(instance interface{}) []string {
	v := reflect.Indirect(reflect.ValueOf(instance))
	t := v.Type()

	if t.Kind() != reflect.Struct {
		return nil
	}

	fields := make([]string, 0, v.NumField())

	for i := 0; i < v.NumField(); i++ {
		vField := v.Field(i)
		tField := v.Type().Field(i)

		// Is exportable?
		if tField.PkgPath != "" {
			continue
		}

		if tField.Type.Kind() == reflect.Struct && tField.Anonymous {
			fields = append(fields, getFieldNames(vField.Interface())...)
		} else {
			fields = append(fields, tField.Name)
		}
	}

	return fields
}

// isNullableType returns true if the given type is a nullable one.
func isNullableType(t reflect.Type) bool {
	return t.ConvertibleTo(reflect.TypeOf((*driver.Valuer)(nil)).Elem())
}
