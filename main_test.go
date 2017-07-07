package main_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"testing"

	_ "github.com/jinzhu/gorm/dialects/sqlite"
	main "github.com/mojlighetsministeriet/identity-provider"
	"github.com/mojlighetsministeriet/identity-provider/users"
	uuid "github.com/satori/go.uuid"
)

var service main.Service

func TestMain(m *testing.M) {
	DatabaseConnectionString := "test-" + uuid.NewV4().String() + ".db"

	service = main.Service{}
	service.Setup("sqlite3", DatabaseConnectionString, true)
	defer os.Remove(DatabaseConnectionString)

	code := m.Run()
	os.Exit(code)
}

func TestEmptyUserTable(t *testing.T) {
	request, err := http.NewRequest("GET", "/user", nil)
	if err != nil {
		t.Error(err)
	}

	response := executeRequest(request)

	checkResponseCode(t, http.StatusOK, response.Code)

	if body := response.Body.String(); body != "[]" {
		t.Errorf("Expected an empty array. Got %s", body)
	}
}

func TestCreateUser(t *testing.T) {
	body := []byte("{ \"email\": \"testing@mojlighetsministeriet.se\", \"password\": \"testing\", \"roles\": \"user\" }")
	request, err := http.NewRequest("POST", "/user", bytes.NewBuffer(body))
	if err != nil {
		t.Error(err)
	}

	request.Header.Set("Content-Type", "application/json")
	response := executeRequest(request)
	checkResponseCode(t, http.StatusCreated, response.Code)

	responseBody := response.Body.String()
	expression := regexp.MustCompile("[0-9a-f]{8}\\-[0-9a-f]{4}\\-[0-9a-f]{4}\\-[0-9a-f]{4}\\-[0-9a-f]{12}")
	maskedBody := expression.ReplaceAllString(responseBody, "*uuid*")
	if maskedBody != "{\"id\":\"*uuid*\"}" {
		t.Errorf("Expected matching {\"id\":\"*uuid*\"} but got %s", responseBody)
	}
}

func TestExpectNotFoundWithInvalidUserID(t *testing.T) {
	request, err := http.NewRequest("GET", "/user/1234", nil)
	if err != nil {
		t.Error(err)
	}

	response := executeRequest(request)
	checkResponseCode(t, http.StatusNotFound, response.Code)

	request, err = http.NewRequest("PUT", "/user/1234", nil)
	if err != nil {
		t.Error(err)
	}

	response = executeRequest(request)
	checkResponseCode(t, http.StatusNotFound, response.Code)

	request, err = http.NewRequest("DELETE", "/user/1234", nil)
	if err != nil {
		t.Error(err)
	}

	response = executeRequest(request)
	checkResponseCode(t, http.StatusNotFound, response.Code)
}

func TestListGetUpdateDeleteUser(t *testing.T) {
	request, err := http.NewRequest("GET", "/user", nil)
	if err != nil {
		t.Error(err)
	}

	response := executeRequest(request)
	checkResponseCode(t, http.StatusOK, response.Code)

	type ListedUser struct {
		ID uuid.UUID `json:"id"`
	}

	var userList []ListedUser
	err = json.Unmarshal(response.Body.Bytes(), &userList)
	if err != nil {
		t.Error(err)
	}

	if len(userList) != 1 {
		t.Error("Expected to get 1 user, got ", len(userList))
	}

	expression := regexp.MustCompile("[0-9a-f]{8}\\-[0-9a-f]{4}\\-[0-9a-f]{4}\\-[0-9a-f]{4}\\-[0-9a-f]{12}")
	if len(userList) == 1 {
		userID := userList[0].ID.String()

		if expression.MatchString(userID) == false {
			t.Error("Expected id parameter to match uuid, got %s", userID)
		}

		request, err = http.NewRequest("GET", "/user/"+userID, nil)
		if err != nil {
			t.Error(err)
		}

		response = executeRequest(request)
		checkResponseCode(t, http.StatusOK, response.Code)

		user := users.User{}
		err = json.Unmarshal(response.Body.Bytes(), &user)
		if err != nil {
			t.Error(err)
		}

		if user.Email != "testing@mojlighetsministeriet.se" {
			t.Error("Expected user email to be testing@mojlighetsministeriet.se, got %s", user.Email)
		}

		if len(user.Roles) != 1 {
			t.Error("Expected user roles count to be 1, got", len(user.Roles))
		}

		if user.Roles[0] != "user" {
			t.Error("Expected user roles to be user, got %s", user.Roles[0])
		}

		if user.Password != "" {
			t.Error("Expected password to be empty, got %s", user.Password)
		}

		user.Email = "updated@mojlighetsministeriet.se"
		body, err := json.Marshal(user)
		if err != nil {
			t.Error(err)
		}

		request, err := http.NewRequest("PUT", "/user/"+userID, bytes.NewBuffer(body))
		if err != nil {
			t.Error(err)
		}

		request.Header.Set("Content-Type", "application/json")
		response := executeRequest(request)
		checkResponseCode(t, http.StatusOK, response.Code)

		responseBody := response.Body.String()
		if responseBody != "" {
			t.Errorf("Expected empty response body but got %s", responseBody)
		}

		request, err = http.NewRequest("GET", "/user/"+userID, nil)
		if err != nil {
			t.Error(err)
		}

		response = executeRequest(request)
		checkResponseCode(t, http.StatusOK, response.Code)

		user = users.User{}
		err = json.Unmarshal(response.Body.Bytes(), &user)
		if err != nil {
			t.Error(err)
		}

		if user.Email != "updated@mojlighetsministeriet.se" {
			t.Error("Expected user email to be updated@mojlighetsministeriet.se, got %s", user.Email)
		}

		if len(user.Roles) != 1 {
			t.Error("Expected user roles count to be 1, got", len(user.Roles))
		}

		if user.Roles[0] != "user" {
			t.Error("Expected user roles to be user, got %s", user.Roles[0])
		}

		if user.Password != "" {
			t.Error("Expected password to be empty, got %s", user.Password)
		}

		request, err = http.NewRequest("DELETE", "/user/"+userID, nil)
		if err != nil {
			t.Error(err)
		}

		response = executeRequest(request)
		checkResponseCode(t, http.StatusOK, response.Code)

		request, err = http.NewRequest("GET", "/user/"+userID, nil)
		if err != nil {
			t.Error(err)
		}

		response = executeRequest(request)
		checkResponseCode(t, http.StatusNotFound, response.Code)
	}
}

func executeRequest(request *http.Request) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	service.Router.ServeHTTP(recorder, request)
	return recorder
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}
