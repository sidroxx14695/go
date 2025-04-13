package utility

import "strings"

func validateString(str string) bool {
	return !strings.HasPrefix(str, "__")
}
