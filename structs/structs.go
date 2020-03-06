// Package structs contains various utilities functions to work with structs.
package structs

import (
	"fmt"

	"github.com/bingoohuang/goreflect"

	"reflect"
)

// Struct encapsulates a struct type to provide several high level functions
// around the struct.
type Struct struct {
	raw    interface{}
	value  reflect.Value
	Option *Option
}

// Option is the options for Validate.
type Option struct {
	TagName string

	OmitNested bool
	OmitEmpty  bool
	Stringer   bool
	Flatten    bool
}

// OptionFn is the function prototype to apply option
type OptionFn func(*Option)

// TagName defines the tag name for validate.
func TagName(s string) OptionFn { return func(o *Option) { o.TagName = s } }

// OmitNested tell the processor omit nested or not.
func OmitNested(b bool) OptionFn { return func(o *Option) { o.OmitNested = b } }

// OmitEmpty tell the processor omit empty or not.
func OmitEmpty(b bool) OptionFn { return func(o *Option) { o.OmitEmpty = b } }

// Stringer tell the processor use Stringer or not.
func Stringer(b bool) OptionFn { return func(o *Option) { o.Stringer = b } }

func createOption(optionFns []OptionFn) *Option {
	option := &Option{}

	for _, fn := range optionFns {
		fn(option)
	}

	if option.TagName == "" {
		// structs is the default tag name for struct fields which provides
		// a more granular to tweak certain structs. Lookup the necessary functions
		// for more info.  struct's field default tag name
		option.TagName = "structs"
	}

	return option
}

// New returns a new *Struct with the struct s. It panics if the s's kind is
// not struct.
func New(s interface{}, optionFns ...OptionFn) *Struct {
	option := createOption(optionFns)

	return &Struct{
		raw:    s,
		value:  strctVal(s),
		Option: option,
	}
}

// Map converts the given struct to a map[string]interface{}, where the keys
// of the map are the field names and the values of the map the associated
// values of the fields. The default key string is the struct field name but
// can be changed in the struct field's tag value. The "structs" key in the
// struct's field tag value is the key name. Example:
//
//   // Field appears in map as key "myName".
//   Name string `structs:"myName"`
//
// A tag value with the content of "-" ignores that particular field. Example:
//
//   // Field is ignored by this package.
//   Field bool `structs:"-"`
//
// A tag value with the content of "string" uses the stringer to get the value. Example:
//
//   // The value will be output of Animal's Stringer() func.
//   // Map will panic if Animal does not implement Stringer().
//   Field *Animal `structs:"field,string"`
//
// A tag value with the option of "flatten" used in a struct field is to flatten its fields
// in the output map. Example:
//
//   // The FieldStruct's fields will be flattened into the output map.
//   FieldStruct time.Time `structs:",flatten"`
//
// A tag value with the option of "omitnested" stops iterating further if the type
// is a struct. Example:
//
//   // Field is not processed further by this package.
//   Field time.Time     `structs:"myName,omitnested"`
//   Field *http.Request `structs:",omitnested"`
//
// A tag value with the option of "omitempty" ignores that particular field if
// the field value is empty. Example:
//
//   // Field appears in map as key "myName", but the field is
//   // skipped if empty.
//   Field string `structs:"myName,omitempty"`
//
//   // Field appears in map as key "Field" (the default), but
//   // the field is skipped if empty.
//   Field string `structs:",omitempty"`
//
// Note that only exported fields of a struct can be accessed, non exported
// fields will be neglected.
func (s *Struct) Map() map[string]interface{} {
	out := make(map[string]interface{})
	s.FillMap(out)

	return out
}

