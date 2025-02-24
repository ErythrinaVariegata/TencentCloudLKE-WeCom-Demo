package entity

const (
	WxMessageSendURL = "https://qyapi.weixin.qq.com/cgi-bin/message/send"
)

// TextMessage 普通文本消息
type TextMessage struct {
	ToUser                 string   `json:"touser,omitempty"`
	ToParty                string   `json:"toparty,omitempty"`
	ToTag                  string   `json:"totag,omitempty"`
	MsgType                string   `json:"msgtype"`
	AgentID                int      `json:"agentid"`
	Text                   TextBody `json:"text"`
	Safe                   int      `json:"safe"`
	EnableIDTrans          int      `json:"enable_id_trans"`
	EnableDuplicateCheck   int      `json:"enable_duplicate_check"`
	DuplicateCheckInterval int      `json:"duplicate_check_interval"`
}

// MarkdownMessage Markdown消息
type MarkdownMessage struct {
	ToUser                 string   `json:"touser,omitempty"`
	ToParty                string   `json:"toparty,omitempty"`
	ToTag                  string   `json:"totag,omitempty"`
	MsgType                string   `json:"msgtype"`
	AgentID                int      `json:"agentid"`
	Markdown               TextBody `json:"markdown"`
	EnableDuplicateCheck   int      `json:"enable_duplicate_check"`
	DuplicateCheckInterval int      `json:"duplicate_check_interval"`
}

// TextBody 文本内容
type TextBody struct {
	Content string `json:"content"`
}

// MessageResponse 企业微信发送应用消息响应体
type MessageResponse struct {
	ErrCode        int    `json:"errcode"`
	ErrMsg         string `json:"errmsg"`
	InvalidUser    string `json:"invaliduser"`
	InvalidParty   string `json:"invalidparty"`
	InvalidTag     string `json:"invalidtag"`
	UnlicensedUser string `json:"unlicenseduser"`
	MsgID          string `json:"msgid"`
	ResponseCode   string `json:"response_code"`
}

// WxBizMsg 企业微信自建应用收到的用户消息结构体
type WxBizMsg struct {
	ToUserName   string  `xml:"ToUserName"`
	FromUserName string  `xml:"FromUserName"`
	CreateTime   int64   `xml:"CreateTime"`
	MsgType      MsgType `xml:"MsgType"`
	MsgId        int64   `xml:"MsgId"`
	AgentID      int64   `xml:"AgentID"`
	Content      string  `xml:"Content,omitempty"`      // 文本消息-内容
	PicUrl       string  `xml:"PicUrl,omitempty"`       // 图片消息-图片链接，链接消息-封面缩略图的url
	MediaId      string  `xml:"MediaId,omitempty"`      // 媒体文件id：图片消息-图片、语音消息-语音、视频消息-视频。可以调用获取媒体文件接口拉取，仅三天内有效
	ThumbMediaId string  `xml:"ThumbMediaId,omitempty"` // 视频消息-视频消息缩略图的媒体id，可以调用获取媒体文件接口拉取数据，仅三天内有效
	Format       string  `xml:"Format,omitempty"`       // 语音消息-语音格式，如amr，speex等
	LocationX    string  `xml:"Location_X,omitempty"`   // 位置消息-地理位置纬度
	LocationY    string  `xml:"Location_Y,omitempty"`   // 位置消息-地理位置经度
	Scale        string  `xml:"Scale,omitempty"`        // 位置消息-地图缩放大小
	Label        string  `xml:"Label,omitempty"`        // 位置消息-地理位置信息
	Title        string  `xml:"Title,omitempty"`        // 链接消息-标题
	Description  string  `xml:"Description,omitempty"`  // 链接消息-描述
	Url          string  `xml:"Url,omitempty"`          // 链接消息-链接跳转的url
	Event        string  `xml:"Event,omitempty"`        // 事件-事件类型
	EventKey     string  `xml:"EventKey,omitempty"`     //  事件-事件内容
}

// MsgType 消息类型
type MsgType string

const (
	MsgTypeText     MsgType = "text"     // 文本消息
	MsgTypeImage    MsgType = "image"    // 图片消息
	MsgTypeVoice    MsgType = "voice"    // 语音消息
	MsgTypeVideo    MsgType = "video"    // 视频消息
	MsgTypeLocation MsgType = "location" // 位置消息
	MsgTypeLink     MsgType = "link"     // 链接消息
	MsgTypeEvent    MsgType = "event"    // 事件类型
)

// WxBizURLParam 企业微信回调链接参数
type WxBizURLParam struct {
	MsgSignature string
	Timestamp    string
	Nonce        string
	EchoStr      string
}

// IsValid 企业微信回调链接是否有效
func (p *WxBizURLParam) IsValid() bool {
	// 验证参数是否存在
	if p.MsgSignature == "" || p.Timestamp == "" || p.Nonce == "" {
		return false
	}
	return true
}
