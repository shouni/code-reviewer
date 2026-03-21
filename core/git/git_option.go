package git

// GitConfigurable は、実行時に Git の動作を構成するためのメソッドを定義します。
type GitConfigurable interface {
	SetInsecureSkipHostKeyCheck(bool)
	SetBaseBranch(string)
}

// Option は GitConfigurable に対する設定関数です。
type Option func(GitConfigurable)

// WithInsecureSkipHostKeyCheck は、ホストキーチェックをスキップするかどうかを構成します。
func WithInsecureSkipHostKeyCheck(skip bool) Option {
	return func(c GitConfigurable) {
		c.SetInsecureSkipHostKeyCheck(skip)
	}
}

// WithBaseBranch は、Git 操作のベース ブランチを設定するオプション関数を提供します。
func WithBaseBranch(branch string) Option {
	return func(c GitConfigurable) {
		c.SetBaseBranch(branch)
	}
}
