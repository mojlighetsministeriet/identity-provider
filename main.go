package main // import "github.com/mojlighetsministeriet/identity-provider"

import (
	_ "github.com/jinzhu/gorm/dialects/mssql"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/mojlighetsministeriet/identity-provider/service"
	"github.com/mojlighetsministeriet/utils"
	"github.com/mojlighetsministeriet/utils/emailtemplates"
)

func main() {
	identityService := service.Service{}

	newAccountTemplate := emailtemplates.Template{
		Subject: utils.GetEnv("EMAIL_ACCOUNT_CREATED_SUBJECT", "Your new account"),
		Body:    utils.GetEnv("EMAIL_ACCOUNT_CREATED_BODY", "You have a new account, choose your password <a href=\"{{.ServiceURL}}/reset-password?token={{.ResetToken}}\" target=\"_blank\">here</a>."),
	}
	resetPasswordTemplate := emailtemplates.Template{
		Subject: utils.GetEnv("EMAIL_ACCOUNT_RESET_SUBJECT", "Password reset"),
		Body:    utils.GetEnv("EMAIL_ACCOUNT_RESET_BODY", "You have requested a password reset, choose your new password <a href=\"{{.ServiceURL}}/reset-password?token={{.ResetToken}}\" target=\"_blank\">here</a>. If you did not request a password reset, please ignore this message."),
	}

	initializeErr := identityService.Initialize(
		utils.GetEnv("DATABASE_TYPE", "mysql"),
		utils.GetFileAsString("/run/secrets/database-connection",
			utils.GetEnv(
				"DATABASE_CONNECTION",
				"user:password@/dbname?charset=utf8mb4,utf8&parseTime=True&loc=Europe/Stockholm",
			),
		),
		utils.GetFileAsString(
			"/run/secrets/private-key",
			utils.GetEnv("PRIVATE_KEY", ""),
		),
		newAccountTemplate,
		resetPasswordTemplate,
	)

	if initializeErr != nil {
		identityService.Log.Error("Failed to initialize the service, make sure that you provided the correct database credentials.")
		identityService.Log.Error(initializeErr)
		panic("Cannot continue due to previous errors.")
	}

	defer identityService.Close()

	listenErr := identityService.Listen(":" + utils.GetEnv("PORT", "80"))
	if listenErr != nil {
		panic(listenErr)
	}
}
