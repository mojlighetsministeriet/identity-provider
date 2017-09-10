package service_test

import (
	"os"
	"testing"

	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/mojlighetsministeriet/identity-provider/service"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func TestServiceInitialize(test *testing.T) {
	os.Setenv("RSA_PRIVATE_KEY_BITS", "512")
	identityService := service.Service{}
	err := identityService.Initialize("sqlite3wq", "test-storage-"+uuid.NewV4().String()+".db")
	assert.NoError(test, err)
}
