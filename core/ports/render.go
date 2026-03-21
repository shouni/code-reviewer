package ports

// PromptBuilder は、プロンプト文字列を生成する責務を定義します。
// data にはテンプレート内で評価可能な構造体またはマップを渡す必要があります。
type PromptBuilder interface {
	Build(mode string, data any) (string, error)
}
