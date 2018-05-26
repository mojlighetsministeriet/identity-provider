package service

import (
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"

	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/log"
	"github.com/mojlighetsministeriet/identity-provider/entity"
	"github.com/mojlighetsministeriet/utils"
	"github.com/mojlighetsministeriet/utils/emailtemplates"
	"github.com/mojlighetsministeriet/utils/httprequest"
	uuid "github.com/satori/go.uuid"
)

// EmailTemplates holds templates for all email types that the service can send, they are defined in subject/body pairs
type EmailTemplates struct {
	NewAccountSubject   string
	NewAccountBody      string
	PasswordRestSubject string
	PasswordRestBody    string
}

// Service is the main service that holds web server and database connections and so on
type Service struct {
	DatabaseConnection *gorm.DB
	Router             *echo.Echo
	PrivateKey         *rsa.PrivateKey
	Log                echo.Logger
	TLSConfig          *tls.Config
	EmailTemplates     emailtemplates.Templates
	HTTPClient         *httprequest.JSONClient
}

// Initialize will prepeare the service by connecting to database and creating a web server instance (but it will not start listening until service.Listen() is run)
func (service *Service) Initialize(databaseType string, databaseConnectionString string, rsaKeyPEMString string, newAccountTemplate emailtemplates.Template, resetPasswordTemplate emailtemplates.Template) (err error) {
	service.TLSConfig, err = utils.GetCACertificatesTLSConfig()
	if err != nil {
		return
	}

	service.Router = echo.New()
	service.Router.Use(middleware.Gzip())

	service.Log = service.Router.Logger
	service.Log.SetLevel(log.INFO)

	service.EmailTemplates = emailtemplates.Templates{}
	newAccountTemplate.Name = "new-account"
	service.EmailTemplates.Add(newAccountTemplate)
	resetPasswordTemplate.Name = "reset-password"
	service.EmailTemplates.Add(resetPasswordTemplate)

	service.HTTPClient, err = httprequest.NewJSONClient()
	if err != nil {
		return
	}

	service.DatabaseConnection, err = gorm.Open(databaseType, databaseConnectionString)
	if err != nil {
		return
	}

	err = service.DatabaseConnection.AutoMigrate(&entity.Account{}).Error
	if err != nil {
		return
	}

	service.setupAdministratorUserIfMissing()

	if service.PrivateKey == nil || service.PrivateKey.Validate() != nil {
		service.setupPrivateKey(rsaKeyPEMString)
	}

	service.accountResource()
	service.tokenResource()
	service.publicKeyResource()
	service.indexResource()

	return
}

// Listen will make the service start listning for incoming requests
func (service *Service) Listen(address string) (err error) {
	err = service.Router.Start(address)
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
	resetToken := uuid.Must(uuid.NewV4()).String()
	administrator.SetPasswordResetToken(resetToken)

	err = service.DatabaseConnection.Create(&administrator).Error
	if err == nil {
		// TODO: Change instructions to work with refactored version of the end points
		service.Log.Info(fmt.Sprintf("No account with administrator found, created a new account with email %s and reset token %s, reset password by POST account/reset-token/password Authorization: Bearer %s { \"password\": \"yournewpassword\" }", administrator.Email, resetToken, resetToken))
	}

	return
}

func (service *Service) setupPrivateKey(pemString string) (err error) {
	privateKey, err := pemStringToPrivateKey(pemString)
	if err != nil {
		return
	}

	service.PrivateKey = privateKey

	return
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
