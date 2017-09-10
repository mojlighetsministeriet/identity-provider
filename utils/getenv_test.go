package utils_test

import (
	"os"
	"testing"

	"github.com/mojlighetsministeriet/identity-provider/utils"
	"github.com/stretchr/testify/assert"
)

func TestGetenvWithMissingValueButWithDefault(test *testing.T) {
	value := utils.Getenv("AN_ENVIRONMENT_VARIABLE", "this is default")
	assert.Equal(test, "this is default", value)
}

func TestGetenvWithSetValueWithDefault(test *testing.T) {
	os.Setenv("ANOTHER_ENVIRONMENT_VARIABLE", "this is a value")
	value := utils.Getenv("ANOTHER_ENVIRONMENT_VARIABLE", "this is default")
	assert.Equal(test, "this is a value", value)
}

func TestGetenvIntWithMissingValueButWithDefault(test *testing.T) {
	value := utils.GetenvInt("AN_INT_ENVIRONMENT_VARIABLE", 7000)
	assert.Equal(test, 7000, value)
}

func TestGetenvIntWithSetValueWithDefault(test *testing.T) {
	os.Setenv("ANOTHER_INT_ENVIRONMENT_VARIABLE", "4000")
	value := utils.GetenvInt("ANOTHER_INT_ENVIRONMENT_VARIABLE", 10000)
	assert.Equal(test, 4000, value)
}
