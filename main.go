package main

import (
	"net/http"

	validator "gopkg.in/go-playground/validator.v9"

	"github.com/jinzhu/copier"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/labstack/echo"
	"github.com/mojlighetsministeriet/identity-provider/entity"
	"github.com/mojlighetsministeriet/identity-provider/service"
	"github.com/mojlighetsministeriet/identity-provider/token"
	"github.com/mojlighetsministeriet/identity-provider/utils"
	uuid "github.com/satori/go.uuid"
)

func main() {
	identityService := service.Service{}
	err := identityService.Initialize(utils.Getenv("DATABASE_TYPE", "sqlite3"), utils.Getenv("DATABASE_CREDENTIALS", "storage.db"))
	if err != nil {
		panic(err)
	}
	defer identityService.Close()

	identityService.DatabaseConnection.AutoMigrate(&entity.Account{})

	accountGroup := identityService.Router.Group("/account")
	accountGroup.Use(token.JWTRequiredRoleMiddleware(&identityService.PrivateKey.PublicKey, "administrator"))

	// TODO: Add better validation error messages
	accountGroup.POST("", func(context echo.Context) error {
		entityWithPassword := entity.AccountWithPassword{}
		err := context.Bind(&entityWithPassword)
		if err != nil {
			return context.JSONBlob(http.StatusBadRequest, []byte("{\"message\":\"Bad Request\"}"))
		}

		account := entity.Account{}
		copier.Copy(&account, &entityWithPassword)

		if account.Password != "" {
			account.SetPassword(account.Password)
		}

		if account.ID.String() == "00000000-0000-0000-0000-000000000000" {
			account.ID = uuid.NewV4()
		}

		validate := validator.New()
		err = validate.Struct(account)
		if err != nil {
			return context.JSONBlob(http.StatusBadRequest, []byte("{\"message\":\"Bad Request\"}"))
		}

		err = identityService.DatabaseConnection.Create(account).Error
		if err != nil {
			identityService.Log.Error(err)
			return context.JSONBlob(http.StatusInternalServerError, []byte("{\"message\":\"Internal Server Error\"}"))
		}

		return context.JSON(http.StatusCreated, account)
	})

	accountGroup.GET("", func(context echo.Context) error {
		var entities []entity.Account

		err := identityService.DatabaseConnection.Find(&entities).Error
		if err == nil {
			return context.JSON(http.StatusOK, entities)
		}

		identityService.Log.Error(err)
		return context.JSONBlob(http.StatusInternalServerError, []byte("{\"message\":\"Internal Server Error\"}"))
	})

	identityService.Listen(":1323")
}
