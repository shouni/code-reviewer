package timeutil_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/shouni/code-reviewer/utils/timeutil"
)

// Helper: UTC時刻とJST時刻のオフセット (9時間) を定義
const jstOffset = 9 * time.Hour

// TestJSTLocation_Initialization は、JSTLocation() が一度だけ初期化され、
// 正しいロケーションオブジェクトを返すことを確認します。
func TestJSTLocation_Initialization(t *testing.T) {
	// 1. 初回の呼び出し
	loc1 := timeutil.JSTLocation()

	if loc1.String() != "Asia/Tokyo" && loc1.String() != "JST" {
		t.Errorf("初回ロードのロケーション名が期待値と異なります: Got %s", loc1.String())
	}

	// 2. 2回目の呼び出し（キャッシュが効いていることの確認）
	loc2 := timeutil.JSTLocation()

	// ポインタが同一であること（==）を確認し、sync.Once によるキャッシュが効いていることをテスト
	if loc1 != loc2 {
		t.Errorf("2回目の呼び出しで同じロケーションポインタが返されませんでした。キャッシュされていません。")
	}

	// JST (UTC+9) のオフセットが正しいことを確認
	_, offset := time.Now().In(loc1).Zone()
	if offset != int(jstOffset.Seconds()) {
		t.Errorf("JST ロケーションのオフセットが正しくありません。Expected %v, Got %d (seconds)", jstOffset, offset)
	}
}

// TestNowJST は、現在時刻が JST (UTC+9) のタイムゾーン情報を持って取得されていることを確認します。
func TestNowJST(t *testing.T) {
	nowJST := timeutil.NowJST()

	// タイムゾーンの名称を確認
	_, offset := nowJST.Zone()
	if nowJST.Location().String() != "Asia/Tokyo" && nowJST.Location().String() != "JST" {
		t.Errorf("NowJST のタイムゾーンが JST ではありません: %s", nowJST.Location().String())
	}

	// オフセットが +9時間であることを確認
	if offset != int(jstOffset.Seconds()) {
		t.Errorf("NowJST のオフセットが正しくありません。Expected %v, Got %d (seconds)", jstOffset, offset)
	}

	// 絶対時刻が現在時刻に近いことを確認 (実行速度を考慮し、1秒以内の誤差を許容)
	diff := time.Since(nowJST)
	if diff > time.Second || diff < -time.Second {
		t.Errorf("NowJST の時刻が現在の絶対時刻と大きくずれています。Diff: %v", diff)
	}
}

// TestToJST は、UTC時刻をJSTに正しく変換することを確認します。
func TestToJST(t *testing.T) {
	// 基準時刻: 2025年1月1日 00:00:00 UTC
	utcTime := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)
	expectedJST := time.Date(2025, time.January, 1, 9, 0, 0, 0, timeutil.JSTLocation()) // 9時間進む

	actualJST := timeutil.ToJST(utcTime)

	// 変換後の絶対時刻 (UTC) が等しいことを確認
	if !actualJST.Equal(expectedJST) {
		t.Errorf("ToJST の変換後の時刻が期待値と異なります。\nExpected: %v\nGot:      %v", expectedJST, actualJST)
	}

	// 変換後のタイムゾーンを確認
	if actualJST.Location().String() != expectedJST.Location().String() {
		t.Errorf("ToJST のロケーション名が期待値と異なります。\nExpected: %s\nGot:      %s", expectedJST.Location().String(), actualJST.Location().String())
	}
}

// TestFormatJSTAndFormatJSTString は、フォーマット機能とパース機能を確認します。
func TestFormatJSTAndFormatJSTString(t *testing.T) {
	// 基準時刻: 2025年1月1日 00:00:00 UTC
	utcTime := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)

	// FormatJST のテスト: 2025-01-01T09:00:00Z+09:00 -> "2025/01/01 09:00"
	layout := "2006/01/02 15:04"
	expectedFormattedJST := "2025/01/01 09:00"

	formatted := timeutil.FormatJST(utcTime, layout)
	if formatted != expectedFormattedJST {
		t.Errorf("FormatJST の結果が期待値と異なります。\nExpected: %s\nGot:      %s", expectedFormattedJST, formatted)
	}

	// FormatJSTString のテスト
	// 時刻文字列 "18:30" をJSTの時刻としてパースし、別の形式にフォーマット
	timeStr := "18:30"
	parseLayout := "15:04"
	formatLayout := "15時04分"

	// 期待値: JST の 18時30分
	expectedOutput := "18時30分"

	output, err := timeutil.FormatJSTString(timeStr, parseLayout, formatLayout)
	if err != nil {
		t.Fatalf("FormatJSTString がエラーを返しました: %v", err)
	}
	if output != expectedOutput {
		t.Errorf("FormatJSTString の結果が期待値と異なります。\nExpected: %s\nGot:      %s", expectedOutput, output)
	}

	// FormatJSTString のエラーテスト
	// 不正な時刻文字列をパース
	invalidTimeStr := "25:00"
	_, err = timeutil.FormatJSTString(invalidTimeStr, parseLayout, formatLayout)
	if err == nil {
		t.Error("FormatJSTString が不正な入力に対してエラーを返しませんでした")
	} else if fmt.Sprintf("%v", err) == "時刻文字列 '25:00' のパースに失敗しました: parsing time \"25:00\" as \"15:04\": cannot parse \"5:00\" as \"04\"" {
		t.Logf("FormatJSTString のエラー処理を確認しました: %v", err)
	}
}
