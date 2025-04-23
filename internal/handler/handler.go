package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"regexp"
	"strings"

	"entitlements/utility"

	"net/http"
)

type JSONMap map[string]any

func ParseGraphQLQuery(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Policies")
	if authHeader == "" {
		respondWithError(w, http.StatusBadRequest, "Missing Policies header")
		return
	}

	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Error reading request body")
		return
	}
	apiRequestBody := string(body)
	if apiRequestBody == "" {
		respondWithError(w, http.StatusBadRequest, "Request body cannot be empty")
		return
	}

	allFieldMap, err := utility.ParseSchema()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error parsing schema: "+err.Error())
		return
	}

	policiesList := splitPoliciesAndRemoveSpace(authHeader, ",")
	if len(policiesList) == 0 {
		respondWithError(w, http.StatusBadRequest, "No valid policies provided")
		return
	}

	var data JSONMap
	err = json.Unmarshal([]byte(apiRequestBody), &data)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Error parsing JSON body: "+err.Error())
		return
	}

	policyMap := make(map[string]map[string]any)
	for _, policy := range policiesList {
		parts := splitPoliciesAndRemoveSpace(policy, ".")
		if len(parts) != 2 {
			respondWithError(w, http.StatusBadRequest, "Invalid policy format")
			return
		}
		typename := parts[0]
		field := parts[1]

		if typename == "Query" {
			if dataField, ok := data["data"].(map[string]any); ok {
				delete(dataField, field)
			}
			continue
		}

		if policyMap[typename] == nil {
			policyMap[typename] = make(map[string]any)
		}

		engineResponse, error := getEngineResponseBasedOnPolicy(policy)
		if error != nil {
			respondWithError(w, http.StatusBadRequest, "Error getting engine response: "+error.Error())
			return
		}
		policyMap[typename]["engineResonse"] = engineResponse
		policyMap[typename][field] = true
	}

	if len(policyMap) != 0 {
		data = traverseAndRedact(data["data"].(map[string]any), allFieldMap, policyMap, "", "")
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	responseSuccess := map[string]any{
		"status":   "success",
		"data":     data,
		"allfield": allFieldMap,
		"message":  "Successfully parsed JSON",
	}
	_ = json.NewEncoder(w).Encode(responseSuccess)
}

func splitPoliciesAndRemoveSpace(policies string, delimeter string) []string {
	parts := strings.Split(policies, delimeter)
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

func normalizeTypeName(name string) string {
	re := regexp.MustCompile(`\[|\]`) // Remove brackets
	return re.ReplaceAllString(name, "")
}

func traverseAndRedact(jsonMap map[string]interface{}, fieldMap map[string]string, policyMap map[string]map[string]any, typename string, refid string) map[string]interface{} {
	for key, value := range jsonMap {

		if typename != "" {
			normalizedType := normalizeTypeName(typename)
			if normalizeTypeName(typename) == "Account" {
				if refIdVal, ok := jsonMap["accountReferenceId"].(string); ok {
					refid = refIdVal
				}
			}
			if normalizeTypeName(typename) == "Card" {
				if refIdVal, ok := jsonMap["cardReferenceId"].(string); ok {
					refid = refIdVal
				}
			}
			if policyMap[normalizedType] != nil && policyMap[normalizedType][key] == true {
				engineResponse := policyMap[normalizedType]["engineResonse"].(map[string]string)
				if !processEngineResonse(engineResponse, refid) {
					delete(jsonMap, key)
				}
				continue
			}
		}

		switch v := value.(type) {
		case map[string]interface{}:
			jsonMap[key] = traverseAndRedact(v, fieldMap, policyMap, fieldMap[key], refid)
		case []interface{}:
			for i, item := range v {
				if obj, ok := item.(map[string]interface{}); ok {
					v[i] = traverseAndRedact(obj, fieldMap, policyMap, fieldMap[key], refid)
				}
			}
			jsonMap[key] = v
		}
	}
	return jsonMap
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status":  "error",
		"message": message,
	})
}

func getEngineResponseBasedOnPolicy(policy string) (map[string]string, error) {
	permissions := map[string]map[string]string{
		"Account.balance": {
			"acc123": "ALLOW",
			"acc456": "DENY",
		},
		"Card.cardNumber": {
			"card123": "ALLOW",
			"card456": "DENY",
		},
		"AvailableCreditAmount.availableSpendingCreditAmount": {
			"acc123":  "ALLOW",
			"acc456":  "DENY",
			"card123": "DENY",
			"card456": "ALLOW",
		},
		"Account.availableCreditAmount": {
			"acc123": "ALLOW",
			"acc456": "DENY",
		},
		"Card.cardReferenceId": {
			"card123": "ALLOW",
			"card456": "DENY",
		},
		"Customer.accounts": {
			"acc123": "ALLOW",
			"acc456": "DENY",
		},
	}

	if val, ok := permissions[policy]; ok {
		return val, nil
	}
	//return nil, errors.New("no engine response found for policy: " + policy)
	return nil, nil
}

