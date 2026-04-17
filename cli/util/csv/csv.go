package csv

import "strings"

// Append splits raw on commas, trims whitespace, drops empty values, and
// appends the remaining items to dst.
func Append(dst []string, raw string) []string {
	for part := range strings.SplitSeq(raw, ",") {
		if value := strings.TrimSpace(part); value != "" {
			dst = append(dst, value)
		}
	}
	return dst
}
