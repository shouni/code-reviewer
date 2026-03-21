package git

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/shouni/code-reviewer/core/ports"
)

// GitLocalAdapter は、ローカルの 'git' コマンドを subprocess/os/exec 経由で実行するアダプタです。
type GitLocalAdapter struct {
	LocalPath                string
	SSHKeyPath               string
	BaseBranch               string
	InsecureSkipHostKeyCheck bool
}

// NewGitLocalAdapter は GitLocalAdapter を初期化します。
func NewGitLocalAdapter(localPath string, sshKeyPath string, opts ...Option) ports.GitService {
	adapter := &GitLocalAdapter{
		LocalPath:  localPath,
		SSHKeyPath: sshKeyPath,
		BaseBranch: "main",
	}

	for _, opt := range opts {
		opt(adapter)
	}

	return adapter
}

// SetInsecureSkipHostKeyCheck は、SSH 接続中にホスト キーのチェックをスキップするかどうかを設定します。
func (ga *GitLocalAdapter) SetInsecureSkipHostKeyCheck(skip bool) {
	ga.InsecureSkipHostKeyCheck = skip
}

// SetBaseBranch は、Git 操作のベース ブランチを設定します。
func (ga *GitLocalAdapter) SetBaseBranch(branch string) {
	ga.BaseBranch = branch
}

// getEnvWithSSH は、現在の環境変数に GIT_SSH_COMMAND を追加したリストを返します。
// GIT_SSH_COMMAND はシェル経由で実行されるため、キーのパスを適切にエスケープします。
func (ga *GitLocalAdapter) getEnvWithSSH() []string {
	env := os.Environ()
	if ga.SSHKeyPath == "" {
		return env
	}

	// コマンドインジェクション脆弱性対策
	safeKeyPath := quotePathForShell(ga.SSHKeyPath)

	// ssh -i '/path/to/key' -F /dev/null ... の形式で構築
	sshCmdParts := []string{"ssh", "-i", safeKeyPath, "-F", "/dev/null"}

	if ga.InsecureSkipHostKeyCheck {
		sshCmdParts = append(sshCmdParts, "-o", "StrictHostKeyChecking=no")
	}

	// スペースで結合してコマンド文字列にする
	sshCmd := strings.Join(sshCmdParts, " ")

	slog.Debug("GIT_SSH_COMMANDを構築", "cmd", sshCmd)

	// GIT_SSH_COMMANDを環境変数に追加
	env = append(env, fmt.Sprintf("GIT_SSH_COMMAND=%s", sshCmd))
	return env
}

// runGitCommand は、指定されたGitコマンドをアダプタの設定（SSH環境変数など）で実行します。
func (ga *GitLocalAdapter) runGitCommand(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = ga.LocalPath

	// 統一された環境変数設定ロジックを使用
	cmd.Env = ga.getEnvWithSSH()

	slog.Debug("Gitコマンドを実行中", "dir", cmd.Dir, "args", args)
	output, err := cmd.CombinedOutput()
	outputStr := strings.TrimSpace(string(output))

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			slog.Error("Gitコマンド実行に失敗しました", "args", args, "stderr", outputStr, "exit", exitErr.ExitCode())
			return "", fmt.Errorf("Gitコマンド実行失敗: %s. 出力:\n%s", exitErr.Error(), outputStr)
		}
		slog.Error("Gitコマンド実行中に予期せぬエラーが発生しました", "args", args, "error", err)
		return "", fmt.Errorf("予期せぬGit実行エラー: %w", err)
	}

	slog.Debug("Gitコマンド成功", "args", args)
	return outputStr, nil
}

// --- coreAdapters.GitService インターフェースの実装 ---

