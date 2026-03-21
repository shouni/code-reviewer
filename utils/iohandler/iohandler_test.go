package iohandler

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestReadInput(t *testing.T) {
	content := []byte("hello world")

	// 1. ファイルからの読み込みテスト
	t.Run("Read from file", func(t *testing.T) {
		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "test.txt")
		if err := os.WriteFile(tmpFile, content, 0644); err != nil {
			t.Fatal(err)
		}

		got, err := ReadInput(tmpFile)
		if err != nil {
			t.Fatalf("ReadInput() error = %v", err)
		}
		if !bytes.Equal(got, content) {
			t.Errorf("got %q, want %q", got, content)
		}
	})

	// 2. 標準入力からの読み込みテスト
	t.Run("Read from stdin", func(t *testing.T) {
		// Stdinを一時的に差し替え
		oldStdin := os.Stdin
		defer func() { os.Stdin = oldStdin }()

		r, w, _ := os.Pipe()
		os.Stdin = r

		// 書き込みを行ってからクローズしないと、io.ReadAll が終わらない
		go func() {
			w.Write(content)
			w.Close()
		}()

		got, err := ReadInput("")
		if err != nil {
			t.Fatalf("ReadInput(\"\") error = %v", err)
		}
		if !bytes.Equal(got, content) {
			t.Errorf("got %q, want %q", got, content)
		}
	})
}

func TestWriteOutput(t *testing.T) {
	content := []byte("output data")

	// 1. ファイルへの書き出しテスト
	t.Run("Write to file", func(t *testing.T) {
		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "out.txt")

		err := WriteOutput(tmpFile, content)
		if err != nil {
			t.Fatalf("WriteOutput() error = %v", err)
		}

		got, _ := os.ReadFile(tmpFile)
		if !bytes.Equal(got, content) {
			t.Errorf("got %q, want %q", got, content)
		}
	})

	// 2. 標準出力への書き出しテスト
	t.Run("Write to stdout", func(t *testing.T) {
		// Stdoutを一時的に差し替え
		oldStdout := os.Stdout
		defer func() { os.Stdout = oldStdout }()

		r, w, _ := os.Pipe()
		os.Stdout = w

		err := WriteOutput("", content)
		w.Close() // 読み取り前にクローズ

		if err != nil {
			t.Fatalf("WriteOutput(\"\") error = %v", err)
		}

		var buf bytes.Buffer
		io.Copy(&buf, r)
		if buf.String() != string(content) {
			t.Errorf("got %q, want %q", buf.String(), string(content))
		}
	})
}

func TestStringWrappers(t *testing.T) {
	// ReadInputString と WriteOutputString の簡易テスト
	tmpFile := filepath.Join(t.TempDir(), "string_test.txt")
	want := "test string"

	err := WriteOutputString(tmpFile, want)
	if err != nil {
		t.Fatal(err)
	}

	got, err := ReadInputString(tmpFile)
	if err != nil {
		t.Fatal(err)
	}

	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
