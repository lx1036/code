package initial

import (
	"github.com/astaxie/beego"
	_ "github.com/go-sql-driver/mysql"
	"k8s-lx1036/k8s-ui/backend/bus"
	"k8s-lx1036/k8s-ui/backend/client"
	"k8s-lx1036/k8s-ui/backend/util"
	"k8s.io/apimachinery/pkg/util/wait"
	"time"
)

func InitClient() {
	// 定期更新client
	go wait.Forever(client.BuildApiServerClient, 30*time.Second)
}

// bus
func InitBus() {
	var err error
	bus.DefaultBus, err = bus.NewBus(beego.AppConfig.String("BusRabbitMQURL"))
	if err != nil {
		panic(err)
	}
}

func InitKubeLabel() {
	util.AppLabelKey = beego.AppConfig.DefaultString("AppLabelKey", "k8s-ui-app")
	util.NamespaceLabelKey = beego.AppConfig.DefaultString("NamespaceLabelKey", "k8s-ui-ns")
	util.PodAnnotationControllerKindLabelKey = beego.AppConfig.DefaultString("PodAnnotationControllerKindLabelKey", "k8s-ui.cloud/controller-kind")
}
