package token_test

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"github.com/mojlighetsministeriet/identity-provider/account"
	"github.com/mojlighetsministeriet/identity-provider/token"
	uuid "github.com/satori/go.uuid"
)

func TestGenerateAndValidateToken(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Error(err)
	}

	account := account.Account{
		ID:    uuid.NewV4().String(),
		Email: "tech+testing@mojlighetsministeriet.se",
		Roles: []string{"user"},
	}

	accessToken, err := token.Generate(privateKey, account)
	if err != nil {
		t.Error("Failed to generate token", err)
	}

	if token.Validate(&privateKey.PublicKey, accessToken) != nil {
		t.Error("Unable to validate newly created token")
	}
}

func TestFailValidateWithBadToken(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Error(err)
	}

	if token.Validate(&privateKey.PublicKey, []byte{1, 2, 3, 4}) == nil {
		t.Error("Should have failed validate with invalid public key")
	}
}
