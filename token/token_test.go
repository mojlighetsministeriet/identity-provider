package token_test

import (
	"io/ioutil"
	"testing"

	"github.com/mojlighetsministeriet/identity-provider/token"
	"github.com/mojlighetsministeriet/identity-provider/users"
	uuid "github.com/satori/go.uuid"
)

func TestGenerateAndValidateToken(t *testing.T) {
	privateKey, err := ioutil.ReadFile("../fixtures/key.private")
	if err != nil {
		t.Error("Failed to load private key fixtures/key.private")
	}

	publicKey, err := ioutil.ReadFile("../fixtures/key.public")
	if err != nil {
		t.Error("Failed to load public key fixtures/key.public")
	}

	user := users.User{
		ID:    uuid.NewV4(),
		Email: "tech+testing@mojlighetsministeriet.se",
		Roles: []string{"user"},
	}

	accessToken, err := token.Generate(privateKey, user)
	if err != nil {
		t.Error("Failed to generate token", err)
	}

	if token.Validate(publicKey, accessToken) != nil {
		t.Error("Unable to validate newly created token")
	}
}

func TestFailGenerateWithBadPrivateKey(t *testing.T) {
	user := users.User{
		ID:    uuid.NewV4(),
		Email: "tech+testing@mojlighetsministeriet.se",
		Roles: []string{"user"},
	}

	_, err := token.Generate([]byte{1, 2, 3, 4}, user)
	if err == nil {
		t.Error("Should have failed to generate token")
	}
}

func TestFailValidateWithBadPublicKey(t *testing.T) {
	privateKey, err := ioutil.ReadFile("../fixtures/key.private")
	if err != nil {
		t.Error("Failed to load private key fixtures/key.private")
	}

	user := users.User{
		ID:    uuid.NewV4(),
		Email: "tech+testing@mojlighetsministeriet.se",
		Roles: []string{"user"},
	}

	accessToken, err := token.Generate(privateKey, user)
	if err != nil {
		t.Error("Failed to generate token", err)
	}

	if token.Validate([]byte{1, 2, 3, 4}, accessToken) == nil {
		t.Error("Should have failed validate with invalid public key")
	}
}

func TestFailValidateWithBadToken(t *testing.T) {
	publicKey, err := ioutil.ReadFile("../fixtures/key.public")
	if err != nil {
		t.Error("Failed to load public key fixtures/key.public")
	}

	if token.Validate(publicKey, []byte{1, 2, 3, 4}) == nil {
		t.Error("Should have failed validate with invalid public key")
	}
}
