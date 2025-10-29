package nextflow

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"unicode"
)

// snakeToCamel converts snake_case to camelCase for Groovy field names
// Example: "error_strategy" -> "errorStrategy"
func snakeToCamel(s string) string {
	if s == "" {
		return ""
	}

	parts := strings.Split(s, "_")
	if len(parts) == 1 {
		return s
	}

	var result strings.Builder
	result.WriteString(parts[0])

	for _, part := range parts[1:] {
		if part == "" {
			continue
		}
		// Capitalize first letter
		runes := []rune(part)
		runes[0] = unicode.ToUpper(runes[0])
		result.WriteString(string(runes))
	}

	return result.String()
}

// escapeGroovyString escapes special characters in Groovy string literals
func escapeGroovyString(s string) string {
	// Single quotes are used in Groovy, so escape any single quotes in the string
	s = strings.ReplaceAll(s, "'", "\\'")
	// Also escape backslashes
	s = strings.ReplaceAll(s, "\\", "\\\\")
	return s
}

// quoteGroovyString wraps a string in single quotes for Groovy
func quoteGroovyString(s string) string {
	return "'" + escapeGroovyString(s) + "'"
}

// GenerateConfigHash generates a SHA256 hash of the config content for the ID
func GenerateConfigHash(config string) string {
	hash := sha256.Sum256([]byte(config))
	return fmt.Sprintf("%x", hash)
}

// writeField writes a Groovy field assignment
func writeField(name string, value string) string {
	groovyName := snakeToCamel(name)
	return fmt.Sprintf("  %s = %s\n", groovyName, value)
}

// writeStringField writes a string field with single quotes
func writeStringField(name string, value string) string {
	return writeField(name, quoteGroovyString(value))
}

// writeNumberField writes a numeric field
func writeNumberField(name string, value interface{}) string {
	return writeField(name, fmt.Sprintf("%v", value))
}

// writeBoolField writes a boolean field
func writeBoolField(name string, value bool) string {
	return writeField(name, fmt.Sprintf("%t", value))
}
