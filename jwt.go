package jwt

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/SermoDigital/jose/crypto"
	"github.com/SermoDigital/jose/jws"
)

// TODO: allow to set custom expiration

// Generate a new JWT token
func Generate(privateKey string) string {
	bytes, _ := ioutil.ReadFile(privateKey)

	claims := jws.Claims{}
	claims.SetExpiration(time.Now().Add(time.Duration(10) * time.Second))

	rsaPrivate, _ := crypto.ParseRSAPrivateKeyFromPEM(bytes)
	jwt := jws.NewJWT(claims, crypto.SigningMethodRS256)

	b, _ := jwt.Serialize(rsaPrivate)
	fmt.Printf("%s", b)

	return b
}
