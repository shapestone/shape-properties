package properties

import (
	"errors"
	"strings"
	"testing"

	"github.com/shapestone/shape-core/pkg/ast"
)

// ============================================================================
// Fast Path Tests
// ============================================================================

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid simple", "host=localhost", false},
		{"valid multiple", "host=localhost\nport=8080", false},
		{"valid with comments", "# comment\nhost=localhost", false},
		{"invalid missing equals", "host localhost", true},
		{"invalid duplicate key", "host=a\nhost=b", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateReader(t *testing.T) {
	input := "host=localhost\nport=8080"
	reader := strings.NewReader(input)

	err := ValidateReader(reader)
	if err != nil {
		t.Errorf("ValidateReader() error = %v", err)
	}
}

func TestLoad(t *testing.T) {
	input := `# Configuration
host=localhost
port=8080
debug=true`

	props, err := Load(input)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(props) != 3 {
		t.Errorf("expected 3 properties, got %d", len(props))
	}

	expected := map[string]string{
		"host":  "localhost",
		"port":  "8080",
		"debug": "true",
	}

	for k, v := range expected {
		if props[k] != v {
			t.Errorf("Load()[%q] = %q, want %q", k, props[k], v)
		}
	}
}

func TestLoadReader(t *testing.T) {
	input := "host=localhost\nport=8080"
	reader := strings.NewReader(input)

	props, err := LoadReader(reader)
	if err != nil {
		t.Fatalf("LoadReader() error = %v", err)
	}

	if len(props) != 2 {
		t.Errorf("expected 2 properties, got %d", len(props))
	}
}

func TestMustLoad(t *testing.T) {
	props := MustLoad("host=localhost")
	if props["host"] != "localhost" {
		t.Errorf("MustLoad()['host'] = %q, want 'localhost'", props["host"])
	}
}

func TestMustLoadPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustLoad() should panic on invalid input")
		}
	}()
	MustLoad("invalid input without equals")
}

// ============================================================================
// AST Path Tests
// ============================================================================

func TestParse(t *testing.T) {
	input := `host=localhost
port=8080`

	node, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	obj, ok := node.(*ast.ObjectNode)
	if !ok {
		t.Fatalf("expected *ast.ObjectNode, got %T", node)
	}

	if len(obj.Properties()) != 2 {
		t.Errorf("expected 2 properties, got %d", len(obj.Properties()))
	}

	hostProp := obj.Properties()["host"].(*ast.LiteralNode)
	if hostProp.Value() != "localhost" {
		t.Errorf("host = %v, want 'localhost'", hostProp.Value())
	}
}

func TestParseReader(t *testing.T) {
	input := "host=localhost\nport=8080"
	reader := strings.NewReader(input)

	node, err := ParseReader(reader)
	if err != nil {
		t.Fatalf("ParseReader() error = %v", err)
	}

	obj := node.(*ast.ObjectNode)
	if len(obj.Properties()) != 2 {
		t.Errorf("expected 2 properties, got %d", len(obj.Properties()))
	}
}

func TestMustParse(t *testing.T) {
	node := MustParse("host=localhost")
	obj := node.(*ast.ObjectNode)

	if _, ok := obj.Properties()["host"]; !ok {
		t.Error("MustParse() missing 'host' property")
	}
}

func TestMustParsePanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustParse() should panic on invalid input")
		}
	}()
	MustParse("invalid input without equals")
}

// ============================================================================
// Conversion Tests
// ============================================================================

func TestMapToNode(t *testing.T) {
	m := map[string]string{
		"host": "localhost",
		"port": "8080",
	}

	node, err := MapToNode(m)
	if err != nil {
		t.Fatalf("MapToNode() error = %v", err)
	}

	obj := node.(*ast.ObjectNode)
	if len(obj.Properties()) != 2 {
		t.Errorf("expected 2 properties, got %d", len(obj.Properties()))
	}
}

func TestMapToNodeInvalidKey(t *testing.T) {
	m := map[string]string{
		"123invalid": "value",
	}

	_, err := MapToNode(m)
	if err == nil {
		t.Error("MapToNode() should fail for invalid key")
	}
}

func TestNodeToMap(t *testing.T) {
	input := "host=localhost\nport=8080"
	node, _ := Parse(input)

	m, err := NodeToMap(node)
	if err != nil {
		t.Fatalf("NodeToMap() error = %v", err)
	}

	if m["host"] != "localhost" || m["port"] != "8080" {
		t.Errorf("NodeToMap() = %v", m)
	}
}

