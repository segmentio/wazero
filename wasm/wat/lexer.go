package wat

import (
	"errors"
	"fmt"
	"io"
	"strconv"
)

type Lexer struct {
	input string
	pos   int // TODO: add line/column for readability
}

func New(input string) *Lexer {
	return &Lexer{input: input}
}

func (l *Lexer) NextToken() (*token, error) {
	c, e := l.nextNonWhiteSpace()
	if e == io.EOF {
		return newToken(tokenEOF, ""), nil
	} else if e != nil {
		return nil, e
	}

	switch c {
	case '(':
		return newToken(tokenLParen, ""), nil
	case ')':
		return newToken(tokenRParen, ""), nil
	}
	return newToken(tokenIllegal, strconv.Itoa(int(c))), nil
}

// nextNonWhiteSpace returns the next character that isn't a white space as defined in the text format or raises an
// error on io.EOF or dangling block comments.
//
// See https://www.w3.org/TR/wasm-core-1/#white-space%E2%91%A0
func (l *Lexer) nextNonWhiteSpace() (c byte, e error) {
	input := l.input
	p := l.pos
	length := len(input)
	for p < length {
		c = input[p]
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			p = p + 1
			c = 0
			continue // next whitespace
		}

		if c == ';' && p+1 < length && input[p+1] == ';' { // line comment
			p = p + 2 // consume ;;
			c = 0
			for p < length && input[p] != '\n' { // skip to end of line or EOF
				p = p + 1
			}
			continue // next whitespace
		}

		if c == '(' && p+1 < length && input[p+1] == ';' { // block comment
			p = p + 2 // consume (;
			c = 0

			level := 1
			for p+1 < length && level > 0{ // skip to end of all block comments
				c1 := input[p]
				c2 := input[p+1]
				p = p + 2
				if c1 == ';' && c2 == ')' {
					level = level - 1
				} else if c1 == '(' && c2 == ';' {
					level = level + 1
				} // else skip
			}
			if level > 0 {
				return 0, errors.New("unbalanced block comment") // TODO: at line/column
			}
			continue // next whitespace
		}

		break // no more whitespace!
	}

	if p < length {
		l.pos = p + 1
		return c, nil
	}
	return 0, io.EOF
}

func newToken(tokenType tokenType, val string) *token {
	return &token{tokenType, val}
}

type token struct {
	tokenType
	value string
}

// String returns the string representation of this token.
func (t *token) String() string {
	if t.value == "" {
		return t.tokenType.String()
	}
	return fmt.Sprintf("%s(%s)", t.tokenType, t.value)
}
