package entity

const (
	TencentLKESSEUrl       = "https://wss.lke.cloud.tencent.com/v1/qbot/chat/sse"
	TencentLKECapiEndpoint = "lke.tencentcloudapi.com"
	TencentLKECapiRegion   = "ap-guangzhou"
)

// SseSendEvent SSE发送事件
type SseSendEvent struct {
	ReqID             string `json:"req_id"`
	Content           string `json:"content"`
	BotAppKey         string `json:"bot_app_key"`
	VisitorBizID      string `json:"visitor_biz_id"`
	SessionID         string `json:"session_id"`
	StreamingThrottle int    `json:"streaming_throttle"`
	Timeout           int64  `json:"timeout"`
	SystemRole        string `json:"system_role"`
	IsEvaluateTest    bool   `json:"is_evaluate_test"` // 是否来自应用评测
}

// SseRecvEvent SSE回复事件
type SseRecvEvent struct {
	ReqID     string    `json:"req_id,omitempty"`
	Type      EventType `json:"type"`
	Payload   Payload   `json:"payload"`
	Error     Error     `json:"error,omitempty"`
	MessageID string    `json:"message_id"`
}

type EventType string

const (
	EventTypeReply     = "reply"      // 回复事件
	EventTypeError     = "error"      // 错误事件
	EventTypeTokenStat = "token_stat" // token统计事件
	EventTypeReference = "reference"  // 引用事件
	EventTypeThought   = "thought"    // 思考事件，推理模型独有
)

// Error 错误
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Payload 事件消息体
type Payload struct {
	// 公共参数
	RequestID       string           `json:"request_id"`
	SessionID       string           `json:"session_id"`
	TraceId         string           `json:"trace_id"`
	RecordID        string           `json:"record_id"`
	RelatedRecordID string           `json:"related_record_id"`
	ReplyMethod     ReplyMethod      `json:"reply_method,omitempty"`
	Content         string           `json:"content,omitempty"`
	Knowledge       []ReplyKnowledge `json:"knowledge,omitempty"`
	IntentCategory  string           `json:"intent_category,omitempty"`
	References      []Reference      `json:"references,omitempty"`
	Timestamp       int64            `json:"timestamp,omitempty"`
	FromName        string           `json:"from_name,omitempty"`
	FromAvatar      string           `json:"from_avatar,omitempty"`
	CanFeedback     bool             `json:"can_feedback,omitempty"`
	CanRating       bool             `json:"can_rating,omitempty"`
	IsFinal         bool             `json:"is_final,omitempty"`
	IsFromSelf      bool             `json:"is_from_self,omitempty"`
	IsEvil          bool             `json:"is_evil,omitempty"`
	IsLLMGenerated  bool             `json:"is_llm_generated,omitempty"`
	Elapsed         int              `json:"elapsed,omitempty"`
	IsWorkflow      bool             `json:"is_workflow,omitempty"`
	Procedures      []Procedure      `json:"procedures,omitempty"`
	WorkflowName    string           `json:"workflow_name,omitempty"`
}

// ReplyMethod 回复方式
type ReplyMethod uint8

// 回复方式
const (
	ReplyMethodModel          ReplyMethod = 1  // 大模型直接回复
	ReplyMethodBare           ReplyMethod = 2  // 保守回复, 未知问题回复
	ReplyMethodRejected       ReplyMethod = 3  // 拒答问题回复
	ReplyMethodEvil           ReplyMethod = 4  // 敏感回复
	ReplyMethodPriorityQA     ReplyMethod = 5  // 问答对直接回复, 已采纳问答对优先回复
	ReplyMethodGreeting       ReplyMethod = 6  // 欢迎语回复
	ReplyMethodBusy           ReplyMethod = 7  // 并发超限回复
	ReplyGlobalKnowledge      ReplyMethod = 8  // 全局干预知识
	ReplyMethodTaskFlow       ReplyMethod = 9  // 任务流程过程回复, 当历史记录中 task_flow.type = 0 时, 为大模型回复
	ReplyMethodTaskAnswer     ReplyMethod = 10 // 任务流程答案回复
	ReplyMethodSearch         ReplyMethod = 11 // 搜索引擎回复
	ReplyMethodDecorator      ReplyMethod = 12 // 知识润色后回复
	ReplyMethodImage          ReplyMethod = 13 // 图片理解回复
	ReplyMethodFile           ReplyMethod = 14 // 实时文档回复
	ReplyMethodClarifyConfirm ReplyMethod = 15 // 澄清确认回复
	ReplyMethodWorkflow       ReplyMethod = 16 // 工作流回复
)

// ReplyKnowledge 回复事件中的知识
type ReplyKnowledge struct {
	ID   string `json:"id"`
	Type uint32 `json:"type"`
}

// Reference 大模型引用信息
type Reference struct {
	DocBizID string `json:"doc_biz_id"`
	DocID    string `json:"doc_id"`
	DocName  string `json:"doc_name"`
	ID       string `json:"id"`
	Name     string `json:"name"`
	QABizID  string `json:"qa_biz_id"`
	Type     int    `json:"type"`
	URL      string `json:"url"`
}

// Procedure 过程步骤
type Procedure struct {
	Debugging    Debugging `json:"debugging"`
	Elapsed      int       `json:"elapsed"`
	Icon         string    `json:"icon"`
	Index        int       `json:"index"`
	Name         string    `json:"name"`
	PluginType   int       `json:"plugin_type"`
	Status       string    `json:"status"`
	Switch       string    `json:"switch"`
	Title        string    `json:"title"`
	WorkflowName string    `json:"workflow_name"`
}

// Debugging 过程步骤的主要内容
type Debugging struct {
	Content string `json:"content"`
}
