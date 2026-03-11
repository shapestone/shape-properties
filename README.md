# shape-properties

![Build Status](https://github.com/shapestone/shape-properties/actions/workflows/ci.yml/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/shapestone/shape-properties)](https://goreportcard.com/report/github.com/shapestone/shape-properties)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![codecov](https://codecov.io/gh/shapestone/shape-properties/branch/main/graph/badge.svg)](https://codecov.io/gh/shapestone/shape-properties)
![Go Version](https://img.shields.io/github/go-mod/go-version/shapestone/shape-properties)
![Latest Release](https://img.shields.io/github/v/release/shapestone/shape-properties)
[![GoDoc](https://pkg.go.dev/badge/github.com/shapestone/shape-properties.svg)](https://pkg.go.dev/github.com/shapestone/shape-properties)
[![CodeQL](https://github.com/shapestone/shape-properties/actions/workflows/codeql.yml/badge.svg)](https://github.com/shapestone/shape-properties/actions/workflows/codeql.yml)
[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/shapestone/shape-properties/badge)](https://securityscorecards.dev/viewer/?uri=github.com/shapestone/shape-properties)
[![Security Policy](https://img.shields.io/badge/Security-Policy-brightgreen)](SECURITY.md)

**Repository:** github.com/shapestone/shape-properties

A Simple Properties Configuration Format parser for the [Shape Parser™](https://github.com/shapestone/shape) ecosystem.

Parses `.properties` configuration files into Shape Parser's™ unified AST representation.

## Installation

```bash
go get github.com/shapestone/shape-properties
```

## Quick Start

```go
import "github.com/shapestone/shape-properties/pkg/properties"

// Load configuration (fast path - recommended for config loading)
props, err := properties.Load(`
host=localhost
port=8080
debug=true
`)
if err != nil {
    log.Fatal(err)
}
fmt.Println(props["host"]) // localhost

// Or parse to AST (for tree manipulation)
node, err := properties.Parse(`host=localhost`)
```

## Format Specification

The Simple Properties Configuration Format uses `key=value` pairs:

```properties
# Database configuration
db.host=localhost
db.port=5432
db.name=myapp

# Application settings
log-level=info
timeout=30
```

### Rules

- Keys must match `[A-Za-z_][A-Za-z0-9_.-]*`
- Keys are case-sensitive
- Values extend to end of line (no inline comments)
- Leading/trailing whitespace around keys and values is trimmed
- Comments start with `#` at the beginning of a line
- Duplicate keys are an error

See [properties-format.md](properties-format.md) for the complete specification.

## API Reference

### Fast Path (Performance Optimized)

Use these functions for configuration loading and validation:

```go
// Validate input without parsing
err := properties.Validate(input)
err := properties.ValidateReader(reader)

// Load into map[string]string
props, err := properties.Load(input)
props, err := properties.LoadReader(reader)

// Panic on error (for tests/init)
props := properties.MustLoad(input)
```

### AST Path (Full Feature Set)

Use these functions for tree manipulation and format conversion:

```go
// Parse to AST
node, err := properties.Parse(input)
node, err := properties.ParseReader(reader)
node := properties.MustParse(input)

// Convert between AST and map
node, err := properties.MapToNode(map[string]string{"host": "localhost"})
props, err := properties.NodeToMap(node)

// Render to text (sorted keys)
text, err := properties.Render(node)
text, err := properties.RenderMap(map[string]string{"host": "localhost"})
```

## Dual-Path Architecture

| Path | Returns | Use Case | Performance |
|------|---------|----------|-------------|
| Fast | `map[string]string` | Config loading, validation | Baseline |
| AST | `ast.SchemaNode` | Tree manipulation, conversion | 5-10x slower |

For configuration loading (the common case), use `Load()` or `Validate()`.
For format conversion or tree manipulation, use `Parse()`.

## Examples

### Loading Configuration

```go
// From string
props, _ := properties.Load("host=localhost\nport=8080")

// From file
file, _ := os.Open("config.properties")
defer file.Close()
props, _ := properties.LoadReader(file)

// Access values
host := props["host"]
port := props["port"]
```

### Validation

```go
if err := properties.Validate(userInput); err != nil {
    return fmt.Errorf("invalid configuration: %w", err)
}
```

### Generating Properties

```go
config := map[string]string{
    "host":     "localhost",
    "port":     "8080",
    "db.name":  "myapp",
}

text, _ := properties.RenderMap(config)
os.WriteFile("config.properties", []byte(text), 0644)
```

### AST Manipulation

```go
// Parse to AST
node, _ := properties.Parse(input)
obj := node.(*ast.ObjectNode)

// Access properties
for key, valueNode := range obj.Properties() {
    lit := valueNode.(*ast.LiteralNode)
    fmt.Printf("%s = %v\n", key, lit.Value())
}

// Render back to text
text, _ := properties.Render(node)
```

## Error Handling

The parser reports detailed errors with line numbers:

```go
_, err := properties.Load("123invalid=value")
// Error: invalid key start character "1" at line 1

_, err := properties.Load("host=localhost\nhost=other")
// Error: duplicate key "host" at line 2

_, err := properties.Load("key=value\x00more")
// Error: NUL byte not allowed
```

## Benchmarks

Run benchmarks:

```bash
make bench
```

Expected performance on typical hardware:

| Operation | Small (10 props) | Medium (500 props) | Large (10K props) |
|-----------|------------------|--------------------|--------------------|
| Load | ~5 µs | ~200 µs | ~4 ms |
| Parse | ~25 µs | ~1 ms | ~20 ms |

Fast path is 5-10x faster than AST path.

## Testing

```bash
# Run all tests
make test

# Run with race detection
go test -race ./...

# Fuzz testing
make fuzz

# Coverage report
make coverage
```

## Thread Safety

All public functions are safe for concurrent use:

- **`Load`, `Validate`, `Parse`** and their variants each create a new parser instance per call — no shared state.
- **`Render`, `RenderMap`** use a `sync.Pool` of `bytes.Buffer` instances for zero-contention buffer reuse.
- No package-level mutable state exists outside the buffer pool, which is itself goroutine-safe.

```go
// Safe to call concurrently from multiple goroutines
var wg sync.WaitGroup
for _, cfg := range configs {
    wg.Add(1)
    go func(input string) {
        defer wg.Done()
        props, _ := properties.Load(input)
        _ = props
    }(cfg)
}
wg.Wait()
```

## Dependencies

- [shape-core](https://github.com/shapestone/shape-core) - AST types (`ast.ObjectNode`, `ast.LiteralNode`)

## License

Apache License 2.0

Copyright 2020-2025 Shapestone
