# Architecture

This document describes the dual-path parser architecture for the Simple Properties
Configuration Format, following the pattern established in shape-json.

## Overview

The shape-properties package implements a dual-path parsing architecture that provides:

- **Fast Path**: Direct byte-level parsing returning `map[string]string`
- **AST Path**: Full AST construction using Shape Core types

This design optimizes for the common case (config loading) while supporting advanced
use cases (format conversion, tree manipulation).

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           Public API                                     │
│  pkg/properties/properties.go                                            │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  Fast Path                           │  AST Path                         │
│  ──────────                          │  ────────                         │
│  Validate(string) error              │  Parse(string) (ast.SchemaNode)   │
│  ValidateReader(io.Reader) error     │  ParseReader(io.Reader)           │
│  Load(string) map[string]string      │                                   │
│  LoadReader(io.Reader)               │  Render(ast.SchemaNode) string    │
│                                       │  MapToNode(map) ast.SchemaNode   │
│                                       │  NodeToMap(ast.SchemaNode) map   │
│                                       │                                   │
└─────────────────────────────────────────────────────────────────────────┘
          │                                      │
          ▼                                      ▼
┌─────────────────────────────┐    ┌─────────────────────────────────────┐
│   internal/fastparser       │    │   internal/parser                    │
│   ─────────────────────     │    │   ───────────────                    │
│   Zero-allocation parser    │    │   LL(1) recursive descent parser     │
│   Direct byte → map         │    │   Tokenizer → AST                    │
│   No AST construction       │    │   Full position tracking             │
│   Optimized for speed       │    │   Rich error messages                │
└─────────────────────────────┘    └─────────────────────────────────────┘
                                                 │
                                                 ▼
                                   ┌─────────────────────────────────────┐
                                   │   internal/tokenizer                 │
                                   │   ──────────────────                 │
                                   │   Token types: KEY, EQUALS, VALUE,   │
                                   │   COMMENT, NEWLINE, EOF              │
                                   │   UTF-8 aware                        │
                                   │   Position tracking                  │
                                   └─────────────────────────────────────┘
```

## Package Structure

```
shape-properties/
├── pkg/
│   └── properties/
│       ├── properties.go     # Public API entry points
│       ├── convert.go        # AST ↔ map conversion
│       └── render.go         # AST → text rendering
├── internal/
│   ├── tokenizer/
│   │   ├── tokens.go         # Token type definitions
│   │   └── tokenizer.go      # Tokenizer implementation
│   ├── parser/
│   │   └── parser.go         # AST parser (LL(1))
│   └── fastparser/
│       └── parser.go         # Fast path parser
├── docs/
│   └── grammar/
│       └── properties.ebnf   # Formal grammar
└── testdata/
    └── benchmarks/
        ├── small.properties  # ~10 properties
        ├── medium.properties # ~500 properties
        └── large.properties  # ~10,000 properties
```

## Dual-Path Design

### When to Use Each Path

| Use Case | Path | Function |
|----------|------|----------|
| Load configuration | Fast | `Load()`, `LoadReader()` |
| Validate input | Fast | `Validate()`, `ValidateReader()` |
| Format conversion | AST | `Parse()` → transform → `Render()` |
| Tree manipulation | AST | `Parse()`, access `ast.ObjectNode` |
| Generate properties | Both | `RenderMap()` or `MapToNode()` → `Render()` |

### Performance Characteristics

| Path | Time | Memory | Allocations | Use Case |
|------|------|--------|-------------|----------|
| Fast | Baseline | Baseline | Baseline | Config loading |
| AST | 5-10x slower | 5-10x more | 3-6x more | Tree operations |

## Token Types

The tokenizer produces these token types:

| Token | Pattern | Example |
|-------|---------|---------|
| KEY | `[A-Za-z_][A-Za-z0-9_.-]*` | `db.host`, `log-level` |
| EQUALS | `=` | `=` |
| VALUE | `.*` (until newline) | `localhost` |
| COMMENT | `#.*` | `# comment` |
| NEWLINE | `\n` or `\r\n` | |
| EOF | End of input | |

## AST Representation

A properties file is represented as a single `*ast.ObjectNode` containing
`*ast.LiteralNode` values (all strings):

```go
// Input:
// host=localhost
// port=8080

// AST:
*ast.ObjectNode{
    Properties: map[string]ast.SchemaNode{
        "host": *ast.LiteralNode{Value: "localhost"},
        "port": *ast.LiteralNode{Value: "8080"},
    }
}
```

## Error Handling

The parser detects and reports these error conditions:

1. **Missing `=` separator** - `expected '=' after key`
2. **Empty key** - `empty key`
3. **Invalid key characters** - `invalid key start character`
4. **Duplicate keys** - `duplicate key "x"`
5. **Control characters** - `invalid control character 0x01`
6. **NUL byte** - `NUL byte not allowed`

Errors include line numbers for easy debugging.

## Validation Flow

```
Input String
     │
     ▼
┌─────────────┐
│ Skip WS     │
└─────────────┘
     │
     ▼
┌─────────────┐     ┌─────────────┐
│ Comment?    │────►│ Skip line   │
└─────────────┘ yes └─────────────┘
     │ no
     ▼
┌─────────────┐
│ Parse Key   │◄─── Must match [A-Za-z_][A-Za-z0-9_.-]*
└─────────────┘
     │
     ▼
┌─────────────┐
│ Expect '='  │
└─────────────┘
     │
     ▼
┌─────────────┐
│ Parse Value │◄─── Everything until \n, validate for control chars
└─────────────┘
     │
     ▼
┌─────────────┐
│ Check Dup   │◄─── Track seen keys
└─────────────┘
     │
     ▼
┌─────────────┐
│ Store       │
└─────────────┘
```

## Design Decisions

### Why Dual-Path?

1. **Performance**: Config loading is hot path, AST overhead unnecessary
2. **Flexibility**: Tree operations need proper AST for traversal
3. **Consistency**: Same pattern as shape-json enables code reuse

### Why No Streaming Tokenizer in Fast Path?

The fast path parses directly from bytes to avoid:
- Token object allocation
- String interning overhead
- Virtual function calls

For small-medium files (<100KB), batch processing is faster than streaming.

### Why Sorted Keys in Render?

Deterministic output enables:
- Reliable diff/comparison
- Reproducible test assertions
- Consistent version control diffs

## Testing Strategy

| Test Type | Location | Purpose |
|-----------|----------|---------|
| Unit tests | `*_test.go` | Correctness |
| Fuzz tests | `*_fuzz_test.go` | Crash resistance |
| Benchmarks | `*_bench_test.go` | Performance tracking |
| Integration | `pkg/properties/properties_test.go` | API behavior |

## Dependencies

- `github.com/shapestone/shape-core/pkg/ast` - AST node types
- Standard library only for internal packages
