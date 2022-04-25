package typemapper

func (q *QueryTypeFactory) ToQueryTypeMapper() QueryTypeMapper {
	var queryMapper *queryTypeFactoryToQueryTypeMapper = (*queryTypeFactoryToQueryTypeMapper)(q)

	return queryMapper
}

func (q *QueryTypeFactory) ToQueryTypeRegexMapper() QueryTypeMapper {
	var queryMapper *queryTypeFactoryToReqexQueryTypeMapper = (*queryTypeFactoryToReqexQueryTypeMapper)(q)

	return queryMapper
}

type queryTypeFactoryToQueryTypeMapper QueryTypeFactory

// Can validate values count or type if needed(string, json, or users type)
func (q *queryTypeFactoryToQueryTypeMapper) ValidateValues(values []string) error {
	return nil
}

// Can validateField format
func (q *queryTypeFactoryToQueryTypeMapper) ValidateField(field string) error {
	return nil
}

func (q *queryTypeFactoryToQueryTypeMapper) Map(field string, values []string) (interface{}, error) {
	var factory *QueryTypeFactory = (*QueryTypeFactory)(q)
	{
		if err := factory.Validate(field, values); err != nil {
			return nil, err
		}

		return factory.MapField(field, values)
	}
}


type queryTypeFactoryToReqexQueryTypeMapper QueryTypeFactory

// Can validate values count or type if needed(string, json, or users type)
func (q *queryTypeFactoryToReqexQueryTypeMapper) ValidateValues(values []string) error {
	return nil
}

// Can validateField format
func (q *queryTypeFactoryToReqexQueryTypeMapper) ValidateField(field string) error {
	return nil
}

func (q *queryTypeFactoryToReqexQueryTypeMapper) Map(field string, values []string) (interface{}, error) {
	var factory *QueryTypeFactory = (*QueryTypeFactory)(q)
	{
		if err := factory.ValidateRegexField(field, values); err != nil {
			return nil, err
		}

		return factory.MapRegexField(field, values)
	}
}