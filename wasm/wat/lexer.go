package wat

import (
	"fmt"
)

// parseToken allows a parser to inspect a token without necessarily allocating strings
// * source is the underlying byte stream: do not modify this
// * tokenType is the token type
// * beginPos is the byte position in the source where the token begins, inclusive
// * endPos is the byte position in the source where the token ends, exclusive
//
// Returning an error will short-circuit any future invocations.
type parseToken func(source []byte, tok tokenType, beginPos, endPos int) error

// lex invokes the parser function for each token, the source is exhausted.
//
// Errors from the parser or during tokenization exit early, such as dangling block comments or unexpected characters.
func lex(source []byte, parser parseToken) error {
	length := len(source)
	p := 0
	for p < length {
		c1 := source[p]
		var c2 byte
		if p+1 < length {
			c2 = source[p+1]
		}

		if c1 == ' ' || c1 == '\t' || c1 == '\n' || c1 == '\r' {
			p = p + 1
			continue // next whitespace
		}

		if c1 == ';' && c2 == ';' { // line comment
			p = p + 2                             // consume ;;
			for p < length && source[p] != '\n' { // skip to end of line or EOF
				p = p + 1
			}
			continue // next whitespace
		}

		if c1 == '(' && c2 == ';' { // block comment
			p = p + 2 // consume (;

			level := 1
			for p+1 < length && level > 0 { // skip to end of all block comments
				c1 = source[p]
				c2 = source[p+1]
				if c1 == ';' && c2 == ')' {
					p = p + 2 // consume ;)
					level = level - 1
				} else if c1 == '(' && c2 == ';' {
					p = p + 2 // consume (;
					level = level + 1
				} else { // move ahead
					p = p + 1
				}
			}
			if level > 0 {
				return fmt.Errorf("expected block comment end ';)' at position %d", p)
			}
			continue // next whitespace
		}

		// no more whitespace: start tokenization
		peekEOFOrWs := c2 == 0 || c2 == ' ' || c2 == '\t' || c2 == '\n' || c2 == '\r'
		switch {
		case c1 == '(' && peekEOFOrWs:
			if e := parser(source, tokenLParen, p, p); e != nil {
				return e
			}
			p = p + 1
			continue
		case c1 == ')' && peekEOFOrWs:
			if e := parser(source, tokenRParen, p, p); e != nil {
				return e
			}
			p = p + 1
			continue
		}
		return fmt.Errorf("unexpected character %s at position %d", string(c1), p)
	}
	return nil // EOF
}
