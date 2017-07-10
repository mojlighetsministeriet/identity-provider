package service

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io/ioutil"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

type Service struct {
	DatabaseConnection *gorm.DB
	Router             *gin.Engine
	PrivateKey         *rsa.PrivateKey
}

func (service *Service) Initialize(databaseType string, databaseConnectionString string) (err error) {
	if service.PrivateKey == nil || service.PrivateKey.Validate() != nil {
		service.setupPrivateKey()
	}

	service.Router = gin.Default()
	service.DatabaseConnection, err = gorm.Open(databaseType, databaseConnectionString)
	return
}

func (service *Service) Listen(address string) (err error) {
	err = service.Router.Run(address)
	return
}

func (service *Service) Close() {
	service.DatabaseConnection.Close()
}

func (service *Service) setupPrivateKey() {
	privateKey, err := pemStringToPrivateKey(os.Getenv("RSA_PRIVATE_KEY"))
	if err == nil {
		service.PrivateKey = privateKey
	}

	if service.PrivateKey == nil {
		privateKeyBytes, err := ioutil.ReadFile("key.private")
		if err == nil {
			privateKey, err := pemStringToPrivateKey(string(privateKeyBytes))

			if err == nil {
				service.PrivateKey = privateKey
			} else {
				log.Print("Unable to find a valid RSA key as environment variable RSA_PRIVATE_KEY or as the file key.private, generating a new key.private file")

				privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
				if err == nil {
					service.PrivateKey = privateKey
					block := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}

					err = ioutil.WriteFile(
						"key.private",
						pem.EncodeToMemory(block),
						0600,
					)

					if err != nil {
						log.Fatal(err)
					}
				} else {
					log.Fatal(err)
				}
			}
		}
	}
}

func pemStringToPrivateKey(pemString string) (privateKey *rsa.PrivateKey, err error) {
	block, _ := pem.Decode([]byte(pemString))
	if block == nil {
		err = errors.New("Unable to find private key in pemString")
		return
	} else if block.Type != "RSA PRIVATE KEY" {
		err = errors.New("Unable to find private key in pemString, type found " + block.Type)
		return
	}

	privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	return
}
