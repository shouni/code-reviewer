package publisher

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/shouni/code-reviewer/core/ports"
	"github.com/shouni/code-reviewer/remoteio"
	"github.com/shouni/code-reviewer/utils/timeutil"
)

const (
	contentTypeHTML = "text/html; charset=utf-8"
	reviewTitle     = "AIコードレビュー結果"
)

type Publisher struct {
	writer remoteio.OutputWriter
	md     ports.MarkdownToHtmlRunner
}

// NewPublisher は Publisher の新しいインスタンスを生成します。
func NewPublisher(writer remoteio.OutputWriter, md ports.MarkdownToHtmlRunner) (ports.Publisher, error) {
	if writer == nil {
		return nil, fmt.Errorf("writer is nil")
	}
	if md == nil {
		return nil, fmt.Errorf("MarkdownToHtmlRunner is nil")
	}

	return &Publisher{
		writer: writer,
		md:     md,
	}, nil
}

// Publish ReviewData からレンダリングされた HTML をリモート ストレージ内の指定された URI にアップロードし、失敗した場合はエラーを返します。
func (p *Publisher) Publish(ctx context.Context, uri string, data ports.ReviewData) error {
	htmlReader, err := p.convertMarkdownToHTML(data)
	if err != nil {
		return fmt.Errorf("HTML変換に失敗しました: %w", err)
	}

	slog.Info("リモートストレージへアップロード開始", "uri", uri)

	if err := p.writer.Write(ctx, uri, htmlReader, contentTypeHTML); err != nil {
		return fmt.Errorf("リモートストレージへの書き込みに失敗しました: %w", err)
	}

	return nil
}

// convertMarkdownToHTML は ReviewData から Markdown を組み立て、HTML に変換して返します。
func (p *Publisher) convertMarkdownToHTML(data ports.ReviewData) (io.Reader, error) {
	nowJST := timeutil.NowJST()
	reviewTimeStr := nowJST.Format("2006/01/02 15:04:05 JST")

	summaryMarkdown := fmt.Sprintf(
		"レビュー対象リポジトリ: `%s`\n\nブランチ差分: `%s` ← `%s`\n\nレビュー実行日時: *%s*\n\n",
		data.RepoURL,
		data.BaseBranch,
		data.FeatureBranch,
		reviewTimeStr,
	)

	var buffer bytes.Buffer
	buffer.WriteString("# " + reviewTitle + "\n\n")
	buffer.WriteString(summaryMarkdown + "\n\n")
	buffer.WriteString(data.ReviewMarkdown)

	return p.md.Run(buffer.Bytes())
}
