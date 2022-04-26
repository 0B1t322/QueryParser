package queryparser_test

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"testing"

	queryparser "github.com/0B1t322/QueryParser"
	"github.com/0B1t322/QueryParser/typemapper"
	"github.com/stretchr/testify/require"
)

func TestFunc_Parser(t *testing.T) {

	t.Run(
		"ParseOffset",
		func(t *testing.T) {
			p := queryparser.New(
				typemapper.NewQueryTypeFactory(),
				queryparser.ParseSchema{
					"offset": queryparser.ParseSchemaItem{
						IsRegex: false,
						TypeMapFunc: func(field string, values []string) (interface{}, error) {
							strOffset := values[0]

							intOffset, err := strconv.Atoi(strOffset)
							if err != nil {
								return nil, err
							}

							return intOffset, nil
						},
						ValidateValuesFunc: func(values []string) error {
							if len(values) > 1 {
								return fmt.Errorf("To much values")
							}
							return nil
						},
						ValidationFunc: func(value interface{}) error {
							if value.(int) < 0 {
								return fmt.Errorf("Offset can't be lower then zero")
							}
							return nil
						},
					},
				},
			)

			url, _ := url.Parse("http://localhost:8080?offset=14&limit=10&offset[eq]=12")
			t.Log(url.Query())
			result := p.ParseUrlValues(
				url.Query(),
			)
			require.False(t, result["offset"].IsError())

			require.Equal(
				t,
				result["offset"].Result,
				14,
			)
		},
	)
}

// Code below show how to parse with recirsive with diffucal user schemas
type FieldOperation struct {
	Op string
	Value string
}

type Root struct {
	Or   []*Root

	And  []*Root

	Name *FieldOperation
}

func (r *Root) NewParser() *queryparser.Parser {
	return queryparser.New(
		typemapper.NewQueryTypeFactory(),
		r.NewParseSchema(),
	)
}

func (r *Root) getOpearationAndField(field string) (op string, cleanField string) {
	regex := regexp.MustCompile(`(?P<field>.*)\[(?P<op>.*)\]`)
	opAt := regex.SubexpIndex("op")
	fieldAt := regex.SubexpIndex("field")
	submatch := regex.FindStringSubmatch(field)

	return submatch[opAt], submatch[fieldAt]
}

func (r *Root) NameParseSchema() queryparser.ParseSchemaItem {
	return queryparser.ParseSchemaItem{
		IsRegex: true,
		TypeMapFunc: func(field string, values []string) (interface{}, error) {
			op, _ := r.getOpearationAndField(field)

			name := values[0]

			r.Name = &FieldOperation{
				Op: op,
				Value: name,
			}

			return nil, nil
		},
	}
}

func (r *Root) OrParseSchema() queryparser.ParseSchemaItem {
	return queryparser.ParseSchemaItem{
		IsRegex: false,
		TypeMapFunc: func(field string, values []string) (interface{}, error) {
			value := values[0]

			for _, value := range strings.Split(value, ",") {
				subRoot := &Root{}
				subParser := subRoot.NewParser()
				subQuery := url.Values{}
				fieldValie := strings.SplitN(value, "=", 2)
				if len(fieldValie) != 2 {
					continue
				}

				subQuery.Add(fieldValie[0], fieldValie[1])
				subParser.ParseUrlValues(subQuery)
				r.Or = append(r.Or, subRoot)
			}

			return nil, nil
		},
	}
}

func (r *Root) NewParseSchema() queryparser.ParseSchema {
	return queryparser.ParseSchema{
		`name\[.*\]`: r.NameParseSchema(),
		"or": r.OrParseSchema(),
	}
}

func TestFunc_ParseRecursive(t *testing.T) {
	r := &Root{}
	url, _ := url.Parse("http://localhost:8080?name[eq]=some_name&or=name[like]=dan*,name[lte]=asd")
	r.NewParser().ParseUrlValues(
		url.Query(),
	)


	require.Equal(
		t,
		&Root{
			Name: &FieldOperation{
				Op: "eq",
				Value: "some_name",
			},
			Or: []*Root{
				{
					Name: &FieldOperation{
						Op: "like",
						Value: "dan*",
					},
				},
				{
					Name: &FieldOperation{
						Op: "lte",
						Value: "asd",
					},
				},
			},
		},
		r,
	)

}
