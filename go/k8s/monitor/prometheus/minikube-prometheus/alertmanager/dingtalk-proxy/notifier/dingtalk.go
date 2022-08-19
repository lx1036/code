package notifier

import (
	"encoding/json"
	"io/ioutil"
	"k8s-lx1036/k8s/monitor/prometheus/minikube-prometheus/alertmanager/dingtalk-proxy/model"
	"k8s-lx1036/k8s/monitor/prometheus/minikube-prometheus/alertmanager/dingtalk-proxy/transformer"
	"net/http"
	"strings"
)

type Notifier interface {
	Send() (Response, error)
}

type Builder struct {
	Notifier Notifier
}

type DingTalk struct {
	Url          string
	Notification model.Notification
}

type Response struct {
	Errcode int    `json:"errcode"`
	Errmsg  string `json:"errmsg"`
}

func (dingTalk *DingTalk) Send() (dingTalkResponse Response, err error) {
	markdown := transformer.TransformToMarkdown(dingTalk.Notification)
	data, err := json.Marshal(&markdown)
	if err != nil {
		return dingTalkResponse, err
	}

	req, _ := http.NewRequest("POST", dingTalk.Url, strings.NewReader(string(data)))
	req.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return dingTalkResponse, err
	}
	defer response.Body.Close()

	responseBody, _ := ioutil.ReadAll(response.Body)
	err = json.Unmarshal(responseBody, &dingTalkResponse)
	if err != nil {
		return dingTalkResponse, err
	}

	return dingTalkResponse, nil
}

func NewNotifier(notifier Notifier) *Builder {
	return &Builder{Notifier: notifier}
}
