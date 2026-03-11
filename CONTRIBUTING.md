# Contributing to shape-properties

Thank you for your interest in contributing to shape-properties! This document provides
guidelines for contributing to the project.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Scope Policy](#scope-policy)
- [How to Contribute](#how-to-contribute)
- [Development Setup](#development-setup)
- [Pull Request Process](#pull-request-process)
- [Testing Guidelines](#testing-guidelines)

## Code of Conduct

This project adheres to the Contributor Covenant [Code of Conduct](CODE_OF_CONDUCT.md).
By participating, you are expected to uphold this code. Please report unacceptable behavior
to conduct@shapestone.com.

## Scope Policy

shape-properties implements a single, well-defined format: the **Simple Properties
Configuration Format** (`key=value` pairs). Contributions must stay within that scope.

### What Belongs in shape-properties

- Bug fixes in parsing, validation, or rendering
- Performance improvements for the fast path or AST path
- Documentation improvements, usage examples, API documentation
- Test coverage for edge cases, including fuzz corpus inputs
- Tooling improvements: CI/CD, Makefile targets, benchmarking

### What Does Not Belong in shape-properties

- Support for Java `.properties` escaping or `:` separators
- dotenv-style variable expansion or quoting
- Multiline value support (this format is intentionally line-oriented)
- New key formats (the key pattern is part of the format specification)
- Nested structures, arrays, or type-annotated values

If you need a richer configuration format, see the
[Shape Ecosystem](https://github.com/shapestone/shape) for YAML, TOML, JSON, and other parsers.

### Format Specification Changes

The Simple Properties Configuration Format is defined in
[properties-format.md](properties-format.md). Changes to the format specification require
discussion before implementation because they affect all consumers of the format, not just
this library.

If you have a proposal to extend the format, open an issue with the label `spec-discussion`
before writing any code.

## How to Contribute

### Types of Contributions Welcome

1. **Bug Fixes**: Incorrect parse results, wrong error messages, line number off-by-one errors
2. **Performance Improvements**: Faster tokenization, reduced allocations in the fast path
3. **Documentation**: Improved examples, clearer API docs, corrections to existing docs
4. **Test Coverage**: Table-driven tests for edge cases, fuzz corpus entries
5. **Tooling**: Makefile improvements, CI configuration, benchmark infrastructure

### Types of Contributions We Generally Do Not Accept

1. **Format extensions**: Multiline values, escaping, variable expansion
2. **Breaking API changes**: Public API signatures are stable after v1.0
3. **Compatibility modes**: Java `.properties`, dotenv, or shell-compatible parsing
4. **External dependencies**: The only permitted external dependency is `shape-core`

## Development Setup

### Requirements

- Go 1.25 or later
- `golangci-lint` for linting (optional but recommended)
- `benchstat` for statistical benchmark analysis (optional)

```bash
# Install benchstat
go install golang.org/x/perf/cmd/benchstat@latest

# Install golangci-lint (see https://golangci-lint.run/usage/install/)
brew install golangci-lint
```

### Quick Setup

```bash
# Clone the repository
git clone https://github.com/shapestone/shape-properties.git
cd shape-properties

# Run tests
go test ./...

# Run with race detection
go test -race ./...

# Run linter (if installed)
golangci-lint run

# Check coverage
make coverage
```

### Repository Structure

```
shape-properties/
├── pkg/
│   └── properties/
│       ├── properties.go     # Public API — the primary surface
│       ├── convert.go        # AST <-> map conversion
│       └── render.go         # AST -> text rendering
├── internal/
│   ├── tokenizer/            # Token types and tokenizer
│   ├── parser/               # AST parser (LL(1) recursive descent)
│   └── fastparser/           # Fast path parser (direct to map)
├── docs/
│   └── grammar/
│       └── properties.ebnf   # Formal grammar specification
├── testdata/
│   └── benchmarks/           # Benchmark input files
└── properties-format.md      # Format specification
```

### Key Design Constraints

**Dual-path architecture**: The fast path (`internal/fastparser`) and AST path
(`internal/parser`) must stay behaviorally consistent. If you fix a parsing bug, apply
the fix to both paths and add tests verifying both.

**No allocations in the fast path for the happy case**: The fast path is optimized for
allocation-free parsing. Avoid introducing `interface{}`, `any`, or heap-allocated
values in `internal/fastparser`.

**Sorted output in Render**: `Render()` and `RenderMap()` output keys in sorted order.
This is a deliberate design choice for deterministic output. Do not change this.

**Error messages include line numbers**: All parse errors must include the line number
where the error occurred.

## Pull Request Process

1. **Fork the repository** and create a feature branch:
   ```bash
   git checkout -b fix/duplicate-key-detection
   git checkout -b feat/benchstat-ci-integration
   ```

2. **Make your changes**:
   - Write clean, idiomatic Go code
   - Add tests for all new or changed behavior
   - Update documentation if the change is user-visible
   - Run `go vet ./...` and ensure no warnings

3. **Run the full test suite**:
   ```bash
   go test -race ./...
   make coverage
   ```

4. **Commit with clear messages** following
   [Conventional Commits](https://www.conventionalcommits.org/):
   ```bash
   git commit -m "fix: correct line number in duplicate key error"
   git commit -m "perf: reduce allocations in fast path for medium files"
   git commit -m "docs: add example for loading config from environment"
   git commit -m "test: add fuzz corpus entry for NUL byte in value"
   ```

   Commit type prefixes:
   - `feat:` - New feature
   - `fix:` - Bug fix
   - `docs:` - Documentation only
   - `test:` - Test additions or changes
   - `perf:` - Performance improvement
   - `refactor:` - Code restructuring without behavior change
   - `chore:` - Build process, tooling, dependency updates

5. **Push and open a PR**:
   ```bash
   git push origin fix/duplicate-key-detection
   ```

   In your PR description:
   - Describe what changed and why
   - Reference any related issues (`Fixes #42`)
   - Include benchmark results if the change affects performance
   - Note which paths are affected (fast path, AST path, or both)

6. **Code review**:
   - Maintainers will review within 3-5 business days
   - Address requested changes and push to the same branch
   - Once approved, a maintainer will merge

## Testing Guidelines

### Test Coverage Requirements

- All new code must include tests
- Bug fixes must include a test that reproduces the bug
- Coverage targets: parser 95%+, tokenizer 95%+, fast parser 90%+, public API 90%+

### Running Tests

```bash
# All tests
make test

# With race detection
go test -race ./...

# Specific package
go test ./internal/parser -v
go test ./internal/fastparser -v
go test ./pkg/properties -v

# Coverage report
make coverage

# Fuzz tests (seed corpus only, used in CI)
go test ./internal/parser -run Fuzz
go test ./internal/fastparser -run Fuzz

# Extended fuzzing
make fuzz
```

### Writing Good Tests

Use table-driven tests:

```go
func TestLoad(t *testing.T) {
    tests := []struct {
        name  string
        input string
        want  map[string]string
        err   bool
    }{
        {
            name:  "simple key-value",
            input: "host=localhost",
            want:  map[string]string{"host": "localhost"},
        },
        {
            name:  "duplicate key is an error",
            input: "host=a\nhost=b",
            err:   true,
        },
        {
            name:  "empty value is allowed",
            input: "key=",
            want:  map[string]string{"key": ""},
        },
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := properties.Load(tt.input)
            if (err != nil) != tt.err {
                t.Fatalf("Load() error = %v, wantErr %v", err, tt.err)
            }
            if !tt.err && !reflect.DeepEqual(got, tt.want) {
                t.Errorf("Load() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

Both the fast path and the AST path must be tested for any parsing behavior:

```go
// Test fast path
props, err := properties.Load(input)

// Test AST path (same behavior expected)
node, err := properties.Parse(input)
props, err = properties.NodeToMap(node)
```

### Benchmark Contributions

If your change affects performance, include before/after benchmark numbers in your PR.

```bash
# Save baseline
make bench-report

# Make your changes, then compare
make bench-compare
benchstat benchmarks/benchstat.txt
```

## Questions?

- **Issues**: [GitHub Issues](https://github.com/shapestone/shape-properties/issues)
- **Discussions**: [GitHub Discussions](https://github.com/shapestone/shape-properties/discussions)
- **Documentation**: [docs/](docs/)
- **Format specification**: [properties-format.md](properties-format.md)
- **Architecture**: [ARCHITECTURE.md](ARCHITECTURE.md)

Thank you for contributing to shape-properties!
