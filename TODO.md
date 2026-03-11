# shape-properties TODO

## Feature: Dual-Path Parser Implementation

Implement a dual-path parser architecture for the Simple Properties Configuration Format,
following the pattern established in shape-json. This provides both high-performance
parsing for common use cases and full AST support for advanced scenarios.

### Reference
- Format specification: `properties-format.md`
- Architecture reference: `shape-json` dual-path implementation

### Requirements
- Fast path: Direct byte-level parsing returning `map[string]string`
- AST path: Full AST construction using Shape Core types
- Shared tokenizer for AST path
- Public API routing based on caller needs
- Streaming support via `io.Reader`
- Full compliance with `properties-format.md` specification

---

## Architecture Overview

### Dual-Path Design

| Path | Purpose | Returns | Used By |
|------|---------|---------|---------|
| **Fast Path** | Validation, config loading | `map[string]string` | `Validate()`, `Load()` |
| **AST Path** | Tree manipulation, conversion | `ast.SchemaNode` | `Parse()`, format tools |

### Public API

```go
// Fast path - performance optimized
func Validate(input string) error
func ValidateReader(r io.Reader) error
func Load(input string) (map[string]string, error)
func LoadReader(r io.Reader) (map[string]string, error)

// AST path - full feature set
func Parse(input string) (ast.SchemaNode, error)
func ParseReader(r io.Reader) (ast.SchemaNode, error)

// Rendering
func Render(node ast.SchemaNode) (string, error)

// Conversion
func MapToNode(m map[string]string) (ast.SchemaNode, error)
func NodeToMap(node ast.SchemaNode) (map[string]string, error)
```

---

## Files to Create/Modify

**New Files:**

### Internal - Fast Parser
- `internal/fastparser/parser.go` - Zero-copy byte-level parser
- `internal/fastparser/parser_test.go` - Fast path tests

### Internal - AST Parser
- `internal/parser/parser.go` - LL(1) recursive descent parser
- `internal/parser/parser_test.go` - AST path tests

### Internal - Tokenizer
- `internal/tokenizer/tokenizer.go` - Shape-based tokenization
- `internal/tokenizer/tokens.go` - Token type definitions (KEY, EQUALS, VALUE, COMMENT, NEWLINE, EOF)
- `internal/tokenizer/tokenizer_test.go` - Tokenizer tests

### Public API
- `pkg/props/convert.go` - AST ↔ map[string]string conversion
- `pkg/props/render.go` - AST → properties text rendering
- `pkg/props/props_test.go` - Integration tests

### Documentation
- `docs/grammar/props.ebnf` - Formal EBNF grammar (extract from properties-format.md)
- `ARCHITECTURE.md` - Architecture documentation

**Modified Files:**
- `pkg/props/props.go` - Add public API routing
- `README.md` - Update with new API documentation
- `go.mod` - Add shape-core dependency if needed

---

## Implementation Phases

### Phase 1: Foundation
- [ ] Set up directory structure
- [ ] Define token types
- [ ] Implement tokenizer with Shape framework
- [ ] Write tokenizer tests

### Phase 2: AST Parser
- [ ] Implement LL(1) parser
- [ ] Define AST structure (ObjectNode with string properties)
- [ ] Write parser tests
- [ ] Implement `Parse()` and `ParseReader()`

### Phase 3: Fast Parser
- [ ] Implement byte-level parser
- [ ] Optimize for common cases (no escapes, short keys/values)
- [ ] Write fast parser tests
- [ ] Implement `Validate()`, `Load()`, and reader variants

### Phase 4: Conversion & Rendering
- [ ] Implement `MapToNode()` and `NodeToMap()`
- [ ] Implement `Render()` for AST → text
- [ ] Write conversion/rendering tests

### Phase 5: Polish
- [ ] Add streaming support for large files
- [ ] Write integration tests
- [ ] Add benchmarks comparing fast vs AST path
- [ ] Write ARCHITECTURE.md
- [ ] Update README.md

---

## Token Types

```go
const (
    TokenKey     = "KEY"      // [A-Za-z_][A-Za-z0-9_.-]*
    TokenEquals  = "="        // Assignment separator
    TokenValue   = "VALUE"    // Everything after = until newline
    TokenComment = "COMMENT"  // # to end of line
    TokenNewline = "NEWLINE"  // \n or \r\n
    TokenEOF     = "EOF"      // End of input
)
```

---

## AST Representation

Properties file maps to a single `*ast.ObjectNode`:

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

---

## Error Cases (from properties-format.md §4)

Must detect and report:
- [ ] Missing `=` separator
- [ ] Empty key
- [ ] Invalid key characters
- [ ] Duplicate keys
- [ ] Control characters (except TAB)
- [ ] NUL byte anywhere

---

## Fuzz Testing

Fuzz tests ensure the parser never panics on malformed input. Following shape-json patterns:

### Fuzz Targets

| Target | Purpose | Location |
|--------|---------|----------|
| `FuzzParser` | General parser fuzzing | `internal/parser/parser_fuzz_test.go` |
| `FuzzFastParser` | Fast path fuzzing | `internal/fastparser/parser_fuzz_test.go` |
| `FuzzTokenizer` | Tokenizer robustness | `internal/tokenizer/tokenizer_fuzz_test.go` |

### Seed Corpus

