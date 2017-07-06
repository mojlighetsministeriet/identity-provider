package main_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"."
	uuid "github.com/satori/go.uuid"
)

var service main.Service

func TestMain(m *testing.M) {
	databaseConnectionString := "test-" + uuid.NewV4().String() + ".db"

	service = main.Service{}
	service.Setup("sqlite3", databaseConnectionString)

	code := m.Run()
	os.Exit(code)
}

func TestEmptyTable(t *testing.T) {
	request, _ := http.NewRequest("GET", "/user", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	if body := response.Body.String(); body != "[]" {
		t.Errorf("Expected an empty array. Got %s", body)
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
