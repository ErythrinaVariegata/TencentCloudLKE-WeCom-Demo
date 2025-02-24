package logic

import (
	"encoding/xml"
	"io"
	"log"
	"net/http"
	"net/url"

	"example.com/play/config"
	lkeClient "example.com/play/repo/tencentlke/client"
	lkeEntity "example.com/play/repo/tencentlke/entity"
	wecomClient "example.com/play/repo/wecom/client"
	wecomEntity "example.com/play/repo/wecom/entity"
	wecomCrypt "example.com/play/repo/wecom/wxbizmsgcrypt"
	"example.com/play/utils"
)

func CallbackHandler(w http.ResponseWriter, r *http.Request) {
	// 解析并解码URL参数
	query, err := url.QueryUnescape(r.URL.RawQuery)
	if err != nil {
		http.Error(w, "DecodeURL failed", http.StatusBadRequest)
		log.Printf("DecodeURL failed, rawQuery: %v, err: %v", r.URL.RawQuery, err)
		return
	} else if len(query) == 0 {
		http.Error(w, "Not found", http.StatusNotFound)
		log.Printf("Get non-query request from %s, url: %s", r.RemoteAddr, r.URL.String())
		return
	}
	// 获取并解码所有参数
	values, err := url.ParseQuery(query)
	if err != nil {
		http.Error(w, "ParseQuery failed", http.StatusBadRequest)
		log.Printf("ParseQuery failed, query: %v, err: %v", query, err)
		return
	}
	// 获取必要参数（已自动URL解码）
	urlParams := wecomEntity.WxBizURLParam{
		MsgSignature: values.Get("msg_signature"),
		Timestamp:    values.Get("timestamp"),
		Nonce:        values.Get("nonce"),
		EchoStr:      values.Get("echostr"),
	}
	// 验证参数是否存在
	if !urlParams.IsValid() {
		http.Error(w, "Missing required parameters", http.StatusBadRequest)
		log.Printf("Validation failed: Missing required parameters, urlParams: %v", urlParams)
		return
	}
	// 存在EchoStr且为Get请求，即验证URL回调接口
	if urlParams.EchoStr != "" {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", "GET")
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			log.Printf("Invalid request method: %s from %s", r.Method, r.RemoteAddr)
			return
		}
		log.Println("\n---------------------------")
		log.Printf("Received GET request from %s", r.RemoteAddr)
		VerifyURLHandler(w, &urlParams)
		return
	}
	// 不存在EchoStr且不为Post请求，报错返回
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "Post")
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		log.Printf("Invalid request method: %s from %s\n", r.Method, r.RemoteAddr)
		return
	}
	// 不存在EchoStr且为Post请求，即接受消息接口
	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		log.Printf("Error reading body: %v", err)
		return
	}
	log.Println("---------------------------")
	log.Printf("Received POST request from %s", r.RemoteAddr)
	// for k, v := range r.Header {
	// 	log.Printf("%s: %s", k, v)
	// }
	// log.Printf("Body (%d bytes):\n%s\n", len(body), body)
	ReceiveMessageHandler(w, &urlParams, body)
}

func VerifyURLHandler(w http.ResponseWriter, p *wecomEntity.WxBizURLParam) {
	// 验证URL
	wxcpt := wecomCrypt.NewWXBizMsgCrypt(config.Config.WxToken, config.Config.WxEncodingAESKey, config.Config.WxCorpID, wecomCrypt.XmlType)
	echoStr, cryptErr := wxcpt.VerifyURL(p.MsgSignature, p.Timestamp, p.Nonce, p.EchoStr)
	if cryptErr != nil {
		http.Error(w, "VerifyURL process failed", http.StatusUnauthorized)
		log.Println("VerifyURL process failed, err:", cryptErr)
	}
	log.Println("VerifyURL process success, echoStr:", string(echoStr))
	// 返回解密后的EchoStr
	w.Write([]byte(echoStr))
}

func ReceiveMessageHandler(w http.ResponseWriter, p *wecomEntity.WxBizURLParam, msgBodyStr []byte) {
	// 解密用户消息
	wxcpt := wecomCrypt.NewWXBizMsgCrypt(config.Config.WxToken, config.Config.WxEncodingAESKey, config.Config.WxCorpID, wecomCrypt.XmlType)
	msgStr, cryptErr := wxcpt.DecryptMsg(p.MsgSignature, p.Timestamp, p.Nonce, msgBodyStr)
	if cryptErr != nil {
		http.Error(w, "DecryptMsg process failed", http.StatusUnauthorized)
		log.Println("DecryptMsg process failed", cryptErr)
	}
	log.Printf("DecryptMsg process success, msg: %s", string(msgStr))
	var msg wecomEntity.WxBizMsg
	err := xml.Unmarshal(msgStr, &msg)
	if err != nil {
		http.Error(w, "ParseMsg process failed", http.StatusInternalServerError)
		log.Println("ParseMsg process failed, err:", err)
	}
	log.Printf("ParseMsg process success, msg: %+v", msg)
	// 目前仅支持文本消息对接大模型知识引擎
	if msg.MsgType == wecomEntity.MsgTypeText {
		// 将用户的消息传入腾讯云大模型知识引擎
		go CallTencentLKEApp(&msg)
		w.Write(nil)
		return
	}
	// 其他消息类型返回提示
	invoice := "抱歉，目前仅支持文本输入，请尝试用文字与我交流 :-/"
	wecomResp, wecomErr := wecomClient.SendTextMessage(int(msg.AgentID), invoice, msg.FromUserName)
	if wecomErr != nil {
		log.Printf("SendBackMessage failed, msgID: %d, err: %v", msg.MsgId, wecomErr)
		return
	}
	log.Printf("SendBackMessage success, msgId: %d, resp: %v", msg.MsgId, *wecomResp)
}

func CallTencentLKEApp(wecomMsg *wecomEntity.WxBizMsg) {
	sessionID := utils.GetSessionID()
	event := &lkeEntity.SseSendEvent{
		Content:           wecomMsg.Content,
		BotAppKey:         config.Config.TencentCloudLKEAppKey,
		VisitorBizID:      wecomMsg.FromUserName,
		SessionID:         sessionID,
		StreamingThrottle: 1,
	}

	replyChan, errChan := lkeClient.SendEvent(event)
	for {
		select {
		case reply := <-replyChan:
			log.Printf("Call TencentLKEApp, msgID: %d, markdown reply:\n%s", wecomMsg.MsgId, reply)
			if len(reply) == 0 {
				continue
			}
			wecomResp, wecomErr := wecomClient.SendMarkdownMessage(int(wecomMsg.AgentID), reply, wecomMsg.FromUserName)
			if wecomErr != nil {
				log.Printf("SendBackMessage failed, msgID: %d, err: %v", wecomMsg.MsgId, wecomErr)
				continue
			}
			log.Printf("SendBackMessage success, msgId: %d, resp: %v", wecomMsg.MsgId, *wecomResp)
		case err := <-errChan:
			if err != nil {
				invoice := "抱歉，调用大模型知识引擎出现了一点问题，请稍后再试 :-<"
				wecomResp, wecomErr := wecomClient.SendTextMessage(int(wecomMsg.AgentID), invoice, wecomMsg.FromUserName)
				if wecomErr != nil {
					log.Printf("SendBackMessage failed, msgID: %d, err: %v", wecomMsg.MsgId, wecomErr)
					return
				}
				log.Printf("SendBackMessage success, msgId: %d, resp: %v", wecomMsg.MsgId, *wecomResp)
			}
			return
		}
	}
}
