package token

import (
	"net/http"
	"strings"

	"github.com/SermoDigital/jose/jws"
	"github.com/labstack/echo"
	"github.com/mojlighetsministeriet/identity-provider/account"
	"github.com/mojlighetsministeriet/identity-provider/service"
)

func RegisterResource(serviceInstance *service.Service) {
	serviceInstance.Router.POST("/token", create(serviceInstance))
	serviceInstance.Router.POST("/token/renew", renew(serviceInstance))
	serviceInstance.Router.POST("/token/validate", validate(serviceInstance))
}

type createTokenBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func create(serviceInstance *service.Service) echo.HandlerFunc {
	return func(context echo.Context) error {
		parameters := createTokenBody{}
		context.Bind(&parameters)

		account, err := account.LoadAccountFromEmailAndPassword(
			serviceInstance.DatabaseConnection,
			parameters.Email,
			parameters.Password,
		)
		if err != nil {
			serviceInstance.Logger.Error(err)
			return context.JSON(http.StatusUnauthorized, false)
		}

		token, err := Generate(serviceInstance.PrivateKey, account)
		if err != nil {
			serviceInstance.Logger.Error(err)
			return context.JSON(http.StatusInternalServerError, false)
		}

		return context.JSON(http.StatusCreated, struct {
			Token string `json:"token"`
		}{Token: string(token)})
	}
}

func renew(serviceInstance *service.Service) echo.HandlerFunc {
	return func(context echo.Context) error {
		token := getTokenFromContext(context)

		if Validate(&serviceInstance.PrivateKey.PublicKey, token) != nil {
			return context.JSON(http.StatusUnauthorized, false)
		}

		jwt, err := jws.ParseJWT(token)
		if err != nil {
			serviceInstance.Logger.Error(err)
			return context.JSON(http.StatusUnauthorized, false)
		}

		account, err := account.LoadAccountFromID(
			serviceInstance.DatabaseConnection,
			jwt.Claims().Get("sub").(string),
		)
		if err != nil {
			serviceInstance.Logger.Error(err)
			return context.JSON(http.StatusUnauthorized, false)
		}

		newToken, err := Generate(serviceInstance.PrivateKey, account)
		if err != nil {
			serviceInstance.Logger.Error(err)
			return context.JSON(http.StatusUnauthorized, false)
		}

		return context.JSON(http.StatusCreated, struct {
			Token string `json:"token"`
		}{Token: string(newToken)})
	}
}

func validate(serviceInstance *service.Service) echo.HandlerFunc {
	return func(context echo.Context) error {
		if Validate(&serviceInstance.PrivateKey.PublicKey, getTokenFromContext(context)) == nil {
			return context.JSON(http.StatusOK, true)
		}

		return context.JSON(http.StatusUnauthorized, false)
	}
}

func getTokenFromContext(context echo.Context) (result []byte) {
	token := context.Request().Header.Get("Authorization")
	token = strings.Replace(token, "Bearer", "", -1)
	token = strings.Trim(strings.Replace(token, "bearer", "", -1), " ")

	if len(token) > 20 {
		result = []byte(token)
	}

	return
}
