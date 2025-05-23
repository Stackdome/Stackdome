package services

import (
	"regexp"
	"strings"
)

func ValidDNSName(input string) string {
	result := strings.ToLower(input)

	// Replace invalid characters with hyphens
	reg := regexp.MustCompile("[^a-z0-9.-]")
	result = reg.ReplaceAllString(result, "-")

	// Replace multiple consecutive hyphens with a single hyphen
	reg = regexp.MustCompile("-+")
	result = reg.ReplaceAllString(result, "-")

	// Replace multiple consecutive dots with a single dot
	reg = regexp.MustCompile("\\.+")
	result = reg.ReplaceAllString(result, ".")

	// Remove leading and trailing hyphens and dots
	result = strings.Trim(result, "-.")

	// Ensure it's not longer than 200 characters(keeping it less than 253 for safety)
	if len(result) > 200 {
		result = result[:200]
		// Ensure we don't end with hyphen or dot after truncating
		result = strings.TrimRight(result, "-.")
	}

	return result
}
