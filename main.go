package main // import "github.com/mojlighetsministeriet/identity-provider"

import (
	_ "github.com/jinzhu/gorm/dialects/mssql"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/mojlighetsministeriet/identity-provider/service"
	"github.com/mojlighetsministeriet/utils"
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

	listenErr := identityService.Listen(":" + utils.GetEnv("PORT", "80"))
	if listenErr != nil {
		panic(listenErr)
	}
}
