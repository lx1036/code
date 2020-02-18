package initial

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
	"k8s-lx1036/k8s-ui/backend/client"
	"k8s-lx1036/k8s-ui/backend/util"
	"k8s.io/apimachinery/pkg/util/wait"
	"time"
)

func InitClient() {
	// 定期更新client
	go wait.Forever(client.BuildApiServerClient, 30*time.Second)
}

func InitKubeLabel() {
	util.AppLabelKey = viper.GetString("default.AppLabelKey")
	util.NamespaceLabelKey = viper.GetString("default.AppLabelKey")
	util.PodAnnotationControllerKindLabelKey = viper.GetString("default.PodAnnotationControllerKindLabelKey")
}
