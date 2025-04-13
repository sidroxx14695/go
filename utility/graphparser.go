package utility

import (
	"os"

	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
)

type TypeMap map[string]map[string]string

type FieldsMap map[string]string

func ParseSchema() (TypeMap, FieldsMap, error) {
	schemaFilePath := "/Users/sharathnavva/Documents/Go/Post enforcement/schema.graphql"
	body, err := os.ReadFile(schemaFilePath)
	if err != nil {
		return nil, nil, err
	}
	doc, err := gqlparser.LoadSchema(&ast.Source{Input: string(body)})
	if err != nil {
		return nil, nil, err
	}

	types := make(TypeMap)
	allFieldMap := make(FieldsMap)
	for typeName, def := range doc.Types {
		if validateString(typeName) && len(def.Fields) > 0 {
			fieldMap := make(map[string]string)
			for _, field := range def.Fields {
				if validateString(field.Name) {
					fieldMap[field.Name] = field.Type.String()
					allFieldMap[field.Name] = field.Type.String()
				}
			}
			types[typeName] = fieldMap
		}
	}
	return types, allFieldMap, nil
}
