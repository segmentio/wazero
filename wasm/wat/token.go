package wat

// token is the set of tokens defined by the WebAssembly Text Format 1.0
// See https://www.w3.org/TR/wasm-core-1/#tokens%E2%91%A0
type tokenType byte

const (
	// tokenKeyword is a potentially empty sequence of asciiTypeId characters prefixed by a lowercase letter.
	//
	// For example, in the below, 'local.get' 'i32.const' and 'i32.lt_s' are keywords:
	//		local.get $y
	//		i32.const 6
	//		i32.lt_s
	//
	// See https://www.w3.org/TR/wasm-core-1/#text-keyword
	tokenKeyword tokenType = iota

	// tokenUN is an unsigned integer in decimal or hexadecimal notation, optionally separated by underscores.
	//
	// For example, the following tokens represent the same number: 10
	//		(i32.const 10)
	//		(i32.const 1_0)
	//		(i32.const 0x0a)
	//		(i32.const 0x0_A)
	//
	// See https://www.w3.org/TR/wasm-core-1/#text-int
	tokenUN

	// tokenSN is a signed integer in decimal or hexadecimal notation, optionally separated by underscores.
	//
	// For example, the following tokens represent the same number: 10
	//		(i32.const +10)
	//		(i32.const +1_0)
	//		(i32.const +0x0a)
	//		(i32.const +0x0_A)
	//
	// See https://www.w3.org/TR/wasm-core-1/#text-int
	tokenSN

	// tokenFN represents an IEEE-754 floating point number in decimal or hexadecimal notation, optionally separated by
	// underscores. This also includes special constants for infinity ('inf') and NaN ('nan').
	//
	// For example, the right-hand side of the following S-expressions are all valid floating point tokens:
	//		(f32.const +nan)
	//		(f64.const -nan:0xfffffffffffff)
	//		(f64.const -inf)
	//      (f64.const +0x0.0p0)
	//      (f32.const 0.0e0)
	//		(i32.const +0x0_A)
	//		(f32.const 1.e10)
	//      (f64.const 0x1.fff_fff_fff_fff_fp+1_023)
	//		(f64.const 1.7976931348623157e+308)
	tokenFN

	// tokenString is a UTF-8 sequence enclosed by quotation marks, representing an encoded byte string. A tokenString
	// can contain any character except ASCII control characters, quotation marks ('"') and backslash ('\'): these must
	// be escaped. tokenString characters correspond to UTF-8 encoding except the special case of '\hh', which allows
	// raw bytes expressed as hexadecimal.
	//
	// For example, the following tokens represent the same raw bytes: 0xe298ba0a
	//		(data (i32.const 0) "☺\n")
	//		(data (i32.const 0) "\u{263a}\u{0a}")
	//		(data (i32.const 0) "\e2\98\ba\0a")
	//
	// See https://www.w3.org/TR/wasm-core-1/#strings%E2%91%A0
	tokenString

	// tokenId is a sequence of asciiTypeId characters prefixed by a '$':
	//
	// For example, in the below, '$y' is an id:
	//		local.get $y
	//		i32.const 6
	//		i32.lt_s
	//
	// See https://www.w3.org/TR/wasm-core-1/#text-id
	tokenId

	// tokenLParen is a left paren: '('
	tokenLParen

	// tokenLParen is a left paren: ')'
	tokenRParen

	// tokenReserved is a sequence of asciiTypeId characters which are neither a tokenId nor a tokenString.
	//
	// For example, '0$y' is a tokenReserved, because it doesn't start with a letter or '$'.
	//
	// See https://www.w3.org/TR/wasm-core-1/#text-reserved
	tokenReserved
)

// tokenNames is index-coordinated with tokenType
var tokenNames = []string{
	"keyword",
	"uN",
	"sN",
	"fN",
	"string",
	"id",
	"(",
	")",
	"reserved",
}

// String returns the string name of this token.
func (t tokenType) String() string {
	return tokenNames[t]
}

// asciiType helps identify a meaningful ASCII character in a UTF-8 string.
type asciiType byte

const (
	asciiTypeUnknown asciiType = iota
	// asciiTypeId is a printable ASCII character that does not contain a space, quotation mark, comma, semicolon, or bracket.
	// See https://www.w3.org/TR/wasm-core-1/#text-idchar
	asciiTypeId
	// wsChar is a space, tab, newline or carriage return
	// See https://www.w3.org/TR/wasm-core-1/#text-space
	asciiTypeWs
)

var asciiMap = buildAsciiMap()

func buildAsciiMap() (result [256]asciiType) {
	for i := 0; i < 128; i++ {
		result[i] = getAsciiType(byte(i))
	}
	return
}

func getAsciiType(ch byte) asciiType {
	switch ch {
	case ' ':
		fallthrough
	case '\t':
		fallthrough
	case '\r':
		fallthrough
	case '\n':
		return asciiTypeWs
	case '!':
		fallthrough
	case '#':
		fallthrough
	case '$':
		fallthrough
	case '%':
		fallthrough
	case '&':
		fallthrough
	case '\'':
		fallthrough
	case '*':
		fallthrough
	case '+':
		fallthrough
	case '-':
		fallthrough
	case '.':
		fallthrough
	case '/':
		fallthrough
	case ':':
		fallthrough
	case '<':
		fallthrough
	case '=':
		fallthrough
	case '>':
		fallthrough
	case '?':
		fallthrough
	case '@':
		fallthrough
	case '\\':
		fallthrough
	case '^':
		fallthrough
	case '_':
		fallthrough
	case '`':
		fallthrough
	case '|':
		fallthrough
	case '~':
		return asciiTypeId
	}
	switch {
	case ch >= '0' && ch <= '9':
		fallthrough
	case ch >= 'a' && ch <= 'z':
		fallthrough
	case ch >= 'A' && ch <= 'Z':
		return asciiTypeId
	}
	return asciiTypeUnknown
}
