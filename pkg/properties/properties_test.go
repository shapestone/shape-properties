package properties

import (
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
