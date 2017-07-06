package main

import (
	"fmt"
	"io/ioutil"

	"github.com/mojlighetsministeriet/identity-provider/jwt"
)

func main() {
	buffer, err := ioutil.ReadFile("sample_key.priv")
	if err != nil {
		fmt.Print(err)
	}
	privateKey := string(buffer)

	jwt.Generate()
}
