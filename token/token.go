package token

import (
	"time"

	"github.com/SermoDigital/jose/crypto"
	"github.com/SermoDigital/jose/jws"
	"github.com/mojlighetsministeriet/identity-provider/account"
)

// Generate a new JWT token from a user
func Generate(privateKey []byte, account account.Account) (token []byte, err error) {
	claims := jws.Claims{}
	claims.SetExpiration(time.Now().Add(time.Duration(60*20) * time.Second))

	claims.Set("id", account.ID)
	claims.Set("email", account.Email)
	claims.Set("roles", account.Roles)

	rsaPrivate, err := crypto.ParseRSAPrivateKeyFromPEM(privateKey)
	if err != nil {
		return
	}

	jwt := jws.NewJWT(claims, crypto.SigningMethodRS256)

	token, err = jwt.Serialize(rsaPrivate)

	return
}

// Validate a JWT token
func Validate(publicKey []byte, token []byte) (err error) {
	rsaPublic, err := crypto.ParseRSAPublicKeyFromPEM(publicKey)
	if err != nil {
		return
	}

	jwt, err := jws.ParseJWT(token)
	if err != nil {
		return
	}

	err = jwt.Validate(rsaPublic, crypto.SigningMethodRS256)

	return
}
