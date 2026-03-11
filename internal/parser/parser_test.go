package parser

import (
	"strings"
	"testing"

	"github.com/shapestone/shape-core/pkg/ast"
)

func TestParser_SimpleProperty(t *testing.T) {
	input := "host=localhost"
	p := NewParser(input)

	node, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	obj, ok := node.(*ast.ObjectNode)
	if !ok {
		t.Fatalf("expected ObjectNode, got %T", node)
	}

	hostProp, ok := obj.Properties()["host"]
	if !ok {
		t.Fatal("missing 'host' property")
	}

	lit, ok := hostProp.(*ast.LiteralNode)
	if !ok {
		t.Fatalf("expected LiteralNode, got %T", hostProp)
	}

	if lit.Value() != "localhost" {
		t.Errorf("expected 'localhost', got %q", lit.Value())
	}
}

func TestParser_MultipleProperties(t *testing.T) {
	input := `host=localhost
port=8080
debug=true`

	p := NewParser(input)
	node, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	obj := node.(*ast.ObjectNode)
	props := obj.Properties()

	if len(props) != 3 {
		t.Errorf("expected 3 properties, got %d", len(props))
	}

	tests := map[string]string{
		"host":  "localhost",
		"port":  "8080",
		"debug": "true",
	}

	for key, expected := range tests {
		prop, ok := props[key]
		if !ok {
			t.Errorf("missing property %q", key)
			continue
		}
		lit := prop.(*ast.LiteralNode)
		if lit.Value() != expected {
			t.Errorf("property %q: expected %q, got %q", key, expected, lit.Value())
		}
	}
}

func TestParser_Comments(t *testing.T) {
	input := `# This is a comment
host=localhost
# Another comment
port=8080`

	p := NewParser(input)
	node, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	obj := node.(*ast.ObjectNode)
	if len(obj.Properties()) != 2 {
		t.Errorf("expected 2 properties, got %d", len(obj.Properties()))
	}
}

func TestParser_EmptyLines(t *testing.T) {
	input := `host=localhost

port=8080

`

	p := NewParser(input)
	node, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	obj := node.(*ast.ObjectNode)
	if len(obj.Properties()) != 2 {
		t.Errorf("expected 2 properties, got %d", len(obj.Properties()))
	}
}

func TestParser_EmptyValue(t *testing.T) {
	input := "empty="

	p := NewParser(input)
	node, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	obj := node.(*ast.ObjectNode)
	prop := obj.Properties()["empty"].(*ast.LiteralNode)

	if prop.Value() != "" {
		t.Errorf("expected empty string, got %q", prop.Value())
	}
}

func TestParser_ValueWithEquals(t *testing.T) {
	input := "equation=a=b+c"

	p := NewParser(input)
	node, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	obj := node.(*ast.ObjectNode)
	prop := obj.Properties()["equation"].(*ast.LiteralNode)

	if prop.Value() != "a=b+c" {
		t.Errorf("expected 'a=b+c', got %q", prop.Value())
	}
}

func TestParser_ValueWithSpaces(t *testing.T) {
	input := "message=hello world"

	p := NewParser(input)
	node, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	obj := node.(*ast.ObjectNode)
	prop := obj.Properties()["message"].(*ast.LiteralNode)

	if prop.Value() != "hello world" {
		t.Errorf("expected 'hello world', got %q", prop.Value())
	}
}

func TestParser_InlineCommentNotSupported(t *testing.T) {
	// Per spec: inline comments are NOT supported
	input := "port=1234   # this is not a comment"

	p := NewParser(input)
	node, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	obj := node.(*ast.ObjectNode)
	prop := obj.Properties()["port"].(*ast.LiteralNode)

	// The entire line after = should be the value (trailing whitespace trimmed)
	expected := "1234   # this is not a comment"
	if prop.Value() != expected {
		t.Errorf("expected %q, got %q", expected, prop.Value())
	}
}

