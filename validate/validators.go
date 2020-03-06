package validate

import (
	"reflect"
	"strconv"
	"time"
)

// Type is used for validator type definitions.
type Type string

// Following validators are available.
// nolint lll
const (
	// Eq (equals) compares a numeric value of a number or compares a count of elements in a string, a map, a slice, or an array.
	// E.g. `validate:"eq=1"`
	Eq Type = "eq"

	// Ne (not equals) compares a numeric value of a number or compares a count of elements in a string, a map, a slice, or an array.
	// E.g. `validate:"ne=0"`
	Ne = "ne"

	// Gt (greater than) compares a numeric value of a number or compares a count of elements in a string, a map, a slice, or an array.
	// E.g. `validate:"gt=-1"`
	Gt = "gt"

	// Lt (less than) compares a numeric value of a number or compares a count of elements in a string, a map, a slice, or an array.
	// E.g. `validate:"lt=11"`
	Lt = "lt"

	// Gte (greater than or equal to) compares a numeric value of a number or compares a count of elements in a string, a map, a slice, or an array.
	// E.g. `validate:"gte=0"`
	Gte = "gte"

	// Lte (less than or equal to) compares a numeric value of a number or compares a count of elements in a string, a map, a slice, or an array.
	// E.g. `validate:"lte=10"`
	Lte = "lte"

	// Empty checks if a string, a map, a slice, or an array is (not) empty.
	// E.g. `validate:"empty=false"`
	Empty = "empty"

	// Nil checks if a pointer is (not) nil.
	// E.g. `validate:"nil=false"`
	Nil = "nil"

	// Enum checks if a number or a string contains any of the given elements.
	// E.g. `validate:"enum=1,2,3"`
	Enum = "enum"

	// Format checks if a string of a given format.
	// E.g. `validate:"format=email"`
	Format = "format"
)

// validatorFunc is an interface for validator func
type validatorFunc func(value reflect.Value, validator string) ErrorField

func getValidatorTypeMap() map[Type]validatorFunc {
	return map[Type]validatorFunc{
		Eq:     validateEq,
		Ne:     validateNe,
		Gt:     validateGt,
		Lt:     validateLt,
		Gte:    validateGte,
		Lte:    validateLte,
		Empty:  validateEmpty,
		Nil:    validateNil,
		Enum:   validateOneOf,
		Format: validateFormat,
	}
}

type validator struct {
	Type  Type
	Value string
}

// nolint dupl
func validateEq(value reflect.Value, validator string) ErrorField {
	kind := value.Kind()
	typ := value.Type()

	errorValidation := ErrorValidation{
		fieldValue:     value,
		validatorType:  Eq,
		validatorValue: validator,
	}

	errorSyntax := ErrorSyntax{
		expression: validator,
		near:       string(Eq),
		comment:    "could not parse or run",
	}

	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if typ == reflect.TypeOf((time.Duration)(0)) {
			if token, err := time.ParseDuration(validator); err != nil {
				return errorSyntax
			} else if time.Duration(value.Int()) != token {
				return errorValidation
			}
		} else {
			if token, err := strconv.ParseInt(validator, 10, 64); err != nil {
				return errorSyntax
			} else if value.Int() != token {
				return errorValidation
			}
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		if token, err := strconv.ParseUint(validator, 10, 64); err != nil {
			return errorSyntax
		} else if value.Uint() != token {
			return errorValidation
		}
	case reflect.Float32, reflect.Float64:
		if token, err := strconv.ParseFloat(validator, 64); err != nil {
			return errorSyntax
		} else if value.Float() != token {
			return errorValidation
		}
	case reflect.String, reflect.Map, reflect.Slice, reflect.Array:
		if token, err := strconv.Atoi(validator); err != nil {
			return errorSyntax
		} else if value.Len() != token {
			return errorValidation
		}
	default:
		return errorSyntax
	}

	return nil
}

