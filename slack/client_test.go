package slack

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/slack-go/slack"
)

type mockRequester struct {
	postJSONFunc func(ctx context.Context, url string, data any) ([]byte, error)
}

// PostJSONAndFetchBytes はテストで主に使用するメソッドです。
func (m *mockRequester) PostJSONAndFetchBytes(ctx context.Context, url string, data any) ([]byte, error) {
	if m.postJSONFunc == nil {
		return nil, nil
	}
	return m.postJSONFunc(ctx, url, data)
}

// Requester インターフェースを満たすための他のメソッド定義。
func (m *mockRequester) DoRequest(req *http.Request) ([]byte, error)                { return nil, nil }
func (m *mockRequester) FetchBytes(ctx context.Context, url string) ([]byte, error) { return nil, nil }
func (m *mockRequester) FetchAndDecodeJSON(ctx context.Context, url string, v any) error {
	return nil
}
func (m *mockRequester) PostRawBodyAndFetchBytes(ctx context.Context, url string, body []byte, contentType string) ([]byte, error) {
	return nil, nil
}

// --- テスト本体 ---

// TestClient_SendTextWithHeader は、ヘッダーとメッセージを明示的に指定した場合の送信テストです。
func TestClient_SendTextWithHeader(t *testing.T) {
	ctx := context.Background()
	webhookURL := "https://hooks.slack.com/services/test"

	tests := []struct {
		name      string
		header    string
		message   string
		setupMock func(m *mockRequester)
		wantErr   bool
	}{
		{
			name:    "正常系: 正しいパラメータで Webhook が呼ばれる",
			header:  "Test Header",
			message: "Test Message",
			setupMock: func(m *mockRequester) {
				m.postJSONFunc = func(ctx context.Context, url string, data any) ([]byte, error) {
					// URLの検証
					if url != webhookURL {
						return nil, errors.New("invalid url")
					}
					// 送信データの型と中身の検証
					msg, ok := data.(slack.WebhookMessage)
					if !ok {
						return nil, errors.New("data is not WebhookMessage")
					}
					// フォールバック用テキスト（Text）と、Block Kitのヘッダーが正しいか
					if msg.Text != "Test Header" {
						return nil, errors.New("unexpected fallback text")
					}
					return []byte("ok"), nil
				}
			},
			wantErr: false,
		},
		{
			name:    "異常系: HTTPリクエストに失敗した場合エラーを返す",
			header:  "Error Test",
			message: "Panic",
			setupMock: func(m *mockRequester) {
				m.postJSONFunc = func(ctx context.Context, url string, data any) ([]byte, error) {
					return nil, errors.New("network error")
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mReq := &mockRequester{}
			tt.setupMock(mReq)

			// NewClient のエラーチェック
			client, err := NewClient(mReq, webhookURL)
			if err != nil {
				t.Fatalf("failed to create client: %v", err)
			}

			err = client.SendTextWithHeader(ctx, tt.header, tt.message)

			if (err != nil) != tt.wantErr {
				t.Errorf("SendTextWithHeader() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestClient_SendText_HeaderGeneration は、本文から自動生成されるヘッダーの加工ロジックをテストします。
func TestClient_SendText_HeaderGeneration(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name           string
		message        string
		expectedHeader string
	}{
		{
			name:           "1行目がヘッダーとして採用される",
			message:        "第一行目\n第二行目",
			expectedHeader: "📢 第一行目",
		},
		{
			name: "長い行は切り捨てられる",
			// 50文字を超えるメッセージ
			message: "これは非常に長いタイトルなので50文字くらいでカットされるはずです。カットされてるかな？どこまで続くかな？",
			// 実装の text.Truncate(trimmedLine, 50, "...") に合わせた期待値。
			// 「📢 」を除いた本文先頭から正確に50文字（Rune数）で切り捨てられます。
			// 「どこまで続く」の「く」までが50文字目となるため、その後にサフィックスが付与されます。
			expectedHeader: "📢 これは非常に長いタイトルなので50文字くらいでカットされるはずです。カットされてるかな？どこまで続く...",
		},
		{
			name:           "空メッセージの場合はデフォルト",
			message:        "",
			expectedHeader: "📢 通知メッセージ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mReq := &mockRequester{}
			var capturedHeader string
			mReq.postJSONFunc = func(ctx context.Context, url string, data any) ([]byte, error) {
				msg, ok := data.(slack.WebhookMessage)
				if ok {
					capturedHeader = msg.Text
				}
				return []byte("ok"), nil
			}

			client, err := NewClient(mReq, "http://dummy")
			if err != nil {
				t.Fatalf("failed to create client: %v", err)
			}

			err = client.SendText(ctx, tt.message)
			if err != nil {
				t.Fatalf("failed to send text: %v", err)
			}

			if capturedHeader != tt.expectedHeader {
				t.Errorf("Header generation failed.\ngot  = %v\nwant = %v", capturedHeader, tt.expectedHeader)
			}
		})
	}
}
