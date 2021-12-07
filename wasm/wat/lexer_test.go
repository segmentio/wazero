package wat

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLex(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    []*token
		expectedErr error
	}{
		{
			name:  "empty",
			input: "",
		},
		{
			name:     "only tokens",
			input:    "( )",
			expected: []*token{lParenAt(1, 1, 0), rParenAt(1, 3, 2)},
		},
		{
			name:  "only white space characters",
			input: " \t\r\n",
		},
		{
			name:     "after white space characters - EOL",
			input:    " \t\n(",
			expected: []*token{lParenAt(2, 1, 3)},
		},
		{
			name:     "after white space characters - Windows EOL",
			input:    " \t\r\n(",
			expected: []*token{lParenAt(2, 1, 4)},
		},
		{
			name:  "only line comment - EOL before EOF",
			input: ";; TODO\n",
		},
		{
			name:  "only line comment - EOF",
			input: ";; TODO",
		},
		{
			name:  "only unicode line comment - EOF",
			input: ";; брэд-ЛГТМ",
		},
		{
			name:     "after line comment",
			input:    ";; TODO\n(",
			expected: []*token{lParenAt(2, 1, 8)},
		},
		{
			name:     "after unicode line comment",
			input:    ";; брэд-ЛГТМ\n(",
			expected: []*token{lParenAt(2, 1, 21)},
		},
		{
			name:     "after line comment - Windows EOL",
			input:    ";; TODO\r\n(",
			expected: []*token{lParenAt(2, 1, 9)},
		},
		{
			name:     "after redundant line comment",
			input:    ";;;; TODO\n(",
			expected: []*token{lParenAt(2, 1, 10)},
		},
		{
			name:     "after line commenting out block comment",
			input:    ";; TODO (; ;)\n(",
			expected: []*token{lParenAt(2, 1, 14)},
		},
		{
			name:     "after line commenting out open block comment",
			input:    ";; TODO (;\n(",
			expected: []*token{lParenAt(2, 1, 11)},
		},
		{
			name:     "after line commenting out close block comment",
			input:    ";; TODO ;)\n(",
			expected: []*token{lParenAt(2, 1, 11)},
		},
		{
			name:        "half line comment",
			input:       "; TODO",
			expectedErr: errors.New("1:1 unexpected character ;"),
		},
		{
			name:  "only block comment - EOL before EOF",
			input: "(; TODO ;)\n",
		},
		{
			name:  "only block comment - Windows EOL before EOF",
			input: "(; TODO ;)\r\n",
		},
		{
			name:  "only block comment - EOF",
			input: "(; TODO ;)",
		},
		{
			name:     "after block comment",
			input:    "(; TODO ;)(",
			expected: []*token{lParenAt(1, 11, 10)},
		},
		{
			name:        "open block comment",
			input:       "(; TODO",
			expectedErr: errors.New("1:7 expected block comment end ';)'"),
		},
		{
			name:        "close block comment",
			input:       ";) TODO",
			expectedErr: errors.New("1:1 unexpected character ;"),
		},
		{
			name:  "only nested block comment - EOL before EOF",
			input: "(; TODO (; (YOLO) ;) ;)\n",
		},
		{
			name:  "only nested block comment - EOF",
			input: "(; TODO (; (YOLO) ;) ;)",
		},
		{
			name:  "only unicode block comment - EOF",
			input: "(; брэд-ЛГТМ ;)",
		},
		{
			name:     "after nested block comment",
			input:    "(; TODO (; (YOLO) ;) ;)(",
			expected: []*token{lParenAt(1, 24, 23)},
		},
		{
			name:     "after nested block comment - EOL",
			input:    "(; TODO (; (YOLO) ;) ;)\n (",
			expected: []*token{lParenAt(2, 2, 25)},
		},
		{
			name:     "after nested block comment - Windows EOL",
			input:    "(; TODO (; (YOLO) ;) ;)\r\n (",
			expected: []*token{lParenAt(2, 2, 26)},
		},
		{
			name:        "unbalanced nested block comment",
			input:       "(; TODO (; (YOLO) ;)",
			expectedErr: errors.New("1:20 expected block comment end ';)'"),
		},
		{
			name:     "white space between parens",
			input:    "( )",
			expected: []*token{lParenAt(1, 1, 0), rParenAt(1, 3, 2)},
		},
	}

	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			var tokens []*token
			e := lex([]byte(tc.input), func(source []byte, tok tokenType, line, col, pos, _ int) error {
				switch tok {
				case tokenLParen:
					tokens = append(tokens, lParenAt(line, col, pos))
					return nil
				case tokenRParen:
					tokens = append(tokens, rParenAt(line, col, pos))
					return nil
				}
				return fmt.Errorf("%d:%d unsupported token: %s at position %d", line, col, tok, pos)
			})
			if tc.expectedErr != nil {
				require.Equal(t, e, tc.expectedErr)
			} else {
				require.NoError(t, e)
				require.Equal(t, tc.expected, tokens)
			}
		})
	}
}

func BenchmarkLex(b *testing.B) {
	benchmarks := []struct {
		name string
		data []byte
	}{
		{"base", []byte("( )")}, // 3 bytes
		{"whitespace chars", []byte("(                        \n)\n")}, // 28 bytes
		{"unicode line comment", []byte("( ;; брэд-ЛГТМ   \n)\n")},     // 28 bytes
		{"unicode block comment", []byte("( (; брэд-ЛГТМ ;)\n)\n")},    // 28 bytes
	}
	var noopParseToken parseToken = func(source []byte, tok tokenType, beginLine, beginCol, beginPos, endPos int) error {
		return nil
	}
	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				err := lex(bm.data, noopParseToken)
				if err != nil {
					panic(err)
				}
			}
		})
	}
}

func lParenAt(line, col, pos int) *token {
	return &token{tokenLParen, line, col, pos, ""}
}

func rParenAt(line, col, pos int) *token {
	return &token{tokenRParen, line, col, pos, ""}
}

type token struct {
	tokenType
	line, col, pos int
	value          string
}

// String is here to allow tests to be easier to troubleshoot
func (t *token) String() string {
	if t.value == "" {
		return fmt.Sprintf("%d:%d %s at position %d", t.line, t.col, t.tokenType, t.pos)
	}
	return fmt.Sprintf("%d:%d %s(%s) at position %d", t.line, t.col, t.tokenType, t.value, t.pos)
}
