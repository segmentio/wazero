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
			name:  "only white space characters",
			input: " \t\r\n",
		},
		{
			name:     "after white space characters",
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
			name:     "after line comment",
			input:    ";; TODO\n(",
			expected: []*token{lParenAt(2, 1, 8)},
		},
		{
			name:  "only block comment - EOL before EOF",
			input: "(; TODO ;)\n",
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
			name:        "unbalanced block comment",
			input:       "(; TODO",
			expectedErr: errors.New("1:7 expected block comment end ';)'"),
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
			name:     "after nested block comment",
			input:    "(; TODO (; (YOLO) ;) ;)(",
			expected: []*token{lParenAt(1, 24, 23)},
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
