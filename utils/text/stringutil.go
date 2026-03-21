package text

import "strings"

// ParseCommaSeparatedList は、カンマ区切りの文字列を入力として受け取り、
// 各要素の先頭と末尾の空白を除去（トリミング）した後、空でない要素のみを含む
// クリーンな文字列スライスを返します。
// 入力が空文字列の場合は空のスライスを返します。
func ParseCommaSeparatedList(s string) []string {
	// カンマで分割
	parts := strings.Split(s, ",")

	// 初期キャパシティを確保することで、appendによる再確保を最小限に抑える
	// 最悪の場合 (全ての要素が有効な場合) に一度の割り当てで済むようにする
	res := make([]string, 0, len(parts))

	// 各要素をトリミングし、空でない要素のみを追加
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			res = append(res, trimmed)
		}
	}
	return res
}
