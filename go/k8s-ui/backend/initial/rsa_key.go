package initial

import (
	"github.com/dgrijalva/jwt-go"
	"io/ioutil"
	"k8s-lx1036/k8s-ui/backend/apikey"
	"path/filepath"
)

func InitRsaKey(privateKeyPath string, publicKeyPath string) {
	//privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(readKey("RsaPrivateKey"))
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(readKey(privateKeyPath))
	if err != nil {
		panic(err)
	}
	apikey.RsaPrivateKey = privateKey

	//publicKey, err := jwt.ParseRSAPublicKeyFromPEM(readKey("RsaPublicKey"))
	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(readKey(publicKeyPath))
	if err != nil {
		panic(err)
	}
	apikey.RsaPublicKey = publicKey
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