// FillMap is the same as Map. Instead of returning the output, it fills the
// given map.
func (s *Struct) FillMap(out map[string]interface{}) {
	if out == nil {
		return
	}

	fields := s.structFields()

	for _, field := range fields {
		name := field.Name
		val := s.value.FieldByName(name)
		isSubStruct := false

		var finalVal interface{}

		tagName, tagOpts := parseTag(s.Option, field.Tag.Get(s.Option.TagName))
		if tagName != "" {
			name = tagName
		}

		// if the value is a zero value and the field is marked as omitempty do
		// not include
		if tagOpts.OmitEmpty() && goreflect.IsEmptyValue(val) {
			continue
		}

		if tagOpts.OmitNested() {
			finalVal = val.Interface()
		} else {
			finalVal = s.nested(val)

			v := reflect.ValueOf(val.Interface())
			if v.Kind() == reflect.Ptr {
				v = v.Elem()
			}

			switch v.Kind() {
			case reflect.Map, reflect.Struct:
				isSubStruct = true
			}
		}

		if tagOpts.Stringer() {
			v := val.Interface()
			if s, ok := v.(fmt.Stringer); ok {
				out[name] = s.String()
			} else {
				out[name] = fmt.Sprintf("%v", v)
			}

			continue
		}

		if isSubStruct && tagOpts.Flatten() {
			for k := range finalVal.(map[string]interface{}) {
				out[k] = finalVal.(map[string]interface{})[k]
			}
		} else {
			out[name] = finalVal
		}
	}
}

// Values converts the given s struct's field values to a []interface{}.  A
// struct tag with the content of "-" ignores the that particular field.
// Example:
//
//   // Field is ignored by this package.
//   Field int `structs:"-"`
//
// A value with the option of "omitnested" stops iterating further if the type
// is a struct. Example:
//
//   // Fields is not processed further by this package.
//   Field time.Time     `structs:",omitnested"`
//   Field *http.Request `structs:",omitnested"`
//
// A tag value with the option of "omitempty" ignores that particular field and
// is not added to the values if the field value is empty. Example:
//
//   // Field is skipped if empty
//   Field string `structs:",omitempty"`
//
// Note that only exported fields of a struct can be accessed, non exported
// fields  will be neglected.
func (s *Struct) Values() []interface{} {
	fields := s.structFields()
	t := make([]interface{}, 0, len(fields))

	for _, field := range fields {
		val := s.value.FieldByName(field.Name)
		_, tagOpts := parseTag(s.Option, field.Tag.Get(s.Option.TagName))

		// if the value is a zero value and the field is marked as omitempty do not include
		if tagOpts.OmitEmpty() && goreflect.IsEmptyValue(val) {
			continue
		}

		if tagOpts.Stringer() {
			v := val.Interface()
			if s, ok := v.(fmt.Stringer); ok {
				t = append(t, s.String())
			} else {
				t = append(t, fmt.Sprintf("%v", v))
			}

			continue
		}

		if IsStruct(val.Interface()) && !tagOpts.OmitNested() {
			// look out for embedded structs, and convert them to a
			// []interface{} to be added to the final values slice
			t = append(t, Values(val.Interface())...)
		} else {
			t = append(t, val.Interface())
		}
	}

	return t
}

// Fields returns a slice of Fields. A struct tag with the content of "-"
// ignores the checking of that particular field. Example:
//
//   // Field is ignored by this package.
//   Field bool `structs:"-"`
//
// It panics if s's kind is not struct.
func (s *Struct) Fields() []*Field {
	return getFields(s.value, s.Option.TagName)
}

// Names returns a slice of field names. A struct tag with the content of "-"
// ignores the checking of that particular field. Example:
//
//   // Field is ignored by this package.
//   Field bool `structs:"-"`
//
// It panics if s's kind is not struct.
func (s *Struct) Names() []string {
	fields := getFields(s.value, s.Option.TagName)

	names := make([]string, len(fields))

	for i, field := range fields {
		names[i] = field.Name()
	}

	return names
}

func getFields(v reflect.Value, tagName string) []*Field {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	t := v.Type()

	var fields []*Field

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		if tag := field.Tag.Get(tagName); tag == "-" {
			continue
		}

		f := &Field{
			field: field,
			value: v.FieldByName(field.Name),
		}

		fields = append(fields, f)
	}

	return fields
}

// Field returns a new Field struct that provides several high level functions
// around a single struct field entity. It panics if the field is not found.
func (s *Struct) Field(name string) *Field {
	f, ok := s.FieldOk(name)
	if !ok {
		panic("field not found")
	}

	return f
}

