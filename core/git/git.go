package git

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"

	"github.com/shouni/code-reviewer/core/ports"
	// NOTE: getAuthMethod の定義があるパッケージをインポートする必要がありますが、
	// ここでは存在を前提とし、外部関数として扱います。
)

// GitAdapter は GitService インターフェースを実装する具体的な構造体です。
type GitAdapter struct {
	LocalPath                string
	SSHKeyPath               string
	BaseBranch               string
	InsecureSkipHostKeyCheck bool
	auth                     transport.AuthMethod
	repo                     *git.Repository
}

// NewGitAdapter は GitAdapter を初期化します。
func NewGitAdapter(localPath string, sshKeyPath string, opts ...Option) ports.GitService {
	adapter := &GitAdapter{
		LocalPath:  localPath,
		SSHKeyPath: sshKeyPath,
	}

	for _, opt := range opts {
		opt(adapter)
	}

	return adapter
}

// SetInsecureSkipHostKeyCheck は、SSH 接続中にホスト キーのチェックをスキップするかどうかを設定します。
func (ga *GitAdapter) SetInsecureSkipHostKeyCheck(skip bool) {
	ga.InsecureSkipHostKeyCheck = skip
}

// SetBaseBranch は、Git 操作のベース ブランチを設定します。
func (ga *GitAdapter) SetBaseBranch(branch string) {
	ga.BaseBranch = branch
}

func (ga *GitAdapter) getRepository() (*git.Repository, error) {
	if ga.repo == nil {
		repo, err := git.PlainOpen(ga.LocalPath)
		if err != nil {
			return nil, fmt.Errorf("内部リポジトリのオープンに失敗: %w", err)
		}
		ga.repo = repo
	}
	return ga.repo, nil
}

// CloneOrUpdate はリポジトリをクローンするか、既に存在する場合は更新します。
func (ga *GitAdapter) CloneOrUpdate(ctx context.Context, repositoryURL string) error {
	localPath := ga.LocalPath
	var repo *git.Repository
	var err error

	auth, err := ga.getAuthMethod(repositoryURL)
	if err != nil {
		return fmt.Errorf("認証情報取得失敗: %w", err)
	}
	ga.auth = auth

	_, err = os.Stat(localPath)
	if os.IsNotExist(err) {
		slog.Info("リポジトリが存在しないため、クローンします。", "url", repositoryURL)
		repo, err = git.PlainCloneContext(ctx, localPath, false, &git.CloneOptions{
			URL:          repositoryURL,
			SingleBranch: false,
			Auth:         ga.auth,
		})
		if err != nil {
			return fmt.Errorf("クローン失敗: %w", err)
		}
	} else if err == nil {
		repo, err = git.PlainOpen(localPath)
		if err != nil {
			return fmt.Errorf("既存リポジトリのオープン失敗: %w", err)
		}
		slog.Info("既存リポジトリをオープンしました。")
	} else {
		return fmt.Errorf("パス確認失敗: %w", err)
	}

	ga.repo = repo
	return nil
}

// Fetch はリモートから最新の変更を取得します。
func (ga *GitAdapter) Fetch(ctx context.Context) error {
	repo, err := ga.getRepository()
	if err != nil {
		return err
	}

	slog.Info("リモートから最新の変更をフェッチしています...")
	err = repo.FetchContext(ctx, &git.FetchOptions{
		Auth:     ga.auth,
		RefSpecs: []config.RefSpec{"+refs/heads/*:refs/remotes/origin/*"},
		Tags:     git.AllTags, // タグもすべて取得
		Progress: io.Discard,
	})

	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("フェッチ失敗: %w", err)
	}

	return nil
}

