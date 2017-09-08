package token

import (
	"crypto/rsa"
	"time"

	"github.com/mojlighetsministeriet/identity-provider/entity"

	"github.com/SermoDigital/jose/crypto"
	"github.com/SermoDigital/jose/jws"
	"github.com/SermoDigital/jose/jwt"
)

// Generate a new JWT token from a user
func Generate(privateKey *rsa.PrivateKey, account entity.Account) (serializedToken []byte, err error) {
	claims := jws.Claims{}
	claims.SetExpiration(time.Now().Add(time.Duration(60*20) * time.Second))

	account.BeforeSave()

	claims.Set("id", account.ID)
	claims.Set("email", account.Email)
	claims.Set("roles", account.RolesSerialized)

	token := jws.NewJWT(claims, crypto.SigningMethodRS256)

	serializedToken, err = token.Serialize(privateKey)

	return
}

// ParseIfValid return a parsed JWT token if it is valid
func ParseIfValid(publicKey *rsa.PublicKey, token []byte) (parsedToken jwt.JWT, err error) {
	parsedToken, err = jws.ParseJWT(token)
	if err != nil {
		return
	}

	err = parsedToken.Validate(publicKey, crypto.SigningMethodRS256)
	if err != nil {
		claims := jws.Claims{}
		claims.SetExpiration(time.Now().Add(time.Duration(60*20) * time.Second))
		parsedToken = jws.NewJWT(claims, crypto.SigningMethodRS256)
	}

	return
}
