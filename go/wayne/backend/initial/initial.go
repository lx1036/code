package initial

import (
	"github.com/astaxie/beego"
	_ "github.com/go-sql-driver/mysql"
	"k8s-lx1036/wayne/backend/bus"
	"k8s-lx1036/wayne/backend/client"
	"k8s-lx1036/wayne/backend/util"
	"k8s.io/apimachinery/pkg/util/wait"
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



func InitKubeLabel() {
	util.AppLabelKey = beego.AppConfig.DefaultString("AppLabelKey", "wayne-app")
	util.NamespaceLabelKey = beego.AppConfig.DefaultString("NamespaceLabelKey", "wayne-ns")
	util.PodAnnotationControllerKindLabelKey = beego.AppConfig.DefaultString("PodAnnotationControllerKindLabelKey", "wayne.cloud/controller-kind")
}


