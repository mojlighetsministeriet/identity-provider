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
	os.Remove("testMain1.db")
	os.Setenv("DATABASE_CONNECTION", "testMain1.db")
	os.Setenv("PRIVATE_KEY", key)
	os.Setenv("EMAIL_ACCOUNT_RESET_BODY", "En återställning av lösenord har begärts, återställningslänken är giltig i en timme. Om det inte är du begärt återställningen så kan du ignorera denna e-post.<br><br><a href=\"{{.ServiceURL}}/reset-password?token={{.ResetToken}}\" target=\"_blank\">Följ länken</a> för att återställa ditt lösenord.<br><br>Allt gott!<br>Möjlighetsministeriet<br><img src=\"{{.ServiceURL}}/assets/images/mojlighetsministeriet-224.png\" alt=\"Möjlighetsministeriets logotyp\">")

	defer func() {
		os.Remove("testMain1.db")
	}()

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

func TestFailMainByListeningToInvalidPort(test *testing.T) {
	key, err := generatePrivateKey()
	if err != nil {
		panic(err)
	}

	os.Setenv("PORT", "1")
	os.Setenv("DATABASE_TYPE", "sqlite3")
	os.Remove("testMain2.db")
	os.Setenv("DATABASE_CONNECTION", "testMain2.db")
	os.Setenv("PRIVATE_KEY", key)

	go func() {
		defer func() {
			if recovery := recover(); recovery != nil {
				assert.Equal(test, true, true)
			}
			os.Remove("testMain2.db")
		}()
		main()
	}()

	time.Sleep(500 * time.Millisecond)

	client, err := httprequest.NewClient()
	assert.NoError(test, err)

	_, err = client.Get("http://localhost:1/")
	assert.Error(test, err)
}

func TestFailMainByInvalidDatabaseType(test *testing.T) {
	os.Setenv("PORT", "5433")
	os.Setenv("DATABASE_TYPE", "invalid")
	os.Remove("testMain3.db")
	os.Setenv("DATABASE_CONNECTION", "testMain3.db")

	go func() {
		defer func() {
			if recovery := recover(); recovery != nil {
				assert.Equal(test, true, true)
			}
			os.Remove("testMain3.db")
		}()
		main()
	}()

	time.Sleep(500 * time.Millisecond)

	client, err := httprequest.NewClient()
	assert.NoError(test, err)

	_, err = client.Get("http://localhost:5433/")
	assert.Error(test, err)
}
