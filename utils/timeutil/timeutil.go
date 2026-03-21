package timeutil

import (
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// JST の Location のキャッシュと、それを保護するための Mutex
var (
	jstLocationCache *time.Location
	jstOnce          sync.Once
)

const jstLocationName = "Asia/Tokyo"

// JSTLocation は、"Asia/Tokyo" の time.Location を一度だけロードし、そのポインタを返します。
// これにより、Location のロード失敗を防ぎ、また呼び出しごとのファイルシステムへのアクセスを避けます。
func JSTLocation() *time.Location {
	// sync.Once を使用して、初期化処理をスレッドセーフかつ一度だけ実行することを保証
	jstOnce.Do(func() {
		loc, err := time.LoadLocation(jstLocationName)
		if err != nil {
			// 警告を slog.Warn で構造化出力
			slog.Warn(
				"Failed to load location, falling back to FixedZone.",
				slog.String("location", jstLocationName),
				slog.String("fallback", "FixedZone (UTC+9)"),
				slog.Any("error", err),
			)
			// FixedZone("JST", 9 * 60 * 60) は JST (UTC+9) を表す。
			loc = time.FixedZone("JST", 9*60*60)
		}
		jstLocationCache = loc
	})
	return jstLocationCache
}

// NowJST は、日本標準時 (JST) における現在の時刻を返します。
// 例: 2025-11-23 15:00:00 +0900 JST
func NowJST() time.Time {
	return time.Now().In(JSTLocation())
}

// ToJST は、引数として渡された time.Time を JST に変換します。
func ToJST(t time.Time) time.Time {
	return t.In(JSTLocation())
}

// FormatJST は、与えられた time.Time オブジェクトを JST に変換した後、指定されたレイアウトでフォーマットします。
func FormatJST(t time.Time, layout string) string {
	return ToJST(t).Format(layout)
}

// FormatJSTString は、与えられた時刻文字列をJSTの time.Time にパースし、指定されたレイアウトでフォーマットします。
// パースに失敗した場合は、空文字列とエラーを返します。
func FormatJSTString(timeStr, parseLayout, formatLayout string) (string, error) {
	// タイムゾーン情報を含まない時刻文字列をJSTとして解釈させるため。
	t, err := time.ParseInLocation(parseLayout, timeStr, JSTLocation())
	if err != nil {
		return "", fmt.Errorf("時刻文字列 '%s' のパースに失敗しました: %w", timeStr, err)
	}
	return FormatJST(t, formatLayout), nil
}
