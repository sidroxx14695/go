package main

import (
	"encoding/json"
	"fmt"
	"regexp"
)

var arrayIndexRegex = regexp.MustCompile(`\[\d+\]`)

func traverseAndRedactMultiple(
	m map[string]interface{},
	prefix string,
	fieldMap map[string]string,
	policy map[string][]string,
	typename string,
) {
	for key, value := range m {
		cleanKey := removeArrayIndices(key)
		currentPath := fmt.Sprintf("%s.%s", prefix, cleanKey)

		if typename == "" {
			if t, ok := fieldMap[key]; ok {
				typename = t
			}
		}

		if fieldsToDelete, ok := policy[typename]; ok {
			for _, field := range fieldsToDelete {
				delete(m, field)
			}
		}

		switch v := value.(type) {
		case map[string]interface{}:
			traverseAndRedactMultiple(v, currentPath, fieldMap, policy, fieldMap[key])
		case []interface{}:
			for i, item := range v {
				if obj, ok := item.(map[string]interface{}); ok {
					traverseAndRedactMultiple(obj, fmt.Sprintf("%s[%d]", currentPath, i), fieldMap, policy, fieldMap[key])
				}
			}
			m[key] = v
		}
	}
}

func removeArrayIndices(s string) string {
	return arrayIndexRegex.ReplaceAllString(s, "")
}

func main() {
	fieldMap := map[string]string{
		"accountReferenceId":            "String!",
		"accounts":                      "[Account]",
		"address":                       "String!",
		"arguments":                     "JSON",
		"availableCashCreditAmount":     "Float",
		"availableCreditAmount":         "AvailableCreditAmount",
		"availableSpendingCreditAmount": "Float",
		"balance":                       "Float",
		"cardNumber":                    "String",
		"cardReferenceId":               "String!",
		"cards":                         "[Card]",
		"customerReferenceId":           "String!",
		"email":                         "String!",
		"expiryDate":                    "String",
		"getAccount":                    "Account",
		"getAccounts":                   "[Account]",
		"getCard":                       "Card",
		"getCards":                      "[Card]",
		"getCustomer":                   "Customer",
		"key":                           "String!",
		"last4ssn":                      "String!",
		"name":                          "String!",
		"node":                          "JSON",
		"status":                        "String",
		"type":                          "String",
	}

	// Sample JSON (same as before)
	jsonData := map[string]interface{}{
		"data": map[string]interface{}{
			"getCustomer": map[string]interface{}{
				"customerReferenceId": "cust123",
				"name":                "John Doe",
				"last4ssn":            "1234",
				"email":               "john.doe@example.com",
				"address":             "123 Main St, Anytown, USA",
				"accounts": []interface{}{
					map[string]interface{}{
						"accountReferenceId": "acc123",
						"status":             "active",
						"type":               "savings",
						"balance":            1000.0,
						"availableCreditAmount": map[string]interface{}{
							"availableSpendingCreditAmount": 500.0,
							"availableCashCreditAmount":     200.0,
						},
						"cards": []interface{}{
							map[string]interface{}{
								"cardReferenceId": "card123",
								"status":          "active",
								"type":            "credit",
								"expiryDate":      "12/25",
								"cardNumber":      "**** **** **** 1234",
								"availableCreditAmount": map[string]interface{}{
									"availableSpendingCreditAmount": 300.0,
									"availableCashCreditAmount":     100.0,
								},
							},
						},
					},
				},
			},
			"getAccount": map[string]interface{}{
				"accountReferenceId": "acc456",
				"status":             "active",
				"type":               "checking",
				"balance":            2000.0,
				"availableCreditAmount": map[string]interface{}{
					"availableSpendingCreditAmount": 1500.0,
					"availableCashCreditAmount":     1000.0,
				},
				"cards": []interface{}{
					map[string]interface{}{
						"cardReferenceId": "card456",
						"status":          "active",
						"type":            "debit",
						"expiryDate":      "11/24",
						"cardNumber":      "**** **** **** 4567",
						"availableCreditAmount": map[string]interface{}{
							"availableSpendingCreditAmount": 800.0,
							"availableCashCreditAmount":     400.0,
						},
					},
				},
			},
		},
	}

	// Define the policy: delete multiple fields
	policy := map[string][]string{
		"Customer":              {"accounts"},
		"AvailableCreditAmount": {"availableSpendingCreditAmount", "availableCashCreditAmount"},
	}

	// Call function to remove based on header policy
	traverseAndRedactMultiple(jsonData["data"].(map[string]interface{}), "data", fieldMap, policy, "")

	// Pretty print the final redacted JSON
	finalJSON, _ := json.MarshalIndent(jsonData, "", "  ")
	fmt.Printf("\nFinal Redacted JSON:\n%s\n", string(finalJSON))
}
