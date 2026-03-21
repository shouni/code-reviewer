package ports

import "context"

// CodeReviewAI は、AIとの通信機能の抽象化を提供し、DIで使用されます。
type CodeReviewAI interface {
	// ReviewCodeDiff は完成されたプロンプトを基にAIにレビューを依頼します。
	ReviewCodeDiff(ctx context.Context, modelName, finalPrompt string) (string, error)
}
