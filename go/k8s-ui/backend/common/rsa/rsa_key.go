package rsa

import (
	"crypto/rsa"
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/spf13/viper"
	"io/ioutil"
	"path/filepath"
)

var (
	RsaPrivateKey *rsa.PrivateKey
	RsaPublicKey  *rsa.PublicKey
)

func InitRsaKey() {
	privateKeyPath := viper.GetString("default.RsaPrivateKey")
	publicKeyPath := viper.GetString("default.RsaPublicKey")
	if len(privateKeyPath) == 0 || len(publicKeyPath) == 0 {
		panic(errors.New("rsa private/public key can't be empty"))
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(readKey(privateKeyPath))
	if err != nil {
		panic(err)
	}
	RsaPrivateKey = privateKey

	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(readKey(publicKeyPath))
	if err != nil {
		panic(err)
	}
	RsaPublicKey = publicKey
}

func readKey(filename string) []byte {
	//filename := beego.AppConfig.String(key)
	//filename := fmt.Sprintf("./apikey/%s", key)
	// get the abs
	// which will try to find the 'filename' from current workind dir too.
	pem, err := filepath.Abs(filename)
	if err != nil {
		panic(err)
	}

	// read the raw contents of the file
	data, err := ioutil.ReadFile(pem)
	if err != nil {
		panic(err)
	}

	return data
}
