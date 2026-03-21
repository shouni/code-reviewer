package text

import (
	"strings"

	"github.com/forPelevin/gomoji"
)

// NormalizeText は、連続する空白文字（改行やタブを含む）を単一のスペースに変換し、
// 前後の空白を削除します。内部で strings.Fields を使用しています。
func NormalizeText(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// RemoveEmojis は、入力文字列からすべての絵文字文字を削除します。
func RemoveEmojis(s string) string {
	return gomoji.RemoveEmojis(s)
}

// CleanStringFromEmojis は、入力文字列からすべての絵文字文字を削除し、空白を正規化します。
// この関数は、最初に絵文字を削除し、次に NormalizeText を使用して空白を正規化します。
func CleanStringFromEmojis(s string) string {
	s = RemoveEmojis(s)
	s = NormalizeText(s)
	return s
}

// Truncate は、指定された文字列を rune (文字) の数で最大長まで切り詰め、必要に応じてサフィックスを追加します。
//
// 注意: 切り詰められた文字列の末尾に空白文字が残った場合、サフィックスを付加する前に
// strings.TrimSpaceを使用してその末尾の空白は無条件に削除されます。
// これは、ログ出力や表示用途で不要な末尾スペースを排除し、サフィックスを綺麗に付加するための仕様です
func Truncate(s string, maxLen int, suffix string) string {
	if maxLen <= 0 {
		return ""
	}

	// 1. 文字列を rune のスライスに変換 (マルチバイト対応)
	runes := []rune(s)

	// 2. 文字数が最大長以下であればそのまま返す
	if len(runes) <= maxLen {
		return s
	}

	// 3. rune の数で切り詰める
	truncatedRuneSlice := runes[:maxLen]

	// 4. rune スライスを文字列に戻し、末尾スペースを削除
	truncatedString := strings.TrimSpace(string(truncatedRuneSlice))

	return truncatedString + suffix
}
