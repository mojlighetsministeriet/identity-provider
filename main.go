package main

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/mojlighetsministeriet/identity-provider/users"
)

// Service will hold the entire HTTP service
type Service struct {
	Router             *mux.Router
	DatabaseConnection *gorm.DB
}

// Setup database connection and all the routes
func (service *Service) Setup(databaseType string, databaseConnectionString string) {
	var err error
	service.DatabaseConnection, err = gorm.Open(databaseType, databaseConnectionString)
	if err != nil {
		panic("Failed to connect database")
	}
	service.DatabaseConnection.AutoMigrate(&users.User{})

	service.Router = mux.NewRouter()
	service.Router.HandleFunc("/", service.apiInfoHandler)

	service.Router.HandleFunc("/token", service.generateTokenHandler).Methods("POST")
	service.Router.HandleFunc("/token/renew", service.renewTokenHandler).Methods("POST")
	service.Router.HandleFunc("/token", service.deleteTokenHandler).Methods("DELETE")

	id := "{id:[0-9a-f]{8}\\-[0-9a-f]{4}\\-[0-9a-f]{4}\\-[0-9a-f]{4}\\-[0-9a-f]{12}}"
	service.Router.HandleFunc("/user", service.listUsersHandler).Methods("GET")
	service.Router.HandleFunc("/user/"+id, service.updateUserHandler).Methods("PUT")
	service.Router.HandleFunc("/user", service.createUserHandler).Methods("POST")
	service.Router.HandleFunc("/user/"+id, service.deleteUserHandler).Methods("DELETE")

	//http.Handle("/", service.Router)
}

// Listen for incoming HTTP calls on a port
func (service *Service) Listen(address string) {
	err := http.ListenAndServe(address, service.Router)
	if err != nil {
		panic("Failed to start HTTP server on address " + address)
	}
}

func main() {
	service := Service{}
	service.Setup("mysql", "root:hej256@/identity-provider?charset=utf8&parseTime=True&loc=Local")
	service.Listen(":8080")
}

func respondWithError(response http.ResponseWriter, code int, message string) {
	respondWithJSON(response, code, map[string]string{"error": message})
}

func respondWithJSON(response http.ResponseWriter, code int, payload interface{}) {
	body, _ := json.Marshal(payload)
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(code)
	response.Write(body)
}