// GetCodeDiff は2つの参照間の差分を取得します。
func (ga *GitAdapter) GetCodeDiff(ctx context.Context, base, head string) (string, error) {
	repo, err := ga.getRepository()
	if err != nil {
		return "", err
	}

	// --- 1. フェッチ処理 ---
	// 複雑な RefSpec 生成はタグ指定時に失敗するため、安全に全ブランチとタグを更新するアプローチに変更。
	slog.Info("差分計算のために、リモートの最新情報をフェッチしています。")
	err = repo.FetchContext(ctx, &git.FetchOptions{
		RemoteName: "origin",
		Auth:       ga.auth,
		Tags:       git.AllTags,
		Progress:   io.Discard,
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return "", fmt.Errorf("フェッチに失敗: %w", err)
	}

	// --- 2. 参照の解決 ---
	baseHash, err := ga.resolveToHash(repo, base)
	if err != nil {
		return "", err
	}

	headHash, err := ga.resolveToHash(repo, head)
	if err != nil {
		return "", err
	}

	// --- 3. コミットオブジェクトの取得と差分生成 ---
	baseCommit, err := repo.CommitObject(baseHash)
	if err != nil {
		return "", fmt.Errorf("ベースコミット取得失敗 (%s): %w", baseHash, err)
	}
	headCommit, err := repo.CommitObject(headHash)
	if err != nil {
		return "", fmt.Errorf("ヘッドコミット取得失敗 (%s): %w", headHash, err)
	}

	mergeBaseCommits, err := baseCommit.MergeBase(headCommit)
	if err != nil {
		return "", fmt.Errorf("マージベースの検索に失敗しました: %w", err)
	}
	if len(mergeBaseCommits) == 0 {
		return "", fmt.Errorf("共通の祖先が見つかりませんでした")
	}

	baseTree, err := mergeBaseCommits[0].Tree()
	if err != nil {
		return "", fmt.Errorf("ベースツリー取得失敗: %w", err)
	}
	headTree, err := headCommit.Tree()
	if err != nil {
		return "", fmt.Errorf("ヘッドツリー取得失敗: %w", err)
	}

	changes, err := baseTree.Diff(headTree)
	if err != nil {
		return "", fmt.Errorf("差分計算失敗: %w", err)
	}

	patch, err := changes.Patch()
	if err != nil {
		return "", fmt.Errorf("パッチ生成失敗: %w", err)
	}
	return patch.String(), nil
}

// CheckRefExists は指定された参照が利用可能か確認します。
func (ga *GitAdapter) CheckRefExists(ctx context.Context, ref string) (bool, error) {
	repo, err := ga.getRepository()
	if err != nil {
		return false, err
	}

	if ref == "" {
		return false, fmt.Errorf("参照名が空です")
	}

	_, err = ga.resolveToHash(repo, ref)
	if err != nil {
		slog.Debug("参照を解決できませんでした。", "ref", ref, "error", err)
		return false, nil
	}

	return true, nil
}

// Cleanup はローカルリポジトリを削除します。
func (ga *GitAdapter) Cleanup(ctx context.Context) error {
	slog.Info("クリーンアップ: ディレクトリを削除します。", "path", ga.LocalPath)
	ga.repo = nil
	return os.RemoveAll(ga.LocalPath)
}

// resolveToHash はリファレンス名、タグ、またはコミットハッシュを plumbing.Hash に解決します。
func (ga *GitAdapter) resolveToHash(repo *git.Repository, ref string) (plumbing.Hash, error) {
	// 1. まずリモートブランチとして解決を試みる (16進数名のブランチ対策)
	refName := plumbing.NewRemoteReferenceName("origin", ref)
	reference, err := repo.Reference(refName, false)
	if err == nil {
		return reference.Hash(), nil
	}

	// 2. リビジョンとして解決
	// ^{commit} を付与することで、アノテートタグの場合もコミットハッシュまで解決(peel)します
	hash, err := repo.ResolveRevision(plumbing.Revision(ref + "^{commit}"))
	if err != nil {
		// フォールバック: 修飾子なしでの解決も試みる (一部の特殊な参照用)
		hash, err = repo.ResolveRevision(plumbing.Revision(ref))
		if err != nil {
			return plumbing.ZeroHash, fmt.Errorf("参照 '%s' を解決できませんでした: %w", ref, err)
		}
	}

	// コミットオブジェクトとして存在するか最終確認
	if _, err = repo.CommitObject(*hash); err != nil {
		return plumbing.ZeroHash, fmt.Errorf("オブジェクト '%s' は有効なコミットではありません: %w", ref, err)
	}

	return *hash, nil
}
