package publisher

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/shouni/code-reviewer/core/ports"
)

// --- Mocks ---

type MockMarkdownToHtmlRunner struct {
	mock.Mock
}

// ports.MarkdownToHtmlRunner の最新シグネチャ: Run(markdownContent []byte) (io.Reader, error)
func (m *MockMarkdownToHtmlRunner) Run(markdown []byte) (io.Reader, error) {
	args := m.Called(markdown)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(io.Reader), args.Error(1)
}

type MockOutputWriter struct {
	mock.Mock
}

func (m *MockOutputWriter) Write(ctx context.Context, uri string, r io.Reader, contentType string) error {
	return m.Called(ctx, uri, r, contentType).Error(0)
}

// インターフェースを満たすかチェック
var _ ports.Publisher = (*Publisher)(nil)

func TestNewPublisher(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockWriter := new(MockOutputWriter)
		mockMD := new(MockMarkdownToHtmlRunner)

		p, err := NewPublisher(mockWriter, mockMD)
		assert.NoError(t, err)
		assert.NotNil(t, p)
	})

	t.Run("Failure_WriterIsNil", func(t *testing.T) {
		mockMD := new(MockMarkdownToHtmlRunner)
		p, err := NewPublisher(nil, mockMD)
		assert.Error(t, err)
		assert.Nil(t, p)
		assert.Contains(t, err.Error(), "writer is nil")
	})

	t.Run("Failure_MarkdownRunnerIsNil", func(t *testing.T) {
		mockWriter := new(MockOutputWriter)
		p, err := NewPublisher(mockWriter, nil)
		assert.Error(t, err)
		assert.Nil(t, p)
		assert.Contains(t, err.Error(), "MarkdownToHtmlRunner is nil")
	})
}

func TestPublisher_Publish(t *testing.T) {
	ctx := context.Background()
	uri := "gs://review-bucket/report.html"
	data := ports.ReviewData{
		RepoURL:        "https://github.com/shouni/test",
		BaseBranch:     "main",
		FeatureBranch:  "feat/fix",
		ReviewMarkdown: "## Good Job",
	}

	t.Run("Success", func(t *testing.T) {
		mockWriter := new(MockOutputWriter)
		mockMD := new(MockMarkdownToHtmlRunner)

		// Run は []byte を受け取り io.Reader を返す
		mockMD.On("Run", mock.Anything).Return(strings.NewReader("<html></html>"), nil)
		mockWriter.On("Write", ctx, uri, mock.Anything, contentTypeHTML).Return(nil)

		p, _ := NewPublisher(mockWriter, mockMD)
		err := p.Publish(ctx, uri, data)
		assert.NoError(t, err)
	})

	t.Run("Error_MarkdownConversionFails", func(t *testing.T) {
		mockMD := new(MockMarkdownToHtmlRunner)
		mockMD.On("Run", mock.Anything).Return(nil, errors.New("markdown error"))

		p := &Publisher{md: mockMD}

		err := p.Publish(ctx, uri, data)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "HTML変換に失敗しました")
	})

	t.Run("Error_RemoteWriteFails", func(t *testing.T) {
		mockWriter := new(MockOutputWriter)
		mockMD := new(MockMarkdownToHtmlRunner)

		mockMD.On("Run", mock.Anything).Return(strings.NewReader("<html></html>"), nil)
		mockWriter.On("Write", ctx, uri, mock.Anything, contentTypeHTML).Return(errors.New("upload failed"))

		p := &Publisher{
			writer: mockWriter,
			md:     mockMD,
		}

		err := p.Publish(ctx, uri, data)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "リモートストレージへの書き込みに失敗しました")
	})
}

func TestPublisher_ConvertMarkdownToHTML_Format(t *testing.T) {
	mockMD := new(MockMarkdownToHtmlRunner)
	p := &Publisher{md: mockMD}

	data := ports.ReviewData{
		RepoURL:        "https://github.com/shouni/repo",
		BaseBranch:     "base",
		FeatureBranch:  "feat",
		ReviewMarkdown: "AI_REVIEW_BODY",
	}

	// 引数の検証: 期待する内容が含まれているか（title 引数は排除されている）
	mockMD.On("Run", mock.MatchedBy(func(b []byte) bool {
		s := string(b)
		return strings.Contains(s, "https://github.com/shouni/repo") &&
			strings.Contains(s, "AI_REVIEW_BODY")
	})).Return(strings.NewReader("html"), nil)

	_, err := p.convertMarkdownToHTML(data)
	assert.NoError(t, err)
	mockMD.AssertExpectations(t)
}
