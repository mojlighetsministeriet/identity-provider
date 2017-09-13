package utils

import (
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"
)

// GetFileAsString returns file contents or a default value
func GetFileAsString(path, fallback string) string {
	content, err := ioutil.ReadFile(path)

	if len(content) == 0 || err != nil {
		return fallback
	}

	return strings.Trim(string(content), "\n ")
}

// GetFileAsInt returns file contents or a default value as integer
func GetFileAsInt(path string, fallback int) int {
	content, err := ioutil.ReadFile(path)

	contentString := strings.Trim(string(content), "\n ")
	if len(contentString) == 0 || err != nil {
		return fallback
	}

	pattern := regexp.MustCompile("[^\\d]+")
	value, _ := strconv.Atoi(pattern.ReplaceAllString(contentString, ""))

	return value
}
