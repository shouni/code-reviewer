package ai

import (
	"context"
	"fmt"

	"github.com/shouni/code-reviewer/core/ports"
	"github.com/shouni/code-reviewer/gemini"
)

const (
	// コードレビューの一貫性を優先するため、低い温度に設定
	defaultGeminiTemperature = float32(0.1)
	// 一時的なネットワークエラーに対応するためのリトライ回数
	defaultGeminiMaxRetries = uint64(1)
)

// GeminiClient は gemini.Client の振る舞いを抽象化する内部インターフェース
type GeminiClient interface {
	GenerateContent(ctx context.Context, model string, prompt string) (*gemini.Response, error)
}

// GeminiAdapter は CodeReviewAI インターフェースを実装する具体的な構造体
type GeminiAdapter struct {
	client GeminiClient
}

// GeminiOptions はアダプター初期化のためのパラメータを保持
type GeminiOptions struct {
	ProjectID  string // GCPプロジェクトID (APIKey未指定時に必須)
	LocationID string // GCPロケーションID (未指定時は "global" にフォールバック)
	APIKey     string // Gemini APIキー (ProjectIDより優先される)
}

// NewGeminiAdapter は GeminiOptions を受け取り、クライアントを初期化します
func NewGeminiAdapter(ctx context.Context, opts GeminiOptions) (ports.CodeReviewAI, error) {
	cfg, err := buildGeminiConfig(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to build gemini config: %w", err)
	}

	gClient, err := gemini.NewClient(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize underlying gemini client: %w", err)
	}

	return &GeminiAdapter{
		client: gClient,
	}, nil
}

// buildGeminiConfig はオプションからgemini.Configを構築する責務を担います
func buildGeminiConfig(opts GeminiOptions) (gemini.Config, error) {
	temp := defaultGeminiTemperature
	cfg := gemini.Config{
		Temperature: &temp,
		MaxRetries:  defaultGeminiMaxRetries,
	}

	if opts.APIKey != "" {
		cfg.APIKey = opts.APIKey
	} else if opts.ProjectID != "" {
		cfg.ProjectID = opts.ProjectID
		if opts.LocationID != "" {
			cfg.LocationID = opts.LocationID
		} else {
			cfg.LocationID = "global"
		}
	} else {
		return cfg, fmt.Errorf("APIKey or ProjectID must be set in GeminiOptions")
	}

	return cfg, nil
}

// ReviewCodeDiff は AI モデルに対してコードレビューの生成を依頼します
func (ga *GeminiAdapter) ReviewCodeDiff(ctx context.Context, modelName, finalPrompt string) (string, error) {
	resp, err := ga.client.GenerateContent(ctx, modelName, finalPrompt)
	if err != nil {
		return "", fmt.Errorf("Gemini API call failed (Model: %s): %w", modelName, err)
	}

	return resp.Text, nil
}
