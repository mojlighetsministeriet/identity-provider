package utils_test

import (
	"os"
	"testing"

	"github.com/mojlighetsministeriet/identity-provider/utils"
	"github.com/stretchr/testify/assert"
)

func TestGetEnvWithMissingValueButWithDefault(test *testing.T) {
	value := utils.GetEnv("AN_ENVIRONMENT_VARIABLE", "this is default")
	assert.Equal(test, "this is default", value)
}

func TestGetEnvWithSetValueWithDefault(test *testing.T) {
	os.Setenv("ANOTHER_ENVIRONMENT_VARIABLE", "this is a value")
	value := utils.GetEnv("ANOTHER_ENVIRONMENT_VARIABLE", "this is default")
	assert.Equal(test, "this is a value", value)
}

func TestGetEnvIntWithMissingValueButWithDefault(test *testing.T) {
	value := utils.GetEnvInt("AN_INT_ENVIRONMENT_VARIABLE", 7000)
	assert.Equal(test, 7000, value)
}

func TestGetEnvIntWithSetValueWithDefault(test *testing.T) {
	os.Setenv("ANOTHER_INT_ENVIRONMENT_VARIABLE", "4000")
	value := utils.GetEnvInt("ANOTHER_INT_ENVIRONMENT_VARIABLE", 10000)
	assert.Equal(test, 4000, value)
}
