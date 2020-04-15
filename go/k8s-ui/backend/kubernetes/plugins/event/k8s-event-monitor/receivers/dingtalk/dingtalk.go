package dingtalk

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


//
func NewDingTalkReceiver(receiver string) *DingTalk  {
	dingTalk := &DingTalk{
		Level: WARNING,
	}

}
