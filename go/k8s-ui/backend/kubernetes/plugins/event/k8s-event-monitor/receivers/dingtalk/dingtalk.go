package dingtalk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"k8s-lx1036/k8s-ui/backend/kubernetes/plugins/event/k8s-event-monitor/common"
	"k8s-lx1036/k8s-ui/backend/kubernetes/plugins/event/k8s-event-monitor/receivers"
	"k8s.io/api/core/v1"
	"net/http"
)

const (
	WARNING = 2
)


type DingTalk struct {
	Endpoint   string
	Namespaces []string
	Kinds      []string
	Token      string
	Level      int
	Labels     []string
	MsgType    string
	ClusterID  string
	Region     string
}

// https://ding-doc.dingtalk.com/doc#/serverapi2/qf2nxq/d535db33
type DingTalkMessage struct {
}

type DingTalkMarkdown struct {
	MsgType  string    `json:"msgtype"` // 此消息类型为固定markdown
	Markdown *Markdown `json:"markdown"`
	At       *At       `json:"at"`
}

type Markdown struct {
	Title string `json:"title"` // 首屏会话透出的展示内容
	Text  string `json:"text"`
}

type At struct {
	AtMobiles []string `json:"atMobiles"` // 被@人的手机号(在text内容里要有@手机号)
	IsAtAll   bool     `json:"isAtAll"`   // @所有人时：true，否则为：false
}

//
func NewDingTalkReceiver(receiver string) *DingTalk  {
	dingTalk := &DingTalk{
		Level: WARNING,
	}

	return dingTalk
}

func (receiver *DingTalk) ExportEvents(events *common.Events) {

}

func (receiver *DingTalk) send(event *v1.Event)  {
	message := transform(event)
	msgBytes, err := json.Marshal(message)
	if err != nil {

	}
	response, err := http.Post(fmt.Sprintf("https://%s?access_token=%s", receiver.Endpoint, receiver.Token), "application/json", bytes.NewBuffer(msgBytes))
	if err != nil {

	}

	if response != nil && response.StatusCode != http.StatusOK {

	}
}

func transform(event *v1.Event) DingTalkMarkdown  {

}
