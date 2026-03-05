package utils

import (
	"regexp"
	"strings"
)

var e164Regex = regexp.MustCompile(`^\+[1-9]\d{7,14}$`)

// NormalizePhone trims spaces before validation/persistence.
func NormalizePhone(phone string) string {
	return strings.TrimSpace(phone)
}

// IsValidE164 validates phone number in E.164 format.
func IsValidE164(phone string) bool {
	return e164Regex.MatchString(phone)
}
