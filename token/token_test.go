package token_test

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"github.com/mojlighetsministeriet/identity-provider/entity"
	"github.com/mojlighetsministeriet/identity-provider/token"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func TestGenerateAndParseIfValid(test *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	assert.NoError(test, err)

	account := entity.Account{
		ID:    uuid.NewV4().String(),
		Email: "tech+testing@mojlighetsministerietest.se",
		Roles: []string{"user"},
	}

	accessToken, err := token.Generate(privateKey, account)
	assert.NoError(test, err)

	parsedToken, err := token.ParseIfValid(&privateKey.PublicKey, accessToken)
	assert.NoError(test, err)
	assert.Equal(test, account.ID, parsedToken.Claims().Get("sub").(string))
	assert.Equal(test, "tech+testing@mojlighetsministerietest.se", parsedToken.Claims().Get("email"))
	assert.Equal(test, "user", parsedToken.Claims().Get("roles"))
}

func TestFailParseIfValidWithBadToken(test *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	assert.NoError(test, err)

	parsedToken, err := token.ParseIfValid(&privateKey.PublicKey, []byte{1, 2, 3, 4})
	assert.Error(test, err)
	assert.Equal(test, nil, parsedToken)
}

func TestFailParseIfValidWithBadPublicKey(test *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	assert.NoError(test, err)

	account := entity.Account{
		ID:    uuid.NewV4().String(),
		Email: "tech+testing@mojlighetsministerietest.se",
		Roles: []string{"user"},
	}

	accessToken, err := token.Generate(privateKey, account)
	assert.NoError(test, err)

	wrongPrivateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	assert.NoError(test, err)

	parsedToken, err := token.ParseIfValid(&wrongPrivateKey.PublicKey, accessToken)
	assert.Error(test, err)
	assert.Equal(test, false, parsedToken.Claims().Has("email"))
}
