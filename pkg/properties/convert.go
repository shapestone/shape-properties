package properties

import (
	"fmt"

	"github.com/shapestone/shape-core/pkg/ast"
)

// MapToNode converts a map[string]string to an AST ObjectNode.
// All values are stored as LiteralNode with string values.
//
// Example:
//
//	m := map[string]string{"host": "localhost", "port": "8080"}
//	node, err := properties.MapToNode(m)
func MapToNode(m map[string]string) (ast.SchemaNode, error) {
	properties := make(map[string]ast.SchemaNode, len(m))
	pos := ast.NewPosition(0, 1, 1)

	for key, value := range m {
		// Validate key format
		if err := validateKey(key); err != nil {
			return nil, fmt.Errorf("invalid key %q: %w", key, err)
		}
		// Validate value
		if err := validateValue(value); err != nil {
			return nil, fmt.Errorf("invalid value for key %q: %w", key, err)
		}
		properties[key] = ast.NewLiteralNode(value, pos)
	}

	return ast.NewObjectNode(properties, pos), nil
}

// NodeToMap converts an AST ObjectNode to a map[string]string.
// Only works with ObjectNode containing LiteralNode values with string values.
//
// Example:
//
//	node, _ := properties.Parse(`host=localhost`)
//	m, err := properties.NodeToMap(node)
//	fmt.Println(m["host"]) // localhost
func NodeToMap(node ast.SchemaNode) (map[string]string, error) {
	obj, ok := node.(*ast.ObjectNode)
	if !ok {
		return nil, fmt.Errorf("expected ObjectNode, got %T", node)
	}

	result := make(map[string]string, len(obj.Properties()))

	for key, valueNode := range obj.Properties() {
		lit, ok := valueNode.(*ast.LiteralNode)
		if !ok {
			return nil, fmt.Errorf("expected LiteralNode for key %q, got %T", key, valueNode)
		}

		strValue, ok := lit.Value().(string)
		if !ok {
			// Convert non-string values to string representation
			strValue = fmt.Sprintf("%v", lit.Value())
		}

		result[key] = strValue
	}

	return result, nil
}

// validateKey validates a property key according to the spec.
func validateKey(key string) error {
	if len(key) == 0 {
		return fmt.Errorf("empty key")
	}

	first := key[0]
	if !isKeyStart(first) {
		return fmt.Errorf("key must start with letter or underscore")
	}

	for i := 1; i < len(key); i++ {
		if !isKeyChar(key[i]) {
			return fmt.Errorf("invalid character %q in key", string(key[i]))
		}
	}

	return nil
}

// validateValue checks for invalid control characters in values.
func validateValue(value string) error {
	for i := 0; i < len(value); i++ {
		c := value[i]
		if c == 0x00 {
			return fmt.Errorf("NUL byte not allowed")
		}
		if c < 0x20 && c != '\t' {
			return fmt.Errorf("invalid control character 0x%02x", c)
		}
	}
	return nil
}

// isKeyStart returns true if the byte can start a key.
func isKeyStart(c byte) bool {
	return (c >= 'A' && c <= 'Z') ||
		(c >= 'a' && c <= 'z') ||
		c == '_'
}

// isKeyChar returns true if the byte can be part of a key.
func isKeyChar(c byte) bool {
	return isKeyStart(c) ||
		(c >= '0' && c <= '9') ||
		c == '-' ||
		c == '.'
}
