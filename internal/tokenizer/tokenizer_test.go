package tokenizer

import (
	"strings"
	"testing"
)

func TestTokenizer_SimpleProperty(t *testing.T) {
	input := "host=localhost"
	tok := NewTokenizer(input)

	// KEY: host
	token := tok.NextToken()
	if token.Kind != TokenKey || token.Value != "host" {
		t.Errorf("expected KEY 'host', got %s '%s'", token.Kind, token.Value)
	}

	// EQUALS: =
	token = tok.NextToken()
	if token.Kind != TokenEquals || token.Value != "=" {
		t.Errorf("expected EQUALS '=', got %s '%s'", token.Kind, token.Value)
	}

	// VALUE: localhost
	token = tok.ScanValue()
	if token.Kind != TokenValue || token.Value != "localhost" {
		t.Errorf("expected VALUE 'localhost', got %s '%s'", token.Kind, token.Value)
	}

	// EOF
	token = tok.NextToken()
	if token.Kind != TokenEOF {
		t.Errorf("expected EOF, got %s", token.Kind)
	}
}

func TestTokenizer_MultipleProperties(t *testing.T) {
	input := "host=localhost\nport=8080"
	tok := NewTokenizer(input)

	// First property
	if tok.NextToken().Kind != TokenKey {
		t.Error("expected KEY")
	}
	if tok.NextToken().Kind != TokenEquals {
		t.Error("expected EQUALS")
	}
	if tok.ScanValue().Kind != TokenValue {
		t.Error("expected VALUE")
	}
	if tok.NextToken().Kind != TokenNewline {
		t.Error("expected NEWLINE")
	}

	// Second property
	if tok.NextToken().Kind != TokenKey {
		t.Error("expected KEY")
	}
	if tok.NextToken().Kind != TokenEquals {
		t.Error("expected EQUALS")
	}
	if tok.ScanValue().Kind != TokenValue {
		t.Error("expected VALUE")
	}

	// EOF
	if tok.NextToken().Kind != TokenEOF {
		t.Error("expected EOF")
	}
}

func TestTokenizer_Comment(t *testing.T) {
	input := "# this is a comment"
	tok := NewTokenizer(input)

	token := tok.NextToken()
	if token.Kind != TokenComment {
		t.Errorf("expected COMMENT, got %s", token.Kind)
	}
	if token.Value != "# this is a comment" {
		t.Errorf("expected '# this is a comment', got '%s'", token.Value)
	}
}

func TestTokenizer_CommentThenProperty(t *testing.T) {
	input := "# comment\nhost=localhost"
	tok := NewTokenizer(input)

	// Comment
	token := tok.NextToken()
	if token.Kind != TokenComment {
		t.Errorf("expected COMMENT, got %s", token.Kind)
	}

	// Newline
	token = tok.NextToken()
	if token.Kind != TokenNewline {
		t.Errorf("expected NEWLINE, got %s", token.Kind)
	}

	// Key
	token = tok.NextToken()
	if token.Kind != TokenKey || token.Value != "host" {
		t.Errorf("expected KEY 'host', got %s '%s'", token.Kind, token.Value)
	}
}

func TestTokenizer_KeyFormats(t *testing.T) {
	tests := []struct {
		input string
		key   string
	}{
		{"host=value", "host"},
		{"HOST=value", "HOST"},
		{"_private=value", "_private"},
		{"db.host=value", "db.host"},
		{"log-level=value", "log-level"},
		{"SERVICE_NAME=value", "SERVICE_NAME"},
		{"key123=value", "key123"},
		{"a.b.c.d=value", "a.b.c.d"},
		{"my-app.server.port=value", "my-app.server.port"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			tok := NewTokenizer(tt.input)
			token := tok.NextToken()

			if token.Kind != TokenKey {
				t.Errorf("expected KEY, got %s", token.Kind)
			}
			if token.Value != tt.key {
				t.Errorf("expected '%s', got '%s'", tt.key, token.Value)
			}
		})
	}
}

func TestTokenizer_ValueWithSpaces(t *testing.T) {
	input := "message=hello world"
	tok := NewTokenizer(input)

	tok.NextToken() // KEY
	tok.NextToken() // EQUALS

	token := tok.ScanValue()
	if token.Value != "hello world" {
		t.Errorf("expected 'hello world', got '%s'", token.Value)
	}
}

func TestTokenizer_ValueWithLeadingSpaces(t *testing.T) {
	input := "key =   value with spaces"
	tok := NewTokenizer(input)

	tok.NextToken() // KEY
	tok.NextToken() // EQUALS

	token := tok.ScanValue()
	if token.Value != "value with spaces" {
		t.Errorf("expected 'value with spaces', got '%s'", token.Value)
	}
}

