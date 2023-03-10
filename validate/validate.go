package validate

import (
	"reflect"
	"regexp"
	"strings"

	"github.com/bingoohuang/gor/defaults"
)

// Validator is an interface for a validated struct.
type Validator interface {
	// Validate is a custom validation function.
	// Validate does not work when the receiver is a reference.
	// Validate does not work for nested types obtained from unexported field.
	Validate() error
}

// Option is the options for Validate.
type Option struct {
	TagName string `default:"validate"`
}

// OptionFn is the function prototype to apply option
type OptionFn func(*Option)

// TagName defines the tag name for validate.
func TagName(tagName string) OptionFn { return func(o *Option) { o.TagName = tagName } }

// Validate validates fields of a struct.
// It accepts a struct or a struct pointer as a parameter.
// It returns an error if a struct does not validate or nil if there are no validation errors.
//
//	err := validate.Validate(struct {
//		field time.Duration `validate:"gte=0s"`
//	}{
//		field: -time.Second,
//	})
//
//	// err contains an error
func Validate(element interface{}, optionFns ...OptionFn) error {
	option, err := createOption(optionFns)
	if err != nil {
		return err
	}

	value := reflect.ValueOf(element)

	return option.validateField(value, "", "")
}

func createOption(optionFns []OptionFn) (*Option, error) {
	option := &Option{}
	if err := defaults.Set(option); err != nil {
		return nil, err
	}

	for _, fn := range optionFns {
		fn(option)
	}

	return option, nil
}

