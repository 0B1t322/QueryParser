package typemapper_test

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/0B1t322/QueryParser/typemapper"
	"github.com/stretchr/testify/require"
)

type UserDefinedType struct {
	Field string
	Value string
}

type FieldOperationString struct {
	Operation string
	Value     string
}

func TestFunc_TypeMapper(t *testing.T) {
	factory := typemapper.NewQueryTypeFactory().
		AddField(
			"field_1",
			typemapper.NewCustomQueryTypeBuilder().
				SetValidateValuesFunc(
					func(values []string) error {
						if len(values) != 1 {
							return fmt.Errorf("expect 1 value")
						}
						return nil
					},
				).
				SetTypeMapperFunc(
					func(field string, values []string) (interface{}, error) {
						value := values[0]

						return value, nil
					},
				).MustBuild(),
		).
		AddField(
			"field_2",
			typemapper.NewCustomQueryTypeBuilder().
				SetValidateValuesFunc(
					func(values []string) error {
						if len(values) <= 0 {
							return fmt.Errorf("Expect slice")
						}
						return nil
					},
				).
				SetTypeMapperFunc(
					func(field string, values []string) (interface{}, error) {
						var slice []UserDefinedType
						for _, value := range values {
							keyValye := strings.Split(value, ":")
							if len(keyValye) != 2 {
								continue
							}
							slice = append(
								slice,
								UserDefinedType{
									Field: keyValye[0],
									Value: keyValye[1],
								},
							)
						}

						return slice, nil
					},
				).
				MustBuild(),
		).AddField(
		"field_3",
		typemapper.NewCustomQueryTypeBuilder().
			SetValidateValuesFunc(
				func(values []string) error {
					if len(values) != 1 {
						return fmt.Errorf("Values should be one")
					}
					return nil
				},
			).SetFieldFunc(
			func(field string) error {
				op := strings.TrimRight(
					strings.TrimLeft(
						strings.TrimLeftFunc(
							field,
							func(r rune) bool {
								return r != '['
							},
						),
						"[",
					),
					"]",
				)

				if !(op == "eq" || op == "lte") {
					return fmt.Errorf("Bad operation")
				}
				return nil
			},
		).SetTypeMapperFunc(
			func(field string, values []string) (interface{}, error) {
				op := strings.TrimRight(
					strings.TrimLeft(
						strings.TrimLeftFunc(
							field,
							func(r rune) bool {
								return r != '['
							},
						),
						"[",
					),
					"]",
				)

				return FieldOperationString{
					Operation: op,
					Value:     values[0],
				}, nil
			},
		).MustBuild(),
	)

	t.Run(
		"Field_1",
		func(t *testing.T) {
			require.NoError(t, factory.Validate("field_1", []string{"value_1"}))
			require.Error(t, factory.Validate("field_1", []string{"value_1", "value_2"}))

			mappedValue, err := factory.MapField("field_1", []string{"value_1"})
			require.NoError(t, err)

			require.Equal(t, "value_1", mappedValue)
		},
	)

	t.Run(
		"Field_2",
		func(t *testing.T) {
			require.NoError(t, factory.Validate("field_2", []string{"value_1"}))
			require.NoError(t, factory.Validate("field_2", []string{"value_1", "value_2"}))
			require.Error(t, factory.Validate("field_2", []string{}))

			mappedValue, err := factory.MapField("field_2", []string{"field_1:value_1", "field_2:value_2"})
			require.NoError(t, err)

			require.Equal(
				t,
				[]UserDefinedType{
					{
						Field: "field_1",
						Value: "value_1",
					},
					{
						Field: "field_2",
						Value: "value_2",
					},
				},
				mappedValue,
			)
		},
	)

	t.Run(
		"Field_3",
		func(t *testing.T) {
			require.NoError(t, factory.ValidateRegexField("field_3[eq]", []string{"value_1"}))
			require.NoError(t, factory.ValidateRegexField("field_3[lte]", []string{"value_1"}))

			require.Error(t, factory.ValidateRegexField("field_3", []string{"value_1"}))
			require.Error(t, factory.ValidateRegexField("field_3", []string{}))

			require.Error(t, factory.ValidateRegexField("field_3[eq]", []string{}))
			require.Error(t, factory.ValidateRegexField("field_3[lte]", []string{}))

			require.Error(t, factory.ValidateRegexField("field_3[eq]", []string{"value_1", "value_2"}))
			require.Error(t, factory.ValidateRegexField("field_3[lte]", []string{"value_1", "value_2"}))

			require.Equal(
				t,
				FieldOperationString{
					Operation: "eq",
					Value:     "value_1",
				},
				func() interface{} {
					value, _ := factory.MapRegexField(
						"field_3[eq]",
						[]string{"value_1"},
					)
					return value
				}(),
			)

			require.Equal(
				t,
				FieldOperationString{
					Operation: "lte",
					Value:     "value_2",
				},
				func() interface{} {
					value, _ := factory.MapRegexField(
						"field_3[lte]",
						[]string{"value_2"},
					)
					return value
				}(),
			)
		},
	)
}