```properties
# Valid inputs
host=localhost
port=8080
empty=
key.with.dots=value
key-with-dashes=value
key_with_underscores=value

# Edge cases
   key = value with spaces
# comment line
=invalid-empty-key
no-equals-sign
key=value=with=equals
key=value # not a comment
```

### Running Fuzz Tests

```bash
# Run for 30 seconds
go test -fuzz=FuzzParser -fuzztime=30s ./internal/parser/

# Run for 5 minutes (thorough)
go test -fuzz=FuzzParser -fuzztime=5m ./internal/parser/

# Run all fuzz targets
make fuzz
```

---

## Benchmarks

Benchmarks compare fast path vs AST path performance across different file sizes.

### Test Data Sizes

| Size | Properties | File Size | Use Case |
|------|------------|-----------|----------|
| **Small** | ~10 properties | ~200 bytes | Config snippets |
| **Medium** | ~500 properties | ~15 KB | Typical config files |
| **Large** | ~10,000 properties | ~300 KB | Large config/i18n files |

### Test Data Location

```
testdata/benchmarks/
├── small.properties    # ~10 properties, ~200 bytes
├── medium.properties   # ~500 properties, ~15 KB
└── large.properties    # ~10,000 properties, ~300 KB
```

### Benchmark Tests

```go
// Fast path benchmarks
BenchmarkLoad_Small
BenchmarkLoad_Medium
BenchmarkLoad_Large
BenchmarkValidate_Small
BenchmarkValidate_Medium
BenchmarkValidate_Large

// AST path benchmarks
BenchmarkParse_Small
BenchmarkParse_Medium
BenchmarkParse_Large

// Reader variants
BenchmarkLoadReader_Large
BenchmarkParseReader_Large

// Dual-path comparison
BenchmarkLoad_vs_Parse_Large  // Side-by-side comparison
```

### Expected Performance Targets

Based on shape-json dual-path results:

| Metric | Fast Path | AST Path | Target Improvement |
|--------|-----------|----------|-------------------|
| Time | baseline | +5-10x | Fast path 5-10x faster |
| Memory | baseline | +5-10x | Fast path 5-10x less |
| Allocations | baseline | +3-6x | Fast path 3-6x fewer |

---

## Makefile

Create `Makefile` with the following targets:

```makefile
.PHONY: test lint build coverage clean all
.PHONY: bench bench-small bench-medium bench-large bench-compare
.PHONY: fuzz fuzz-parser fuzz-fast fuzz-tokenizer

# Core targets
test:
	go test -v -race ./internal/... ./pkg/...

lint:
	golangci-lint run

build:
	go build ./...

coverage:
	@mkdir -p coverage
	go test -coverprofile=coverage/coverage.out ./internal/... ./pkg/...
	go tool cover -html=coverage/coverage.out -o coverage/coverage.html

clean:
	rm -rf coverage/ benchmarks/

all: test lint build coverage

# Benchmark targets
bench:
	go test -bench=. -benchmem ./pkg/props/

bench-small:
	go test -bench=Small -benchmem ./pkg/props/

bench-medium:
	go test -bench=Medium -benchmem ./pkg/props/

bench-large:
	go test -bench=Large -benchmem ./pkg/props/

bench-compare:
	@mkdir -p benchmarks
	go test -bench=. -benchmem -count=10 ./pkg/props/ | tee benchmarks/results.txt

bench-profile:
	@mkdir -p benchmarks
	go test -bench=Large -benchmem -cpuprofile=benchmarks/cpu.prof ./pkg/props/
	go test -bench=Large -benchmem -memprofile=benchmarks/mem.prof ./pkg/props/

# Fuzz targets
fuzz:
	go test -fuzz=Fuzz -fuzztime=30s ./internal/parser/
	go test -fuzz=Fuzz -fuzztime=30s ./internal/fastparser/

fuzz-parser:
	go test -fuzz=FuzzParser -fuzztime=60s ./internal/parser/

fuzz-fast:
	go test -fuzz=FuzzFastParser -fuzztime=60s ./internal/fastparser/

fuzz-tokenizer:
	go test -fuzz=FuzzTokenizer -fuzztime=60s ./internal/tokenizer/
```

---

## Files to Create (Updated)

**New Files:**

### Build & Test Infrastructure
- `Makefile` - Build, test, benchmark, and fuzz targets
- `testdata/benchmarks/small.properties` - ~10 properties
- `testdata/benchmarks/medium.properties` - ~500 properties
- `testdata/benchmarks/large.properties` - ~10,000 properties

### Fuzz Tests
- `internal/parser/parser_fuzz_test.go` - AST parser fuzz tests
- `internal/fastparser/parser_fuzz_test.go` - Fast parser fuzz tests
- `internal/tokenizer/tokenizer_fuzz_test.go` - Tokenizer fuzz tests

### Benchmark Tests
- `pkg/props/props_bench_test.go` - All benchmark tests

---

## Definition of Done
- [ ] make build passes
- [ ] make test passes (all unit tests)
- [ ] make lint passes
- [ ] make fuzz runs without crashes (minimum 100K executions each)
- [ ] make bench shows fast path 5x+ faster than AST path
- [ ] 100% compliance with properties-format.md specification
- [ ] README.md documents all public APIs
- [ ] ARCHITECTURE.md explains dual-path design
