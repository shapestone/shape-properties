// Package fastparser implements a high-performance properties parser without AST construction.
//
// This parser is optimized for the common case of loading properties directly into a map.
// It bypasses tokenization and AST construction, parsing directly from bytes to values.
//
// Performance targets (vs AST parser):
//   - 5-10x faster parsing
//   - 5-10x less memory
//   - 3-6x fewer allocations
package fastparser

import (
	"errors"
	"fmt"
	"io"
)

// Parser implements a high-performance properties parser that builds map[string]string directly.
type Parser struct {
	data     []byte
	pos      int
	length   int
	line     int
	seenKeys map[string]struct{} // For duplicate detection
}

// NewParser creates a new fast parser for the given data.
func NewParser(data []byte) *Parser {
	return &Parser{
		data:     data,
		pos:      0,
		length:   len(data),
		line:     1,
		seenKeys: make(map[string]struct{}),
	}
}

// NewParserFromReader creates a new fast parser from an io.Reader.
func NewParserFromReader(r io.Reader) (*Parser, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return NewParser(data), nil
}

// Parse parses the properties data and returns map[string]string.
func (p *Parser) Parse() (map[string]string, error) {
	result := make(map[string]string)

	for p.pos < p.length {
		// Skip whitespace at start of line
		p.skipWhitespace()

		if p.pos >= p.length {
			break
		}

		c := p.data[p.pos]

		// Handle newlines (empty lines)
		if c == '\n' {
			p.pos++
			p.line++
			continue
		}
		if c == '\r' {
			p.pos++
			if p.pos < p.length && p.data[p.pos] == '\n' {
				p.pos++
			}
			p.line++
			continue
		}

		// Handle comments
		if c == '#' {
			p.skipToEndOfLine()
			continue
		}

		// Parse key=value
		key, value, err := p.parseProperty()
		if err != nil {
			return nil, err
		}

		// Check for duplicate keys
		if _, exists := p.seenKeys[key]; exists {
			return nil, fmt.Errorf("duplicate key %q at line %d", key, p.line)
		}
		p.seenKeys[key] = struct{}{}
		result[key] = value
	}

	return result, nil
}

// Validate parses and validates without returning the result.
// Returns nil if valid, error otherwise.
func (p *Parser) Validate() error {
	_, err := p.Parse()
	return err
}

// parseProperty parses a single key=value line.
func (p *Parser) parseProperty() (string, string, error) {
	startLine := p.line

	// Parse key
	key, err := p.parseKey()
	if err != nil {
		return "", "", err
	}

	// Skip whitespace before =
	p.skipWhitespace()

	// Expect =
	if p.pos >= p.length || p.data[p.pos] != '=' {
		return "", "", fmt.Errorf("expected '=' after key %q at line %d", key, startLine)
	}
	p.pos++ // skip '='

	// Skip whitespace after =
	p.skipWhitespace()

	// Parse value
	value, err := p.parseValue()
	if err != nil {
		return "", "", fmt.Errorf("invalid value for key %q at line %d: %w", key, startLine, err)
	}

	return key, value, nil
}

// parseKey parses and validates a property key.
func (p *Parser) parseKey() (string, error) {
	start := p.pos

	// First character must be letter or underscore
	if p.pos >= p.length {
		return "", fmt.Errorf("unexpected end of input at line %d", p.line)
	}

	c := p.data[p.pos]
	if !isKeyStart(c) {
		return "", fmt.Errorf("invalid key start character %q at line %d", string(c), p.line)
	}
	p.pos++

	// Subsequent characters: letter, digit, underscore, dash, dot
	for p.pos < p.length {
		c := p.data[p.pos]
		if !isKeyChar(c) {
			break
		}
		p.pos++
	}

	if p.pos == start {
		return "", fmt.Errorf("empty key at line %d", p.line)
	}

	return string(p.data[start:p.pos]), nil
}

// parseValue parses the value until end of line.
func (p *Parser) parseValue() (string, error) {
	start := p.pos

	// Find end of line
	for p.pos < p.length {
		c := p.data[p.pos]
		if c == '\n' || c == '\r' {
			break
		}
		// Validate: no NUL bytes
		if c == 0x00 {
			return "", errors.New("NUL byte not allowed")
		}
		// Validate: no control characters except TAB
		if c < 0x20 && c != '\t' {
			return "", fmt.Errorf("invalid control character 0x%02x", c)
		}
		p.pos++
	}

	// Get the raw value
	value := p.data[start:p.pos]

	// Trim trailing whitespace
	end := len(value)
	for end > 0 && (value[end-1] == ' ' || value[end-1] == '\t') {
		end--
	}

	// Skip the newline
	if p.pos < p.length {
		if p.data[p.pos] == '\r' {
			p.pos++
			if p.pos < p.length && p.data[p.pos] == '\n' {
				p.pos++
			}
		} else if p.data[p.pos] == '\n' {
			p.pos++
		}
		p.line++
	}

	return string(value[:end]), nil
}

// skipWhitespace skips spaces and tabs (not newlines).
func (p *Parser) skipWhitespace() {
	for p.pos < p.length {
		c := p.data[p.pos]
		if c != ' ' && c != '\t' {
			break
		}
		p.pos++
	}
}

// skipToEndOfLine skips to the end of the current line.
func (p *Parser) skipToEndOfLine() {
	for p.pos < p.length {
		c := p.data[p.pos]
		if c == '\n' {
			p.pos++
			p.line++
			return
		}
		if c == '\r' {
			p.pos++
			if p.pos < p.length && p.data[p.pos] == '\n' {
				p.pos++
			}
			p.line++
			return
		}
		p.pos++
	}
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
