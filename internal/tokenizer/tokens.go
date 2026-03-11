// Package tokenizer provides token types and tokenization for the properties format.
package tokenizer

// Token kinds for the properties format
const (
	TokenKey     = "KEY"     // Property key: [A-Za-z_][A-Za-z0-9_.-]*
	TokenEquals  = "="       // Assignment separator
	TokenValue   = "VALUE"   // Property value: everything after = until newline
	TokenComment = "COMMENT" // Comment: # to end of line
	TokenNewline = "NEWLINE" // Line terminator: \n or \r\n
	TokenEOF     = "EOF"     // End of input
)

// Token represents a lexical token in the properties format.
type Token struct {
	Kind   string // Token type (KEY, EQUALS, VALUE, COMMENT, NEWLINE, EOF)
	Value  string // Token text
	Line   int    // 1-based line number
	Column int    // 1-based column number
	Offset int    // Byte offset from start of input
}

// NewToken creates a new token.
func NewToken(kind, value string, line, column, offset int) Token {
	return Token{
		Kind:   kind,
		Value:  value,
		Line:   line,
		Column: column,
		Offset: offset,
	}
}

// IsEOF returns true if this is an EOF token.
func (t Token) IsEOF() bool {
	return t.Kind == TokenEOF
}