func TestParser_KeyFormats(t *testing.T) {
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
			p := NewParser(tt.input)
			node, err := p.Parse()
			if err != nil {
				t.Fatalf("Parse() error: %v", err)
			}

			obj := node.(*ast.ObjectNode)
			if _, ok := obj.Properties()[tt.key]; !ok {
				t.Errorf("missing property %q", tt.key)
			}
		})
	}
}

func TestParser_PathValue(t *testing.T) {
	input := "path=/var/log/app"

	p := NewParser(input)
	node, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	obj := node.(*ast.ObjectNode)
	prop := obj.Properties()["path"].(*ast.LiteralNode)

	if prop.Value() != "/var/log/app" {
		t.Errorf("expected '/var/log/app', got %q", prop.Value())
	}
}

func TestParser_UTF8Value(t *testing.T) {
	input := "greeting=こんにちは"

	p := NewParser(input)
	node, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	obj := node.(*ast.ObjectNode)
	prop := obj.Properties()["greeting"].(*ast.LiteralNode)

	if prop.Value() != "こんにちは" {
		t.Errorf("expected 'こんにちは', got %q", prop.Value())
	}
}

func TestParser_FromReader(t *testing.T) {
	input := "host=localhost\nport=8080"
	reader := strings.NewReader(input)

	p, err := NewParserFromReader(reader)
	if err != nil {
		t.Fatalf("NewParserFromReader() error: %v", err)
	}

	node, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	obj := node.(*ast.ObjectNode)
	if len(obj.Properties()) != 2 {
		t.Errorf("expected 2 properties, got %d", len(obj.Properties()))
	}
}

func TestParser_EmptyInput(t *testing.T) {
	p := NewParser("")
	node, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	obj := node.(*ast.ObjectNode)
	if len(obj.Properties()) != 0 {
		t.Errorf("expected 0 properties, got %d", len(obj.Properties()))
	}
}

func TestParser_OnlyComments(t *testing.T) {
	input := `# comment 1
# comment 2
# comment 3`

	p := NewParser(input)
	node, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	obj := node.(*ast.ObjectNode)
	if len(obj.Properties()) != 0 {
		t.Errorf("expected 0 properties, got %d", len(obj.Properties()))
	}
}

// Error cases

func TestParser_DuplicateKey(t *testing.T) {
	input := `host=localhost
host=example.com`

	p := NewParser(input)
	_, err := p.Parse()
	if err == nil {
		t.Fatal("expected error for duplicate key")
	}
	if !strings.Contains(err.Error(), "duplicate") {
		t.Errorf("expected 'duplicate' in error, got: %v", err)
	}
}

func TestParser_MissingEquals(t *testing.T) {
	input := "host localhost"

	p := NewParser(input)
	_, err := p.Parse()
	if err == nil {
		t.Fatal("expected error for missing equals")
	}
}

func TestParser_InvalidKeyStart(t *testing.T) {
	input := "123invalid=value"

	p := NewParser(input)
	_, err := p.Parse()
	if err == nil {
		t.Fatal("expected error for invalid key")
	}
}

func TestParser_ControlCharacterInValue(t *testing.T) {
	// Value with control character (except TAB which is allowed)
	input := "key=value\x01more"

	p := NewParser(input)
	_, err := p.Parse()
	if err == nil {
		t.Fatal("expected error for control character")
	}
	if !strings.Contains(err.Error(), "control character") {
		t.Errorf("expected 'control character' in error, got: %v", err)
	}
}

func TestParser_NULByte(t *testing.T) {
	input := "key=value\x00more"

	p := NewParser(input)
	_, err := p.Parse()
	if err == nil {
		t.Fatal("expected error for NUL byte")
	}
	if !strings.Contains(err.Error(), "NUL") {
		t.Errorf("expected 'NUL' in error, got: %v", err)
	}
}

func TestParser_TabInValue(t *testing.T) {
	// TAB is explicitly allowed
	input := "key=value\twith\ttabs"

	p := NewParser(input)
	node, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	obj := node.(*ast.ObjectNode)
	prop := obj.Properties()["key"].(*ast.LiteralNode)

	if prop.Value() != "value\twith\ttabs" {
		t.Errorf("expected 'value\\twith\\ttabs', got %q", prop.Value())
	}
}
