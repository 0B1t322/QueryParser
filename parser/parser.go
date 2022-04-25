package parser

type TypeMapper interface {
	Validate(field string, values []string) error

	ValidateRegexField(field string, values []string)

	MapField(field string, values []string) (interface{}, error)

	MapRegexField(field string, values []string) (interface{}, error)
}

type Validator interface {
	
}

type Parser struct {
}
