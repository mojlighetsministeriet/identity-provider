package main

import (
	"log"

	_ "github.com/jinzhu/gorm/dialects/mysql"

	"github.com/mojlighetsministeriet/identity-provider/account"
	"github.com/mojlighetsministeriet/identity-provider/service"
	"github.com/mojlighetsministeriet/identity-provider/token"
)

func main() {
	serviceInstance := service.Service{}
	defer serviceInstance.Close()

	err := serviceInstance.Initialize("mysql", "root:hej256@/identity-provider?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		log.Fatal(err)
	}

	account.RegisterResource(&serviceInstance)
	token.RegisterResource(&serviceInstance)

	err = serviceInstance.Listen(":8080")
	if err != nil {
		log.Fatal(err)
	}
}
