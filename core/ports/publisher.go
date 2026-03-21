package ports

import (
	"context"
	"io"
)

// ReviewData はレポート生成に必要なすべての情報をまとめた構造体です。
type ReviewData struct {
	RepoURL        string
	BaseBranch     string
	FeatureBranch  string
	ReviewMarkdown string
}

// Publisher はレビューレポートを指定されたURIに公開する責務を定義します。
type Publisher interface {
	Publish(ctx context.Context, uri string, data ReviewData) error
}

// MarkdownToHtmlRunner は、Markdown コンテンツを完全な HTML ドキュメントに変換する契約です。
type MarkdownToHtmlRunner interface {
	// Run は Markdown コンテンツをバイトスライスで受け取り、HTML コンテンツを含む io.Reader を返します。
	Run(markdownContent []byte) (io.Reader, error)
}
