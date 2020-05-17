package model

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