// validateField validates a struct field
// nolint:gocognit
func (o *Option) validateField(value reflect.Value, fieldName string, validators string) error {
	kind := value.Kind()

	// Get validator type Map
	validatorTypeMap := getValidatorTypeMap()

	// Get validators
	keyValidators, valueValidators, validators, err := splitValidators(validators)
	if err != nil {
		err = setFieldName(err, fieldName)
		return err
	}

	// Call a custom validator
	if err := callCustomValidator(value); err != nil {
		return err
	}

	// Parse validators
	validatorsOr, err := parseValidators(valueValidators)
	if err != nil {
		err = setFieldName(err, fieldName)
		return err
	}

	// Perform validators
	for _, validatorsAnd := range validatorsOr {
		for _, validator := range validatorsAnd {
			if validatorFunc, ok := validatorTypeMap[validator.Type]; ok {
				if err = validatorFunc(value, validator.Value); err != nil {
					err = setFieldName(err, fieldName)
					break
				}
			} else {
				return ErrorSyntax{
					fieldName:  fieldName,
					expression: string(validator.Type),
					near:       valueValidators,
					comment:    "could not find a validator",
				}
			}
		}
		if err == nil {
			break
		}
	}

	if err != nil {
		return err
	}

	// Dive one level deep into arrays and pointers
	switch kind {
	case reflect.Struct:
		if err := o.validateStruct(value); err != nil {
			return err
		}
	case reflect.Map:
		for _, key := range value.MapKeys() {
			if err := o.validateField(key, fieldName, keyValidators); err != nil {
				return err
			}
			if err := o.validateField(value.MapIndex(key), fieldName, validators); err != nil {
				return err
			}
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < value.Len(); i++ {
			if err := o.validateField(value.Index(i), fieldName, validators); err != nil {
				return err
			}
		}
	case reflect.Ptr:
		if !value.IsNil() {
			if err := o.validateField(value.Elem(), fieldName, validators); err != nil {
				return err
			}
		}
	}

	if kind != reflect.Map {
		if len(keyValidators) > 0 {
			return ErrorSyntax{
				fieldName:  fieldName,
				expression: validators,
				near:       "",
				comment:    "unexpexted expression",
			}
		}
	}

	if kind != reflect.Map && kind != reflect.Slice && kind != reflect.Array && kind != reflect.Ptr {
		if len(validators) > 0 {
			return ErrorSyntax{
				fieldName:  fieldName,
				expression: validators,
				near:       "",
				comment:    "unexpexted expression",
			}
		}
	}

	return nil
}

// validateStruct validates a struct
func (o *Option) validateStruct(value reflect.Value) error {
	typ := value.Type()

	// Iterate over struct fields
	for i := 0; i < typ.NumField(); i++ {
		validators := o.getValidators(typ.Field(i).Tag)
		fieldName := typ.Field(i).Name

		if err := o.validateField(value.Field(i), fieldName, validators); err != nil {
			return err
		}
	}

	return nil
}

// getValidators gets validators
func (o *Option) getValidators(tag reflect.StructTag) string {
	return tag.Get(o.TagName)
}

// splitValidators splits validators into key validators, value validators and remaning validators of the next level
// nolint:nakedret
func splitValidators(validators string) (keyValidators string, valValidators string, remaningValidators string, err ErrorField) {
	gt := 0
	bracket := 0
	bracketStart := 0
	bracketEnd := -1

	i := 0
loop:
	for ; i < len(validators); i++ {
		switch validators[i] {
		case '>':
			if bracket == 0 {
				gt++
				break loop
			}
		case '[':
			if bracket == 0 {
				bracketStart = i
			}
			bracket++
		case ']':
			bracket--
			if bracket == 0 {
				bracketEnd = i
			}
		}
	}

	if bracket > 0 {
		err = ErrorSyntax{
			expression: "",
			near:       validators,
			comment:    "expected \"]\"",
		}

		return
	}

	if bracket < 0 {
		err = ErrorSyntax{
			expression: "",
			near:       validators,
			comment:    "unexpected \"]\"",
		}

		return
	}

	if bracketStart <= len(validators) {
		valValidators += validators[:bracketStart]
	}

	if bracketEnd+1 <= len(validators) {
		if valValidators != "" {
			valValidators += " "
		}

		valValidators += validators[bracketEnd+1 : i]
	}

	if bracketStart+1 <= len(validators) && bracketEnd >= 0 && bracketStart+1 <= bracketEnd {
		keyValidators = validators[bracketStart+1 : bracketEnd]
	}

	if i+1 <= len(validators) {
		remaningValidators = validators[i+1:]
	}

	if gt > 0 && len(remaningValidators) == 0 {
		err = ErrorSyntax{
			expression: "",
			near:       validators,
			comment:    "expected expression",
		}

		return
	}

	return
}

// parseValidator parses validators into the slice of slices.
// First slice acts as AND logic, second array acts as OR logic.
func parseValidators(validators string) (validatorsOr [][]validator, err ErrorField) {
	regexpType := regexp.MustCompile(`[[:alnum:]_]+`)
	regexpValue := regexp.MustCompile(`[^=\s]+[^=]*[^=\s]+|[^=\s]+`)

	if len(validators) == 0 {
		return
	}

	entriesOr := strings.Split(validators, "|")
	validatorsOr = make([][]validator, 0, len(entriesOr))

	for _, entryOr := range entriesOr {
		entriesAnd := strings.Split(entryOr, "&")
		validatorsAnd := make([]validator, 0, len(entriesAnd))

		for _, entryOr := range entriesAnd {
			entries := strings.Split(entryOr, "=")
			if len(entries) == 0 || len(entries) > 2 {
				err = ErrorSyntax{
					expression: validators,
					comment:    "could not parse",
				}

				return validatorsOr, err
			}

			t := regexpType.FindString(entries[0])
			if len(t) == 0 {
				err = ErrorSyntax{
					expression: entries[0],
					near:       validators,
					comment:    "could not parse",
				}

				return
			}

			v := ""

			if len(entries) == 2 { // nolint:gomnd
				v = regexpValue.FindString(entries[1])
			}

			validatorsAnd = append(validatorsAnd, validator{Type(t), v})
		}

		if len(validatorsAnd) > 0 {
			validatorsOr = append(validatorsOr, validatorsAnd)
		}
	}

	return validatorsOr, err
}

// parseTokens parses tokens into array
func parseTokens(str string) []interface{} {
	tokenStrings := strings.Split(str, ",")
	tokens := make([]interface{}, 0, len(tokenStrings))

	for i := range tokenStrings {
		token := strings.TrimSpace(tokenStrings[i])
		if token != "" {
			tokens = append(tokens, token)
		}
	}

	return tokens
}

// tokenOneOf check if a token is one of tokens
func tokenOneOf(token interface{}, tokens []interface{}) bool {
	for _, t := range tokens {
		if t == token {
			return true
		}
	}

	return false
}

// Call a custom validator
func callCustomValidator(value reflect.Value) error {
	if !value.CanInterface() {
		return nil
	}

	// Following code won't work in case if Validate is implemented by reference and value is passed by value
	if customValidator, ok := value.Interface().(Validator); ok {
		return customValidator.Validate()
	}

	// Following code is a fallback if value is passed by value
	valueCopyPointer := reflect.New(value.Type())

	valueCopyPointer.Elem().Set(value)

	if customValidator, ok := valueCopyPointer.Interface().(Validator); ok {
		return customValidator.Validate()
	}

	return nil
}
