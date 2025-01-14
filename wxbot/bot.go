package wxbot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	conf "api-pay/config"
)

var WxBot *Bot

type Bot struct {
	webhookURL string
}

type Message struct {
	MsgType  string                 `json:"msgtype"`
	Markdown map[string]interface{} `json:"markdown"`
}

// InitBot 初始化机器人
func InitBot() {
	WxBot = NewBot()
}

// NewBot 创建一个新的机器人实例
func NewBot() *Bot {
	return &Bot{
		webhookURL: fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=%s", conf.AppConfig.BotKey),
	}
}

// SendMarkdown 发送markdown格式消息
func (b *Bot) SendMarkdown(content string) error {
	msg := Message{
		MsgType: "markdown",
		Markdown: map[string]interface{}{
			"content": content,
		},
	}

	jsonData, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal message failed: %w", err)
	}

	resp, err := http.Post(b.webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("send message failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