// nolint dupl
func validateNe(value reflect.Value, validator string) ErrorField {
	kind := value.Kind()
	typ := value.Type()

	errorValidation := ErrorValidation{
		fieldValue:     value,
		validatorType:  Ne,
		validatorValue: validator,
	}

	errorSyntax := ErrorSyntax{
		expression: validator,
		near:       Ne,
		comment:    "could not parse or run",
	}

	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if typ == reflect.TypeOf((time.Duration)(0)) {
			if token, err := time.ParseDuration(validator); err != nil {
				return errorSyntax
			} else if time.Duration(value.Int()) == token {
				return errorValidation
			}
		} else {
			if token, err := strconv.ParseInt(validator, 10, 64); err != nil {
				return errorSyntax
			} else if value.Int() == token {
				return errorValidation
			}
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		if token, err := strconv.ParseUint(validator, 10, 64); err != nil {
			return errorSyntax
		} else if value.Uint() == token {
			return errorValidation
		}
	case reflect.Float32, reflect.Float64:
		if token, err := strconv.ParseFloat(validator, 64); err != nil {
			return errorSyntax
		} else if value.Float() == token {
			return errorValidation
		}
	case reflect.String, reflect.Map, reflect.Slice, reflect.Array:
		if token, err := strconv.Atoi(validator); err != nil {
			return errorSyntax
		} else if value.Len() == token {
			return errorValidation
		}
	default:
		return errorSyntax
	}

	return nil
}

// nolint dupl
func validateGt(value reflect.Value, validator string) ErrorField {
	kind := value.Kind()
	typ := value.Type()

	errorValidation := ErrorValidation{
		fieldValue:     value,
		validatorType:  Gt,
		validatorValue: validator,
	}

	errorSyntax := ErrorSyntax{
		expression: validator,
		near:       Gt,
		comment:    "could not parse or run",
	}

	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if typ == reflect.TypeOf((time.Duration)(0)) {
			if token, err := time.ParseDuration(validator); err != nil {
				return errorSyntax
			} else if time.Duration(value.Int()) <= token {
				return errorValidation
			}
		} else {
			if token, err := strconv.ParseInt(validator, 10, 64); err != nil {
				return errorSyntax
			} else if value.Int() <= token {
				return errorValidation
			}
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		if token, err := strconv.ParseUint(validator, 10, 64); err != nil {
			return errorSyntax
		} else if value.Uint() <= token {
			return errorValidation
		}
	case reflect.Float32, reflect.Float64:
		if token, err := strconv.ParseFloat(validator, 64); err != nil {
			return errorSyntax
		} else if value.Float() <= token {
			return errorValidation
		}
	case reflect.String, reflect.Map, reflect.Slice, reflect.Array:
		if token, err := strconv.Atoi(validator); err != nil {
			return errorSyntax
		} else if value.Len() <= token {
			return errorValidation
		}
	default:
		return errorSyntax
	}

	return nil
}

// nolint dupl
func validateLt(value reflect.Value, validator string) ErrorField {
	kind := value.Kind()
	typ := value.Type()

	errorValidation := ErrorValidation{
		fieldValue:     value,
		validatorType:  Lt,
		validatorValue: validator,
	}

	errorSyntax := ErrorSyntax{
		expression: validator,
		near:       Lt,
		comment:    "could not parse or run",
	}

	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if typ == reflect.TypeOf((time.Duration)(0)) {
			if token, err := time.ParseDuration(validator); err != nil {
				return errorSyntax
			} else if time.Duration(value.Int()) >= token {
				return errorValidation
			}
		} else {
			if token, err := strconv.ParseInt(validator, 10, 64); err != nil {
				return errorSyntax
			} else if value.Int() >= token {
				return errorValidation
			}
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		if token, err := strconv.ParseUint(validator, 10, 64); err != nil {
			return errorSyntax
		} else if value.Uint() >= token {
			return errorValidation
		}
	case reflect.Float32, reflect.Float64:
		if token, err := strconv.ParseFloat(validator, 64); err != nil {
			return errorSyntax
		} else if value.Float() >= token {
			return errorValidation
		}
	case reflect.String, reflect.Map, reflect.Slice, reflect.Array:
		if token, err := strconv.Atoi(validator); err != nil {
			return errorSyntax
		} else if value.Len() >= token {
			return errorValidation
		}
	default:
		return errorSyntax
	}

	return nil
}

