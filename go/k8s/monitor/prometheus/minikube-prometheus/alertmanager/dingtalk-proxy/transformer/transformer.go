package transformer

import (
	"bytes"
	"fmt"
	"k8s-lx1036/k8s/monitor/prometheus/minikube-prometheus/alertmanager/dingtalk-proxy/model"
)

func TransformToMarkdown(notification model.Notification) model.DingTalkMarkdown {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("#### 通知组 %s(当前状态: %s)\n", notification.GroupKey, notification.Status))
	for _, alert := range notification.Alerts {
		buffer.WriteString(fmt.Sprintf("> %s\n: %s\n", alert.Annotations["summary"], alert.Annotations["description"]))
		buffer.WriteString(fmt.Sprintf("###### 开始时间: %s, 结束时间: %s\n", alert.StartsAt.String(), alert.EndsAt.String()))
	}

	markdown := model.DingTalkMarkdown{
		MsgType: "markdown",
		Markdown: &model.Markdown{
			Title: fmt.Sprintf("[通知组]"),
			Text:  fmt.Sprintf("#### [prometheus]:\n %s\n", buffer.String()),
		},
		At: &model.At{
			IsAtAll: false,
		},
	}

	return markdown
}
