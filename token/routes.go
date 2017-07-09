package token

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/mojlighetsministeriet/identity-provider/account"
	"github.com/mojlighetsministeriet/identity-provider/service"
)

func RegisterResource(serviceInstance *service.Service, privateKey []byte, publicKey []byte) {
	serviceInstance.Router.POST("/token", create(serviceInstance))
}

type createTokenBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func create(serviceInstance *service.Service) gin.HandlerFunc {
	return func(context *gin.Context) {
		fmt.Print("create token")
		parameters := createTokenBody{}
		context.BindJSON(&parameters)
		fmt.Print(parameters)

		account, err := account.LoadAccountFromEmailAndPassword(
			serviceInstance.DatabaseConnection,
			parameters.Email,
			parameters.Password,
		)

		if err != nil {
			context.AbortWithStatus(400)
			return
		}

		token, err := Generate(, account)

		if err != nil {
			log.Fatal(err)
			context.AbortWithStatus(500)
			return
		}

		context.BindJSON(struct {
			Token string `json:"token"`
		}{Token: string(token)})
	}
}
