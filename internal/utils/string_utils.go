package utils

import (
	"fmt"
	"strings"
)

// JoinQuoted joins a slice of strings with commas and quotes.
func JoinQuoted(items []string) string {
	quotedItems := make([]string, 0, len(items))
	for _, s := range items {
		quotedItems = append(quotedItems, fmt.Sprintf("%q", s))
	}
	return strings.Join(quotedItems, ", ")
}

// ToYamlBlock renders a YAML block scalar with safe indentation.
func ToYamlBlock(block string, indent int) string {
	trim := strings.TrimRight(block, "\n")
	lines := strings.Split(trim, "\n")
	prefix := strings.Repeat(" ", indent)
	if len(lines) == 1 && !strings.Contains(lines[0], "\n") {
	}
	var builder strings.Builder
	builder.WriteString("|\n")
	for _, ln := range lines {
		builder.WriteString(prefix)
		builder.WriteString(ln)
		builder.WriteString("\n")
	}
	return builder.String()
}
