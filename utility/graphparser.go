package utility

import (
	"os"
	"regexp"

	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
)

type FieldsMap map[string]string

type EntitlementIdMap map[string]string

func ParseSchema() (FieldsMap, EntitlementIdMap, error) {
	schemaFilePath := "../schema.graphql"
	body, err := os.ReadFile(schemaFilePath)
	if err != nil {
		return nil, nil, err
	}
	doc, err := gqlparser.LoadSchema(&ast.Source{Input: string(body)})
	if err != nil {
		return nil, nil, err
	}
	entitlementIdMap, err := ExtractEntitlementIdentifiers(string(body))
	if err != nil {
		return nil, nil, err
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
	return allFieldMap, entitlementIdMap, nil
}

func ExtractEntitlementIdentifiers(schema string) (EntitlementIdMap, error) {
	result := make(EntitlementIdMap)

	regex := regexp.MustCompile(`key:\s*"(.*?)"(?:[^{}]|{[^{}]*})*?node:\s*{\s*entitlementIdentifier:\s*"(.*?)"`)

	matches := regex.FindAllStringSubmatch(schema, -1)
	for _, match := range matches {
		key := match[1]
		entitlementIdentifier := match[2]
		result[key] = entitlementIdentifier
	}

	return result, nil
}
