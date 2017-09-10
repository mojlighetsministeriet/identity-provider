package main // import "github.com/mojlighetsministeriet/identity-provider"

import (
	"net/http"
	"strings"

	validator "gopkg.in/go-playground/validator.v9"

	"github.com/jinzhu/copier"
	_ "github.com/jinzhu/gorm/dialects/mssql"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/labstack/echo"
	"github.com/mojlighetsministeriet/identity-provider/entity"
	"github.com/mojlighetsministeriet/identity-provider/service"
	"github.com/mojlighetsministeriet/identity-provider/token"
	"github.com/mojlighetsministeriet/identity-provider/utils"
	uuid "github.com/satori/go.uuid"
)

func main() {
	identityService := service.Service{}
	err := identityService.Initialize(
		utils.Getenv("DATABASE_TYPE", "mysql"),
		utils.Getenv("DATABASE", "user:password@/dbname?charset=utf8mb4,utf8&parseTime=True&loc=Local"),
	)
	if err != nil {
		identityService.Log.Error("Failed to initialize the service, make sure that you provided the correct database credentials.")
		identityService.Log.Error(err)
		panic("Cannot continue due to above errors.")
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

	tokenGroup := identityService.Router.Group("/token")

	tokenGroup.POST("", func(context echo.Context) error {
		type createTokenBody struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		parameters := createTokenBody{}
		context.Bind(&parameters)

		account, err := entity.LoadAccountFromEmailAndPassword(identityService.DatabaseConnection, parameters.Email, parameters.Password)
		if err != nil {
			return context.JSONBlob(http.StatusUnauthorized, []byte("{\"message\":\"Unauthorized\"}"))
		}

		newToken, err := token.Generate(identityService.PrivateKey, account)
		if err != nil {
			identityService.Log.Error(err)
			return context.JSONBlob(http.StatusUnauthorized, []byte("{\"message\":\"Internal Server Error\"}"))
		}

		return context.JSON(http.StatusCreated, struct {
			Token string `json:"token"`
		}{Token: string(newToken)})
	})

	tokenGroup.POST("/renew", func(context echo.Context) error {
		parsedToken, err := token.ParseIfValid(&identityService.PrivateKey.PublicKey, token.GetTokenFromContext(context))
		if err != nil {
			return context.JSONBlob(http.StatusUnauthorized, []byte("{\"message\":\"Unauthorized\"}"))
		}

		idClaim := parsedToken.Claims().Get("id").(string)
		id, err := uuid.FromString(idClaim)
		if err != nil {
			return context.JSONBlob(http.StatusUnauthorized, []byte("{\"message\":\"Unauthorized\"}"))
		}

		account, err := entity.LoadAccountFromID(identityService.DatabaseConnection, id)
		if err != nil {
			return context.JSONBlob(http.StatusUnauthorized, []byte("{\"message\":\"Unauthorized\"}"))
		}

		newToken, err := token.Generate(identityService.PrivateKey, account)
		if err != nil {
			identityService.Log.Error(err)
			return context.JSONBlob(http.StatusUnauthorized, []byte("{\"message\":\"Internal Server Error\"}"))
		}

		return context.JSON(http.StatusCreated, struct {
			Token string `json:"token"`
		}{Token: string(newToken)})
	})

	tokenGroup.POST("/validate", func(context echo.Context) error {
		_, err := token.ParseIfValid(&identityService.PrivateKey.PublicKey, token.GetTokenFromContext(context))
		if err == nil {
			return context.JSONBlob(http.StatusOK, []byte("{\"message\":\"The token is valid\"}"))
		}

		return context.JSONBlob(http.StatusUnauthorized, []byte("{\"message\":\"The token is invalid\"}"))
	})

	type routeInfo struct {
		Path   string `json:"path"`
		Method string `json:"method"`
	}
	var registeredRoutes []routeInfo
	for _, route := range identityService.Router.Routes() {
		if !strings.HasSuffix(route.Path, "/*") {
			registeredRoute := routeInfo{
				Path:   route.Path,
				Method: route.Method,
			}
			registeredRoutes = append(registeredRoutes, registeredRoute)
		}
	}

	identityService.Router.GET("/", func(context echo.Context) error {
		return context.JSON(http.StatusOK, registeredRoutes)
	})

	identityService.Listen(":1323")
}
