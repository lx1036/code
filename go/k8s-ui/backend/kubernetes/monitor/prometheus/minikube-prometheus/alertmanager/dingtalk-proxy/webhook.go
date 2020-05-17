package main

import (
	"flag"
	"github.com/gin-gonic/gin"
	"k8s-lx1036/k8s-ui/backend/kubernetes/prometheus/minikube-prometheus/alertmanager/dingtalk-proxy/model"
	"k8s-lx1036/k8s-ui/backend/kubernetes/prometheus/minikube-prometheus/alertmanager/dingtalk-proxy/notifier"
	"net/http"
)

var (
	h   bool
	url string
)

func init() {
	flag.BoolVar(&h, "h", false, "help")
	flag.StringVar(&url, "url", "", "global dingtalk robot webhook, you can overwrite by alert rule with annotations url")
}

func main() {
	flag.Parse()

	if h {
		flag.Usage()
		return
	}

	router := gin.Default()
	router.POST("/webhook", func(context *gin.Context) {
		var notification model.Notification
		err := context.BindJSON(&notification)
		if err != nil {
			context.JSON(http.StatusBadRequest, gin.H{
				"code":    -1,
				"message": err.Error(),
			})
			return
		}

		dingTalk := &notifier.DingTalk{
			Url:          url,
			Notification: notification,
		}
		resp, err := notifier.NewNotifier(dingTalk).Notifier.Send()
		if err != nil {
			context.JSON(http.StatusBadRequest, gin.H{
				"code":    resp.Errcode,
				"message": err.Error() + " " + resp.Errmsg,
			})
			return
		}

		context.JSON(http.StatusOK, gin.H{
			"code":    resp.Errcode,
			"message": resp.Errmsg,
		})
	})

	_ = router.Run()
}
