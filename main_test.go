package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/mojlighetsministeriet/utils/httprequest"
	"github.com/stretchr/testify/assert"
)

func generatePrivateKey() (result string, err error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 128)
	if err != nil {
		return
	}

	block := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}

	result = string(pem.EncodeToMemory(block))

	return
}

func TestMain(test *testing.T) {
	key, err := generatePrivateKey()
	if err != nil {
		panic(err)
	}

	os.Setenv("PORT", "3526")
	os.Setenv("DATABASE_TYPE", "sqlite3")
	os.Remove("testMain.db")
	os.Setenv("DATABASE_CONNECTION", "testMain.db")
	os.Setenv("PRIVATE_KEY", key)
	os.Setenv("EMAIL_ACCOUNT_RESET_BODY", "En återställning av lösenord har begärts, återställningslänken är giltig i en timme. Om det inte är du begärt återställningen så kan du ignorera denna e-post.<br><br><a href=\"{{.ServiceURL}}/reset-password?token={{.ResetToken}}\" target=\"_blank\">Följ länken</a> för att återställa ditt lösenord.<br><br>Allt gott!<br>Möjlighetsministeriet<br><img src=\"{{.ServiceURL}}/assets/images/mojlighetsministeriet-224.png\" alt=\"Möjlighetsministeriets logotyp\">")

	go func() {
		main()
	}()

	time.Sleep(500 * time.Millisecond)

	client, err := httprequest.NewClient()
	assert.NoError(test, err)

	index, err := client.Get("http://localhost:3526/")
	assert.NoError(test, err)
	fmt.Println(string(index))
	assert.Equal(test, true, len(string(index)) > 100)
}
