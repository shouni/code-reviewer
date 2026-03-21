package clibase

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Config は共通フラグの値を保持する内部構造体です。
type Config struct {
	Verbose    bool
	ConfigFile string
}

// globalConfig はパッケージ内でのみ変更可能な設定情報の格納先です。
var globalConfig Config

// GetConfig は現在の設定情報のコピーを返します。
// これにより、利用側は読み取り専用として安全に設定を参照できます。
func GetConfig() Config {
	return globalConfig
}

// App は CLI アプリケーションの構成を定義します。
type App struct {
	Name     string
	AddFlags func(cmd *cobra.Command)                      // 独自フラグ登録用
	PreRunE  func(cmd *cobra.Command, args []string) error // 実行前バリデーション/初期化用
	PostRun  func(cmd *cobra.Command, args []string)       // 実行後のリソース解放用
	Commands []*cobra.Command                              // サブコマンド群
}

// Execute は、アプリケーションの構築と実行をワンストップで行います。
func Execute(app App) {
	rootCmd := &cobra.Command{
		Use:   app.Name,
		Short: fmt.Sprintf("%s CLI tool", app.Name),
		Long:  fmt.Sprintf("%s is a CLI application built with shouni/clibase.", app.Name),

		// アプリケーション固有の実行前処理を呼び出す
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// アプリ固有の PreRunE 処理を実行
			if app.PreRunE != nil {
				return app.PreRunE(cmd, args)
			}
			return nil
		},

		// コマンド実行後に必ず呼び出されるクリーンアップ処理
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			if app.PostRun != nil {
				app.PostRun(cmd, args)
			}
		},

		// 引数なしで実行された場合にヘルプを表示
		// RunE にすることで、エラーハンドリングを上位の Execute() に委ねます
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	// 共通フラグの登録
	rootCmd.PersistentFlags().BoolVarP(&globalConfig.Verbose, "verbose", "V", false, "Enable verbose output")
	rootCmd.PersistentFlags().StringVarP(&globalConfig.ConfigFile, "config", "C", "", "Config file path")

	// アプリ固有フラグの登録
	if app.AddFlags != nil {
		app.AddFlags(rootCmd)
	}

	// サブコマンドの追加
	if len(app.Commands) > 0 {
		rootCmd.AddCommand(app.Commands...)
	}

	// 実行
	if err := rootCmd.Execute(); err != nil {
		// Cobraがエラーを出力するため、ここでは適切な終了コードで終了します
		os.Exit(1)
	}
}