// nolint dupl
func validateGte(value reflect.Value, validator string) ErrorField {
	kind := value.Kind()
	typ := value.Type()

	errorValidation := ErrorValidation{
		fieldValue:     value,
		validatorType:  Gte,
		validatorValue: validator,
	}

	errorSyntax := ErrorSyntax{
		expression: validator,
		near:       Gte,
		comment:    "could not parse or run",
	}

	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if typ == reflect.TypeOf((time.Duration)(0)) {
			if token, err := time.ParseDuration(validator); err != nil {
				return errorSyntax
			} else if time.Duration(value.Int()) < token {
				return errorValidation
			}
		} else {
			if token, err := strconv.ParseInt(validator, 10, 64); err != nil {
				return errorSyntax
			} else if value.Int() < token {
				return errorValidation
			}
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		if token, err := strconv.ParseUint(validator, 10, 64); err != nil {
			return errorSyntax
		} else if value.Uint() < token {
			return errorValidation
		}
	case reflect.Float32, reflect.Float64:
		if token, err := strconv.ParseFloat(validator, 64); err != nil {
			return errorSyntax
		} else if value.Float() < token {
			return errorValidation
		}
	case reflect.String, reflect.Map, reflect.Slice, reflect.Array:
		if token, err := strconv.Atoi(validator); err != nil {
			return errorSyntax
		} else if value.Len() < token {
			return errorValidation
		}
	default:
		return errorSyntax
	}

	return nil
}

// nolint dupl
func validateLte(value reflect.Value, validator string) ErrorField {
	kind := value.Kind()
	typ := value.Type()

	errorValidation := ErrorValidation{
		fieldValue:     value,
		validatorType:  Lte,
		validatorValue: validator,
	}

	errorSyntax := ErrorSyntax{
		expression: validator,
		near:       Lte,
		comment:    "could not parse or run",
	}

	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if typ == reflect.TypeOf((time.Duration)(0)) {
			if token, err := time.ParseDuration(validator); err != nil {
				return errorSyntax
			} else if time.Duration(value.Int()) > token {
				return errorValidation
			}
		} else {
			if token, err := strconv.ParseInt(validator, 10, 64); err != nil {
				return errorSyntax
			} else if value.Int() > token {
				return errorValidation
			}
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		if token, err := strconv.ParseUint(validator, 10, 64); err != nil {
			return errorSyntax
		} else if value.Uint() > token {
			return errorValidation
		}
	case reflect.Float32, reflect.Float64:
		if token, err := strconv.ParseFloat(validator, 64); err != nil {
			return errorSyntax
		} else if value.Float() > token {
			return errorValidation
		}
	case reflect.String, reflect.Map, reflect.Slice, reflect.Array:
		if token, err := strconv.Atoi(validator); err != nil {
			return errorSyntax
		} else if value.Len() > token {
			return errorValidation
		}
	default:
		return errorSyntax
	}

	return nil
}

func validateEmpty(value reflect.Value, validator string) ErrorField {
	kind := value.Kind()

	errorValidation := ErrorValidation{
		fieldValue:     value,
		validatorType:  Empty,
		validatorValue: validator,
	}

	errorSyntax := ErrorSyntax{
		expression: validator,
		near:       Empty,
		comment:    "could not parse or run",
	}

	switch kind {
	case reflect.String, reflect.Map, reflect.Slice, reflect.Array:
		isEmpty, err := strconv.ParseBool(validator)
		if err != nil {
			return errorSyntax
		}

		if isEmpty && value.Len() > 0 {
			return errorValidation
		}

		if !isEmpty && value.Len() == 0 {
			return errorValidation
		}
	default:
		return errorSyntax
	}

	return nil
}

