package service_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"testing"

	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/mojlighetsministeriet/identity-provider/service"
	"github.com/mojlighetsministeriet/utils/emailtemplates"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func TestServiceInitialize(test *testing.T) {
	storage := "test-storage-" + uuid.NewV4().String() + ".db"
	defer os.Remove(storage)

	privateKey, err := rsa.GenerateKey(rand.Reader, 512)
	assert.NoError(test, err)
	block := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}
	pem := string(pem.EncodeToMemory(block))

	identityService := service.Service{}
	err = identityService.Initialize("sqlite3", storage, pem, emailtemplates.Template{}, emailtemplates.Template{})
	assert.NoError(test, err)
}