func TestRoundTrip(t *testing.T) {
	original := map[string]string{
		"host":     "localhost",
		"port":     "8080",
		"db.name":  "myapp",
		"log-level": "info",
	}

	// map -> node -> map
	node, err := MapToNode(original)
	if err != nil {
		t.Fatalf("MapToNode() error = %v", err)
	}

	result, err := NodeToMap(node)
	if err != nil {
		t.Fatalf("NodeToMap() error = %v", err)
	}

	for k, v := range original {
		if result[k] != v {
			t.Errorf("round trip: %q = %q, want %q", k, result[k], v)
		}
	}
}

// ============================================================================
// Render Tests
// ============================================================================

func TestRender(t *testing.T) {
	input := "host=localhost\nport=8080"
	node, _ := Parse(input)

	text, err := Render(node)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	// Keys are sorted, so output is deterministic
	expected := "host=localhost\nport=8080\n"
	if text != expected {
		t.Errorf("Render() = %q, want %q", text, expected)
	}
}

func TestRenderMap(t *testing.T) {
	m := map[string]string{
		"port": "8080",
		"host": "localhost",
	}

	text, err := RenderMap(m)
	if err != nil {
		t.Fatalf("RenderMap() error = %v", err)
	}

	// Keys are sorted alphabetically
	expected := "host=localhost\nport=8080\n"
	if text != expected {
		t.Errorf("RenderMap() = %q, want %q", text, expected)
	}
}

func TestRenderEmpty(t *testing.T) {
	node, _ := Parse("")

	text, err := Render(node)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	if text != "" {
		t.Errorf("Render() = %q, want empty", text)
	}
}

func TestRenderMapEmpty(t *testing.T) {
	text, err := RenderMap(map[string]string{})
	if err != nil {
		t.Fatalf("RenderMap() error = %v", err)
	}

	if text != "" {
		t.Errorf("RenderMap() = %q, want empty", text)
	}
}

func TestRenderRoundTrip(t *testing.T) {
	original := `db.host=localhost
db.port=5432
log-level=info`

	node, _ := Parse(original)
	text, _ := Render(node)
	node2, _ := Parse(text)
	text2, _ := Render(node2)

	// After sorting, both renders should be identical
	if text != text2 {
		t.Errorf("round trip mismatch:\ngot:  %q\nwant: %q", text2, text)
	}
}

// ============================================================================
// E2E Tests
// ============================================================================

func TestLoadParseParity(t *testing.T) {
	input := "# Config\nhost=localhost\nport=8080\ndb.name=myapp\nlog-level=info"

	fast, err := Load(input)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	node, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	slow, err := NodeToMap(node)
	if err != nil {
		t.Fatalf("NodeToMap() error = %v", err)
	}

	if len(fast) != len(slow) {
		t.Fatalf("Load() returned %d keys, Parse()+NodeToMap() returned %d", len(fast), len(slow))
	}

	for k, v := range fast {
		if slow[k] != v {
			t.Errorf("key %q: Load()=%q, Parse()+NodeToMap()=%q", k, v, slow[k])
		}
	}
}

func TestCRLFAtPublicAPI(t *testing.T) {
	input := "host=localhost\r\nport=8080\r\ndebug=true"

	props, err := Load(input)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(props) != 3 {
		t.Errorf("expected 3 properties, got %d", len(props))
	}
	if props["host"] != "localhost" {
		t.Errorf("host = %q, want 'localhost'", props["host"])
	}
	if props["port"] != "8080" {
		t.Errorf("port = %q, want '8080'", props["port"])
	}
	if props["debug"] != "true" {
		t.Errorf("debug = %q, want 'true'", props["debug"])
	}

	if err := Validate(input); err != nil {
		t.Errorf("Validate() error = %v", err)
	}

	if _, err := Parse(input); err != nil {
		t.Errorf("Parse() error = %v", err)
	}
}

func TestErrorConsistency(t *testing.T) {
	cases := []string{
		"host=a\nhost=b",  // duplicate key
		"host localhost",  // missing =
		"123invalid=value", // invalid key start
	}

	for _, input := range cases {
		_, errLoad := Load(input)
		_, errParse := Parse(input)

		if errLoad == nil {
			t.Errorf("Load(%q) expected error, got nil", input)
		}
		if errParse == nil {
			t.Errorf("Parse(%q) expected error, got nil", input)
		}
	}
}

// errReader always returns an error on Read.
type errReader struct{}

func (e *errReader) Read(p []byte) (int, error) {
	return 0, errors.New("read failed")
}

func TestFullWorkflow(t *testing.T) {
	original := "host=localhost\nport=8080"

	props, err := Load(original)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	props["region"] = "us-east-1"

	node, err := MapToNode(props)
	if err != nil {
		t.Fatalf("MapToNode() error = %v", err)
	}

	text, err := Render(node)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	result, err := Load(text)
	if err != nil {
		t.Fatalf("Load(rendered) error = %v", err)
	}

	if result["host"] != "localhost" {
		t.Errorf("host = %q, want 'localhost'", result["host"])
	}
	if result["port"] != "8080" {
		t.Errorf("port = %q, want '8080'", result["port"])
	}
	if result["region"] != "us-east-1" {
		t.Errorf("region = %q, want 'us-east-1'", result["region"])
	}
}

