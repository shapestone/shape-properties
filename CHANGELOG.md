# Changelog

All notable changes to the shape-properties project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2026-03-10

### Initial Release

This is the initial public release of shape-properties, a high-performance parser for the
Simple Properties Configuration Format in the Shape Parser ecosystem.

### Highlights

**Dual-Path Architecture**

- Fast Path: Direct byte-level parsing to `map[string]string` — optimized for
  configuration loading, typically 5-10x faster than the AST path
- AST Path: Full LL(1) recursive descent parser producing `ast.SchemaNode` — for tree
  manipulation, format conversion, and round-trip rendering
- Automatic path selection based on which API you call; no configuration needed

**Format Compliance**

- Full implementation of the Simple Properties Configuration Format specification
  (`properties-format.md`)
- Strict key validation: `[A-Za-z_][A-Za-z0-9_.-]*`
- Duplicate key detection with line numbers
- UTF-8 aware; NUL bytes and control characters (except TAB) are rejected
- Both `\n` and `\r\n` line endings accepted

**Thread Safety**

- All public functions are safe for concurrent use from multiple goroutines
- Each call to `Load`, `Validate`, or `Parse` creates an independent parser instance
- `Render` and `RenderMap` use a `sync.Pool` of `bytes.Buffer` for zero-contention
  buffer reuse

### Added

**Public API** (`pkg/properties`)

- `Validate(string) error` — validate without returning parsed result
- `ValidateReader(io.Reader) error` — validate from a reader
- `Load(string) (map[string]string, error)` — fast-path configuration loading
- `LoadReader(io.Reader) (map[string]string, error)` — fast-path loading from a reader
- `MustLoad(string) map[string]string` — load or panic (for tests and init)
- `Parse(string) (ast.SchemaNode, error)` — AST-path parsing
- `ParseReader(io.Reader) (ast.SchemaNode, error)` — AST-path parsing from a reader
- `MustParse(string) ast.SchemaNode` — parse or panic (for tests and init)
- `MapToNode(map[string]string) (ast.SchemaNode, error)` — map to AST conversion
- `NodeToMap(ast.SchemaNode) (map[string]string, error)` — AST to map conversion
- `Render(ast.SchemaNode) (string, error)` — render AST to sorted properties text
- `RenderMap(map[string]string) (string, error)` — render map to sorted properties text

**Internal Packages**

- `internal/tokenizer` — token types (KEY, EQUALS, VALUE, COMMENT, NEWLINE, EOF)
  and UTF-8 aware tokenizer with position tracking
- `internal/parser` — LL(1) recursive descent AST parser with rich error messages
- `internal/fastparser` — zero-allocation fast path parser, direct bytes to map

**Grammar**

- `docs/grammar/properties.ebnf` — formal EBNF grammar specification
- Grammar verification tests implementing Shape ADR 0005

**Testing**

- Table-driven unit tests for all parsing constructs
- Fuzz tests for the AST parser, fast parser, and tokenizer
- Grammar verification tests
- Benchmark suite with small (10 props), medium (500 props), and large (10K props)
  test data
- Historical benchmark tracking with `make bench-history`

**Documentation**

- `README.md` — quick start, API reference, examples, benchmarks
- `ARCHITECTURE.md` — dual-path design, package structure, design decisions
- `properties-format.md` — complete format specification
- `PERFORMANCE_REPORT.md` — benchmark methodology and skeleton report
- `docs/TESTING.md` — testing strategy and coverage targets

### Performance Benchmarks

Expected performance on typical hardware (run `make bench` to measure on your system):

| Operation | Small (10 props) | Medium (500 props) | Large (10K props) |
|-----------|------------------|--------------------|--------------------|
| Load      | ~5 µs            | ~200 µs            | ~4 ms              |
| Parse     | ~25 µs           | ~1 ms              | ~20 ms             |

Fast path (`Load`) is 5-10x faster than AST path (`Parse`).

### Dependencies

- `github.com/shapestone/shape-core v0.9.3` — AST node types (`ast.ObjectNode`,
  `ast.LiteralNode`)

### License

Apache License 2.0
Copyright 2020-2025 Shapestone

### Links

- Repository: https://github.com/shapestone/shape-properties
- Documentation: https://pkg.go.dev/github.com/shapestone/shape-properties
- Issues: https://github.com/shapestone/shape-properties/issues
- Format Specification: [properties-format.md](properties-format.md)

[Unreleased]: https://github.com/shapestone/shape-properties/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/shapestone/shape-properties/releases/tag/v0.1.0
