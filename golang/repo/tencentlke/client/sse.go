package client

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"example.com/play/repo/tencentlke/entity"
)

func SendEvent(event *entity.SseSendEvent) (<-chan string, <-chan error) {
	replyChan := make(chan string, 10) // 使用带缓冲的channel
	errChan := make(chan error, 1)     // 错误channel只需要1个缓冲

	go func() {
		defer close(replyChan)
		defer close(errChan)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()
		client := &http.Client{
			Timeout: 10 * time.Minute,
		}

		payloadBytes, err := json.Marshal(&event)
		if err != nil {
			log.Println("JsonMarshal failed, err:", err)
			errChan <- err
			return
		}

		req, err := http.NewRequest("POST", entity.TencentLKESSEUrl, bytes.NewBuffer(payloadBytes))
		if err != nil {
			log.Println("HttpNewRequest failed, err:", err)
			errChan <- err
			return
		}

		// 添加context支持
		req = req.WithContext(ctx)

		resp, err := client.Do(req)
		if err != nil {
			log.Println("DoHttpRequest failed, err:", err)
			errChan <- err
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Printf("Get http response failed, statusCode: %d", resp.StatusCode)
			errChan <- errors.New("DoHttpRequest failed")
			return
		}

		// 读取数据
		content := ""
		contentPreviousNewlinesPos := 0
		contentCurrentNewlinesPos := 0
		reasoningContent := ""
		reasoningContentSnapshot := ""
		reasoningProcedureName := ""
		reasoningElapsed := 0.0
		reasoningElapsedSent := false
		references := []entity.Reference{}
		lastestProcedureName := ""
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if len(line) == 0 {
				continue
			}
			if strings.HasPrefix(line, "event:") {
				eventType := strings.TrimPrefix(line, "event:")
				switch eventType {
				case entity.EventTypeTokenStat, entity.EventTypeError, entity.EventTypeReply, entity.EventTypeReference, entity.EventTypeThought:
					// 什么都不做
				default:
					// 如果有新的事件类型，提示适配
					log.Println("Inspect event:", line)
				}
			} else if strings.HasPrefix(line, "data:") {
				log.Printf("Recv event data:\n%+v", line)
				data := strings.TrimPrefix(line, "data:")
				eventData := &entity.SseRecvEvent{}
				if err := json.Unmarshal([]byte(data), eventData); err != nil {
					log.Println("JsonUnmarshal response failed, err:", err)
					errChan <- err
					return
				}
				switch eventData.Type {
				case entity.EventTypeTokenStat:
					procedureCount := len(eventData.Payload.Procedures)
					if procedureCount > 0 {
						procedureName := strings.TrimSpace(eventData.Payload.Procedures[procedureCount-1].Title)
						if procedureName != lastestProcedureName {
							lastestProcedureName = procedureName
							procedureInvoice := fmt.Sprintf("> %s，请稍等...", lastestProcedureName)
							select {
							case replyChan <- procedureInvoice:
							case <-ctx.Done():
								errChan <- ctx.Err()
								return
							}
						}
					}
				case entity.EventTypeError:
					log.Printf("Get error from response: %+v", *eventData)
					errChan <- errors.New(eventData.Error.Message)
					return
				case entity.EventTypeReference:
					references = append(references, eventData.Payload.References...)
				case entity.EventTypeThought:
					if len(eventData.Payload.Procedures) > 0 {
						reasoningContent = eventData.Payload.Procedures[0].Debugging.Content
						reasoningElapsed = float64(eventData.Payload.Procedures[0].Elapsed)
						reasoningProcedureName = strings.TrimSpace(eventData.Payload.Procedures[0].Title)
						procedureInvoice := fmt.Sprintf("> %s中...", reasoningProcedureName)
						// 每遇到一段思考，输出一次，避免等待过久体验不佳以及回答过长企微强制截断。
						// 先处理段落，再处理单句。单句一般就是思考的最后一段话。
						shouldYieldReply := false
						if strings.HasSuffix(reasoningContent, "\n\n") {
							shouldYieldReply = true
						}
						if shouldYieldReply {
							// 先裁剪出新增的文字
							reasoningContentCut, _ := strings.CutPrefix(reasoningContent, reasoningContentSnapshot)
							formattedReasoningContentCut := strings.TrimSpace(reasoningContentCut)
							// 再保存快照
							reasoningContentSnapshot = reasoningContent
							// 输出思考过程片段
							if len(formattedReasoningContentCut) != 0 {
								procedureInvoice = fmt.Sprintf("%s\n> \n> %s", procedureInvoice, formattedReasoningContentCut)
								select {
								case replyChan <- formatReferences(procedureInvoice, references):
								case <-ctx.Done():
									errChan <- ctx.Err()
									return
								}
							}
						}
					}
				case entity.EventTypeReply:
					if !eventData.Payload.IsFromSelf {
						if eventData.Payload.IsFinal {
							log.Printf("Get final event, traceId: %s, data: %+v", eventData.Payload.TraceId, *eventData)
						}
						prefix := ""
						// 思考过程的最后才是回复，需要打印思考耗时
						if !reasoningElapsedSent && reasoningElapsed > 0 {
							prefix = fmt.Sprintf("> <font color=\"comment\">%s共用时%.3f秒</font>%s",
								reasoningProcedureName, reasoningElapsed/1000, prefix)
							reasoningElapsedSent = true
						}
						// 如果思考内容还有部分没输出，需要输出
						if len(reasoningContent) > len(reasoningContentSnapshot) {
							reasoningContentCut, _ := strings.CutPrefix(reasoningContent, reasoningContentSnapshot)
							// 快照保存为最终的思考内容，避免重复发送
							reasoningContentSnapshot = reasoningContent
							// 输出思考过程片段
							formattedReasoningContentCut := strings.TrimSpace(reasoningContentCut)
							if len(formattedReasoningContentCut) != 0 {
								procedureInvoice := fmt.Sprintf("> %s中...\n> \n> %s\n%s", reasoningProcedureName, formattedReasoningContentCut, prefix)
								select {
								case replyChan <- formatReferences(procedureInvoice, references):
								case <-ctx.Done():
									errChan <- ctx.Err()
									return
								}
							}
						}
						content = eventData.Payload.Content
						// 每遇到一段回答，输出一次，避免等待过久体验不佳以及回答过长企微强制截断。
						newlinesPos := strings.LastIndex(content, "\n\n")
						if newlinesPos != -1 && newlinesPos != contentCurrentNewlinesPos {
							contentCurrentNewlinesPos = newlinesPos
							// 裁剪出完整段落的文字
							contentCut := content[contentPreviousNewlinesPos:contentCurrentNewlinesPos]
							contentPreviousNewlinesPos = contentCurrentNewlinesPos
							// 输出回复片段
							formattedContentCut := strings.TrimSpace(contentCut)
							if len(prefix) > 0 {
								formattedContentCut = fmt.Sprintf("%s\n\n%s", prefix, formattedContentCut)
							}
							if len(formattedContentCut) != 0 {
								select {
								case replyChan <- formatReferences(formattedContentCut, references):
								case <-ctx.Done():
									errChan <- ctx.Err()
									return
								}
							}
						}
					} else {
						log.Printf("Get input event, traceId: %s, data: %+v", eventData.Payload.TraceId, *eventData)
					}
				default:
					log.Println("Inspect event:", *eventData)
				}
			}
		}
		log.Printf("Get http response done, err: %v", scanner.Err())
		if scanner.Err() != nil {
			return
		}
		// 发送最后的一段回复
		contentCut := content[contentCurrentNewlinesPos:]
		formattedContentCut := strings.TrimSpace(contentCut)
		if !reasoningElapsedSent && reasoningElapsed > 0 {
			formattedContentCut = fmt.Sprintf("> <font color=\"comment\">%s共用时%.3f秒</font>\n\n%s",
				reasoningProcedureName, reasoningElapsed/1000, formattedContentCut)
			reasoningElapsedSent = true
		}
		if len(formattedContentCut) != 0 {
			select {
			case replyChan <- formatReferences(formattedContentCut, references):
			case <-ctx.Done():
				errChan <- ctx.Err()
				return
			}
		}
	}()

	return replyChan, errChan
}

func formatReferences(text string, refs []entity.Reference) string {
	formattedText := text
	if len(refs) > 0 {
		for _, ref := range refs {
			refHolder := fmt.Sprintf("[%s]", ref.ID)
			refLink := fmt.Sprintf("[【资料%s】](%s)", ref.ID, ref.URL)
			formattedText = strings.ReplaceAll(formattedText, refHolder, refLink)
		}
	}
	return formattedText
}
