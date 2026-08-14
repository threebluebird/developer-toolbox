package jsontool

import (
	"encoding/json"
	"strings"
)

// FormatJSON returns a pretty-printed JSON string using the specified indent.
func FormatJSON(input string, indent int) (string, error) {
	var value any
	if err := json.Unmarshal([]byte(input), &value); err != nil {
		return "", err
	}

	indented, err := json.MarshalIndent(value, "", strings.Repeat(" ", indent))
	if err != nil {
		return "", err
	}

	return string(indented), nil
}

// MinifyJSON returns a compact JSON string.
func MinifyJSON(input string) (string, error) {
	var value any
	if err := json.Unmarshal([]byte(input), &value); err != nil {
		return "", err
	}

	minified, err := json.Marshal(value)
	if err != nil {
		return "", err
	}

	return string(minified), nil
}