// ============================================================================
// Coverage Gap Tests
// ============================================================================

// Reader error paths

func TestValidateReaderError(t *testing.T) {
	err := ValidateReader(&errReader{})
	if err == nil {
		t.Fatal("expected error from failing reader")
	}
}

func TestLoadReaderError(t *testing.T) {
	_, err := LoadReader(&errReader{})
	if err == nil {
		t.Fatal("expected error from failing reader")
	}
}

func TestParseReaderError(t *testing.T) {
	_, err := ParseReader(&errReader{})
	if err == nil {
		t.Fatal("expected error from failing reader")
	}
}

// NodeToMap edge cases

func TestNodeToMapNonObjectNode(t *testing.T) {
	pos := ast.NewPosition(0, 1, 1)
	lit := ast.NewLiteralNode("value", pos)
	_, err := NodeToMap(lit)
	if err == nil {
		t.Error("expected error when passing non-ObjectNode to NodeToMap")
	}
}

func TestNodeToMapNonLiteralValue(t *testing.T) {
	pos := ast.NewPosition(0, 1, 1)
	// ObjectNode as a value — not a LiteralNode
	inner := ast.NewObjectNode(map[string]ast.SchemaNode{}, pos)
	outer := ast.NewObjectNode(map[string]ast.SchemaNode{"key": inner}, pos)
	_, err := NodeToMap(outer)
	if err == nil {
		t.Error("expected error when ObjectNode value is not a LiteralNode")
	}
}

func TestNodeToMapNonStringLiteral(t *testing.T) {
	pos := ast.NewPosition(0, 1, 1)
	// LiteralNode with integer value — triggers the non-string conversion path
	lit := ast.NewLiteralNode(42, pos)
	obj := ast.NewObjectNode(map[string]ast.SchemaNode{"count": lit}, pos)
	m, err := NodeToMap(obj)
	if err != nil {
		t.Fatalf("NodeToMap() error = %v", err)
	}
	if m["count"] != "42" {
		t.Errorf("count = %q, want '42'", m["count"])
	}
}

// MapToNode edge cases (validateKey, validateValue)

func TestMapToNodeEmptyKey(t *testing.T) {
	_, err := MapToNode(map[string]string{"": "value"})
	if err == nil {
		t.Error("expected error for empty key")
	}
}

func TestMapToNodeControlCharValue(t *testing.T) {
	_, err := MapToNode(map[string]string{"key": "val\x01ue"})
	if err == nil {
		t.Error("expected error for control character in value")
	}
}

func TestMapToNodeNULValue(t *testing.T) {
	_, err := MapToNode(map[string]string{"key": "val\x00ue"})
	if err == nil {
		t.Error("expected error for NUL byte in value")
	}
}

// Render edge cases

func TestRenderNonObjectNode(t *testing.T) {
	pos := ast.NewPosition(0, 1, 1)
	lit := ast.NewLiteralNode("value", pos)
	_, err := Render(lit)
	if err == nil {
		t.Error("expected error when passing non-ObjectNode to Render")
	}
}

func TestRenderNonLiteralValue(t *testing.T) {
	pos := ast.NewPosition(0, 1, 1)
	inner := ast.NewObjectNode(map[string]ast.SchemaNode{}, pos)
	outer := ast.NewObjectNode(map[string]ast.SchemaNode{"key": inner}, pos)
	_, err := Render(outer)
	if err == nil {
		t.Error("expected error when ObjectNode value is not a LiteralNode")
	}
}

func TestRenderMapInvalidKey(t *testing.T) {
	_, err := RenderMap(map[string]string{"123bad": "value"})
	if err == nil {
		t.Error("expected error for invalid key in RenderMap")
	}
}

func TestRenderMapInvalidValue(t *testing.T) {
	_, err := RenderMap(map[string]string{"key": "val\x01ue"})
	if err == nil {
		t.Error("expected error for control character in RenderMap value")
	}
}

func TestRenderMapLargeBuffer(t *testing.T) {
	// Produce >64KB output to trigger the large-buffer discard path in putBuffer
	m := map[string]string{
		"key": strings.Repeat("x", 65537),
	}
	text, err := RenderMap(m)
	if err != nil {
		t.Fatalf("RenderMap() error = %v", err)
	}
	if !strings.HasPrefix(text, "key=") {
		t.Errorf("unexpected output prefix: %q", text[:10])
	}
}
