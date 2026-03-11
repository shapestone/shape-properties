// Package properties provides a parser for the Simple Properties Configuration Format.
//
// This package implements a dual-path parsing architecture:
//   - Fast path: Direct parsing to map[string]string for validation and loading
//   - AST path: Full AST construction for tree manipulation and format conversion
//
// The format supports key=value pairs with comments:
//
//	# Database configuration
//	host=localhost
//	port=8080
//	db.name=myapp
//
// See properties-format.md for the complete specification.
package properties

import (
	"io"

	"github.com/shapestone/shape-core/pkg/ast"
	"github.com/shapestone/shape-properties/internal/fastparser"
	"github.com/shapestone/shape-properties/internal/parser"
)

// ============================================================================
// Fast Path API - Performance optimized
// ============================================================================

// Validate validates a properties string without returning the parsed result.
// This is the fastest way to check if input is valid.
//
// Example:
//
//	if err := properties.Validate(input); err != nil {
//	    log.Fatal(err)
//	}
func Validate(input string) error {
	p := fastparser.NewParserFromString(input)
	return p.Validate()
}

// ValidateReader validates properties from an io.Reader without returning the parsed result.
//
// Example:
//
//	file, _ := os.Open("config.properties")
//	if err := properties.ValidateReader(file); err != nil {
//	    log.Fatal(err)
//	}
func ValidateReader(r io.Reader) error {
	p, err := fastparser.NewParserFromReader(r)
	if err != nil {
		return err
	}
	return p.Validate()
}

// Load parses a properties string and returns a map[string]string.
// This is the recommended way to load configuration files.
//
// Example:
//
//	props, err := properties.Load(`host=localhost\nport=8080`)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(props["host"]) // localhost
func Load(input string) (map[string]string, error) {
	p := fastparser.NewParserFromString(input)
	return p.Parse()
}

// LoadReader parses properties from an io.Reader and returns a map[string]string.
//
// Example:
//
//	file, _ := os.Open("config.properties")
//	props, err := properties.LoadReader(file)
//	if err != nil {
//	    log.Fatal(err)
//	}
func LoadReader(r io.Reader) (map[string]string, error) {
	p, err := fastparser.NewParserFromReader(r)
	if err != nil {
		return nil, err
	}
	return p.Parse()
}

// ============================================================================
// AST Path API - Full feature set
// ============================================================================

// Parse parses a properties string and returns an AST.
// Use this when you need tree manipulation or format conversion.
//
// Example:
//
//	node, err := properties.Parse(`host=localhost`)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	obj := node.(*ast.ObjectNode)
//	for key, val := range obj.Properties() {
//	    fmt.Printf("%s = %v\n", key, val.(*ast.LiteralNode).Value())
//	}
func Parse(input string) (ast.SchemaNode, error) {
	p := parser.NewParser(input)
	return p.Parse()
}

// ParseReader parses properties from an io.Reader and returns an AST.
//
// Example:
//
//	file, _ := os.Open("config.properties")
//	node, err := properties.ParseReader(file)
func ParseReader(r io.Reader) (ast.SchemaNode, error) {
	p, err := parser.NewParserFromReader(r)
	if err != nil {
		return nil, err
	}
	return p.Parse()
}

// MustParse parses or panics (useful for tests and initialization).
//
// Example:
//
//	node := properties.MustParse(`host=localhost`)
func MustParse(input string) ast.SchemaNode {
	node, err := Parse(input)
	if err != nil {
		panic(err)
	}
	return node
}

// MustLoad loads or panics (useful for tests and initialization).
//
// Example:
//
//	props := properties.MustLoad(`host=localhost`)
func MustLoad(input string) map[string]string {
	props, err := Load(input)
	if err != nil {
		panic(err)
	}
	return props
}
