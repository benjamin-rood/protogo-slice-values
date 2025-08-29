package parser

import (
	"strings"
	"unicode"
)

// toCamelCase converts snake_case to CamelCase
func toCamelCase(s string) string {
	if s == "" {
		return s
	}

	parts := strings.Split(s, "_")
	if len(parts) == 1 {
		return capitalizeFirst(s)
	}

	result := capitalizeFirst(parts[0])
	for _, part := range parts[1:] {
		if part != "" {
			result += capitalizeFirst(part)
		}
	}
	return result
}

// capitalizeFirst capitalizes the first character of a string
func capitalizeFirst(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}