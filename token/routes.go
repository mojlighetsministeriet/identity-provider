package token

import (
	"log"
	"strings"

	"github.com/SermoDigital/jose/jws"
	"github.com/gin-gonic/gin"
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

func create(serviceInstance *service.Service) gin.HandlerFunc {
	return func(context *gin.Context) {
		parameters := createTokenBody{}
		context.BindJSON(&parameters)

		account, err := account.LoadAccountFromEmailAndPassword(
			serviceInstance.DatabaseConnection,
			parameters.Email,
			parameters.Password,
		)

		if err != nil {
			context.AbortWithStatus(401)
			return
		}

		token, err := Generate(serviceInstance.PrivateKey, account)

		if err != nil {
			log.Fatal(err)
			context.AbortWithStatus(500)
			return
		}

		context.JSON(201, struct {
			Token string `json:"token"`
		}{Token: string(token)})
	}
}

func renew(serviceInstance *service.Service) gin.HandlerFunc {
	return func(context *gin.Context) {
		token := getTokenFromContext(context)

		if Validate(&serviceInstance.PrivateKey.PublicKey, token) != nil {
			context.AbortWithStatus(401)
			return
		}

		jwt, err := jws.ParseJWT(token)
		if err != nil {
			context.AbortWithStatus(401)
			return
		}

		account, err := account.LoadAccountFromID(
			serviceInstance.DatabaseConnection,
			jwt.Claims().Get("sub").(string),
		)

		if err != nil {
			log.Fatal(err)
			context.AbortWithStatus(500)
			return
		}

		newToken, err := Generate(serviceInstance.PrivateKey, account)

		if err != nil {
			log.Fatal(err)
			context.AbortWithStatus(500)
			return
		}

		context.JSON(201, struct {
			Token string `json:"token"`
		}{Token: string(newToken)})
	}
}

func validate(serviceInstance *service.Service) gin.HandlerFunc {
	return func(context *gin.Context) {
		if Validate(&serviceInstance.PrivateKey.PublicKey, getTokenFromContext(context)) == nil {
			context.AbortWithStatus(200)
		} else {
			context.AbortWithStatus(401)
		}
	}
}

func getTokenFromContext(context *gin.Context) (result []byte) {
	token := context.GetHeader("Authorization")
	token = strings.Replace(token, "Bearer", "", -1)
	token = strings.Trim(strings.Replace(token, "bearer", "", -1), " ")

	if len(token) > 20 {
		result = []byte(token)
	}

	return
}
