package service

import (
	"crypto/x509"
	"encoding/pem"
	"net/http"

	"github.com/labstack/echo"
)

func (service *Service) publicKeyResource() {
	service.Router.GET("/public-key", func(context echo.Context) error {
		body, err := x509.MarshalPKIXPublicKey(&service.PrivateKey.PublicKey)
		if err != nil {
			return context.JSONBlob(http.StatusInternalServerError, []byte("{\"message\":\"Internal Server Error\"}"))
		}
		block := pem.Block{
			Type:    "PUBLIC KEY",
			Headers: nil,
			Bytes:   body,
		}
		key := pem.EncodeToMemory(&block)
		return context.Blob(http.StatusOK, "application/x-pem-file", key)
	})
}