// CloneOrUpdate はリポジトリをクローンするか、既に存在する場合は更新を試みます。
func (ga *GitLocalAdapter) CloneOrUpdate(ctx context.Context, repositoryURL string) error {
	localPath := ga.LocalPath
	info, err := os.Stat(localPath)
	if err == nil {
		// localPath が存在する
		if !info.IsDir() {
			return fmt.Errorf("ローカルパス '%s' はディレクトリではありません。Gitリポジトリをクローンできません。", localPath)
		}
		// localPath がディレクトリの場合、.git ディレクトリの存在を確認
		_, gitDirErr := os.Stat(filepath.Join(localPath, ".git"))
		if os.IsNotExist(gitDirErr) {
			// localPath はディレクトリだが、.git がない -> Gitリポジトリではない
			return fmt.Errorf("ローカルパス '%s' は存在しますが、Gitリポジトリではありません。手動で削除するか、別のパスを指定してください。", localPath)
		} else if gitDirErr != nil {
			// .git ディレクトリの確認中にエラー
			return fmt.Errorf("ローカルパス '%s' 内の .git ディレクトリの確認に失敗しました: %w", localPath, gitDirErr)
		}
		// .git ディレクトリが存在する -> 既存リポジトリとして扱う
		slog.Info("既存リポジトリをオープンしました。後続の Fetch に更新を委ねます。", "path", localPath)
		return nil
	} else if os.IsNotExist(err) {
		// localPath が存在しない場合はクローン
		slog.Info("リポジトリが存在しないため、クローンします。", "url", repositoryURL, "path", localPath, "branch", ga.BaseBranch)

		parentDir := filepath.Dir(localPath)
		repoDir := filepath.Base(localPath)

		// 親ディレクトリが存在しない場合は作成
		if _, err := os.Stat(parentDir); os.IsNotExist(err) {
			if err := os.MkdirAll(parentDir, 0755); err != nil {
				return fmt.Errorf("親ディレクトリの作成に失敗しました: %w", err)
			}
		}

		// クローン実行
		cloneArgs := []string{"clone", repositoryURL, repoDir}

		cmd := exec.CommandContext(ctx, "git", cloneArgs...)
		cmd.Dir = parentDir

		// SSH認証環境変数を引き継ぐ
		cmd.Env = ga.getEnvWithSSH()

		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("リポジトリのクローンに失敗しました: %s\n%s", err.Error(), string(output))
		}
		slog.Info("リポジトリのクローンに成功しました。", "path", localPath)
		return nil
	} else {
		// その他の os.Stat エラー
		return fmt.Errorf("ローカルパス '%s' の確認に失敗しました: %w", localPath, err)
	}
}

// Fetch はリモートから最新の変更を取得します。
func (ga *GitLocalAdapter) Fetch(ctx context.Context) error {
	_, err := ga.runGitCommand(ctx, "fetch", "origin", "--prune")
	if err != nil {
		return fmt.Errorf("リモートからのフェッチに失敗しました: %w", err)
	}
	return nil
}

// GetCodeDiff は指定された2つのブランチまたはコミットハッシュ間の純粋な差分を、ローカルの 'git diff' コマンドで取得します。
// ブランチ名の場合は "origin/" プレフィックスを付け、ハッシュ値の場合はそのまま使用します。
func (ga *GitLocalAdapter) GetCodeDiff(ctx context.Context, base, head string) (string, error) {
	// resolveRef を廃止し、検証済みの参照を取得
	baseRef, err := ga.resolveToHash(ctx, base)
	if err != nil {
		return "", fmt.Errorf("ベース参照の解決失敗: %w", err)
	}

	featureRef, err := ga.resolveToHash(ctx, head)
	if err != nil {
		return "", fmt.Errorf("フィーチャー参照の解決失敗: %w", err)
	}

	slog.Info("差分を計算します", "base", baseRef, "feature", featureRef)

	// 3点比較 Diff の実行
	diffArgs := []string{
		"diff",
		fmt.Sprintf("%s...%s", baseRef, featureRef),
		"--unified=10",
	}

	return ga.runGitCommand(ctx, diffArgs...)
}

// CheckRefExists は指定されたブランチまたはコミットハッシュがリモート 'origin' に存在するか確認します。
// ブランチ名の場合は "origin/" プレフィックスを付け、ハッシュ値の場合はそのまま使用します。
func (ga *GitLocalAdapter) CheckRefExists(ctx context.Context, ref string) (bool, error) {
	if ref == "" {
		return false, fmt.Errorf("参照の存在確認に失敗しました: 参照名が空です")
	}

	// resolveToHash 内部で rev-parse --verify 相当の解決と検証が行われている
	_, err := ga.resolveToHash(ctx, ref)
	if err != nil {
		// 解決に失敗した＝存在しない、と判断して正常に false を返す
		slog.Debug("参照を解決できませんでした（存在しない可能性があります）", "ref", ref, "error", err)
		return false, nil
	}

	// 解決できたのであれば、それは「存在する」とみなしてOK！
	return true, nil
}

