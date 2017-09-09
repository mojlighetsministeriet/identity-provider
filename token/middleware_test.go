package token_test

import (
	"crypto/rand"
	"crypto/rsa"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
)

func TestGetTokenFromContext(test *testing.T) {
	request := httptest.NewRequest(echo.GET, "/", nil)
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	assert.NoError(test, err)

	request.Header.Set(echo.HeaderAuthorization)
	recorder := httptest.NewRecorder()
	context := server.NewContext(request, recorder)

	GetTokenFromContext()
}
