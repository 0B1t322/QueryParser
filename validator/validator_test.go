package validator_test

import (
	"fmt"
	"testing"

	"github.com/0B1t322/QueryParser/validator"
	"github.com/stretchr/testify/require"
)

func TestFunc_Validator(t *testing.T) {
	validations := validator.Validations{
		"field_1": nil,
		"field_2": validator.TypeIs[int](),
		"field_3": func(value interface{}) error {
			number := value.(int)
			{
				if number < 0 {
					return fmt.Errorf("value should be greater then zero")
				}
	
				return nil
			}
		},
		"field_4": validator.MergeValidationsFunc(
			validator.TypeIs[string](),
			func(value interface{}) error {
				v := value.(string)
				{
					if v != "string" {
						return fmt.Errorf("Expect string")
					}

					return nil
				}
			},
		),
	}

	t.Run(
		"ValidateField",
		func(t *testing.T) {
			require.NoError(t, validations.Validate("field_1", nil))
			require.NoError(t, validations.Validate("field_1", "string"))

			require.NoError(t, validations.Validate("field_2", 12))
			require.Error(t, validations.Validate("field_2", "string"))

			require.NoError(t, validations.Validate("field_3", 12))
			require.Error(t, validations.Validate("field_3", -1))

			require.NoError(t, validations.Validate("field_4", "string"))
			require.Error(t, validations.Validate("field_4", "string123"))
			require.Error(t, validations.Validate("field_4", 12))
		},
	)
}
