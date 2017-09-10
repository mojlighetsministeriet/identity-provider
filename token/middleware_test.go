package token_test

import (
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/SermoDigital/jose/crypto"
	"github.com/SermoDigital/jose/jws"
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

	assert.Equal(test, http.StatusUnauthorized, recorder.Result().StatusCode)
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

	assert.Equal(test, http.StatusForbidden, recorder.Result().StatusCode)
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

	assert.Equal(test, http.StatusOK, recorder.Result().StatusCode)
}

func TestJWTRequiredRoleMiddlewareWithInvalidTokenFormat(test *testing.T) {
	server := echo.New()
	request := httptest.NewRequest(echo.GET, "/", nil)
	recorder := httptest.NewRecorder()
	context := server.NewContext(request, recorder)

	privateKey, err := rsa.GenerateKey(rand.Reader, 512)
	assert.NoError(test, err)

	claims := jws.Claims{}
	claims.SetExpiration(time.Now().Add(time.Duration(60*20) * time.Second))

	claims.Set("sub", uuid.NewV4())
	claims.Set("email", "email")

	serializedToken, err := jws.NewJWT(claims, crypto.SigningMethodRS256).Serialize(privateKey)

	assert.NoError(test, err)
	request.Header.Set(echo.HeaderAuthorization, "Bearer "+string(serializedToken))

	middleware := token.JWTRequiredRoleMiddleware(&privateKey.PublicKey, "administrator")
	handler := middleware(echo.HandlerFunc(func(context echo.Context) error {
		return context.NoContent(http.StatusOK)
	}))
	handler(context)
	assert.Equal(test, http.StatusForbidden, recorder.Result().StatusCode)
}