// Cleanup はクリーンアップを実行します。
func (ga *GitLocalAdapter) Cleanup(ctx context.Context) error {
	slog.Info("クリーンアップ: fetch -> checkout -f -> clean を実行します。", "path", ga.LocalPath)

	if _, err := ga.runGitCommand(ctx, "fetch", "origin"); err != nil {
		return fmt.Errorf("クリーンアップ中のフェッチに失敗: %w", err)
	}

	baseRef, isBranchRef, err := ga.resolveToRef(ctx, ga.BaseBranch)
	if err != nil {
		return fmt.Errorf("クリーンアップ用のベース参照を解決できませんでした: %w", err)
	}

	if isBranchRef {
		localBranch := branchNameFromResolvedRef(baseRef)
		// -f を付与し、ローカルの変更を強制破棄してチェックアウト
		checkoutArgs := []string{"checkout", "-f", "-B", localBranch, baseRef}
		if _, err := ga.runGitCommand(ctx, checkoutArgs...); err != nil {
			return fmt.Errorf("クリーンアップ中のチェックアウト/リセットに失敗: %w", err)
		}
	} else {
		slog.Info("ベース参照がコミット指定のため直接チェックアウトします。", "base", ga.BaseBranch, "resolved", baseRef)
		// コミット指定時も -f を付与
		if _, err := ga.runGitCommand(ctx, "checkout", "-f", baseRef); err != nil {
			return fmt.Errorf("クリーンアップ中のコミットチェックアウトに失敗: %w", err)
		}
	}

	if _, err := ga.runGitCommand(ctx, "clean", "-f", "-d"); err != nil {
		return fmt.Errorf("クリーンアップ中のクリーンに失敗: %w", err)
	}

	slog.Info("クリーンアップ完了。", "base", ga.BaseBranch)
	return nil
}

// resolveToHash は、指定された文字列をリモートブランチ(origin/)または コミットハッシュとして解決・検証します。
func (ga *GitLocalAdapter) resolveToHash(ctx context.Context, ref string) (string, error) {
	resolved, _, err := ga.resolveToRef(ctx, ref)
	if err != nil {
		return "", err
	}
	return resolved, nil
}

func (ga *GitLocalAdapter) resolveToRef(ctx context.Context, ref string) (resolved string, isBranchRef bool, err error) {
	if ref == "" {
		return "", false, fmt.Errorf("参照名が空です")
	}

	candidates := buildRefCandidates(ref)
	for _, candidate := range candidates {
		if _, err := ga.runGitCommand(ctx, "rev-parse", "--verify", candidate.ref+"^{commit}"); err == nil {
			slog.Debug("参照を解決しました", "input", ref, "resolved", candidate.ref, "branch_ref", candidate.isBranchRef)
			return candidate.ref, candidate.isBranchRef, nil
		}
	}

	return "", false, fmt.Errorf("参照 '%s' を解決できませんでした（リモートブランチまたは有効なコミットではありません）", ref)
}

type refCandidate struct {
	ref         string
	isBranchRef bool
}

func buildRefCandidates(ref string) []refCandidate {
	trimmed := strings.TrimSpace(ref)
	if trimmed == "" {
		return nil
	}

	// origin/ が付いている場合はそれのみ
	if strings.HasPrefix(trimmed, "origin/") {
		return []refCandidate{{ref: trimmed, isBranchRef: true}}
	}

	// 条件分岐を削除し、常に「リモートブランチ -> コミット」の順で解決を試みる。
	// これにより、数字のみのブランチ名(12345等)がコミットと誤認されるのを防ぎ、かつコードが最小化される。
	return []refCandidate{
		{ref: fmt.Sprintf("origin/%s", trimmed), isBranchRef: true},
		{ref: trimmed, isBranchRef: false},
	}
}

func branchNameFromResolvedRef(ref string) string {
	return strings.TrimPrefix(ref, "origin/")
}
