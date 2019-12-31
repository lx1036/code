package initial

import (
	"github.com/astaxie/beego"
	"github.com/dgrijalva/jwt-go"
	"io/ioutil"
	"k8s-lx1036/k8s-ui/backend/apikey"
	"path/filepath"
)


func InitRsaKey() {
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(readKey("RsaPrivateKey"))
	if err != nil {
		panic(err)
	}
	apikey.RsaPrivateKey = privateKey

	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(readKey("RsaPublicKey"))
	if err != nil {
		panic(err)
	}
	apikey.RsaPublicKey = publicKey
}

func readKey(key string) []byte {
	filename := beego.AppConfig.String(key)
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
