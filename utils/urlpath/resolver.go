package urlpath

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
)

// URIスキームの定義
const (
	SchemeGCS = "gs://"
	SchemeS3  = "s3://"
)

// IsRemoteURI は、指定されたURIがクラウドストレージ（GCSまたはS3）を指しているか判定します。
func IsRemoteURI(uri string) bool {
	return IsGCSURI(uri) || IsS3URI(uri)
}

// IsGCSURI は、URIが Google Cloud Storage (gs://) を指しているか判定します。
func IsGCSURI(uri string) bool {
	return strings.HasPrefix(strings.ToLower(uri), SchemeGCS)
}

// IsS3URI は、URIが S3 (s3://) を指しているか判定します。
func IsS3URI(uri string) bool {
	return strings.HasPrefix(strings.ToLower(uri), SchemeS3)
}

// ResolveBaseDir は、入力パスから親ディレクトリのパスを抽出し、
// 末尾がセパレータ（URLなら /、ローカルならOS依存）で終わるように正規化します。
func ResolveBaseDir(rawPath string) string {
	if rawPath == "" {
		return ""
	}

	u, err := url.Parse(rawPath)
	// スキームがある場合はURLとして処理
	if err == nil && u.Scheme != "" {
		// ディレクトリ構造を取得するためにパスの末尾を調整
		if !strings.HasSuffix(u.Path, "/") {
			u.Path = filepath.Dir(u.Path)
		}
		baseURL := u.String()
		if !strings.HasSuffix(baseURL, "/") {
			baseURL += "/"
		}
		return baseURL
	}

	// ローカルファイルパスとして処理
	baseDir := filepath.Dir(rawPath)
	if baseDir == "." {
		return "." + string(filepath.Separator)
	}

	if !strings.HasSuffix(baseDir, string(filepath.Separator)) {
		baseDir += string(filepath.Separator)
	}
	return baseDir
}

// ResolvePath は、ベースディレクトリとファイル名を結合します。
// リモートURIの場合はURLとして、それ以外はローカルパスとして結合します。
func ResolvePath(baseDir, fileName string) (string, error) {
	if IsRemoteURI(baseDir) {
		result, err := url.JoinPath(baseDir, fileName)
		if err != nil {
			return "", fmt.Errorf("リモートストレージパスの結合に失敗: %w", err)
		}
		return result, nil
	}

	return filepath.Join(baseDir, fileName), nil
}

// GenerateIndexedPath は、指定されたパスの拡張子の前に連番を挿入します。
// 例: "path/to/image.png", 1 -> "path/to/image_1.png"
func GenerateIndexedPath(basePath string, index int) (string, error) {
	if index <= 0 {
		return "", fmt.Errorf("インデックスは1以上の整数である必要があります: %d", index)
	}

	// URLの場合はPath部分のみ、ローカルなら全体から拡張子を取得
	ext := filepath.Ext(basePath)
	mainPath := strings.TrimSuffix(basePath, ext)

	return fmt.Sprintf("%s_%d%s", mainPath, index, ext), nil
}
