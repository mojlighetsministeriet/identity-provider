package account

import (
	"net/http"

	validator "gopkg.in/go-playground/validator.v9"

	"github.com/jinzhu/copier"
	"github.com/labstack/echo"
	"github.com/mojlighetsministeriet/identity-provider/service"
	uuid "github.com/satori/go.uuid"
)

// RegisterResource will register this resource to the HTTP service
func RegisterResource(serviceInstance *service.Service) {
	serviceInstance.DatabaseConnection.AutoMigrate(&Account{})

	serviceInstance.Router.GET("/account", list(serviceInstance))
	serviceInstance.Router.POST("/account", create(serviceInstance))
}

func list(serviceInstance *service.Service) echo.HandlerFunc {
	return func(context echo.Context) error {
		var entities []Account

		err := serviceInstance.DatabaseConnection.Find(&entities).Error
		if err == nil {
			return context.JSON(http.StatusOK, entities)
		}

		serviceInstance.Logger.Error(err)
		return context.String(http.StatusInternalServerError, "Internal Server Error")
	}
}

func create(serviceInstance *service.Service) echo.HandlerFunc {
	return func(context echo.Context) error {
		entityWithPassword := AccountWithPassword{}
		err := context.Bind(&entityWithPassword)
		if err != nil {
			return context.String(http.StatusBadRequest, err.Error())
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
			return context.String(http.StatusBadRequest, err.Error())
		}

		err = serviceInstance.DatabaseConnection.Create(entity).Error

		if err != nil {
			return context.String(http.StatusBadRequest, err.Error())
		}

		return context.JSON(http.StatusCreated, entity)
	}
}
