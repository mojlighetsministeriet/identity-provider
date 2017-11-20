package service

import (
	"net/http"
	"strings"
	"time"

	"github.com/jinzhu/copier"
	"github.com/labstack/echo"
	"github.com/mojlighetsministeriet/identity-provider/entity"
	"github.com/mojlighetsministeriet/utils"
	"github.com/mojlighetsministeriet/utils/jwt"
	uuid "github.com/satori/go.uuid"
	validator "gopkg.in/go-playground/validator.v9"
)

func (service *Service) accountResource() {
	accountGroup := service.Router.Group("/account")
	accountGroup.Use(jwt.RequiredRoleMiddleware(&service.PrivateKey.PublicKey, "administrator"))

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
				service.Log.Error(err)
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

		err = service.DatabaseConnection.Create(&account).Error
		if err != nil {
			// TODO: handle for non-MySQL databases as well
			if strings.HasPrefix(err.Error(), "Error 1062") {
				return context.JSONBlob(http.StatusBadRequest, []byte("{\"message\":\"The email or id was already taken\"}"))
			}

			service.Log.Error(err)
			return context.JSONBlob(http.StatusInternalServerError, []byte("{\"message\":\"Internal Server Error\"}"))
		}

		if account.Password == "" && account.PasswordResetToken != "" {
			email, err := service.EmailTemplates.Render(
				"new-account",
				account.Email,
				nil,
				struct {
					ServiceURL string
					ResetToken string
				}{utils.GetOriginalSystemURLFromContext(context), resetToken},
			)
			if err != nil {
				service.Log.Error(err)
				return context.JSONBlob(http.StatusInternalServerError, []byte("{\"message\":\"Internal Server Error\"}"))
			}

			err = service.HTTPClient.Post("http://email", email, nil)
			if err != nil {
				service.Log.Error(err)
				return context.JSONBlob(http.StatusInternalServerError, []byte("{\"message\":\"Internal Server Error\"}"))
			}
		}

		return context.JSONBlob(http.StatusCreated, []byte("{\"message\":\"Created\"}"))
	})

	accountGroup.GET("", func(context echo.Context) error {
		var entities []entity.Account

		err := service.DatabaseConnection.Find(&entities).Error
		if err == nil {
			return context.JSON(http.StatusOK, entities)
		}

		service.Log.Error(err)
		return context.JSONBlob(http.StatusInternalServerError, []byte("{\"message\":\"Internal Server Error\"}"))
	})

	service.Router.POST("/account/reset-token/password", func(context echo.Context) error {
		type resetPasswordBody struct {
			Password string `json:"password"`
		}

		parameters := resetPasswordBody{}
		context.Bind(&parameters)

		// TODO: Add validation to input parameters
		if parameters.Password == "" {
			return context.JSONBlob(http.StatusBadRequest, []byte("{\"message\":\"Bad Request\"}"))
		}

		claims, err := jwt.GetClaimsFromContextIfValid(&service.PrivateKey.PublicKey, context)
		if err != nil {
			return context.JSONBlob(http.StatusUnauthorized, []byte("{\"message\":\"Unauthorized\"}"))
		}

		resetToken := jwt.GetTokenFromContext(context)
		account, err := entity.LoadAccountFromEmail(service.DatabaseConnection, claims.Get("email").(string))
		if err != nil || account.CompareHashedPasswordResetTokenAgainst(string(resetToken)) != nil {
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

		err = service.DatabaseConnection.Save(&account).Error
		if err != nil {
			service.Log.Error(err)
			return context.JSONBlob(http.StatusInternalServerError, []byte("{\"message\":\"Internal Server Error\"}"))
		}

		return context.JSONBlob(http.StatusOK, []byte("{\"message\":\"Password was reset\"}"))
	})

	service.Router.POST("/account/reset-token", func(context echo.Context) error {
		type emailBody struct {
			Email string `json:"email"`
		}

		parameters := emailBody{}
		context.Bind(&parameters)

		account, err := entity.LoadAccountFromEmail(service.DatabaseConnection, parameters.Email)
		if err != nil {
			return context.JSONBlob(http.StatusBadRequest, []byte("{\"message\":\"Bad Request\"}"))
		}

		if account.Email != parameters.Email {
			return context.JSONBlob(http.StatusBadRequest, []byte("{\"message\":\"Bad Request\"}"))
		}

		expiration := time.Now().Add(time.Duration(3600) * time.Second)
		resetToken, err := jwt.GenerateWithCustomExpiration(
			"identity-provider",
			service.PrivateKey,
			&entity.Account{Email: parameters.Email},
			expiration,
		)
		if err != nil {
			return context.JSONBlob(http.StatusBadRequest, []byte("{\"message\":\"Bad Request\"}"))
		}

		err = account.SetPasswordResetToken(string(resetToken))
		if err != nil {
			return context.JSONBlob(http.StatusBadRequest, []byte("{\"message\":\"Bad Request\"}"))
		}

		validate := validator.New()
		err = validate.Struct(account)
		if err != nil {
			return context.JSONBlob(http.StatusBadRequest, []byte("{\"message\":\"Bad Request\"}"))
		}

		err = service.DatabaseConnection.Save(&account).Error
		if err != nil {
			service.Log.Error(err)
			return context.JSONBlob(http.StatusInternalServerError, []byte("{\"message\":\"Internal Server Error\"}"))
		}

		email, err := service.EmailTemplates.Render(
			"reset-password",
			account.Email,
			nil,
			struct {
				ServiceURL string
				ResetToken string
			}{utils.GetOriginalSystemURLFromContext(context), string(resetToken)},
		)
		if err != nil {
			service.Log.Error(err)
			return context.JSONBlob(http.StatusInternalServerError, []byte("{\"message\":\"Internal Server Error\"}"))
		}

		err = service.HTTPClient.Post("http://email", email, nil)
		if err != nil {
			service.Log.Error(err)
			return context.JSONBlob(http.StatusInternalServerError, []byte("{\"message\":\"Internal Server Error\"}"))
		}

		return context.JSONBlob(http.StatusOK, []byte("{\"message\":\"Reset token created\"}"))
	})
}
