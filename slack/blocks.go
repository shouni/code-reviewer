package slack

import (
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/slack-go/slack"

	"github.com/shouni/code-reviewer/utils/text"
	"github.com/shouni/code-reviewer/utils/timeutil"
)

const (
	// maxSectionLength はSlackセクションブロックの最大文字数（Slackの制限は3000）
	maxSectionLength = 2900
	// maxBlocks はSlackメッセージあたりの最大ブロック数（Slackの制限は50）
	maxBlocks = 50
	// truncationSuffix はメッセージ切り捨て時に追加されるサフィックス
	truncationSuffix = "\n\n... (メッセージが長すぎるため省略されました)"
	// footerTimeFormat はSlackメッセージのフッターに表示する時刻のフォーマット
	footerTimeFormat = "2006/01/02 15:04:05 JST"
)

// 正規表現はグローバル変数として一度だけコンパイル
var (
	boldRegex     = regexp.MustCompile(`\*\*(.*?)\*\*`)   // **text** -> *text*
	headerRegex   = regexp.MustCompile(`(?m)^##\s*(.*)$`) // ## Title -> *Title*
	listItemRegex = regexp.MustCompile(`(?m)^\s*-\s+`)    // - item -> • item
)

// buildMessageBlocks は、SlackのBlock Kitスライスを構築する責務を担います。
func buildMessageBlocks(headerText string, message string) ([]slack.Block, error) {
	if headerText == "" {
		return nil, errors.New("headerText は必須です")
	}

	// ヘッダーブロックを作成
	blocks := []slack.Block{
		slack.NewHeaderBlock(
			slack.NewTextBlockObject("plain_text", headerText, true, false),
		),
		slack.NewDividerBlock(),
	}

	// 元のロジックを維持するため、1要素のスライスでループ
	reviewSections := []string{message}

	for _, sectionText := range reviewSections {
		if len(blocks) >= maxBlocks-2 { // 2 = この後追加するSection/Divider + 最後のFooter
			slog.Warn("Notification message is too long, truncating message.")
			blocks = append(blocks, slack.NewSectionBlock(
				slack.NewTextBlockObject("mrkdwn", truncationSuffix, false, false), nil, nil))
			break
		}
		if strings.TrimSpace(sectionText) != "" {
			// Markdown整形処理
			processedText := sectionText
			processedText = boldRegex.ReplaceAllString(processedText, "*$1*")
			processedText = headerRegex.ReplaceAllString(processedText, "*$1*")
			processedText = listItemRegex.ReplaceAllString(processedText, "• ")

			// 文字数制限の適用
			textLen := utf8.RuneCountInString(processedText)
			if textLen > maxSectionLength {
				slog.Warn("The notification message is too long, truncating.",
					"current_runes", textLen,
					"max_runes", maxSectionLength)
				processedText = text.Truncate(processedText, maxSectionLength-utf8.RuneCountInString(truncationSuffix), truncationSuffix)
			}

			blocks = append(blocks, slack.NewSectionBlock(
				slack.NewTextBlockObject("mrkdwn", processedText, false, false), nil, nil),
				slack.NewDividerBlock(),
			)
		}
	}

	// 最後の余分なDividerを削除
	if len(blocks) > 2 { // ヘッダーブロック分を考慮
		blocks = blocks[:len(blocks)-1]
	}

	// フッターには送信時刻を含める
	footerBlock := slack.NewContextBlock(
		"notification-context",
		slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("送信時刻: %s",
			timeutil.NowJST().Format(footerTimeFormat)), false, false),
	)
	blocks = append(blocks, footerBlock)

	return blocks, nil
}
