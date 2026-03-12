package scheduler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"sop-chat/internal/config"
)

// SendToWebhook 根据 Webhook 类型将消息发送到对应平台（公开，供外部触发测试使用）
func SendToWebhook(cfg config.WebhookConfig, content string) error {
	return sendToWebhook(cfg, content)
}

// sendToWebhook 根据 Webhook 类型将消息发送到对应平台
func sendToWebhook(cfg config.WebhookConfig, content string) error {
	if cfg.URL == "" {
		return fmt.Errorf("webhook URL 未配置")
	}

	msgType := strings.ToLower(cfg.MsgType)
	if msgType == "" {
		msgType = "text"
	}

	var payload interface{}

	switch strings.ToLower(cfg.Type) {
	case "dingtalk":
		payload = buildDingTalkPayload(msgType, cfg.Title, content)
	case "feishu":
		payload = buildFeishuPayload(msgType, cfg.Title, content)
	case "wecom":
		payload = buildWeComPayload(msgType, content)
	default:
		return fmt.Errorf("不支持的 webhook 类型: %q（支持：dingtalk、feishu、wecom）", cfg.Type)
	}

	return postJSON(cfg.URL, payload)
}

// buildDingTalkPayload 构造钉钉机器人 Webhook 消息体
func buildDingTalkPayload(msgType, title, content string) interface{} {
	switch msgType {
	case "markdown":
		if title == "" {
			title = "通知"
		}
		return map[string]interface{}{
			"msgtype": "markdown",
			"markdown": map[string]string{
				"title": title,
				"text":  content,
			},
		}
	default: // text
		return map[string]interface{}{
			"msgtype": "text",
			"text": map[string]string{
				"content": content,
			},
		}
	}
}

// buildFeishuPayload 构造飞书机器人 Webhook 消息体
func buildFeishuPayload(msgType, title, content string) interface{} {
	switch msgType {
	case "markdown", "post":
		if title == "" {
			title = "通知"
		}
		return map[string]interface{}{
			"msg_type": "post",
			"content": map[string]interface{}{
				"post": map[string]interface{}{
					"zh_cn": map[string]interface{}{
						"title": title,
						"content": [][]map[string]string{
							{{"tag": "text", "text": content}},
						},
					},
				},
			},
		}
	default: // text
		return map[string]interface{}{
			"msg_type": "text",
			"content": map[string]string{
				"text": content,
			},
		}
	}
}

// buildWeComPayload 构造企业微信机器人 Webhook 消息体
func buildWeComPayload(msgType, content string) interface{} {
	switch msgType {
	case "markdown":
		return map[string]interface{}{
			"msgtype": "markdown",
			"markdown": map[string]string{
				"content": content,
			},
		}
	default: // text
		return map[string]interface{}{
			"msgtype": "text",
			"text": map[string]string{
				"content": content,
			},
		}
	}
}

// postJSON 向指定 URL 发送 JSON POST 请求
func postJSON(url string, payload interface{}) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("序列化消息体失败: %w", err)
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("发送 webhook 请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook 返回非成功状态码: %d", resp.StatusCode)
	}
	return nil
}
