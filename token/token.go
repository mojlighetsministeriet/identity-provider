package token

import (
	"crypto/rsa"
	"time"

	"github.com/SermoDigital/jose/crypto"
	"github.com/SermoDigital/jose/jws"
	"github.com/mojlighetsministeriet/identity-provider/account"
)

// Generate a new JWT token from a user
func Generate(privateKey *rsa.PrivateKey, account account.Account) (token []byte, err error) {
	claims := jws.Claims{}
	claims.SetExpiration(time.Now().Add(time.Duration(60*20) * time.Second))

	claims.Set("sub", account.ID)
	claims.Set("email", account.Email)
	claims.Set("roles", account.Roles)

	jwt := jws.NewJWT(claims, crypto.SigningMethodRS256)

	token, err = jwt.Serialize(privateKey)

	return
}

// Validate a JWT token
func Validate(publicKey *rsa.PublicKey, token []byte) (err error) {
	jwt, err := jws.ParseJWT(token)
	if err != nil {
		return
	}

	err = jwt.Validate(publicKey, crypto.SigningMethodRS256)

	return
}
