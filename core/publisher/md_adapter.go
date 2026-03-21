package publisher

import (
	"fmt"
	"io"

	"github.com/shouni/code-reviewer/md/builder"
	"github.com/shouni/code-reviewer/md/ports"
)

// MarkdownConverterAdapter は、md2htmlのロジックをラップしたアダプターです。
type MarkdownConverterAdapter struct {
	md2htmlRunner ports.Runner
}

// NewMarkdownConverterAdapter は、MarkdownToHtmlRunner インスタンスを初期化して返します。
func NewMarkdownConverterAdapter() (*MarkdownConverterAdapter, error) {
	opts := []builder.Option{
		builder.WithEnableUnsafeHTML(true),
	}

	b, err := builder.New(opts...)
	if err != nil {
		return nil, fmt.Errorf("Markdown builderの初期化に失敗: %w", err)
	}

	md2htmlRunner, err := b.BuildRunner()
	if err != nil {
		return nil, fmt.Errorf("MarkdownToHtmlRunnerの構築に失敗: %w", err)
	}

	return &MarkdownConverterAdapter{
		md2htmlRunner: md2htmlRunner,
	}, nil
}

// Run は MarkdownToHtmlRunner インターフェースを満たします。
func (m *MarkdownConverterAdapter) Run(markdownContent []byte) (io.Reader, error) {
	// Run はタイトルを受け取りますが、
	// ここではタイトルはレビュー結果内で #H1 として提供されるため、空文字を渡します。
	buffer, err := m.md2htmlRunner.Run("", markdownContent)
	if err != nil {
		return nil, fmt.Errorf("MarkdownからHTMLへの変換に失敗: %w", err)
	}

	return buffer, nil
}
