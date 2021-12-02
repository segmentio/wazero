package wat

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestTokenType_String(t *testing.T) {
	tests := []struct {
		input    tokenType
		expected string
	}{
		{tokenIllegal, "<illegal>"},
		{tokenEOF, "EOF"},
		{tokenKeyword, "keyword"},
		{tokenUN, "uN"},
		{tokenSN, "sN"},
		{tokenFN, "fN"},
		{tokenString, "string"},
		{tokenId, "id"},
		{tokenLParen, "("},
		{tokenRParen, ")"},
		{tokenReserved, "reserved"},
	}

	for _, tt := range tests {
		tc := tt
		t.Run(tc.expected, func(t *testing.T) {
			require.Equal(t, tc.expected, tc.input.String())
		})
	}
}
