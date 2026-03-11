package fastparser

import (
	"errors"
	"strings"
	"testing"
)

func TestFastParser_SimpleProperty(t *testing.T) {
	input := []byte("host=localhost")
	p := NewParser(input)

	result, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if result["host"] != "localhost" {
		t.Errorf("expected 'localhost', got %q", result["host"])
	}
}

func TestFastParser_MultipleProperties(t *testing.T) {
	input := []byte(`host=localhost
port=8080
debug=true`)

	p := NewParser(input)
	result, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if len(result) != 3 {
		t.Errorf("expected 3 properties, got %d", len(result))
	}

	tests := map[string]string{
		"host":  "localhost",
		"port":  "8080",
		"debug": "true",
	}

	for key, expected := range tests {
		if result[key] != expected {
			t.Errorf("property %q: expected %q, got %q", key, expected, result[key])
		}
	}
}

func TestFastParser_Comments(t *testing.T) {
	input := []byte(`# This is a comment
host=localhost
# Another comment
port=8080`)

	p := NewParser(input)
	result, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("expected 2 properties, got %d", len(result))
	}
}

func TestFastParser_EmptyLines(t *testing.T) {
	input := []byte(`host=localhost

port=8080

`)

	p := NewParser(input)
	result, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("expected 2 properties, got %d", len(result))
	}
}

func TestFastParser_EmptyValue(t *testing.T) {
	input := []byte("empty=")

	p := NewParser(input)
	result, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if result["empty"] != "" {
		t.Errorf("expected empty string, got %q", result["empty"])
	}
}

func TestFastParser_ValueWithEquals(t *testing.T) {
	input := []byte("equation=a=b+c")

	p := NewParser(input)
	result, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if result["equation"] != "a=b+c" {
		t.Errorf("expected 'a=b+c', got %q", result["equation"])
	}
}

func TestFastParser_ValueWithSpaces(t *testing.T) {
	input := []byte("message=hello world")

	p := NewParser(input)
	result, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if result["message"] != "hello world" {
		t.Errorf("expected 'hello world', got %q", result["message"])
	}
}

func TestFastParser_InlineCommentNotSupported(t *testing.T) {
	// Per spec: inline comments are NOT supported
	input := []byte("port=1234   # this is not a comment")

	p := NewParser(input)
	result, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	expected := "1234   # this is not a comment"
	if result["port"] != expected {
		t.Errorf("expected %q, got %q", expected, result["port"])
	}
}

func TestFastParser_KeyFormats(t *testing.T) {
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
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			p := NewParser([]byte(tt.input))
			result, err := p.Parse()
			if err != nil {
				t.Fatalf("Parse() error: %v", err)
			}

			if _, ok := result[tt.key]; !ok {
				t.Errorf("missing property %q", tt.key)
			}
		})
	}
}

func TestFastParser_PathValue(t *testing.T) {
	input := []byte("path=/var/log/app")

	p := NewParser(input)
	result, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if result["path"] != "/var/log/app" {
		t.Errorf("expected '/var/log/app', got %q", result["path"])
	}
}

func TestFastParser_UTF8Value(t *testing.T) {
	input := []byte("greeting=こんにちは")

	p := NewParser(input)
	result, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if result["greeting"] != "こんにちは" {
		t.Errorf("expected 'こんにちは', got %q", result["greeting"])
	}
}

func TestFastParser_FromReader(t *testing.T) {
	input := "host=localhost\nport=8080"
	reader := strings.NewReader(input)

	p, err := NewParserFromReader(reader)
	if err != nil {
		t.Fatalf("NewParserFromReader() error: %v", err)
	}

	result, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("expected 2 properties, got %d", len(result))
	}
}

func TestFastParser_EmptyInput(t *testing.T) {
	p := NewParser([]byte(""))
	result, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("expected 0 properties, got %d", len(result))
	}
}

func TestFastParser_OnlyComments(t *testing.T) {
	input := []byte(`# comment 1
# comment 2
# comment 3`)

	p := NewParser(input)
	result, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("expected 0 properties, got %d", len(result))
	}
}

func TestFastParser_Validate(t *testing.T) {
	input := []byte("host=localhost\nport=8080")

	p := NewParser(input)
	err := p.Validate()
	if err != nil {
		t.Fatalf("Validate() error: %v", err)
	}
}

