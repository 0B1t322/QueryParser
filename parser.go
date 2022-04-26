package queryparser

import (
	"fmt"
	"net/url"
	"regexp"

	"github.com/0B1t322/QueryParser/typemapper"
	"github.com/0B1t322/QueryParser/validator"
)

type Factory interface {
	MapRegexField(field string, values []string) (interface{}, error)

	MapField(field string, values []string) (interface{}, error)

	ValidateRegexField(field string, values []string) error

	Validate(field string, values []string) error

	AddQueryTypeMapperField(field string, queryType typemapper.QueryTypeMapper)
}

type ParseSchemaItem struct {
	// Func to validate after parse
	ValidationFunc validator.ValidateFunc

	// Func to map type
	TypeMapFunc typemapper.TypeMapperFunc

	// Func to validate field
	ValidateFieldFunc typemapper.ValidateFieldFunc

	// Func to validate values
	ValidateValuesFunc typemapper.ValidateValuesFunc

	// Check if field is regex
	IsRegex bool

	// Finalize
	FinalizeParseFunc	FinalizeParseFunc
}

type FinalizeParseFunc func(result ParseResult, field string, value ResultOrError)

type ParseSchema map[string]ParseSchemaItem

type ResultOrError struct {
	Err error
	Result interface{}
}

func (r ResultOrError) IsError() bool {
	return r.Err != nil
}

type ParseResult map[string]ResultOrError

type Parser struct {
	ParseSchema ParseSchema

	Factory Factory
}

func New(
	Factory Factory,
	Schema ParseSchema,
) *Parser {
	return &Parser{
		ParseSchema: Schema,
		Factory: Factory,
	}
}

func (p *Parser) ParseUrlValues(
	urlValues url.Values,
) ParseResult {
	result := ParseResult{}

	for field, values := range urlValues {
		var item *ParseSchemaItem
		{
			if findedItem, err := p.findField(field); err == nil {
				item = findedItem
			} else if err != nil {
				continue
			}
		}

		p.Factory.AddQueryTypeMapperField(
			p.findFieldNameInSchema(field, item),
			typemapper.NewCustomQueryTypeBuilder().
			SetFieldFunc(item.ValidateFieldFunc).
			SetValidateValuesFunc(item.ValidateValuesFunc).
			SetTypeMapperFunc(item.TypeMapFunc).
			MustBuild(),
		)

		resultItem := p.getResultItem(item, field, values)

		if resultItem.Err != nil || resultItem.Result != nil {
			if item.FinalizeParseFunc != nil {
				item.FinalizeParseFunc(result, field, resultItem)
			} else {
				result[field] = resultItem
			}
		}
	}

	return result
}

func (p *Parser) getResultItem(item *ParseSchemaItem, field string, values []string) ResultOrError {
	var resultItem ResultOrError
	{
		mapResult, err := p.mapItem(item, field, values)
		if err != nil {
			resultItem.Err = err
			return resultItem
		}

		if validate := item.ValidationFunc; validate != nil {
			if err := validate(mapResult); err != nil {
				resultItem.Err = err
				return resultItem
			}
		}

		resultItem.Result = mapResult
	}

	return resultItem
}

func (p *Parser) mapItem(item *ParseSchemaItem, field string, values []string) (interface{}, error) {
	if item.IsRegex {
		return p.mapRegexField(field, values)
	}

	return p.mapDirectField(field, values)
}

func (p *Parser) mapDirectField(field string, values []string) (interface{}, error) {
	return p.Factory.MapField(field, values)
}

func (p *Parser) mapRegexField(field string, values []string) (interface{}, error) {
	return p.Factory.MapRegexField(field, values)
}


func (p *Parser) findRegexField(field string) (*ParseSchemaItem, error) {
	for key, value := range p.ParseSchema {
		if match, err := regexp.MatchString(key, field); match && value.IsRegex {
			return &value, nil
		} else if err != nil {
			return nil, err
		}
	}

	return nil, fmt.Errorf("Not found field")
}

func (p *Parser) findFieldNameInSchema(field string, item *ParseSchemaItem) (string) {
	if item.IsRegex {
		for key, value := range p.ParseSchema {
			if match, _ := regexp.MatchString(key, field); match && value.IsRegex {
				return key
			}
		}
	}

	return field
}

func (p *Parser) findDirectField(field string) (*ParseSchemaItem, error) {
	item, find := p.ParseSchema[field]
	if !find {
		return nil, fmt.Errorf("Not found field")
	}
	return &item, nil
}

func (p *Parser) findField(field string) (*ParseSchemaItem, error) {
	if regexItem, err := p.findRegexField(field); err == nil {
		return regexItem, nil
	}

	if directItem, err := p.findDirectField(field); err == nil {
		return directItem, nil
	}

	return nil, fmt.Errorf("Field not found")
}