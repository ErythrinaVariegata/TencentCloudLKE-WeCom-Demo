package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"example.com/play/repo/wecom/cron"
	"example.com/play/repo/wecom/entity"
)

// SendTextMessage sends a text message using the WeChat Work API
func SendTextMessage(agentID int, content string, userID string) (*entity.MessageResponse, error) {
	msg := entity.TextMessage{
		ToUser:                 userID,
		MsgType:                "text",
		AgentID:                agentID,
		Text:                   entity.TextBody{Content: content},
		Safe:                   0,
		EnableIDTrans:          0,
		EnableDuplicateCheck:   0,
		DuplicateCheckInterval: 1800,
	}

	payloadBytes, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %v", err)
	}
	return doRequest(payloadBytes)
}

// SendMarkdownMessage sends a text message using the WeChat Work API
func SendMarkdownMessage(agentID int, content string, userID string) (*entity.MessageResponse, error) {
	msg := entity.MarkdownMessage{
		ToUser:                 userID,
		MsgType:                "markdown",
		AgentID:                agentID,
		Markdown:               entity.TextBody{Content: content},
		EnableDuplicateCheck:   0,
		DuplicateCheckInterval: 1800,
	}

	payloadBytes, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %v", err)
	}
	return doRequest(payloadBytes)
}

func doRequest(payloadBytes []byte) (*entity.MessageResponse, error) {
	accessToken := cron.GetAccessToken()
	url := fmt.Sprintf("%s?access_token=%s", entity.WxMessageSendURL, accessToken)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var result entity.MessageResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	if result.ErrCode != 0 {
		log.Printf("Message sending failed with error: %s", result.ErrMsg)
		return &result, fmt.Errorf("API error: %s", result.ErrMsg)
	}

	return &result, nil
}
