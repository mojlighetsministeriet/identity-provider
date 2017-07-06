package token

import (
	"time"

	"github.com/SermoDigital/jose/crypto"
	"github.com/SermoDigital/jose/jws"
	"github.com/mojlighetsministeriet/identity-provider/users"
)

// Generate a new JWT token from a user
func Generate(privateKey []byte, user users.User) (token []byte, err error) {
	claims := jws.Claims{}
	claims.SetExpiration(time.Now().Add(time.Duration(60*20) * time.Second))

	claims.Set("id", user.ID)
	claims.Set("email", user.Email)
	claims.Set("roles", user.Roles)

	rsaPrivate, err := crypto.ParseRSAPrivateKeyFromPEM(privateKey)
	if err != nil {
		return
	}

	jwt := jws.NewJWT(claims, crypto.SigningMethodRS256)

	token, err = jwt.Serialize(rsaPrivate)

	return
}

// Validate a JWT token
func Validate(publicKey []byte, token []byte) error {
	rsaPublic, err := crypto.ParseRSAPublicKeyFromPEM(publicKey)
	if err != nil {
		return err
	}

	jwt, err := jws.ParseJWT(token)
	if err != nil {
		return err
	}

	err = jwt.Validate(rsaPublic, crypto.SigningMethodRS256)
	if err != nil {
		return err
	}

	return err
}