// FieldOK returns a new Field struct that provides several high level functions
// around a single struct field entity. The boolean returns true if the field
// was found.
func (s *Struct) FieldOk(name string) (*Field, bool) {
	t := s.value.Type()

	field, ok := t.FieldByName(name)
	if !ok {
		return nil, false
	}

	return &Field{
		field:      field,
		value:      s.value.FieldByName(name),
		defaultTag: s.Option.TagName,
	}, true
}

// IsZero returns true if all fields in a struct is a zero value (not
// initialized) A struct tag with the content of "-" ignores the checking of
// that particular field. Example:
//
//   // Field is ignored by this package.
//   Field bool `structs:"-"`
//
// A value with the option of "omitnested" stops iterating further if the type
// is a struct. Example:
//
//   // Field is not processed further by this package.
//   Field time.Time     `structs:"myName,omitnested"`
//   Field *http.Request `structs:",omitnested"`
//
// Note that only exported fields of a struct can be accessed, non exported
// fields  will be neglected. It panics if s's kind is not struct.
func (s *Struct) IsZero() bool {
	fields := s.structFields()

	for _, field := range fields {
		val := s.value.FieldByName(field.Name)

		_, tagOpts := parseTag(s.Option, field.Tag.Get(s.Option.TagName))

		if IsStruct(val.Interface()) && !tagOpts.OmitNested() {
			ok := IsZero(val.Interface())
			if !ok {
				return false
			}

			continue
		}

		// zero value of the given field, such as "" for string, 0 for int
		zero := reflect.Zero(val.Type()).Interface()

		//  current value of the given field
		current := val.Interface()

		if !reflect.DeepEqual(current, zero) {
			return false
		}
	}

	return true
}

// HasZero returns true if a field in a struct is not initialized (zero value).
// A struct tag with the content of "-" ignores the checking of that particular
// field. Example:
//
//   // Field is ignored by this package.
//   Field bool `structs:"-"`
//
// A value with the option of "omitnested" stops iterating further if the type
// is a struct. Example:
//
//   // Field is not processed further by this package.
//   Field time.Time     `structs:"myName,omitnested"`
//   Field *http.Request `structs:",omitnested"`
//
// Note that only exported fields of a struct can be accessed, non exported
// fields  will be neglected. It panics if s's kind is not struct.
func (s *Struct) HasZero() bool {
	fields := s.structFields()

	for _, field := range fields {
		val := s.value.FieldByName(field.Name)

		_, tagOpts := parseTag(s.Option, field.Tag.Get(s.Option.TagName))

		if IsStruct(val.Interface()) && !tagOpts.OmitNested() {
			ok := HasZero(val.Interface())
			if ok {
				return true
			}

			continue
		}

		// zero value of the given field, such as "" for string, 0 for int
		zero := reflect.Zero(val.Type()).Interface()

		//  current value of the given field
		current := val.Interface()

		if reflect.DeepEqual(current, zero) {
			return true
		}
	}

	return false
}

// Name returns the structs's type name within its package. For more info refer
// to Name() function.
func (s *Struct) Name() string {
	return s.value.Type().Name()
}

// structFields returns the exported struct fields for a given s struct. This
// is a convenient helper method to avoid duplicate code in some of the functions.
func (s *Struct) structFields() []reflect.StructField {
	t := s.value.Type()

	f := make([]reflect.StructField, 0, t.NumField())

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		// we can't access the value of unexported fields
		if field.PkgPath != "" {
			continue
		}

		// don't check if it's omitted
		if tag := field.Tag.Get(s.Option.TagName); tag == "-" {
			continue
		}

		f = append(f, field)
	}

	return f
}

func strctVal(s interface{}) reflect.Value {
	v := goreflect.IndirectAll(reflect.ValueOf(s))

	if v.Kind() != reflect.Struct {
		panic("not struct")
	}

	return v
}

// Map converts the given struct to a map[string]interface{}. For more info
// refer to Struct types Map() method. It panics if s's kind is not struct.
func Map(s interface{}, optionFns ...OptionFn) map[string]interface{} {
	return New(s, optionFns...).Map()
}

