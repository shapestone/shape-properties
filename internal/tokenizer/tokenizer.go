package tokenizer

import (
	"io"
	"strings"
	"unicode/utf8"
)

// Tokenizer tokenizes properties format input.
type Tokenizer struct {
	input  string
	pos    int    // Current byte position
	line   int    // Current line number (1-based)
	column int    // Current column number (1-based)
	start  int    // Start position of current token
}

// NewTokenizer creates a new tokenizer for the given input string.
func NewTokenizer(input string) *Tokenizer {
	return &Tokenizer{
		input:  input,
		pos:    0,
		line:   1,
		column: 1,
	}
}

// NewTokenizerFromReader creates a new tokenizer from an io.Reader.
func NewTokenizerFromReader(r io.Reader) (*Tokenizer, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return NewTokenizer(string(data)), nil
}

// NextToken returns the next token from the input.
func (t *Tokenizer) NextToken() Token {
	t.skipWhitespace()

	if t.pos >= len(t.input) {
		return NewToken(TokenEOF, "", t.line, t.column, t.pos)
	}

	t.start = t.pos
	startLine := t.line
	startColumn := t.column

	ch := t.peek()

	// Newline
	if ch == '\n' {
		t.advance()
		return NewToken(TokenNewline, "\n", startLine, startColumn, t.start)
	}
	if ch == '\r' {
		t.advance()
		if t.peek() == '\n' {
			t.advance()
			return NewToken(TokenNewline, "\r\n", startLine, startColumn, t.start)
		}
		return NewToken(TokenNewline, "\r", startLine, startColumn, t.start)
	}

	// Comment
	if ch == '#' {
		return t.scanComment(startLine, startColumn)
	}

	// Equals
	if ch == '=' {
		t.advance()
		return NewToken(TokenEquals, "=", startLine, startColumn, t.start)
	}

	// Key (must start with letter or underscore)
	if isKeyStart(ch) {
		return t.scanKey(startLine, startColumn)
	}

	// If we're after an equals sign on the same token stream, this would be a value
	// But values are scanned by scanValue which is called explicitly
	// For invalid characters, we return them as-is for error handling
	t.advance()
	return NewToken("INVALID", string(ch), startLine, startColumn, t.start)
}

// ScanValue scans a value after the equals sign.
// Values extend to the end of the line and preserve internal whitespace.
func (t *Tokenizer) ScanValue() Token {
	t.skipWhitespace()

	startLine := t.line
	startColumn := t.column
	t.start = t.pos

	// Find the end of the line
	end := t.pos
	for end < len(t.input) {
		ch := t.input[end]
		if ch == '\n' || ch == '\r' {
			break
		}
		end++
	}

	// Extract the raw value
	rawValue := t.input[t.pos:end]

	// Update position - we need to properly track line/column for UTF-8
	for t.pos < end {
		r, size := utf8.DecodeRuneInString(t.input[t.pos:])
		if r == utf8.RuneError && size == 1 {
			// Invalid UTF-8, advance one byte
			t.pos++
			t.column++
		} else {
			t.pos += size
			t.column++
		}
	}

	// Trim trailing whitespace from value
	result := strings.TrimRight(rawValue, " \t")

	return NewToken(TokenValue, result, startLine, startColumn, t.start)
}

// scanComment scans a comment from # to end of line.
func (t *Tokenizer) scanComment(startLine, startColumn int) Token {
	var value strings.Builder
	value.WriteByte(t.peek()) // Include the #
	t.advance()

	for t.pos < len(t.input) {
		ch := t.peek()
		if ch == '\n' || ch == '\r' {
			break
		}
		value.WriteByte(ch)
		t.advance()
	}

	return NewToken(TokenComment, value.String(), startLine, startColumn, t.start)
}

// scanKey scans a property key.
func (t *Tokenizer) scanKey(startLine, startColumn int) Token {
	var value strings.Builder

	for t.pos < len(t.input) {
		ch := t.peek()
		if !isKeyChar(ch) {
			break
		}
		value.WriteByte(ch)
		t.advance()
	}

	return NewToken(TokenKey, value.String(), startLine, startColumn, t.start)
}

// skipWhitespace skips spaces and tabs (not newlines).
func (t *Tokenizer) skipWhitespace() {
	for t.pos < len(t.input) {
		ch := t.peek()
		if ch != ' ' && ch != '\t' {
			break
		}
		t.advance()
	}
}

// peek returns the current byte without advancing.
func (t *Tokenizer) peek() byte {
	if t.pos >= len(t.input) {
		return 0
	}
	return t.input[t.pos]
}

// advance moves forward one byte, updating line and column.
func (t *Tokenizer) advance() {
	if t.pos >= len(t.input) {
		return
	}

	ch := t.input[t.pos]
	if ch == '\n' {
		t.line++
		t.column = 1
	} else {
		// Handle multi-byte UTF-8 characters correctly
		_, size := utf8.DecodeRuneInString(t.input[t.pos:])
		if size > 1 {
			t.pos += size - 1 // We'll add 1 below
		}
		t.column++
	}
	t.pos++
}

// isKeyStart returns true if the byte can start a key.
func isKeyStart(ch byte) bool {
	return (ch >= 'A' && ch <= 'Z') ||
		(ch >= 'a' && ch <= 'z') ||
		ch == '_'
}

// isKeyChar returns true if the byte can be part of a key.
func isKeyChar(ch byte) bool {
	return isKeyStart(ch) ||
		(ch >= '0' && ch <= '9') ||
		ch == '-' ||
		ch == '.'
}

// Position returns the current position information.
func (t *Tokenizer) Position() (line, column, offset int) {
	return t.line, t.column, t.pos
}
