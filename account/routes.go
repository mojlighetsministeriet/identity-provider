package account

import (
	"log"

	validator "gopkg.in/go-playground/validator.v9"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"github.com/mojlighetsministeriet/identity-provider/service"
	uuid "github.com/satori/go.uuid"
)

func RegisterResource(serviceInstance *service.Service) {
	serviceInstance.DatabaseConnection.AutoMigrate(&Account{})

	serviceInstance.Router.GET("/account", list(serviceInstance))
	serviceInstance.Router.POST("/account", create(serviceInstance))
}

func list(serviceInstance *service.Service) gin.HandlerFunc {
	return func(context *gin.Context) {
		var entities []Account

		err := serviceInstance.DatabaseConnection.Find(&entities).Error
		if err == nil {
			context.JSON(200, entities)
		} else {
			log.Fatal(err)
			context.AbortWithStatus(500)
		}
	}
}

func create(serviceInstance *service.Service) gin.HandlerFunc {
	return func(context *gin.Context) {
		entityWithPassword := AccountWithPassword{}
		err := context.BindJSON(&entityWithPassword)
		if err != nil {
			context.AbortWithError(400, err)
			return
		}

		entity := Account{}
		copier.Copy(&entity, &entityWithPassword)

		if entity.Password != "" {
			entity.SetPassword(entity.Password)
		}

		if entity.ID == "" {
			entity.ID = uuid.NewV4().String()
		}

		validate := validator.New()
		err = validate.Struct(entity)

		if err != nil {
			context.AbortWithError(400, err)
			return
		}

		err = serviceInstance.DatabaseConnection.Create(entity).Error

		if err != nil {
			context.AbortWithError(400, err)
			return
		}

		context.BindJSON(entity)
	}
}
