package utils_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/mojlighetsministeriet/identity-provider/utils"
	"github.com/stretchr/testify/assert"
)

func TestGetFileAsStringWithMissingValueButWithDefault(test *testing.T) {
	value := utils.GetFileAsString("/tmp/a-secret", "this is default")
	assert.Equal(test, "this is default", value)
}

func TestGetFileAsStringWithSetValueWithDefault(test *testing.T) {
	err := ioutil.WriteFile("/tmp/another-secret", []byte("this is a value"), 0500)
	assert.NoError(test, err)
	value := utils.GetFileAsString("/tmp/another-secret", "this is default")
	defer os.Remove("/tmp/another-secret")
	assert.Equal(test, "this is a value", value)
}

func TestGetFileAsIntWithMissingValueButWithDefault(test *testing.T) {
	value := utils.GetFileAsInt("/tmp/an-int-secret", 7000)
	assert.Equal(test, 7000, value)
}

func TestGetFileAsIntWithSetValueWithDefault(test *testing.T) {
	err := ioutil.WriteFile("/tmp/another-int-secret", []byte("4000"), 0500)
	assert.NoError(test, err)
	value := utils.GetFileAsInt("/tmp/another-int-secret", 10000)
	defer os.Remove("/tmp/another-int-secret")
	assert.Equal(test, 4000, value)
}

func TestGetFileAsIntWithEmptyContent(test *testing.T) {
	err := ioutil.WriteFile("/tmp/another-empty-int-secret", []byte(""), 0500)
	assert.NoError(test, err)
	value := utils.GetFileAsInt("/tmp/another-empty-int-secret", 20000)
	defer os.Remove("/tmp/another-empty-int-secret")
	assert.Equal(test, 20000, value)
}