func processEngineResonse(engineResonse map[string]string, refids string) bool {
	valueInEngineResponse := engineResonse[refids]
	if valueInEngineResponse == "ALLOW" {
		return true
	} else if valueInEngineResponse == "DENY" {
		return false
	} else {
		return false
	}
}

func traverseAndRedactCopy(jsonMap map[string]interface{}, fieldMap map[string]string, policyMap map[string]map[string]any, typename string, refid string) map[string]interface{} {
	for key, value := range jsonMap {
		if typename != "" {
			normalizedType := normalizeTypeName(typename)
			if normalizeTypeName(typename) == "Account" {
				if refIdVal, ok := jsonMap["accountReferenceId"].(string); ok {
					refid = refIdVal
				}
			}
			if normalizeTypeName(typename) == "Card" {
				if refIdVal, ok := jsonMap["cardReferenceId"].(string); ok {
					refid = refIdVal
				}
			}
			if policyMap[normalizedType] != nil && policyMap[normalizedType][key] == true {
				engineResponse := policyMap[normalizedType]["engineResonse"].(map[string]string)
				fmt.Println(reflect.TypeOf(value))
				switch accounValue := value.(type) {
				case []interface{}:
					var filtered []interface{}
					for _, obj := range accounValue {
						if m, ok := obj.(map[string]interface{}); ok {
							refid, _ := m["accountReferenceId"].(string)
							if processEngineResonse(engineResponse, refid) {
								filtered = append(filtered, m)
							}
						}
					}
					jsonMap[key] = filtered
				case any:
					if !processEngineResonse(engineResponse, refid) {
						delete(jsonMap, key)
						continue
					}
				}
			}
		}
		switch v := jsonMap[key].(type) {
		case map[string]interface{}:
			jsonMap[key] = traverseAndRedactCopy(v, fieldMap, policyMap, fieldMap[key], refid)
		case []interface{}:
			for i, item := range v {
				if obj, ok := item.(map[string]interface{}); ok {
					v[i] = traverseAndRedactCopy(obj, fieldMap, policyMap, fieldMap[key], refid)
				}
			}
			jsonMap[key] = v
		}
	}
	return jsonMap
}

func ParseGraphQLQueryCopy(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Policies")
	if authHeader == "" {
		respondWithError(w, http.StatusBadRequest, "Missing Policies header")
		return
	}

	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Error reading request body")
		return
	}
	apiRequestBody := string(body)
	if apiRequestBody == "" {
		respondWithError(w, http.StatusBadRequest, "Request body cannot be empty")
		return
	}

	allFieldMap, err := utility.ParseSchema()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error parsing schema: "+err.Error())
		return
	}

	policiesList := splitPoliciesAndRemoveSpace(authHeader, ",")
	if len(policiesList) == 0 {
		respondWithError(w, http.StatusBadRequest, "No valid policies provided")
		return
	}

	var data JSONMap
	err = json.Unmarshal([]byte(apiRequestBody), &data)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Error parsing JSON body: "+err.Error())
		return
	}
	policyMap := make(map[string]map[string]any)
	for _, policy := range policiesList {
		parts := splitPoliciesAndRemoveSpace(policy, ".")
		if len(parts) != 2 {
			respondWithError(w, http.StatusBadRequest, "Invalid policy format")
			return
		}
		typename := parts[0]
		field := parts[1]

		if typename == "Query" {
			if dataField, ok := data["data"].(map[string]any); ok {
				delete(dataField, field)
			}
			continue
		}

		if policyMap[typename] == nil {
			policyMap[typename] = make(map[string]any)
		}
		engineResponse, error := getEngineResponseBasedOnPolicy(policy)
		if error != nil {
			respondWithError(w, http.StatusBadRequest, "Error getting engine response: "+error.Error())
			return
		}
		policyMap[typename]["engineResonse"] = engineResponse
		policyMap[typename][field] = true
	}

	data = traverseAndRedactCopy(data["data"].(map[string]any), allFieldMap, policyMap, "", "")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	responseSuccess := map[string]any{
		"status":   "success",
		"data":     data,
		"allfield": allFieldMap,
		"message":  "Successfully parsed JSON",
	}
	_ = json.NewEncoder(w).Encode(responseSuccess)
}
