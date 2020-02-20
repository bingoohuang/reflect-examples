package copystruct

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"strings"
)

const (
	// tagName is the deepcopier struct tag name.
	tagName = "copystruct"
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

// tagOptions is a map that contains extracted struct tag options.
type tagOptions map[string]string

// options are copier options.
type options struct {
	// Context given to WithContext() method.
	Context map[string]interface{}
	// Reversed reverses struct tag checkings.
	Reversed bool
}

// DeepCopier deep copies a struct to/from a struct.
type DeepCopier struct {
	dst interface{}
	src interface{}
	ctx map[string]interface{}
}

// Copy sets source or destination.
func Copy(src interface{}) *DeepCopier { return &DeepCopier{src: src} }

// WithContext injects the given context into the builder instance.
func (dc *DeepCopier) WithContext(ctx map[string]interface{}) *DeepCopier {
	dc.ctx = ctx
	return dc
}

// To sets the destination.
func (dc *DeepCopier) To(dst interface{}) error {
	dc.dst = dst

	return process(dc.dst, dc.src, options{Context: dc.ctx})
}

// From sets the given the source as destination and destination as source.
func (dc *DeepCopier) From(src interface{}) error {
	dc.dst = dc.src
	dc.src = src

	return process(dc.dst, dc.src, options{Context: dc.ctx, Reversed: true})
}

// process handles copy.
func process(dst, src interface{}, options options) error {
	dstValue := reflect.Indirect(reflect.ValueOf(dst))
	if !dstValue.CanAddr() {
		return fmt.Errorf("destination %+v is unaddressable", dstValue.Interface())
	}

	srcValue := reflect.Indirect(reflect.ValueOf(src))

	for _, f := range getFieldNames(src) {
		copyFields(srcValue, dstValue, f, options, dst)
	}

	for _, m := range getMethodNames(src) {
		if err := copyMethods(dstValue, src, dst, m, options); err != nil {
			return err
		}
	}

	return nil
}

func copyMethods(dstValue reflect.Value, src, dst interface{}, m string, opts options) error {
	name, tagOptions := getRelatedField(dst, m)
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
		args = []reflect.Value{reflect.ValueOf(opts.Context)}
	}

	resultValue := method.Call(args)[0]
	resultType := resultValue.Type()

	// Value -> Ptr
	if dstFieldValue.Kind() == reflect.Ptr && force {
		ptr := reflect.New(resultType)
		ptr.Elem().Set(resultValue)

		if ptr.Type().AssignableTo(dstFieldType.Type) {
			dstFieldValue.Set(ptr)
		}

		return nil
	}

	// Ptr -> value
	if resultValue.Kind() == reflect.Ptr && force {
		if resultValue.Elem().Type().AssignableTo(dstFieldType.Type) {
			dstFieldValue.Set(resultValue.Elem())
		}

		return nil
	}

	if resultValue.IsValid() {
		if resultType.AssignableTo(dstFieldType.Type) {
			dstFieldValue.Set(resultValue)
			return nil
		}

		if _, ok := tagOptions[optionConvert]; ok && resultType.ConvertibleTo(dstFieldType.Type) {
			dstFieldValue.Set(resultValue.Convert(dstFieldType.Type))
		}
	}
	return nil
}

