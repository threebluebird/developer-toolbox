package jsontool

import "encoding/json"

// ValidateJSON checks whether the provided input is valid JSON.
func ValidateJSON(input string) error {
	var value any
	return json.Unmarshal([]byte(input), &value)
}
