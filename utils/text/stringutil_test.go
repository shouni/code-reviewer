package text_test

import (
	"reflect"
	"testing"

	"github.com/shouni/code-reviewer/utils/text"
)

func TestParseCommaSeparatedList(t *testing.T) {
	// テストケースを構造体スライスで定義
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "基本的なケース",
			input:    "apple, banana, cherry",
			expected: []string{"apple", "banana", "cherry"},
		},
		{
			name:     "要素間の余分な空白",
			input:    "  apple  ,banana,  cherry ",
			expected: []string{"apple", "banana", "cherry"},
		},
		{
			name:     "空文字列の入力",
			input:    "",
			expected: []string{},
		},
		{
			name:     "空白文字のみの入力",
			input:    "   ",
			expected: []string{},
		},
		{
			name:     "カンマの前後の空要素",
			input:    ",apple,,banana,",
			expected: []string{"apple", "banana"},
		},
		{
			name:     "要素がすべて空または空白",
			input:    " , ,  , ",
			expected: []string{},
		},
		{
			name:     "単一の要素",
			input:    "only one item",
			expected: []string{"only one item"},
		},
		{
			name:     "要素間の連続したカンマと空白",
			input:    "item1,, item2,  ,item3 ,",
			expected: []string{"item1", "item2", "item3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := text.ParseCommaSeparatedList(tt.input)

			// 結果のスライスが期待値と等しいか（順序も含む）をディープ比較
			if !reflect.DeepEqual(actual, tt.expected) {
				t.Errorf("ParseCommaSeparatedList(%q) = %v, 期待値 %v", tt.input, actual, tt.expected)
			}
		})
	}
}
