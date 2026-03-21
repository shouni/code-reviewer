package ai

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/shouni/code-reviewer/gemini"
)

// mockGeminiClient は GeminiClient インターフェースのモック
type mockGeminiClient struct {
	generateContentFunc func(ctx context.Context, model string, prompt string) (*gemini.Response, error)
}

func (m *mockGeminiClient) GenerateContent(ctx context.Context, model string, prompt string) (*gemini.Response, error) {
	if m.generateContentFunc != nil {
		return m.generateContentFunc(ctx, model, prompt)
	}
	return &gemini.Response{Text: "mock response"}, nil
}

// ユニットテスト用にモックを注入するヘルパーコンストラクタ
func newTestGeminiAdapter(client GeminiClient) *GeminiAdapter {
	return &GeminiAdapter{
		client: client,
	}
}

// NewGeminiAdapter の正常系テスト
func TestNewGeminiAdapter_Success(t *testing.T) {
	opts := GeminiOptions{
		APIKey: "dummy-api-key",
	}
	adapter, err := NewGeminiAdapter(context.Background(), opts)
	assert.NoError(t, err)
	assert.NotNil(t, adapter)
}

// NewGeminiAdapter 自体のエラーハンドリングをテスト
func TestNewGeminiAdapter_Error(t *testing.T) {
	opts := GeminiOptions{} // すべて未設定
	adapter, err := NewGeminiAdapter(context.Background(), opts)

	assert.Error(t, err)
	assert.Nil(t, adapter)
	assert.Contains(t, err.Error(), "failed to build gemini config")
}

func TestNewGeminiAdapter_Config(t *testing.T) {
	t.Run("BuildGeminiConfig_Success", func(t *testing.T) {
		opts := GeminiOptions{APIKey: "key"}
		cfg, err := buildGeminiConfig(opts)
		assert.NoError(t, err)
		assert.Equal(t, "key", cfg.APIKey)
	})

	t.Run("BuildGeminiConfig_Failure", func(t *testing.T) {
		opts := GeminiOptions{}
		_, err := buildGeminiConfig(opts)
		assert.Error(t, err)
	})
}

func TestGeminiAdapter_ReviewCodeDiff(t *testing.T) {
	ctx := context.Background()
	modelName := "gemini-1.5-pro"

	t.Run("Success", func(t *testing.T) {
		expectedResponse := "Review result text"
		mock := &mockGeminiClient{
			generateContentFunc: func(ctx context.Context, model string, prompt string) (*gemini.Response, error) {
				return &gemini.Response{Text: expectedResponse}, nil
			},
		}

		adapter := newTestGeminiAdapter(mock)

		result, err := adapter.ReviewCodeDiff(ctx, modelName, "prompt")
		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, result)
	})

	t.Run("APIError", func(t *testing.T) {
		mock := &mockGeminiClient{
			generateContentFunc: func(ctx context.Context, model string, prompt string) (*gemini.Response, error) {
				return nil, errors.New("api error")
			},
		}

		adapter := newTestGeminiAdapter(mock)

		_, err := adapter.ReviewCodeDiff(ctx, modelName, "prompt")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Gemini API call failed")
	})
}

func TestGeminiAdapter_Constants(t *testing.T) {
	assert.Equal(t, float32(0.1), defaultGeminiTemperature)
	assert.Equal(t, uint64(1), defaultGeminiMaxRetries)
}
