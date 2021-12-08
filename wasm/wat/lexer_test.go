package wat

import (
	"errors"
	"fmt"
	"testing"
	"unicode/utf8"

	"github.com/bytecodealliance/wasmtime-go"
	"github.com/stretchr/testify/require"
)

// exampleWat was at one time in the wasmtime-go README
const exampleWat = `
      (module
        (import "" "hello" (func $hello))
        (func (export "run")
          (call $hello))
      )
    `

func TestLex_Example(t *testing.T) {
	tokens, e := lexTokens(exampleWat)
	require.NoError(t, e)
	require.Equal(t, []*token{
		lParenAt(2, 7, 7),
		keywordAt(2, 8, 8, "module"),
		lParenAt(3, 9, 23),
		keywordAt(3, 10, 24, "import"),
		stringAt(3, 17, 31, `""`),
		stringAt(3, 20, 34, `"hello"`),
		lParenAt(3, 28, 42),
		keywordAt(3, 29, 43, "func"),
		reservedAt(3, 34, 48, "$hello"),
		rParenAt(3, 40, 54),
		rParenAt(3, 41, 55),
		lParenAt(4, 9, 65),
		keywordAt(4, 10, 66, "func"),
		lParenAt(4, 15, 71),
		keywordAt(4, 16, 72, "export"),
		stringAt(4, 23, 79, `"run"`),
		rParenAt(4, 28, 84),
		lParenAt(5, 11, 96),
		keywordAt(5, 12, 97, "call"),
		reservedAt(5, 17, 102, "$hello"),
		rParenAt(5, 23, 108),
		rParenAt(5, 24, 109),
		rParenAt(6, 7, 117),
	}, tokens)
}

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
			name:     "only parens",
			input:    "()",
			expected: []*token{lParenAt(1, 1, 0), rParenAt(1, 2, 1)},
		},
		{
			name:     "shortest keywords",
			input:    "a z",
			expected: []*token{keywordAt(1, 1, 0, "a"), keywordAt(1, 3, 2, "z")},
		},
		{
			name:     "shortest tokens - EOL",
			input:    "(a)\n",
			expected: []*token{lParenAt(1, 1, 0), keywordAt(1, 2, 1, "a"), rParenAt(1, 3, 2)},
		},
		{
			name:     "only tokens",
			input:    "(module)",
			expected: []*token{lParenAt(1, 1, 0), keywordAt(1, 2, 1, "module"), rParenAt(1, 8, 7)},
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
		{
			name:     "empty string",
			input:    `""`,
			expected: []*token{stringAt(1, 1, 0, `""`)},
		},
		{
			name:     "string inside tokens with newline",
			input:    `("\n")`,
			expected: []*token{lParenAt(1, 1, 0), stringAt(1, 2, 1, `"\n"`), rParenAt(1, 6, 5)},
		},
	}

	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			tokens, e := lexTokens(tc.input)
			if tc.expectedErr != nil {
				require.Equal(t, e, tc.expectedErr)
			} else {
				require.NoError(t, e)
				require.Equal(t, tc.expected, tokens)
			}
		})
	}
}

func lexTokens(input string) ([]*token, error) {
	var tokens []*token
	e := lex([]byte(input), func(source []byte, tok tokenType, line, col, beginPos, endPos int) (err error) {
		switch tok {
		case tokenLParen:
			tokens = append(tokens, lParenAt(line, col, beginPos))
		case tokenRParen:
			tokens = append(tokens, rParenAt(line, col, beginPos))
		case tokenKeyword:
			tokens = append(tokens, keywordAt(line, col, beginPos, string(source[beginPos:endPos])))
		case tokenReserved:
			tokens = append(tokens, reservedAt(line, col, beginPos, string(source[beginPos:endPos])))
		case tokenString:
			tokens = append(tokens, stringAt(line, col, beginPos, string(source[beginPos:endPos])))
		default:
			err = fmt.Errorf("%d:%d unsupported token: %s at position %d:%d", line, col, tok, beginPos, endPos)
		}
		return
	})
	return tokens, e
}

func BenchmarkLex(b *testing.B) {
	benchmarks := []struct {
		name string
		data []byte
	}{
		{"example", []byte(exampleWat)},
		{"whitespace chars", []byte("(                        \nmodule)\n")}, // 34 bytes
		{"unicode line comment", []byte("( ;; брэд-ЛГТМ   \nmodule)\n")},     // 28 bytes
		{"unicode block comment", []byte("( (; брэд-ЛГТМ ;)\nmodule)\n")},    // 28 bytes
	}
	var noopParseToken parseToken = func(source []byte, tok tokenType, beginLine, beginCol, beginPos, endPos int) error {
		return nil
	}
	for _, bm := range benchmarks {
		b.Run(bm.name+" vs utf8.ValidString", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				utf8.ValidString(string(bm.data))
			}
		})
		// Not a fair comparison as we are only lexing and not writing back %.wasm
		// If possible, we should find a way to isolate only the lexing C functions.
		b.Run(bm.name+" vs wasmtime.Wat2Wasm", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, err := wasmtime.Wat2Wasm(string(bm.data))
				if err != nil {
					panic(err)
				}
			}
		})
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

func keywordAt(line, col, pos int, value string) *token {
	return &token{tokenKeyword, line, col, pos, value}
}

func reservedAt(line, col, pos int, value string) *token {
	return &token{tokenReserved, line, col, pos, value}
}

func stringAt(line, col, pos int, value string) *token {
	return &token{tokenString, line, col, pos, value}
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
