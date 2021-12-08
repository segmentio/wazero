package wat

import (
	"fmt"
	"unicode/utf8"
)

// parseToken allows a parser to inspect a token without necessarily allocating strings
// * source is the underlying byte stream: do not modify this
// * tokenType is the token type
// * beginPos is the byte position in the source where the token begins, inclusive
// * endPos is the byte position in the source where the token ends, exclusive
//
// Returning an error will short-circuit any future invocations.
type parseToken func(source []byte, tok tokenType, beginLine, beginCol, beginPos, endPos int) error

// lex invokes the parser function for each token, the source is exhausted.
//
// Errors from the parser or during tokenization exit early, such as dangling block comments or unexpected characters.
func lex(source []byte, parser parseToken) error {
	// One design-affecting constraint is that all characters must be 7-bit ASCII, except when in a string (enclosed by
	// '"'), or comments (whitespace). This simplifies line and column counting, as well boundaries otherwise.
	//
	// See https://www.w3.org/TR/wasm-core-1/#characters%E2%91%A0
	length := len(source)
	p := 0
	line := 1
	col := 0
	inLineComment := false
	blockCommentLevel := 0

	for ; p < length; p = p + 1 {
		b1 := source[p]

		// The spec does not consider newlines apart from '\n'. Notably, a bare '\r' is not a newline here.
		// See https://www.w3.org/TR/wasm-core-1/#text-comment
		if b1 == '\n' {
			line = line + 1
			inLineComment = false
			col = 0
			continue // next line
		}

		col = col + 1                              // the current character is at least one byte long
		if b1 == ' ' || b1 == '\t' || b1 == '\r' { // fast path ASCII whitespace
			continue // next whitespace
		}

		// check UTF-8 size as we may need to affect position without column!
		size := utf8Size(b1)
		switch {
		case size == -1:
			return fmt.Errorf("%d:%d unexpected character %x", line, col, b1)
		case size == 1: // ASCII
		default:
			if !inLineComment && blockCommentLevel == 0 { // non-ASCII is only allowed in comments or strings
				r, _ := utf8.DecodeRune(source[line:])
				return fmt.Errorf("%d:%d expected an ASCII character, not %s", line, col, string(r))
			}
			p = p + size - 1
			continue // skip to next character start or EOF
		}

		// From here on, we know b1 is ASCII

		var b2 byte
		if p+1 < length {
			b2 = source[p+1]
		}

		if b1 == '(' && b2 == ';' { // block comment
			p = p + 1 // consume (
			col = col + 1

			if !inLineComment {
				blockCommentLevel = blockCommentLevel + 1
			}
			continue
		}

		if blockCommentLevel > 0 && b1 == ';' && b2 == ')' {
			p = p + 1 // consume )
			col = col + 1

			if !inLineComment {
				blockCommentLevel = blockCommentLevel - 1
			}
			continue
		}

		if b1 == ';' && b2 == ';' { // line comment
			p = p + 1 // consume ;
			col = col + 1

			inLineComment = true
			continue // next whitespace
		}

		if inLineComment || blockCommentLevel > 0 {
			continue // skip validation as comments can contain line comments or any UTF-8
		}

		// no more whitespace: start tokenization!
		switch { // TODO: classify the first ASCII in a bitflag table
		case b1 == '(':
			if e := parser(source, tokenLParen, line, col, p, p); e != nil {
				return e
			}
		case b1 == ')':
			if e := parser(source, tokenRParen, line, col, p, p); e != nil {
				return e
			}
		case b1 >= 'a' && b1 <= 'z': // keyword
			p0 := p
			col0 := col
			for p+1 < length { // run until the end
				b1 = source[p+1]
				if asciiMap[b1] != asciiTypeId {
					break // end of this token (or malformed, which the next loop will notice)
				}
				p = p + 1
				col = col + 1
			}
			if e := parser(source, tokenKeyword, line, col0, p0, p+1); e != nil {
				return e
			}
		default:
			return fmt.Errorf("%d:%d unexpected character %s", line, col, string(b1))
		}
	}
	if blockCommentLevel > 0 {
		return fmt.Errorf("%d:%d expected block comment end ';)'", line, col)
	}
	return nil // EOF
}

// utf8Size returns the UTF-8 size (cheaper than utf8.DecodeRune), or -1 if invalid
func utf8Size(b1 byte) int { // inlinable
	switch {
	case b1&0x80 == 0x00:
		return 1 // 7-bit ASCII character
	case b1&0xe0 == 0xc0:
		return 2
	case b1&0xf0 == 0xe0:
		return 3
	case b1&0xf8 == 0xf0:
		return 4
	}
	return -1
}
