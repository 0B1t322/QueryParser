package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/0B1t322/QueryParser"
	"github.com/0B1t322/QueryParser/typemapper"
)

// Code below show how to parse with recirsive with diffucal user schemas
type FieldOperation struct {
	Op string	`json:"op"`
	Value string `json:"value"`
}

type Root struct {
	Or   []*Root `json:"or,omitempty"`

	And  []*Root `json:"and,omitempty"`

	Name *FieldOperation `json:"name,omitempty"`
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


// Prints:
	/*
{
	"or": [
			{
					"name": {
							"op": "like",
							"value": "dan*"
					}
			},
			{
					"name": {
							"op": "lte",
							"value": "asd"
					}
			}
	],
	"name": {
			"op": "eq",
			"value": "some_name"
	}
}
*/
func main() {
	r := &Root{}
	url, _ := url.Parse("http://localhost:8080?name[eq]=some_name&or=name[like]=dan*,name[lte]=asd")
	r.NewParser().ParseUrlValues(
		url.Query(),
	)

	data, _ := json.MarshalIndent(r, "", "\t")

	fmt.Println(string(data))
}