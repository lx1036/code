package initial

import (
	"github.com/astaxie/beego"
	"github.com/dgrijalva/jwt-go"
	_ "github.com/go-sql-driver/mysql"
	"io/ioutil"
	"k8s-lx1036/wayne/backend/apikey"
	"k8s-lx1036/wayne/backend/bus"
	"k8s-lx1036/wayne/backend/client"
	"k8s-lx1036/wayne/backend/util"
	"k8s.io/apimachinery/pkg/util/wait"
	"path/filepath"
	"time"
)

func InitClient() {
	// 定期更新client
	go wait.Forever(client.BuildApiserverClient, 5*time.Second)
}

// bus
func InitBus() {
	var err error
	bus.DefaultBus, err = bus.NewBus(beego.AppConfig.String("BusRabbitMQURL"))
	if err != nil {
		panic(err)
	}
}

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

func InitKubeLabel() {
	util.AppLabelKey = beego.AppConfig.DefaultString("AppLabelKey", "wayne-app")
	util.NamespaceLabelKey = beego.AppConfig.DefaultString("NamespaceLabelKey", "wayne-ns")
	util.PodAnnotationControllerKindLabelKey = beego.AppConfig.DefaultString("PodAnnotationControllerKindLabelKey", "wayne.cloud/controller-kind")
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
