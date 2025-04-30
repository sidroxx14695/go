package utility

import (
	"os"
	"regexp"
	"strings"

	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
)

type FieldsMap map[string]string

type EntitlementIdMap map[string]string

var parsedSchema *ast.Schema // global cache to be reused if needed

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

	parsedSchema = doc

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

func ResolveRefIdNameFallback(policyKey string, entitlementIdMap EntitlementIdMap) string {
	// Priority 1: check entitlementIdMap
	if ref, ok := entitlementIdMap[policyKey]; ok && ref != "" {
		return ref
	}

	// Priority 2: check @key directive in parsedSchema
	parts := strings.Split(policyKey, ".")
	if len(parts) != 2 || parsedSchema == nil {
		return ""
	}
	typename := parts[0]

	if typeDef, ok := parsedSchema.Types[typename]; ok {
		for _, dir := range typeDef.Directives {
			if dir.Name == "key" {
				for _, arg := range dir.Arguments {
					if arg.Name == "fields" {
						return arg.Value.Raw
					}
				}
			}
		}
	}

	return ""
}