func TestFastParser_TrailingWhitespace(t *testing.T) {
	input := []byte("key=value   ")

	p := NewParser(input)
	result, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if result["key"] != "value" {
		t.Errorf("expected 'value', got %q", result["key"])
	}
}

func TestFastParser_CRLFLineEndings(t *testing.T) {
	input := []byte("a=1\r\nb=2\r\n")

	p := NewParser(input)
	result, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("expected 2 properties, got %d", len(result))
	}
	if result["a"] != "1" || result["b"] != "2" {
		t.Errorf("unexpected values: a=%q, b=%q", result["a"], result["b"])
	}
}

// Error cases

func TestFastParser_DuplicateKey(t *testing.T) {
	input := []byte(`host=localhost
host=example.com`)

	p := NewParser(input)
	_, err := p.Parse()
	if err == nil {
		t.Fatal("expected error for duplicate key")
	}
	if !strings.Contains(err.Error(), "duplicate") {
		t.Errorf("expected 'duplicate' in error, got: %v", err)
	}
}

func TestFastParser_MissingEquals(t *testing.T) {
	input := []byte("host localhost")

	p := NewParser(input)
	_, err := p.Parse()
	if err == nil {
		t.Fatal("expected error for missing equals")
	}
}

func TestFastParser_InvalidKeyStart(t *testing.T) {
	input := []byte("123invalid=value")

	p := NewParser(input)
	_, err := p.Parse()
	if err == nil {
		t.Fatal("expected error for invalid key")
	}
}

func TestFastParser_ControlCharacterInValue(t *testing.T) {
	input := []byte("key=value\x01more")

	p := NewParser(input)
	_, err := p.Parse()
	if err == nil {
		t.Fatal("expected error for control character")
	}
	if !strings.Contains(err.Error(), "control character") {
		t.Errorf("expected 'control character' in error, got: %v", err)
	}
}

func TestFastParser_NULByte(t *testing.T) {
	input := []byte("key=value\x00more")

	p := NewParser(input)
	_, err := p.Parse()
	if err == nil {
		t.Fatal("expected error for NUL byte")
	}
	if !strings.Contains(err.Error(), "NUL") {
		t.Errorf("expected 'NUL' in error, got: %v", err)
	}
}

func TestFastParser_TabInValue(t *testing.T) {
	// TAB is explicitly allowed
	input := []byte("key=value\twith\ttabs")

	p := NewParser(input)
	result, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if result["key"] != "value\twith\ttabs" {
		t.Errorf("expected 'value\\twith\\ttabs', got %q", result["key"])
	}
}

func TestFastParser_ValidateError(t *testing.T) {
	input := []byte("host localhost") // missing =

	p := NewParser(input)
	err := p.Validate()
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestFastParser_NewParserFromString(t *testing.T) {
	p := NewParserFromString("host=localhost\nport=8080")
	result, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 properties, got %d", len(result))
	}
	if result["host"] != "localhost" || result["port"] != "8080" {
		t.Errorf("unexpected values: %v", result)
	}
}

func TestFastParser_NewParserFromReaderError(t *testing.T) {
	_, err := NewParserFromReader(&errReader{})
	if err == nil {
		t.Fatal("expected error from failing reader")
	}
}

func TestFastParser_CommentCRLF(t *testing.T) {
	// Comment with CRLF ending exercises the \r branch in skipToEndOfLine
	input := []byte("# comment\r\nhost=localhost")
	p := NewParser(input)
	result, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if result["host"] != "localhost" {
		t.Errorf("expected 'localhost', got %q", result["host"])
	}
}

func TestFastParser_CommentCROnly(t *testing.T) {
	// Comment with bare CR ending (no following \n)
	input := []byte("# comment\rhost=localhost")
	p := NewParser(input)
	result, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if result["host"] != "localhost" {
		t.Errorf("expected 'localhost', got %q", result["host"])
	}
}

// errReader always returns an error on Read.
type errReader struct{}

func (e *errReader) Read(p []byte) (int, error) {
	return 0, errReadFailed
}

var errReadFailed = errors.New("read failed")
