package service

import (
	"net/http"
	"strings"

	"github.com/labstack/echo"
)

func (service *Service) indexResource() {
	type routeInfo struct {
		Path   string `json:"path"`
		Method string `json:"method"`
	}
	var registeredRoutes []routeInfo
	for _, route := range service.Router.Routes() {
		if !strings.HasSuffix(route.Path, "/*") {
			registeredRoute := routeInfo{
				Path:   route.Path,
				Method: route.Method,
			}
			registeredRoutes = append(registeredRoutes, registeredRoute)
		}
	}

	service.Router.GET("/", func(context echo.Context) error {
		return context.JSON(http.StatusOK, registeredRoutes)
	})
}
