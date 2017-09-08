package token

import (
	"testing"

	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/mojlighetsministeriet/identity-provider/service"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func TestRegisterResource(t *testing.T) {
	serviceInstance := service.Service{}
	defer serviceInstance.Close()

	err := serviceInstance.Initialize("sqlite3", "/tmp/identity-provider-test-"+uuid.NewV4().String()+".db")
	assert.NoError(t, err)
}
