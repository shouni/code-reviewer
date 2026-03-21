package git

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	cryptossh "golang.org/x/crypto/ssh"
)

// getAuthMethod は go-git が使用する認証方法を返します。
// GitAdapter の設定に基づいて SSH 認証を構築します。
func (ga *GitAdapter) getAuthMethod(repoURL string) (transport.AuthMethod, error) {
	if isSSHRepoURL(repoURL) {
		username, err := sshUsernameFromRepoURL(repoURL)
		if err != nil {
			return nil, fmt.Errorf("リポジトリURLからSSHユーザー名を解決できませんでした: %w", err)
		}

		// 2. SSHキーパスの展開とファイルの読み込み
		sshKeyPath, err := expandTilde(ga.SSHKeyPath)
		if err != nil {
			return nil, fmt.Errorf("SSHキーパスの展開に失敗しました: %w", err)
		}

		if _, err := os.Stat(sshKeyPath); os.IsNotExist(err) {
			return nil, fmt.Errorf("SSHキーファイルが見つかりません: %s", sshKeyPath)
		}

		sshKey, err := os.ReadFile(sshKeyPath)
		if err != nil {
			return nil, fmt.Errorf("SSHキーファイルの読み込みに失敗しました: %w", err)
		}

		// 3. PublicKeys 認証メソッドの生成
		// パスフレーズが空 ("") の SSH キーを想定
		auth, err := ssh.NewPublicKeys(username, sshKey, "")
		if err != nil {
			return nil, fmt.Errorf("SSH認証キーのロードに失敗しました: %w", err)
		}

		// 4. HostKeyCallback の設定
		if ga.InsecureSkipHostKeyCheck {
			auth.HostKeyCallback = cryptossh.InsecureIgnoreHostKey()
		} else {
			auth.HostKeyCallback = nil
		}

		return auth, nil
	}

	// SSH ではないリポジトリURLの場合（例：https://）
	// go-git は通常、HTTP/HTTPSリポジトリに対しては認証なし（nil）でアクセスを試みます。
	// 必要であれば、HTTP Basic Auth などのロジックを追加できます。
	return nil, nil
}

// isSSHRepoURL はリポジトリURLが SSH 形式かどうかを判定します。
// HTTP/HTTPS スキームを明示的に除外することで、Basic認証を含むURLとの誤認を防ぎます。
func isSSHRepoURL(repoURL string) bool {
	// 1. HTTP/HTTPS は明示的に除外
	if strings.HasPrefix(repoURL, "http://") || strings.HasPrefix(repoURL, "https://") {
		return false
	}

	// 2. ssh:// スキーム、または scp 形式 (git@github.com:...) をチェック
	return strings.HasPrefix(repoURL, "ssh://") || strings.Contains(repoURL, "@")
}

func sshUsernameFromRepoURL(repoURL string) (string, error) {
	if strings.HasPrefix(repoURL, "ssh://") {
		u, err := url.Parse(repoURL)
		if err != nil {
			return "", err
		}
		if u.User != nil && u.User.Username() != "" {
			return u.User.Username(), nil
		}
		return "git", nil
	}

	at := strings.Index(repoURL, "@")
	colon := strings.Index(repoURL, ":")
	if at > 0 && colon > at {
		return repoURL[:at], nil
	}

	return "", fmt.Errorf("未対応のSSHリポジトリURL形式です: %s", repoURL)
}