type FieldOperation struct {
	Op    string      `json:"op"`
	Value interface{} `json:"value"`
}

type Root struct {
	Or  []*Root `json:"or,omitempty"`
	And []*Root `json:"and,omitempty"`

	Name  *FieldOperation `json:"name,omitempty"`
	Email *FieldOperation `json:"email,omitempty"`
	Phone *FieldOperation `json:"phone,omitempty"`
	Age   *FieldOperation `json:"age,omitempty"`
}

func (r *Root) getOpearationAndField(field string) (op string, cleanField string) {
	regex := regexp.MustCompile(`(?P<field>.*)\[(?P<op>.*)\]`)
	opAt := regex.SubexpIndex("op")
	fieldAt := regex.SubexpIndex("field")
	submatch := regex.FindStringSubmatch(field)

	return submatch[opAt], submatch[fieldAt]
}

func (query *Root) Factory() *typemapper.QueryTypeFactory {
	factory := typemapper.NewQueryTypeFactory().
	AddField(
		"age",
		query.AgeQueryType(),
	).AddField(
		"phone",
		query.PhoneQueryType(),
	).AddField(
		"name",
		query.NameQueryType(),
	).AddField(
		"email",
		query.EmailQueryType(),
	).AddField(
		"or",
		query.OrQueryType(),
	).AddField(
		"and",
		query.AndQueryType(),
	)

	return factory
}

func (r *Root) OrQueryType() typemapper.QueryTypeMapper {
	return typemapper.NewCustomQueryTypeBuilder().
		SetTypeMapperFunc(
			func(field string, values []string) (interface{}, error) {
				newQuery := &Root{}
				value := values[0]

				subQuery := map[string][]string{}
				{
					for _, value := range strings.Split(value, ";") {
						fieldValue := strings.SplitN(value, "=", 2)
						if len(fieldValue) != 2 {
							continue
						}
						subQuery[fieldValue[0]] = append(subQuery[fieldValue[0]], fieldValue[1])
					}
				}
				f := newQuery.Factory()
				for subField, subQuerys := range subQuery {
					_, err := f.MapRegexField(subField, subQuerys)
					if err != nil {
						return nil, err
					}
				}

				r.Or = append(r.Or, newQuery)
				return nil, nil
			},
		).
		MustBuild()
}

func (r *Root) AndQueryType() typemapper.QueryTypeMapper {
	return typemapper.NewCustomQueryTypeBuilder().
		SetTypeMapperFunc(
			func(field string, values []string) (interface{}, error) {
				newQuery := &Root{}
				value := values[0]

				subQuery := map[string][]string{}
				{
					for _, value := range strings.Split(value, ";") {
						fieldValue := strings.SplitN(value, "=", 2)
						if len(fieldValue) != 2 {
							continue
						}
						subQuery[fieldValue[0]] = append(subQuery[fieldValue[0]], fieldValue[1])
					}
				}

				for subField, subQuerys := range subQuery {
					_, err := newQuery.Factory().MapRegexField(subField, subQuerys)
					if err != nil {
						return nil, err
					}
				}

				r.And = append(r.And, newQuery)
				return nil, nil
			},
		).
		MustBuild()
}

