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
			expected: []*token{lParenAt(4)},
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
			expected: []*token{lParenAt(8)},
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
			expected: []*token{lParenAt(10)},
		},
		{
			name:        "unbalanced block comment",
			input:       "(; TODO",
			expectedErr: errors.New("expected block comment end ';)' at position 6"),
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
			expected: []*token{lParenAt(23)},
		},
		{
			name:        "unbalanced nested block comment",
			input:       "(; TODO (; (YOLO) ;)",
			expectedErr: errors.New("expected block comment end ';)' at position 20"),
		},
		{
			name:     "white space between parens",
			input:    "( )",
			expected: []*token{lParenAt(0), rParenAt(2)},
		},
	}

	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			var tokens []*token
			e := lex([]byte(tc.input), func(source []byte, tok tokenType, beginPos, endPos int) error {
				switch tok {
				case tokenLParen:
					tokens = append(tokens, lParenAt(beginPos))
					return nil
				case tokenRParen:
					tokens = append(tokens, rParenAt(beginPos))
					return nil
				}
				return fmt.Errorf("unsupported token: %s at position %d", tok, beginPos)
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

func lParenAt(pos int) *token {
	return newToken(tokenLParen, pos, "")
}

func rParenAt(pos int) *token {
	return newToken(tokenRParen, pos, "")
}

func newToken(tokenType tokenType, beginPos int, val string) *token {
	return &token{tokenType, beginPos, val}
}

type token struct {
	tokenType
	beginPos int
	value    string
}

// String is here to allow tests to be easier to troubleshoot
func (t *token) String() string {
	if t.value == "" {
		return fmt.Sprintf("%s at position %d", t.tokenType, t.beginPos)
	}
	return fmt.Sprintf("%s(%s) at position %d", t.tokenType, t.value, t.beginPos)
}
