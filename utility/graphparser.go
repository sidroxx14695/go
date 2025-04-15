package utility

import (
	"os"

	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
)

type FieldsMap map[string]string

func ParseSchema() (FieldsMap, error) {
	schemaFilePath := "../schema.graphql"
	body, err := os.ReadFile(schemaFilePath)
	if err != nil {
		return nil, err
	}
	doc, err := gqlparser.LoadSchema(&ast.Source{Input: string(body)})
	if err != nil {
		return nil, err
	}

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
		}
	}
	return allFieldMap, nil
}
