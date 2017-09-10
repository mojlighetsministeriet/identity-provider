package utils

import (
	"os"
	"regexp"
	"strconv"
)

// Getenv returns an environment variable or a default value
func Getenv(key, fallback string) string {
	value := os.Getenv(key)

	if len(value) == 0 {
		return fallback
	}

	return value
}

// GetenvInt returns an environment variable or a default value as integer
func GetenvInt(key string, fallback int) int {
	valueAsString := os.Getenv(key)
	pattern := regexp.MustCompile("[^\\d]+")
	value, err := strconv.Atoi(pattern.ReplaceAllString(valueAsString, ""))

	if valueAsString == "" || err != nil {
		return fallback
	}

	return value
}