func copyFields(srcValue, dstValue reflect.Value, f string, opts options, dst interface{}) {
	srcFieldType, srcFieldFound := srcValue.Type().FieldByName(f)
	if !srcFieldFound {
		return
	}

	srcFieldValue := srcValue.FieldByName(f)
	srcFieldName := srcFieldType.Name

	dstFieldName, tagOptions := parseDstFieldName(srcFieldName, opts, srcFieldType, dst)
	if _, ok := tagOptions[optionSkip]; ok {
		return
	}

	dstFieldType, dstFieldFound := dstValue.Type().FieldByName(dstFieldName)
	if !dstFieldFound {
		return
	}

	dstFieldValue := dstValue.FieldByName(dstFieldName)

	// Force option for empty interfaces and nullable types
	_, force := tagOptions[optionForce]

	dstKind := dstFieldValue.Kind()
	if isNullableType(srcFieldType.Type) {
		if dstKind == reflect.Ptr && force { // Valuer -> ptr
			processNullableTypeValuer2Ptr(srcFieldValue, dstFieldValue, dstFieldType)
		} else { // Valuer -> value
			processNullableTypeValuer2Value(srcFieldValue, dstFieldValue, dstFieldType, force)
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
	if srcFieldType.Type.Kind() == reflect.Ptr && !srcFieldValue.IsNil() && dstKind != reflect.Ptr {
		indirect := reflect.Indirect(srcFieldValue)

		if indirect.Type().AssignableTo(dstFieldType.Type) {
			dstFieldValue.Set(indirect)

			return
		}
	}

	// Other types
	if srcFieldType.Type.AssignableTo(dstFieldType.Type) {
		dstFieldValue.Set(srcFieldValue)
		return
	}

	if _, ok := tagOptions[optionConvert]; ok && srcFieldType.Type.ConvertibleTo(dstFieldType.Type) {
		dstFieldValue.Set(srcFieldValue.Convert(dstFieldType.Type))
	}
}

func processNullableTypeValuer2Ptr(srcFieldValue, dstFieldValue reflect.Value, dstFieldType reflect.StructField) {
	// We have same nullable type on both sides
	if srcFieldValue.Type().AssignableTo(dstFieldType.Type) {
		dstFieldValue.Set(srcFieldValue)
		return
	}

	v, _ := srcFieldValue.Interface().(driver.Valuer).Value()
	if v == nil {
		return
	}

	valueType := reflect.TypeOf(v)

	ptr := reflect.New(valueType)
	ptr.Elem().Set(reflect.ValueOf(v))

	if valueType.AssignableTo(dstFieldType.Type.Elem()) {
		dstFieldValue.Set(ptr)
	}
}

func processNullableTypeValuer2Value(srcFieldValue, dstFieldValue reflect.Value,
	dstFieldType reflect.StructField, force bool) {
	// We have same nullable type on both sides
	if srcFieldValue.Type().AssignableTo(dstFieldType.Type) {
		dstFieldValue.Set(srcFieldValue)
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
	if rv.Type().AssignableTo(dstFieldType.Type) {
		dstFieldValue.Set(rv)
	}
}

func parseDstFieldName(srcFieldName string, options options,
	srcFieldType reflect.StructField, dst interface{}) (string, tagOptions) {
	dstFieldName := srcFieldName

	if options.Reversed {
		tagOptions := parseTagOptions(srcFieldType.Tag.Get(tagName))
		if v, ok := tagOptions[optionField]; ok && v != "" {
			dstFieldName = v
		}

		return dstFieldName, tagOptions
	}

	if name, opts := getRelatedField(dst, srcFieldName); name != "" {
		return name, opts
	}

	return dstFieldName, tagOptions{}
}

// parseTagOptions parses deepcopier tag field and returns options.
// nolint gomnd
func parseTagOptions(value string) tagOptions {
	options := tagOptions{}

	for _, opt := range strings.Split(value, ";") {
		o := strings.SplitN(opt, ":", 2)

		switch len(o) { // nolint gomnd
		case 1: // copystruct:"keyword; without; value;"
			options[o[0]] = ""
		case 2: // copystruct:"key:value; anotherkey:anothervalue"
			options[strings.TrimSpace(o[0])] = strings.TrimSpace(o[1])
		}
	}

	return options
}

// getRelatedField returns first matching field.
func getRelatedField(instance interface{}, name string) (string, tagOptions) {
	value := reflect.Indirect(reflect.ValueOf(instance))
	fieldName := ""
	tagOptions := tagOptions{}

	for i := 0; i < value.NumField(); i++ {
		vField := value.Field(i)
		tField := value.Type().Field(i)
		tagOptions := parseTagOptions(tField.Tag.Get(tagName))

		if tField.Type.Kind() == reflect.Struct && tField.Anonymous {
			if n, o := getRelatedField(vField.Interface(), name); n != "" {
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
