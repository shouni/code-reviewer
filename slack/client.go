package slack

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/slack-go/slack"

	"github.com/shouni/code-reviewer/httpkit"
	"github.com/shouni/code-reviewer/utils/text"
)

const (
	// defaultUsername はデフォルトのユーザー名です。
	defaultUsername = "Bot"
	// defaultIconEmoji はデフォルトの絵文字アイコンを表します。
	defaultIconEmoji = ":robot_face:"
)

// Client は Slack Webhook API と連携するためのクライアントです。
type Client struct {
	client     httpkit.RequestExecutor
	WebhookURL string
	Username   string
	IconEmoji  string
	Channel    string
}

// NewClient は必須項目のみを引数に受け取り、オプションで設定をカスタマイズします
func NewClient(client httpkit.RequestExecutor, webhookURL string, opts ...Option) (*Client, error) {
	if client == nil {
		return nil, errors.New("http client cannot be nil")
	}
	if webhookURL == "" {
		return nil, errors.New("webhookURL is required")
	}

	c := &Client{
		client:     client,
		WebhookURL: webhookURL,
		Username:   defaultUsername,
		IconEmoji:  defaultIconEmoji,
	}

	// オプションを適用
	for _, opt := range opts {
		opt(c)
	}

	return c, nil
}

// SendTextWithHeader はメッセージを構築し、Slackに送信します。
func (s *Client) SendTextWithHeader(ctx context.Context, headerText string, message string) error {
	// Block Kit の構築をヘルパー関数に委譲
	blocks, err := buildMessageBlocks(headerText, message)
	if err != nil {
		return fmt.Errorf("Slack Block Kitの構築に失敗しました: %w", err)
	}

	// Webhookメッセージの作成
	msg := slack.WebhookMessage{
		Text:      headerText, // フォールバック用テキスト
		Username:  s.Username,
		IconEmoji: s.IconEmoji,
		Channel:   s.Channel,
		Blocks: &slack.Blocks{
			BlockSet: blocks,
		},
	}

	// メッセージの送信
	_, err = s.client.PostJSONAndFetchBytes(ctx, s.WebhookURL, msg)
	if err != nil {
		return fmt.Errorf("Slack Webhookメッセージの送信に失敗しました: %w", err)
	}

	return nil
}

// SendText は、プレーンテキストメッセージを通知します。（自動タイトル生成）
func (s *Client) SendText(ctx context.Context, message string) error {
	header := "📢 通知メッセージ" // デフォルトヘッダー

	// message本文からヘッダーを生成
	if len(message) > 0 {
		// 1. 最初の行を抽出
		firstLine := strings.SplitN(message, "\n", 2)[0]
		// 2. 抽出した行から前後の空白を削除
		trimmedLine := strings.TrimSpace(firstLine)
		trimmedLine = text.Truncate(trimmedLine, 50, "...")

		if trimmedLine != "" {
			// 抽出した行をヘッダーとして採用
			header = fmt.Sprintf("📢 %s", trimmedLine)
		}
	}

	return s.SendTextWithHeader(ctx, header, message)
}