// MapString converts the given struct to a map[string]string. For more info
// refer to Struct types Map() method. It panics if s's kind is not struct.
func MapString(s interface{}, optionFns ...OptionFn) map[string]string {
	fns := make([]OptionFn, 0, len(optionFns)+1) // nolint gomnd
	fns = append(fns, optionFns...)
	fns = append(fns, Stringer(true))

	m := make(map[string]string)

	for k, v := range New(s, fns...).Map() {
		m[k] = v.(string)
	}

	return m
}

// FillMap is the same as Map. Instead of returning the output, it fills the
// given map.
func FillMap(s interface{}, out map[string]interface{}, optionFns ...OptionFn) {
	New(s, optionFns...).FillMap(out)
}

// Values converts the given struct to a []interface{}. For more info refer to
// Struct types Values() method.  It panics if s's kind is not struct.
func Values(s interface{}, optionFns ...OptionFn) []interface{} {
	return New(s, optionFns...).Values()
}

// Fields returns a slice of *Field. For more info refer to Struct types
// Fields() method.  It panics if s's kind is not struct.
func Fields(s interface{}, optionFns ...OptionFn) []*Field {
	return New(s, optionFns...).Fields()
}

// Names returns a slice of field names. For more info refer to Struct types
// Names() method.  It panics if s's kind is not struct.
func Names(s interface{}, optionFns ...OptionFn) []string {
	return New(s, optionFns...).Names()
}

// IsZero returns true if all fields is equal to a zero value. For more info
// refer to Struct types IsZero() method.  It panics if s's kind is not struct.
func IsZero(s interface{}, optionFns ...OptionFn) bool {
	return New(s, optionFns...).IsZero()
}

// HasZero returns true if any field is equal to a zero value. For more info
// refer to Struct types HasZero() method.  It panics if s's kind is not struct.
func HasZero(s interface{}, optionFns ...OptionFn) bool {
	return New(s, optionFns...).HasZero()
}

// IsStruct returns true if the given variable is a struct or a pointer to
// struct.
func IsStruct(s interface{}) bool {
	v := reflect.Indirect(reflect.ValueOf(s))

	return v.Kind() == reflect.Struct
}

// Name returns the structs's type name within its package. It returns an
// empty string for unnamed types. It panics if s's kind is not struct.
func Name(s interface{}, optionFns ...OptionFn) string {
	return New(s, optionFns...).Name()
}

// nested retrieves recursively all types for the given value and returns the
// nested value.
func (s *Struct) nested(val reflect.Value) interface{} {
	v := reflect.Indirect(reflect.ValueOf(val.Interface()))

	switch v.Kind() {
	case reflect.Struct:
		n := New(val.Interface())
		n.Option = s.Option
		m := n.Map()

		// do not add the converted value if there are no exported fields, ie: time.Time
		if len(m) > 0 {
			return m
		}
	case reflect.Map:
		// get the element type of the map
		mapElem := val.Type()

		switch mapElem.Kind() {
		case reflect.Ptr, reflect.Array, reflect.Map, reflect.Slice, reflect.Chan:
			mapElem = mapElem.Elem()
			if mapElem.Kind() == reflect.Ptr {
				mapElem = mapElem.Elem()
			}
		}

		// only iterate over struct types, ie: map[string]StructType,
		// map[string][]StructType,
		if mapElem.Kind() == reflect.Struct ||
			(mapElem.Kind() == reflect.Slice && mapElem.Elem().Kind() == reflect.Struct) {
			m := make(map[string]interface{}, val.Len())
			for _, k := range val.MapKeys() {
				m[k.String()] = s.nested(val.MapIndex(k))
			}

			return m
		}
	case reflect.Slice, reflect.Array:
		if val.Type().Kind() == reflect.Interface {
			return val.Interface()
		}

		// do not iterate of non struct types, just pass the value. i. e: []int,
		// []string, co... We only iterate further if it's a struct. i.e []foo or []*foo
		vte := val.Type().Elem()
		if vte.Kind() != reflect.Struct &&
			!(vte.Kind() == reflect.Ptr && vte.Elem().Kind() == reflect.Struct) {
			return val.Interface()
		}

		slices := make([]interface{}, val.Len())
		for x := 0; x < val.Len(); x++ {
			slices[x] = s.nested(val.Index(x))
		}

		return slices
	}

	return val.Interface()
}
