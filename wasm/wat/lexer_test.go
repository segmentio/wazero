package wat

import (
	"github.com/stretchr/testify/require"
	"io"
	"testing"
)

func TestToken_String(t *testing.T) {
	tests := []struct {
		name     string
		input    *token
		expected string
	}{
		{
			name:     "no value",
			input:    newToken(tokenEOF, ""),
			expected: "EOF",
		},
		{
			name:     "has value",
			input:    newToken(tokenUN, "123"),
			expected: "uN(123)",
		},
	}

	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, tc.input.String())
		})
	}
}

func TestNextToken(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    *token
		expectedErr string
	}{
		{
			name:     "empty",
			input:    "",
			expected: newToken(tokenEOF, ""),
		},
		{
			name:     "only white space",
			input:    " \t \r\n;;foo  \n(; oh (; (boy) \n;)\r\n;) \t",
			expected: newToken(tokenEOF, ""),
		},
		{
			name:     "(",
			input:    "(",
			expected: newToken(tokenLParen, ""),
		},
		{
			name:     ")",
			input:    ")",
			expected: newToken(tokenRParen, ""),
		},
	}

	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			l := New(tc.input)
			tok, e := l.NextToken()
			if e != nil {
				require.Equal(t, tc.expectedErr, e)
				require.Equal(t, nil, tok)
			} else {
				require.NoError(t, e)
				require.Equal(t, tc.expected, tok)
			}
		})
	}
}

func TestNextNonWhiteSpace(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    byte
		expectedErr error
	}{
		{
			name:        "empty",
			input:       "",
			expectedErr: io.EOF,
		},
		{
			name:        "only white space",
			input:       " \t \r\n;;foo  \n(; oh (; (boy) \n;)\r\n;) \t",
			expectedErr: io.EOF,
		},
		{
			name:     "(",
			input:    "(",
			expected: '(',
		},
		{
			name:     ")",
			input:    ")",
			expected: ')',
		},
	}

	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			l := New(tc.input)
			c, e := l.nextNonWhiteSpace()
			require.Equal(t, tc.expected, c)
			require.Equal(t, tc.expectedErr, e)
		})
	}
}
