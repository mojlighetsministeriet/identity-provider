package main

import (
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	uuid "github.com/satori/go.uuid"

	"github.com/mojlighetsministeriet/identity-provider/account"
	"github.com/mojlighetsministeriet/identity-provider/service"
	"github.com/mojlighetsministeriet/identity-provider/token"
)

func main() {
	//databaseConnectionString := os.Getenv("DATABASE_CONNECTION")

	serviceInstance := service.Service{}
	defer serviceInstance.Close()

	//err := serviceInstance.Initialize("mysql", databaseConnectionString)
	err := serviceInstance.Initialize("sqlite3", "/tmp/identity-provider-test-"+uuid.NewV4().String()+".db")
	if err != nil {
		serviceInstance.Logger.Error("Unable to connect to database, please verify the connection string")
	}

	account.RegisterResource(&serviceInstance)
	token.RegisterResource(&serviceInstance)

	err = serviceInstance.Listen(":1323")
	if err != nil {
		serviceInstance.Logger.Error(err)
	}
}
