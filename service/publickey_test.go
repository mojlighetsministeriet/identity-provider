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
	"github.com/mojlighetsministeriet/utils/httprequest"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func TestServicePublicKey(test *testing.T) {
	storage := "test-storage-" + uuid.Must(uuid.NewV4()).String() + ".db"
	defer os.Remove(storage)

	privateKey, err := rsa.GenerateKey(rand.Reader, 512)
	assert.NoError(test, err)
	block := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}
	privateKeyString := string(pem.EncodeToMemory(block))

	identityService := service.Service{}
	defer identityService.Close()
	identityService.Initialize("sqlite3", storage, privateKeyString, emailtemplates.Template{}, emailtemplates.Template{})
	assert.NoError(test, err)

	go func() {
		err = identityService.Listen(":10001")
		assert.NoError(test, err)
	}()

	client, err := httprequest.NewClient()
	assert.NoError(test, err)

	publicKey, err := client.Get("http://localhost:10001/public-key")
	assert.NoError(test, err)
	assert.Equal(test, true, len(publicKey) > 0)

	body, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	assert.NoError(test, err)
	expectedPublicKey := pem.EncodeToMemory(&pem.Block{
		Type:    "PUBLIC KEY",
		Headers: nil,
		Bytes:   body,
	})

	assert.Equal(test, expectedPublicKey, publicKey)
}
