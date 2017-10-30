package main // import "github.com/mojlighetsministeriet/identity-provider"

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
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
	"github.com/mojlighetsministeriet/utils"
	"github.com/mojlighetsministeriet/utils/jwt"
	uuid "github.com/satori/go.uuid"
)

func main() {
	identityService := service.Service{}
	initializeErr := identityService.Initialize(
		utils.GetEnv("DATABASE_TYPE", "mysql"),
		utils.GetEnv(
			"DATABASE_CONNECTION",
			utils.GetFileAsString("/run/secrets/database-connection", "user:password@/dbname?charset=utf8mb4,utf8&parseTime=True&loc=Europe/Stockholm"),
		),
		utils.GetEnv("SMTP_HOST", ""),
		utils.GetEnvInt("SMTP_PORT", 0),
		utils.GetEnv("SMTP_EMAIL", ""),
		utils.GetFileAsString("/run/secrets/smtp-password", ""),
		utils.GetFileAsString("/run/secrets/private-key", ""),
	)
	if initializeErr != nil {
		identityService.Log.Error("Failed to initialize the service, make sure that you provided the correct database credentials.")
		identityService.Log.Error(initializeErr)
		panic("Cannot continue due to previous errors.")
	}
	defer identityService.Close()

	accountGroup := identityService.Router.Group("/account")
	accountGroup.Use(jwt.RequiredRoleMiddleware(&identityService.PrivateKey.PublicKey, "administrator"))

	// TODO: Add better validation error messages
	accountGroup.POST("", func(context echo.Context) error {
		entityWithPassword := entity.AccountWithPassword{}
		err := context.Bind(&entityWithPassword)
		if err != nil {
			return context.JSONBlob(http.StatusBadRequest, []byte("{\"message\":\"Bad Request\"}"))
		}

		account := entity.Account{}
		copier.Copy(&account, &entityWithPassword)

		resetToken := uuid.NewV4().String()
		if account.Password == "" {
			err = account.SetPasswordResetToken(resetToken)
			if err != nil {
				identityService.Log.Error(err)
				return context.JSONBlob(http.StatusBadRequest, []byte("{\"message\":\"Internal Server Error\"}"))
			}
		} else {
			err = account.SetPassword(account.Password)
			if err != nil {
				return context.JSONBlob(http.StatusBadRequest, []byte("{\"message\":\"Bad Request\"}"))
			}
		}

		if account.ID == "" {
			account.ID = uuid.NewV4().String()
		}

		validate := validator.New()
		err = validate.Struct(account)
		if err != nil {
			return context.JSONBlob(http.StatusBadRequest, []byte("{\"message\":\"Bad Request\"}"))
		}

		err = identityService.DatabaseConnection.Create(&account).Error
		if err != nil {
			// TODO: handle for non-MySQL databases as well
			if strings.HasPrefix(err.Error(), "Error 1062") {
				return context.JSONBlob(http.StatusBadRequest, []byte("{\"message\":\"The email or id was already taken\"}"))
			}

			identityService.Log.Error(err)
			return context.JSONBlob(http.StatusInternalServerError, []byte("{\"message\":\"Internal Server Error\"}"))
		}

		if account.Password == "" && account.PasswordResetToken != "" {
			// TODO: Email templates should be taken from environment variables
			err = identityService.Email.Send(
				account.Email,
				utils.GetFileAsString("EMAIL_ACCOUNT_CREATED_SUBJECT", "Your new account"),
				utils.GetFileAsString(
					"EMAIL_ACCOUNT_CREATED_BODY",
					fmt.Sprintf(
						"You have a new account, choose your password by visiting <a href=\"%s/reset-password/%s\" target=\"_blank\">%s/reset-password/%s</a>",
						identityService.ExternalURL,
						resetToken,
						identityService.ExternalURL,
						resetToken,
					),
				),
			)

			if err != nil {
				identityService.Log.Error(err)
			}
		}

		return context.JSONBlob(http.StatusCreated, []byte("{\"message\":\"Created\"}"))
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

	identityService.Router.POST("/account/:id/password-reset", func(context echo.Context) error {
		type resetPasswordBody struct {
			ResetToken string `json:"resetToken"`
			Password   string `json:"password"`
		}

		parameters := resetPasswordBody{}
		context.Bind(&parameters)

		// TODO: Add validation to input parameters
		if parameters.ResetToken == "" || parameters.Password == "" {
			return context.JSONBlob(http.StatusBadRequest, []byte("{\"message\":\"Bad Request\"}"))
		}

		// TODO: Add validation to input parameters
		if parameters.ResetToken == "" {
			parameters.ResetToken = uuid.NewV4().String()
		}

		account, err := entity.LoadAccountFromID(identityService.DatabaseConnection, context.Param("id"))
		if err != nil || account.CompareHashedPasswordResetTokenAgainst(parameters.ResetToken) != nil {
			return context.JSONBlob(http.StatusUnauthorized, []byte("{\"message\":\"Unauthorized\"}"))
		}

		err = account.SetPassword(parameters.Password)
		if err != nil {
			return context.JSONBlob(http.StatusBadRequest, []byte("{\"message\":\"Bad Request\"}"))
		}
		account.PasswordResetToken = ""

		validate := validator.New()
		err = validate.Struct(account)

		if err != nil {
			return context.JSONBlob(http.StatusBadRequest, []byte("{\"message\":\"Bad Request\"}"))
		}

		err = identityService.DatabaseConnection.Save(&account).Error
		if err != nil {
			identityService.Log.Error(err)
			return context.JSONBlob(http.StatusInternalServerError, []byte("{\"message\":\"Internal Server Error\"}"))
		}

		return context.JSONBlob(http.StatusOK, []byte("{\"message\":\"Password was reset\"}"))
	})

	identityService.Router.POST("/account/:id/reset-token", func(context echo.Context) error {
		type emailBody struct {
			Email string `json:"email"`
		}

		parameters := emailBody{}
		context.Bind(&parameters)

		account, err := entity.LoadAccountFromID(identityService.DatabaseConnection, context.Param("id"))
		if err != nil {
			return context.JSONBlob(http.StatusBadRequest, []byte("{\"message\":\"Bad Request\"}"))
		}

		if account.Email != parameters.Email {
			return context.JSONBlob(http.StatusBadRequest, []byte("{\"message\":\"Bad Request\"}"))
		}

		resetToken := uuid.NewV4().String()
		err = account.SetPasswordResetToken(resetToken)
		if err != nil {
			return context.JSONBlob(http.StatusBadRequest, []byte("{\"message\":\"Bad Request\"}"))
		}

		validate := validator.New()
		err = validate.Struct(account)
		if err != nil {
			return context.JSONBlob(http.StatusBadRequest, []byte("{\"message\":\"Bad Request\"}"))
		}

		// TODO: Send email to user

		err = identityService.DatabaseConnection.Save(&account).Error
		if err != nil {
			identityService.Log.Error(err)
			return context.JSONBlob(http.StatusInternalServerError, []byte("{\"message\":\"Internal Server Error\"}"))
		}

		// TODO: Email templates should be taken from environment variables
		err = identityService.Email.Send(
			account.Email,
			utils.GetFileAsString("EMAIL_ACCOUNT_RESET_SUBJECT", "Reset your password"),
			utils.GetFileAsString(
				"EMAIL_ACCOUNT_RESET_BODY",
				fmt.Sprintf(
					"You have requested to reset your password, choose your new password by visiting <a href=\"%s/reset-password/%s\" target=\"_blank\">%s/reset-password/%s</a>. If you did not request a password reset you can ignore this message.",
					identityService.ExternalURL,
					resetToken,
					identityService.ExternalURL,
					resetToken,
				),
			),
		)

		return context.JSON(http.StatusOK, []byte("{\"message\":\"Reset token created\"}"))
	})

	tokenGroup := identityService.Router.Group("/token")

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

		account, err := entity.LoadAccountFromEmailAndPassword(identityService.DatabaseConnection, parameters.Email, parameters.Password)
		if err != nil {
			return context.JSONBlob(http.StatusUnauthorized, []byte("{\"message\":\"Unauthorized\"}"))
		}

		newToken, err := jwt.Generate("identity-provider", identityService.PrivateKey, account)
		if err != nil {
			identityService.Log.Error(err)
			return context.JSONBlob(http.StatusInternalServerError, []byte("{\"message\":\"Internal Server Error\"}"))
		}

		return context.JSON(http.StatusCreated, struct {
			Token string `json:"token"`
		}{Token: string(newToken)})
	})

	tokenGroup.POST("/renew", func(context echo.Context) error {
		parsedToken, err := jwt.ParseIfValid(&identityService.PrivateKey.PublicKey, jwt.GetTokenFromContext(context))
		if err != nil {
			return context.JSONBlob(http.StatusUnauthorized, []byte("{\"message\":\"Unauthorized\"}"))
		}

		id := parsedToken.Claims().Get("sub").(string)
		account, err := entity.LoadAccountFromID(identityService.DatabaseConnection, id)
		if err != nil {
			return context.JSONBlob(http.StatusUnauthorized, []byte("{\"message\":\"Unauthorized\"}"))
		}

		newToken, err := jwt.Generate("identity-provider", identityService.PrivateKey, account)
		if err != nil {
			identityService.Log.Error(err)
			return context.JSONBlob(http.StatusInternalServerError, []byte("{\"message\":\"Internal Server Error\"}"))
		}

		return context.JSON(http.StatusCreated, struct {
			Token string `json:"token"`
		}{Token: string(newToken)})
	})

	tokenGroup.POST("/decode", func(context echo.Context) error {
		parsedToken, err := jwt.ParseIfValid(&identityService.PrivateKey.PublicKey, jwt.GetTokenFromContext(context))
		if err != nil {
			return context.JSONBlob(http.StatusUnauthorized, []byte("{\"message\":\"The token is invalid\"}"))
		}

		json, err := parsedToken.Claims().MarshalJSON()
		if err != nil {
			return context.JSONBlob(http.StatusUnauthorized, []byte("{\"message\":\"The token is invalid\"}"))
		}

		return context.JSONBlob(http.StatusOK, json)
	})

	identityService.Router.GET("/public-key", func(context echo.Context) error {
		body, err := x509.MarshalPKIXPublicKey(&identityService.PrivateKey.PublicKey)
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

	listenErr := identityService.Listen(":" + utils.GetEnv("PORT", "80"))
	if listenErr != nil {
		panic(listenErr)
	}
}