func TestTokenizer_ValueWithEqualsSign(t *testing.T) {
	input := "equation=a=b+c"
	tok := NewTokenizer(input)

	tok.NextToken() // KEY
	tok.NextToken() // EQUALS

	token := tok.ScanValue()
	if token.Value != "a=b+c" {
		t.Errorf("expected 'a=b+c', got '%s'", token.Value)
	}
}

func TestTokenizer_EmptyValue(t *testing.T) {
	input := "empty="
	tok := NewTokenizer(input)

	tok.NextToken() // KEY
	tok.NextToken() // EQUALS

	token := tok.ScanValue()
	if token.Value != "" {
		t.Errorf("expected empty string, got '%s'", token.Value)
	}
}

func TestTokenizer_LineNumbers(t *testing.T) {
	input := "a=1\nb=2\nc=3"
	tok := NewTokenizer(input)

	// Line 1
	token := tok.NextToken() // KEY: a
	if token.Line != 1 || token.Column != 1 {
		t.Errorf("expected line 1, col 1, got line %d, col %d", token.Line, token.Column)
	}

	tok.NextToken() // EQUALS
	tok.ScanValue() // VALUE
	tok.NextToken() // NEWLINE

	// Line 2
	token = tok.NextToken() // KEY: b
	if token.Line != 2 || token.Column != 1 {
		t.Errorf("expected line 2, col 1, got line %d, col %d", token.Line, token.Column)
	}
}

func TestTokenizer_CRLFLineEndings(t *testing.T) {
	input := "a=1\r\nb=2"
	tok := NewTokenizer(input)

	tok.NextToken() // KEY
	tok.NextToken() // EQUALS
	tok.ScanValue() // VALUE

	token := tok.NextToken() // NEWLINE
	if token.Kind != TokenNewline || token.Value != "\r\n" {
		t.Errorf("expected NEWLINE '\\r\\n', got %s '%s'", token.Kind, token.Value)
	}

	token = tok.NextToken() // KEY: b
	if token.Kind != TokenKey || token.Value != "b" {
		t.Errorf("expected KEY 'b', got %s '%s'", token.Kind, token.Value)
	}
}

func TestTokenizer_TrailingWhitespaceInValue(t *testing.T) {
	input := "key=value   "
	tok := NewTokenizer(input)

	tok.NextToken() // KEY
	tok.NextToken() // EQUALS

	token := tok.ScanValue()
	if token.Value != "value" {
		t.Errorf("expected 'value' (trimmed), got '%s'", token.Value)
	}
}

func TestTokenizer_ValueNotComment(t *testing.T) {
	// Per spec: inline comments are not supported
	input := "port=1234   # not a comment"
	tok := NewTokenizer(input)

	tok.NextToken() // KEY
	tok.NextToken() // EQUALS

	token := tok.ScanValue()
	expected := "1234   # not a comment"
	if token.Value != expected {
		t.Errorf("expected '%s', got '%s'", expected, token.Value)
	}
}

func TestTokenizer_EmptyInput(t *testing.T) {
	tok := NewTokenizer("")

	token := tok.NextToken()
	if token.Kind != TokenEOF {
		t.Errorf("expected EOF, got %s", token.Kind)
	}
}

func TestTokenizer_WhitespaceOnly(t *testing.T) {
	tok := NewTokenizer("   \t   ")

	token := tok.NextToken()
	if token.Kind != TokenEOF {
		t.Errorf("expected EOF, got %s", token.Kind)
	}
}

func TestTokenizer_FromReader(t *testing.T) {
	input := "host=localhost"
	reader := strings.NewReader(input)

	tok, err := NewTokenizerFromReader(reader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	token := tok.NextToken()
	if token.Kind != TokenKey || token.Value != "host" {
		t.Errorf("expected KEY 'host', got %s '%s'", token.Kind, token.Value)
	}
}

func TestTokenizer_UTF8Value(t *testing.T) {
	input := "greeting=こんにちは"
	tok := NewTokenizer(input)

	tok.NextToken() // KEY
	tok.NextToken() // EQUALS

	token := tok.ScanValue()
	if token.Value != "こんにちは" {
		t.Errorf("expected 'こんにちは', got '%s'", token.Value)
	}
}

func TestTokenizer_PathValue(t *testing.T) {
	input := "path=/var/log/app"
	tok := NewTokenizer(input)

	tok.NextToken() // KEY
	tok.NextToken() // EQUALS

	token := tok.ScanValue()
	if token.Value != "/var/log/app" {
		t.Errorf("expected '/var/log/app', got '%s'", token.Value)
	}
}