func validateNil(value reflect.Value, validator string) ErrorField {
	kind := value.Kind()

	errorValidation := ErrorValidation{
		fieldValue:     value,
		validatorType:  Nil,
		validatorValue: validator,
	}

	errorSyntax := ErrorSyntax{
		expression: validator,
		near:       string(Nil),
		comment:    "could not parse or run",
	}

	switch kind {
	case reflect.Ptr:
		isNil, err := strconv.ParseBool(validator)
		if err != nil {
			return errorSyntax
		}

		if isNil && !value.IsNil() {
			return errorValidation
		}

		if !isNil && value.IsNil() {
			return errorValidation
		}
	default:
		return errorSyntax
	}

	return nil
}

// nolint gocognit
func validateOneOf(value reflect.Value, validator string) ErrorField {
	kind := value.Kind()
	typ := value.Type()

	errorValidation := ErrorValidation{
		fieldValue:     value,
		validatorType:  Enum,
		validatorValue: validator,
	}

	errorSyntax := ErrorSyntax{
		expression: validator,
		near:       Enum,
		comment:    "could not parse or run",
	}

	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if typ == reflect.TypeOf((time.Duration)(0)) {
			var tokens []interface{}
			if tokens = parseTokens(validator); len(tokens) == 0 {
				return errorSyntax
			}

			for i, token := range tokens {
				tokens[i] = nil
				token, err := time.ParseDuration(token.(string))
				if err != nil {
					return errorSyntax
				}
				tokens[i] = token
			}

			if !tokenOneOf(time.Duration(value.Int()), tokens) {
				return errorValidation
			}
		} else {
			var tokens []interface{}
			if tokens = parseTokens(validator); len(tokens) == 0 {
				return errorSyntax
			}
			for i, token := range tokens {
				tokens[i] = nil

				token, err := strconv.ParseInt(token.(string), 10, 64)
				if err != nil {
					return errorSyntax
				}

				tokens[i] = token
			}

			if !tokenOneOf(value.Int(), tokens) {
				return errorValidation
			}
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		var tokens []interface{}
		if tokens = parseTokens(validator); len(tokens) == 0 {
			return errorSyntax
		}

		for i, token := range tokens {
			tokens[i] = nil

			token, err := strconv.ParseUint(token.(string), 10, 64)
			if err != nil {
				return errorSyntax
			}

			tokens[i] = token
		}

		if !tokenOneOf(value.Uint(), tokens) {
			return errorValidation
		}
	case reflect.Float32, reflect.Float64:
		var tokens []interface{}
		if tokens = parseTokens(validator); len(tokens) == 0 {
			return errorSyntax
		}

		for i, token := range tokens {
			tokens[i] = nil

			token, err := strconv.ParseFloat(token.(string), 64)
			if err != nil {
				return errorSyntax
			}

			tokens[i] = token
		}

		if !tokenOneOf(value.Float(), tokens) {
			return errorValidation
		}
	case reflect.String:
		var tokens []interface{}
		if tokens = parseTokens(validator); len(tokens) == 0 {
			return errorSyntax
		}

		if !tokenOneOf(value.String(), tokens) {
			return errorValidation
		}
	default:
		return errorSyntax
	}

	return nil
}

func validateFormat(value reflect.Value, validator string) ErrorField {
	kind := value.Kind()

	errorValidation := ErrorValidation{
		fieldValue:     value,
		validatorType:  Format,
		validatorValue: validator,
	}

	errorSyntax := ErrorSyntax{
		expression: validator,
		near:       string(Format),
		comment:    "could not find format",
	}

	switch kind {
	case reflect.String:
		formatTypeMap := getFormatTypeMap()
		if formatFunc, ok := formatTypeMap[FormatType(validator)]; !ok {
			return errorSyntax
		} else if !formatFunc(value.String()) {
			return errorValidation
		}
	default:
		return errorSyntax
	}

	return nil
}
