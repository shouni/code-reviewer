package envutil

import (
	"os"
	"strconv"
)

// GetEnv は環境変数を取得し、存在しない場合はデフォルト値を返します。
func GetEnv(key string, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// GetEnvAsBool は環境変数からbool値を読み込みます。
// 環境変数が未設定、またはブール値として解釈できない場合はdefaultValueを返します。
func GetEnvAsBool(key string, defaultValue bool) bool {
	// os.LookupEnv を使用し、値が存在するかどうかを明確にチェック
	val, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue // 環境変数が設定されていない場合
	}

	b, err := strconv.ParseBool(val)
	if err != nil {
		return defaultValue
	}

	return b
}

// GetEnvAsInt は環境変数を整数として取得し、存在しないか変換に失敗した場合はデフォルト値を返します。
func GetEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		// 文字列から整数（10進数）への変換を試みる
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
		// 変換に失敗した場合はデフォルト値を返す
	}
	return defaultValue
}
