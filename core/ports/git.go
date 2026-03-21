package ports

import "context"

// GitService はGitリポジトリ操作の抽象化を提供します。
type GitService interface {
	// CloneOrUpdate はリポジトリをクローンまたは更新し、成功時に nil を返します。
	CloneOrUpdate(ctx context.Context, repositoryURL string) error
	// Fetch はリモートから最新の変更を取得します。
	Fetch(ctx context.Context) error
	// CheckRefExists は指定された参照（ブランチ、タグ、またはコミットハッシュ）が利用可能か確認します。
	CheckRefExists(ctx context.Context, ref string) (bool, error)
	// GetCodeDiff は指定された2つの参照間の純粋な差分を文字列として取得します。
	GetCodeDiff(ctx context.Context, base, head string) (string, error)
	// Cleanup は処理後にローカルリポジトリをクリーンな状態に戻します。
	Cleanup(ctx context.Context) error
}
