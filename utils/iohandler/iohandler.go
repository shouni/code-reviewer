// Package iohandler は、ファイルまたは標準入力からのコンテンツの読み込みと、
// ファイルまたは標準出力へのコンテンツの書き出しを行うユーティリティ関数を提供します。
package iohandler

import (
	"fmt"
	"io"
	"os"
)

// ReadInput は、指定されたファイル、またはファイル名が空の場合は標準入力から
// コンテンツを読み込み、バイトスライス ([]byte) で返します。
// 注意: 現在、大きなファイルをサポートしていません。コンテンツ全体がメモリに読み込まれます。
func ReadInput(filename string) ([]byte, error) {
	if filename != "" {
		return os.ReadFile(filename)
	}

	content, err := io.ReadAll(os.Stdin)
	if err != nil {
		// エラーメッセージも日本語に修正します
		return nil, fmt.Errorf("標準入力からの読み込みに失敗しました: %w", err)
	}
	return content, nil
}

// ReadInputString は、指定されたファイル、または標準入力からコンテンツを読み込み、
// 文字列 (string) で返します。
// 内部で ReadInput を呼び出し、結果を string にキャストします。
// 注意: 大きなコンテンツの場合、バイトスライスから文字列への変換により、
// 追加のメモリ割り当てとコピーが発生する可能性があります。
// パフォーマンスが重要な場合は、ReadInput 関数を直接使用し、[]byte で処理することを検討してください。
func ReadInputString(filename string) (string, error) {
	content, err := ReadInput(filename)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// WriteOutput は、コンテンツ (バイトスライス) をファイル、または標準出力に出力します。
// 注意: 現在、大きなファイルをサポートしていません。コンテンツ全体がメモリから書き込まれます。
func WriteOutput(filename string, content []byte) error {
	if filename != "" {
		return os.WriteFile(filename, content, 0644)
	}

	// os.Stdout.Write を使用し、標準出力にバイトスライスを直接書き込みます。
	_, err := os.Stdout.Write(content)
	if err != nil {
		// エラーメッセージも日本語に修正します
		return fmt.Errorf("標準出力への書き込みに失敗しました: %w", err)
	}
	return nil
}

// WriteOutputString は、コンテンツ (文字列) をファイル、または標準出力に出力します。
// 内部で content を []byte にキャストし、WriteOutput を呼び出します。
// 注意: 大きなコンテンツの場合、文字列からバイトスライスへの変換により、
// 追加のメモリ割り当てとコピーが発生する可能性があります。
// パフォーマンスが重要な場合は、WriteOutput 関数を直接使用し、[]byte で処理することを検討してください。
func WriteOutputString(filename string, content string) error {
	return WriteOutput(filename, []byte(content))
}
