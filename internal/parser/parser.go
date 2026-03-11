// Package parser implements LL(1) recursive descent parsing for the properties format.
// Each production rule in the grammar corresponds to a parse function.
package parser

import (
	"fmt"
	"io"

	"github.com/shapestone/shape-core/pkg/ast"
	"github.com/shapestone/shape-properties/internal/tokenizer"
)

// Parser implements LL(1) recursive descent parsing for properties files.
type Parser struct {
	tokenizer *tokenizer.Tokenizer
	current   tokenizer.Token
	hasToken  bool
	seenKeys  map[string]bool // For duplicate key detection
}

// NewParser creates a new properties parser for the given input string.
func NewParser(input string) *Parser {
	p := &Parser{
		tokenizer: tokenizer.NewTokenizer(input),
		seenKeys:  make(map[string]bool),
	}
	p.advance() // Load first token
	return p
}

// NewParserFromReader creates a new properties parser from an io.Reader.
func NewParserFromReader(r io.Reader) (*Parser, error) {
	tok, err := tokenizer.NewTokenizerFromReader(r)
	if err != nil {
		return nil, err
	}
	p := &Parser{
		tokenizer: tok,
		seenKeys:  make(map[string]bool),
	}
	p.advance()
	return p, nil
}

// Parse parses the input and returns an AST representing the properties file.
//
// Grammar:
//   file = { line } ;
//   line = ws | comment | assignment ;
//   assignment = ws key ws "=" ws value ws ;
//
// Returns *ast.ObjectNode with string literal values.
func (p *Parser) Parse() (ast.SchemaNode, error) {
	properties := make(map[string]ast.SchemaNode)
	startPos := ast.NewPosition(0, 1, 1)

	for p.hasToken {
		switch p.current.Kind {
		case tokenizer.TokenComment:
			// Skip comments
			p.advance()

		case tokenizer.TokenNewline:
			// Skip empty lines
			p.advance()

		case tokenizer.TokenKey:
			// Parse property assignment
			key, value, err := p.parseAssignment()
			if err != nil {
				return nil, err
			}

			// Check for duplicate keys
			if p.seenKeys[key] {
				return nil, fmt.Errorf("duplicate key %q at line %d", key, p.current.Line)
			}
			p.seenKeys[key] = true
			properties[key] = value

		default:
			return nil, fmt.Errorf("unexpected token %q at line %d, column %d",
				p.current.Kind, p.current.Line, p.current.Column)
		}
	}

	return ast.NewObjectNode(properties, startPos), nil
}

// parseAssignment parses a property assignment: key = value
func (p *Parser) parseAssignment() (string, ast.SchemaNode, error) {
	// Key
	if p.current.Kind != tokenizer.TokenKey {
		return "", nil, fmt.Errorf("expected key at line %d, column %d, got %q",
			p.current.Line, p.current.Column, p.current.Kind)
	}

	key := p.current.Value
	keyPos := ast.NewPosition(p.current.Offset, p.current.Line, p.current.Column)

	// Validate key format
	if err := validateKey(key); err != nil {
		return "", nil, fmt.Errorf("invalid key %q at line %d: %w",
			key, p.current.Line, err)
	}

	p.advance()

	// Equals sign
	if p.current.Kind != tokenizer.TokenEquals {
		return "", nil, fmt.Errorf("expected '=' after key %q at line %d, column %d",
			key, p.current.Line, p.current.Column)
	}
	// Don't call advance() here - ScanValue() will read directly from the tokenizer
	// at its current position (right after the '=')

	// Value (scan until end of line)
	valueToken := p.tokenizer.ScanValue()
	value := valueToken.Value

	// Validate value for control characters
	if err := validateValue(value); err != nil {
		return "", nil, fmt.Errorf("invalid value for key %q at line %d: %w",
			key, valueToken.Line, err)
	}

	// Create literal node for value
	valueNode := ast.NewLiteralNode(value, keyPos)

	// Skip to next line
	p.advance()

	return key, valueNode, nil
}

// advance moves to the next token.
func (p *Parser) advance() {
	p.current = p.tokenizer.NextToken()
	p.hasToken = p.current.Kind != tokenizer.TokenEOF
}

// validateKey validates a property key according to the spec.
// Keys must match: [A-Za-z_][A-Za-z0-9_.-]*
func validateKey(key string) error {
	if len(key) == 0 {
		return fmt.Errorf("empty key")
	}

	first := key[0]
	if !isKeyStart(first) {
		return fmt.Errorf("key must start with letter or underscore, got %q", string(first))
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
		// NUL byte is always invalid (check first for specific error message)
		if c == 0x00 {
			return fmt.Errorf("NUL byte not allowed")
		}
		// Control characters other than TAB are invalid
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
