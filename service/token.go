package service

import (
	"net/http"

	"github.com/labstack/echo"
	"github.com/mojlighetsministeriet/identity-provider/entity"
	"github.com/mojlighetsministeriet/utils/jwt"
	uuid "github.com/satori/go.uuid"
)

func (service *Service) tokenResource() {
	tokenGroup := service.Router.Group("/token")

	tokenGroup.POST("", func(context echo.Context) error {
		type createTokenBody struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		parameters := createTokenBody{}
		context.Bind(&parameters)

		// TODO: Add validation to input parameters
		// Set an invalid password if password was empty
		if parameters.Password == "" {
			parameters.Password = uuid.NewV4().String()
		}

		account, err := entity.LoadAccountFromEmailAndPassword(service.DatabaseConnection, parameters.Email, parameters.Password)
		if err != nil {
			return context.JSONBlob(http.StatusUnauthorized, []byte("{\"message\":\"Unauthorized\"}"))
		}

		newToken, err := jwt.Generate("identity-provider", service.PrivateKey, &account)
		if err != nil {
			service.Log.Error(err)
			return context.JSONBlob(http.StatusInternalServerError, []byte("{\"message\":\"Internal Server Error\"}"))
		}

		return context.JSON(http.StatusCreated, struct {
			Token string `json:"token"`
		}{Token: string(newToken)})
	})

	tokenGroup.POST("/renew", func(context echo.Context) error {
		parsedToken, err := jwt.ParseIfValid(&service.PrivateKey.PublicKey, jwt.GetTokenFromContext(context))
		if err != nil {
			return context.JSONBlob(http.StatusUnauthorized, []byte("{\"message\":\"Unauthorized\"}"))
		}

		id := parsedToken.Claims().Get("sub").(string)
		account, err := entity.LoadAccountFromID(service.DatabaseConnection, id)
		if err != nil {
			return context.JSONBlob(http.StatusUnauthorized, []byte("{\"message\":\"Unauthorized\"}"))
		}

		newToken, err := jwt.Generate("identity-provider", service.PrivateKey, &account)
		if err != nil {
			service.Log.Error(err)
			return context.JSONBlob(http.StatusInternalServerError, []byte("{\"message\":\"Internal Server Error\"}"))
		}

		return context.JSON(http.StatusCreated, struct {
			Token string `json:"token"`
		}{Token: string(newToken)})
	})

	tokenGroup.POST("/decode", func(context echo.Context) error {
		parsedToken, err := jwt.ParseIfValid(&service.PrivateKey.PublicKey, jwt.GetTokenFromContext(context))
		if err != nil {
			return context.JSONBlob(http.StatusUnauthorized, []byte("{\"message\":\"The token is invalid\"}"))
		}

		json, err := parsedToken.Claims().MarshalJSON()
		if err != nil {
			return context.JSONBlob(http.StatusUnauthorized, []byte("{\"message\":\"The token is invalid\"}"))
		}

		return context.JSONBlob(http.StatusOK, json)
	})
}
