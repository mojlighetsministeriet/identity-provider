package token_test

import (
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo"
	"github.com/mojlighetsministeriet/identity-provider/entity"
	"github.com/mojlighetsministeriet/identity-provider/token"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func TestGetTokenFromContext(test *testing.T) {
	request := httptest.NewRequest(echo.GET, "/", nil)
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	privateKey, err := rsa.GenerateKey(rand.Reader, 512)
	assert.NoError(test, err)

	account := entity.Account{
		ID:    uuid.NewV4(),
		Email: "tech+testing@mojlighetsministerietest.se",
		Roles: []string{"user"},
	}
	accessToken, err := token.Generate(privateKey, account)
	assert.NoError(test, err)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+string(accessToken))

	server := echo.New()
	recorder := httptest.NewRecorder()
	context := server.NewContext(request, recorder)

	extractedToken := token.GetTokenFromContext(context)
	assert.Equal(test, accessToken, extractedToken)
}

func TestJWTRequiredRoleMiddlewareWhenMissingAuthorizationToken(test *testing.T) {
	server := echo.New()
	request := httptest.NewRequest(echo.GET, "/", nil)
	recorder := httptest.NewRecorder()
	context := server.NewContext(request, recorder)

	privateKey, err := rsa.GenerateKey(rand.Reader, 512)
	assert.NoError(test, err)

	middleware := token.JWTRequiredRoleMiddleware(&privateKey.PublicKey, "administrator")
	handler := middleware(echo.HandlerFunc(func(context echo.Context) error {
		return context.NoContent(http.StatusOK)
	}))
	handler(context)

	assert.Equal(test, recorder.Result().StatusCode, http.StatusUnauthorized)
}

func TestJWTRequiredRoleMiddlewareWhenMissingRequiredAdministratorRole(test *testing.T) {
	server := echo.New()
	request := httptest.NewRequest(echo.GET, "/", nil)
	recorder := httptest.NewRecorder()
	context := server.NewContext(request, recorder)

	privateKey, err := rsa.GenerateKey(rand.Reader, 512)
	assert.NoError(test, err)

	account := entity.Account{
		ID:    uuid.NewV4(),
		Email: "tech+testing@mojlighetsministerietest.se",
		Roles: []string{"user"},
	}
	accessToken, err := token.Generate(privateKey, account)
	assert.NoError(test, err)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+string(accessToken))

	middleware := token.JWTRequiredRoleMiddleware(&privateKey.PublicKey, "administrator")
	handler := middleware(echo.HandlerFunc(func(context echo.Context) error {
		return context.NoContent(http.StatusOK)
	}))
	handler(context)

	assert.Equal(test, recorder.Result().StatusCode, http.StatusForbidden)
}

func TestJWTRequiredRoleMiddlewareWithRequiredAdministratorRole(test *testing.T) {
	server := echo.New()
	request := httptest.NewRequest(echo.GET, "/", nil)
	recorder := httptest.NewRecorder()
	context := server.NewContext(request, recorder)

	privateKey, err := rsa.GenerateKey(rand.Reader, 512)
	assert.NoError(test, err)

	account := entity.Account{
		ID:    uuid.NewV4(),
		Email: "tech+testing@mojlighetsministerietest.se",
		Roles: []string{"user", "administrator"},
	}
	accessToken, err := token.Generate(privateKey, account)
	assert.NoError(test, err)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+string(accessToken))

	middleware := token.JWTRequiredRoleMiddleware(&privateKey.PublicKey, "administrator")
	handler := middleware(echo.HandlerFunc(func(context echo.Context) error {
		return context.NoContent(http.StatusOK)
	}))
	handler(context)

	assert.Equal(test, recorder.Result().StatusCode, http.StatusOK)
}
