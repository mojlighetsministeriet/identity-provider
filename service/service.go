package service

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/labstack/gommon/log"
	"github.com/mojlighetsministeriet/identity-provider/email"
	"github.com/mojlighetsministeriet/identity-provider/entity"
	"github.com/mojlighetsministeriet/identity-provider/utils"
	uuid "github.com/satori/go.uuid"
)

// Service is the main service that holds web server and database connections and so on
type Service struct {
	ExternalURL        string
	DatabaseConnection *gorm.DB
	Router             *echo.Echo
	PrivateKey         *rsa.PrivateKey
	Log                echo.Logger
	Email              *email.SMTPSender
}

// Initialize will prepeare the service by connecting to database and creating a web server instance (but it will not start listening until service.Listen() is run)
func (service *Service) Initialize(databaseType string, databaseConnectionString string, smtpHost string, smtpPort int, smtpEmail string, smtpPassword string) (err error) {
	service.Router = echo.New()

	service.Log = service.Router.Logger
	service.Log.SetLevel(log.INFO)

	service.Email = &email.SMTPSender{
		Host:     smtpHost,
		Port:     smtpPort,
		Email:    smtpEmail,
		Password: smtpPassword,
	}

	service.DatabaseConnection, err = gorm.Open(databaseType, databaseConnectionString)
	if err != nil {
		return
	}

	service.DatabaseConnection.Debug()

	err = service.DatabaseConnection.AutoMigrate(&entity.Account{}).Error
	if err != nil {
		return
	}

	service.setupAdministratorUserIfMissing()

	if service.PrivateKey == nil || service.PrivateKey.Validate() != nil {
		service.setupPrivateKey()
	}

	return
}

// Listen will make the service start listning for incoming requests
func (service *Service) Listen(address string) (err error) {
	service.Router.Logger.Error(service.Router.Start(address))
	return
}

// Close will shut down the service and any of it's related components
func (service *Service) Close() {
	service.DatabaseConnection.Close()
}

func (service *Service) setupAdministratorUserIfMissing() (err error) {
	administrator := entity.Account{}

	err = service.DatabaseConnection.Where("roles_serialized LIKE (?)", "administrator").First(&administrator).Error
	if err == nil || err.Error() != "record not found" {
		return
	}

	administrator.Email = "administrator@identity-provider.localhost"
	administrator.Roles = []string{"user", "administrator"}
	resetToken := uuid.NewV4().String()
	administrator.SetPasswordResetToken(resetToken)

	err = service.DatabaseConnection.Create(&administrator).Error
	if err == nil {
		service.Log.Info(fmt.Sprintf("No account with administrator found, created a new account with email %s and reset token %s, reset password by POST account/%s/password-reset { \"resetToken\": \"%s\", \"password\": \"yournewpassword\" }", administrator.Email, resetToken, administrator.ID, resetToken))
	}

	return
}

func (service *Service) setupPrivateKey() {
	privateKey, err := pemStringToPrivateKey(os.Getenv("RSA_PRIVATE_KEY"))
	if err == nil {
		service.PrivateKey = privateKey
	}

	if service.PrivateKey == nil {
		privateKeyBytes, err := ioutil.ReadFile("key.private")

		if err == nil {
			privateKey, err = pemStringToPrivateKey(string(privateKeyBytes))
			if err == nil {
				service.PrivateKey = privateKey
			}
		}

		if service.PrivateKey == nil {
			service.Log.Info("Unable to find a valid RSA key as environment variable RSA_PRIVATE_KEY or as the file key.private, generating a new key.private file")

			privateKey, err = rsa.GenerateKey(rand.Reader, utils.GetenvInt("RSA_PRIVATE_KEY_BITS", 4096))
			if err == nil {
				service.PrivateKey = privateKey
				block := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}

				err = ioutil.WriteFile(
					"key.private",
					pem.EncodeToMemory(block),
					0600,
				)

				if err != nil {
					service.Log.Error(err)
				}
			} else {
				service.Log.Error(err)
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
