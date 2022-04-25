package typemapper

import (
	"errors"
	"fmt"
	"regexp"
)

var (
	NotFoundField = errors.New("Field not found")
)

type TypeMapperFunc func(field string, values []string) (interface{}, error)

type QueryTypeMapper interface {

	// Can validate values count or type if needed(string, json, or users type)
	ValidateValues(values []string) error

	// Can validateField format
	ValidateField(field string) error

	Map(field string, values []string) (interface{}, error)
}

type CustomQueryTypeBuilder interface {
	// Optional method to build QueryType
	SetValidateValuesFunc(f ValidateValuesFunc) CustomQueryTypeBuilder
	// Optional method to build QueryType
	SetFieldFunc(f ValidateFieldFunc) CustomQueryTypeBuilder
	// Required method to build QueryType
	SetTypeMapperFunc(f TypeMapperFunc) CustomQueryTypeBuilder
	// Must build panic if some field are not set
	MustBuild() QueryTypeMapper

	Build() (QueryTypeMapper, error)
}

type ValidateValuesFunc func(values []string) error

type ValidateFieldFunc func(field string) error

type customQueryTypeValidators struct {
	validateValuesFunc ValidateValuesFunc

	validateFieldFunc ValidateFieldFunc
}

type customQueryTypeMapper struct {
	typeMapperFunc TypeMapperFunc
}

type customQueryType struct {
	customQueryTypeValidators

	customQueryTypeMapper
}

func (c *customQueryType) Map(field string, values []string) (interface{}, error) {
	return c.typeMapperFunc(field, values)
}

// Can validate values count or type if needed(string, json, or users type)
func (c *customQueryType) ValidateValues(values []string) error {
	if c.validateValuesFunc == nil {
		return nil
	}
	return c.validateValuesFunc(values)
}

// Can validateField format
func (c *customQueryType) ValidateField(field string) error {
	if c.validateFieldFunc == nil {
		return nil
	}
	return c.validateFieldFunc(field)
}
func (c *customQueryType) SetValidateValuesFunc(f ValidateValuesFunc) CustomQueryTypeBuilder {
	c.validateValuesFunc = f
	return c
}

func (c *customQueryType) SetFieldFunc(f ValidateFieldFunc) CustomQueryTypeBuilder {
	c.validateFieldFunc = f
	return c
}

func (c *customQueryType) SetTypeMapperFunc(f TypeMapperFunc) CustomQueryTypeBuilder {
	c.typeMapperFunc = f
	return c
}

func (c *customQueryType) Build() (QueryTypeMapper, error) {
	// if c.field == nil {
	// 	return nil, fmt.Errorf("Field not set")
	// }

	if c.typeMapperFunc == nil {
		return nil, fmt.Errorf("MapFunc not set")
	}

	return c, nil
}

func (c *customQueryType) MustBuild() QueryTypeMapper {
	build, err := c.Build()
	if err != nil {
		panic(err)
	}

	return build
}

func NewCustomQueryTypeBuilder() CustomQueryTypeBuilder {
	return &customQueryType{}
}

type QueryTypeFactory struct {
	Querys map[string]QueryTypeMapper
}

func NewQueryTypeFactory() *QueryTypeFactory {
	return &QueryTypeFactory{
		Querys: map[string]QueryTypeMapper{},
	}
}

func (q *QueryTypeFactory) AddField(field string, queryType QueryTypeMapper) *QueryTypeFactory {
	q.Querys[field] = queryType
	return q
}

func (q *QueryTypeFactory) Validate(field string, values []string) error {
	typeMapper, find := q.Querys[field]
	if !find {
		return NotFoundField
	}

	if err := typeMapper.ValidateField(field); err != nil {
		return err
	}

	if err := typeMapper.ValidateValues(values); err != nil {
		return err
	}

	return nil
}

func (q *QueryTypeFactory) ValidateRegexField(field string, values []string) error {
	for key, value := range q.Querys {
		if match, err := regexp.MatchString(key, field); match {
			if err := value.ValidateField(field); err != nil {
				return err
			}
		
			if err := value.ValidateValues(values); err != nil {
				return err
			}

			return nil
		} else if err != nil {
			return err
		}
	}
	return NotFoundField
}

// MapField map a field
// 
// cathable errors:
// 	NotFoundField
func (q *QueryTypeFactory) MapField(field string, values []string) (interface{}, error) {
	typeMapper, find := q.Querys[field]
	if !find {
		return nil, NotFoundField
	}

	return typeMapper.Map(field, values)
}

func (q *QueryTypeFactory) MapRegexField(field string, values []string) (interface{}, error) {
	for key, value := range q.Querys {
		if match, err := regexp.MatchString(key, field); match {
			return value.Map(field, values)
		} else if err != nil {
			return nil, err
		}
	}
	return nil, NotFoundField
}
