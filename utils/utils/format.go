package utils

import(
	"fmt"
	"regexp"
	"strings"
	"time"
)


func formatTimeUnit(diff time.Duration, unit time.Duration, singular string) string {
	count := int(diff / unit)
	if count == 1 {
		return "1 " + singular + " ago"
	}
	return strings.Join([]string{string(count), singular + "s ago"}, " ")
}

// FormatStringWithRegex formats a string using a regular expression and a replacement string.
func FormatStringWithRegex(input, pattern, replacement string) string {
	re := regexp.MustCompile(pattern)
	return re.ReplaceAllString(input, replacement)
}

// FormatBytes formats a byte size value to a human-readable string representation.
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return strconv.FormatInt(bytes, 10) + " B"
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}