func (r *Root) AgeQueryType() typemapper.QueryTypeMapper {
	return typemapper.NewCustomQueryTypeBuilder().
		SetTypeMapperFunc(
			func(field string, values []string) (interface{}, error) {
				op, _ := r.getOpearationAndField(field)

				age, _ := strconv.Atoi(values[0])

				r.SetAge(
					op,
					age,
				)

				return r.Age, nil
			},
		).
		MustBuild()
}

func (r *Root) PhoneQueryType() typemapper.QueryTypeMapper {
	return typemapper.NewCustomQueryTypeBuilder().
		SetTypeMapperFunc(
			func(field string, values []string) (interface{}, error) {
				op, _ := r.getOpearationAndField(field)

				r.SetPhone(
					op,
					values[0],
				)

				return r.Phone, nil
			},
		).
		MustBuild()
}

func (r *Root) EmailQueryType() typemapper.QueryTypeMapper {
	return typemapper.NewCustomQueryTypeBuilder().
		SetTypeMapperFunc(
			func(field string, values []string) (interface{}, error) {
				op, _ := r.getOpearationAndField(field)

				r.SetEmail(
					op,
					values[0],
				)

				return r.Email, nil
			},
		).
		MustBuild()
}

func (r *Root) NameQueryType() typemapper.QueryTypeMapper {
	return typemapper.NewCustomQueryTypeBuilder().
		SetTypeMapperFunc(
			func(field string, values []string) (interface{}, error) {
				op, _ := r.getOpearationAndField(field)

				r.SetName(
					op,
					values[0],
				)

				return r.Name, nil
			},
		).
		MustBuild()
}

func (r *Root) SetName(op string, value string) {
	r.Name = &FieldOperation{
		Op:    op,
		Value: value,
	}
}

func (r *Root) SetEmail(op string, value string) {
	r.Email = &FieldOperation{
		Op:    op,
		Value: value,
	}
}

func (r *Root) SetPhone(op string, value string) {
	r.Phone = &FieldOperation{
		Op:    op,
		Value: value,
	}
}

func (r *Root) SetAge(op string, value int) {
	r.Age = &FieldOperation{
		Op:    op,
		Value: value,
	}
}

func (r *Root) AddOr(root *Root) {
	r.Or = append(r.Or, root)
}

func (r *Root) AddAnd(root *Root) {
	r.And = append(r.And, root)
}

type RootQuery struct {
	Root
}

func TestFunc_TypeMapperAndOrFormat(t *testing.T) {
	// Decribe format ?or=name[eq]=value;and=email[like]="dan.*"&and=phone[eq]="89991234567&age[lte]=15"
	// Format is recursive Root scheme
	/*
		{
			"or": [
				{
					"name": {
						"op": eq,
						"value": "some"
					}
				},
				{
					"and": [
						{
							"email": {
								"op": like,
								"value": "dan.*"
							}
						}
					]
				}
			],
			"and": [
				{
					"phone": {
						"op": eq,
						"value": "89991234567"
					}
				}
			],
			"age": {
				"op": "lte",
				"value": 15
			}
		}
	*/

	query := &Root{}
	f := query.Factory()

	f.MapRegexField("or", []string{"name[eq]=value;and=email[like]=dan.*"})

	f.MapRegexField("and", []string{"phone[eq]=89991234567"})

	f.MapRegexField("age[lte]", []string{"15"})

	expect := &Root{
		Or: []*Root{
			{
				Name: &FieldOperation{
					Op: "eq",
					Value: "value",
				},
			},
			{
				And: []*Root{
					{
						Email: &FieldOperation{
							Op: "like",
							Value: "dan.*",
						},
					},
				},
			},
		},
		And: []*Root{
			{
				Phone: &FieldOperation{
					Op: "eq",
					Value: "89991234567",
				},
			},
		},
		Age: &FieldOperation{
			Op: "lte",
			Value: 15,
		},
	}

	require.Equal(
		t,
		expect,
		query,
	)
}
