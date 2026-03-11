package properties

import (
	"bytes"
	"fmt"
	"sort"
	"sync"

	"github.com/shapestone/shape-core/pkg/ast"
)

// renderPool pools bytes.Buffer instances to avoid repeated allocations in
// high-throughput rendering scenarios (e.g. config generation loops).
var renderPool = sync.Pool{
	New: func() interface{} {
		return bytes.NewBuffer(make([]byte, 0, 512))
	},
}

// putBuffer returns a buffer to the pool, discarding oversized buffers to
// prevent the pool from holding onto large allocations permanently.
func putBuffer(buf *bytes.Buffer) {
	if buf.Cap() > 64*1024 {
		return
	}
	buf.Reset()
	renderPool.Put(buf)
}

// Render converts an AST ObjectNode back to properties format text.
// Keys are sorted alphabetically for deterministic output.
//
// Example:
//
//	node, _ := properties.Parse(`host=localhost\nport=8080`)
//	text, err := properties.Render(node)
//	// text = "host=localhost\nport=8080\n"
func Render(node ast.SchemaNode) (string, error) {
	obj, ok := node.(*ast.ObjectNode)
	if !ok {
		return "", fmt.Errorf("expected ObjectNode, got %T", node)
	}

	props := obj.Properties()
	if len(props) == 0 {
		return "", nil
	}

	// Sort keys for deterministic output
	keys := make([]string, 0, len(props))
	for k := range props {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	buf := renderPool.Get().(*bytes.Buffer)
	defer putBuffer(buf)

	for _, key := range keys {
		valueNode := props[key]

		lit, ok := valueNode.(*ast.LiteralNode)
		if !ok {
			return "", fmt.Errorf("expected LiteralNode for key %q, got %T", key, valueNode)
		}

		var value string
		if lit.Value() == nil {
			value = ""
		} else if strVal, ok := lit.Value().(string); ok {
			value = strVal
		} else {
			value = fmt.Sprintf("%v", lit.Value())
		}

		buf.WriteString(key)
		buf.WriteByte('=')
		buf.WriteString(value)
		buf.WriteByte('\n')
	}

	return buf.String(), nil
}

// RenderMap converts a map[string]string to properties format text.
// Keys are sorted alphabetically for deterministic output.
//
// Example:
//
//	m := map[string]string{"host": "localhost", "port": "8080"}
//	text, err := properties.RenderMap(m)
//	// text = "host=localhost\nport=8080\n"
func RenderMap(m map[string]string) (string, error) {
	if len(m) == 0 {
		return "", nil
	}

	// Validate all entries
	for key, value := range m {
		if err := validateKey(key); err != nil {
			return "", fmt.Errorf("invalid key %q: %w", key, err)
		}
		if err := validateValue(value); err != nil {
			return "", fmt.Errorf("invalid value for key %q: %w", key, err)
		}
	}

	// Sort keys for deterministic output
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	buf := renderPool.Get().(*bytes.Buffer)
	defer putBuffer(buf)

	for _, key := range keys {
		buf.WriteString(key)
		buf.WriteByte('=')
		buf.WriteString(m[key])
		buf.WriteByte('\n')
	}

	return buf.String(), nil
}
