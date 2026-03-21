package git

import (
	"fmt"
	"os/user"
	"path/filepath"
	"strings"
)

// quotePathForShell は、パスをシェルで安全に使用できるようにシングルクォートで囲みます。
// パス内のシングルクォートもエスケープします。
func quotePathForShell(path string) string {
	// シングルクォートを '\' に置換してエスケープ
	return "'" + strings.ReplaceAll(path, "'", "'\\''") + "'"
}

// expandTilde はクロスプラットフォームなチルダ展開をサポートする
func expandTilde(path string) (string, error) {
	if !strings.HasPrefix(path, "~/") {
		return path, nil
	}
	// os/userパッケージの利用は、クロスコンパイル環境によっては問題になる可能性がありますが、
	// 通常のアプリケーションでは標準的なアプローチです。
	currentUser, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("現在のユーザーのホームディレクトリの取得に失敗しました: %w", err)
	}
	return filepath.Join(currentUser.HomeDir, path[2:]), nil
}
