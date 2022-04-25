package validator

import (
	"fmt"

	"github.com/pkg/errors"
)

// Validations describe map of query valiues with their validations
type Validations map[string]ValidateFunc

// ValidateFunc desctibe a func that take parsed value and validate it
type ValidateFunc func(value interface{}) error

// Validate try to validate field by passing given object to ValidateFunc
// 
// If field not found return nil
// 
// Don't validate that value is nil
func (v Validations) Validate(
	field string,
	value interface{},
) error {
	validator, find := v[field]
	if find && validator != nil {
		return validator(value)
	}

	return nil
}

func MergeValidationsFunc(funcs ...ValidateFunc) ValidateFunc {
	return func(value interface{}) error {
		for _, f := range funcs {
			if err := f(value); err != nil {
				return err
			}
		}

		return nil
	}
}

// TypeIs check the type of value passed to ValidateFunc
// 
// catchable errors:
// 	UnexpectedType as target
// Target errors catch by:
// 		errors.Is(err, UnexpectedType)
func TypeIs[T any]() ValidateFunc {
	return func(value interface{}) error {
		if _, ok := value.(T); ok {
			return nil
		}

		return errors.Wrap(UnexpectedType, fmt.Sprintf("Expected type: %T, actual type: %T", *new(T), value))
	}
}