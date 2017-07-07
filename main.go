package main

import (
	"encoding/json"
	"log"
	"net/http"

	validator "gopkg.in/go-playground/validator.v9"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/mojlighetsministeriet/identity-provider/users"
	uuid "github.com/satori/go.uuid"
)

// Service will hold the entire HTTP service
type Service struct {
	Router             *mux.Router
	DatabaseConnection *gorm.DB
}

func (service *Service) getUsersListHandler(response http.ResponseWriter, request *http.Request) {
	var usersList []users.User
	err := service.DatabaseConnection.Find(&usersList).Error
	if err != nil {
		log.Fatal(err)
		respondWithError(response, 500, "Internal Server Error")
	} else {
		respondWithJSON(response, 200, usersList)
	}
}

func (service *Service) getUserHandler(response http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	user := users.User{}
	result := service.DatabaseConnection.Where("id = ?", vars["id"]).First(&user)
	if result.Error != nil {
		if result.RecordNotFound() {
			respondWithError(response, 404, "Not Found")
		} else {
			log.Fatal(result.Error)
			respondWithError(response, 500, "Internal Server Error")
		}
	} else {
		respondWithJSON(response, 200, user)
	}
}

func (service *Service) updateUserHandler(response http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	userID := vars["id"]
	user := users.User{}
	decoder := json.NewDecoder(request.Body)
	defer request.Body.Close()

	err := decoder.Decode(&user)
	if err != nil {
		respondWithError(response, 400, "Bad Request")
		return
	}

	user.ID, err = uuid.FromString(userID)
	if err != nil {
		respondWithError(response, 400, "Bad Request")
		return
	}

	validate := validator.New()
	err = validate.Struct(user)
	if err != nil {
		respondWithJSON(response, 400, err)
		return
	}

	err = service.DatabaseConnection.Update(user).Error
	if err != nil {
		log.Fatal(err)
		respondWithError(response, 500, "Internal Server Error")
	} else {
		respondWithJSON(response, 200, user)
	}
}

func (service *Service) createUserHandler(response http.ResponseWriter, request *http.Request) {
	user := users.User{}
	decoder := json.NewDecoder(request.Body)
	defer request.Body.Close()

	err := decoder.Decode(&user)
	if err != nil {
		respondWithError(response, 400, "Bad Request")
		return
	}

	validate := validator.New()
	err = validate.Struct(user)
	if err != nil {
		respondWithJSON(response, 400, err)
		return
	}

	err = service.DatabaseConnection.Save(user).Error
	if err != nil {
		log.Fatal(err)
		respondWithError(response, 500, "Internal Server Error")
	} else {
		respondWithJSON(response, 201, user)
	}
}

func (service *Service) deleteUserHandler(response http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	userID := vars["id"]
	user := users.User{}
	var err error
	user.ID, err = uuid.FromString(userID)
	if err != nil {
		respondWithError(response, 400, "Bad Request")
		return
	}

	err = service.DatabaseConnection.Delete(user).Error
	if err != nil {
		log.Fatal(err)
		respondWithError(response, 500, "Internal Server Error")
	} else {
		respondWithJSON(response, 200, user)
	}
}

// Setup database connection and all the routes
func (service *Service) Setup(databaseType string, DatabaseConnectionString string, debug bool) {
	var err error
	service.DatabaseConnection, err = gorm.Open(databaseType, DatabaseConnectionString)
	if err != nil {
		panic("Failed to connect database")
	}

	if debug == true {
		service.DatabaseConnection.Debug()
	}

	service.DatabaseConnection.AutoMigrate(&users.User{})

	service.Router = mux.NewRouter()

	/*
		service.Router.HandleFunc("/", service.apiInfoHandler)

		service.Router.HandleFunc("/token", service.generateTokenHandler).Methods("POST")
		service.Router.HandleFunc("/token/renew", service.renewTokenHandler).Methods("POST")
		service.Router.HandleFunc("/token", service.deleteTokenHandler).Methods("DELETE")
	*/
	service.Router.HandleFunc("/user", service.getUsersListHandler).Methods("GET")
	service.Router.HandleFunc("/user/{id}", service.getUserHandler).Methods("GET")
	service.Router.HandleFunc("/user/{id}", service.updateUserHandler).Methods("PUT")
	service.Router.HandleFunc("/user", service.createUserHandler).Methods("POST")
	service.Router.HandleFunc("/user/{id}", service.deleteUserHandler).Methods("DELETE")

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
	service.Setup("mysql", "root:hej256@/identity-provider?charset=utf8&parseTime=True&loc=Local", false)
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
