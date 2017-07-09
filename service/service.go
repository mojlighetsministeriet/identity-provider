package service

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
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
	if service.PrivateKey.Validate() != nil {
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
	var err error
	var privateKey *rsa.PrivateKey

	privateKeyString := os.Getenv("RSA_KEY")
	if privateKeyString != "" {
		privateKey, err = x509.ParsePKCS1PrivateKey([]byte(privateKeyString))
		if err == nil {
			service.PrivateKey = privateKey
		}
	}

	privateKeyBytes, err := ioutil.ReadFile("key.private")
	if len(privateKeyBytes) > 0 && err == nil {
		privateKey, err = x509.ParsePKCS1PrivateKey(privateKeyBytes)
		if err == nil {
			service.PrivateKey = privateKey
		}
	}

	if service.PrivateKey.Validate() != nil {
		log.Print("Unable to find an RSA key as environment variable RSA_KEY or as the file key.private, generating a new key.private file")

		service.PrivateKey, err = rsa.GenerateKey(rand.Reader, 4096)
		if err == nil {
			err = ioutil.WriteFile(
				"key.private",
				x509.MarshalPKCS1PrivateKey(service.PrivateKey),
				0600,
			)

			if err != nil {
				log.Fatal(err)
			}
		} else {
			log.Fatal(err)
		}

		err = service.PrivateKey.Validate()
		if err != nil {
			log.Fatal(err)
		}
	}
}
