# Testing Guide

This document describes the testing strategy and coverage for shape-properties.

## Test Coverage Summary

| Package | Coverage | Status |
|---------|----------|--------|
| `internal/parser` | — | Run `make coverage` to update |
| `internal/tokenizer` | — | Run `make coverage` to update |
| `internal/fastparser` | — | Run `make coverage` to update |
| `pkg/properties` | — | Run `make coverage` to update |

Run `make coverage` to generate an up-to-date HTML report at `coverage/coverage.html`.

## Test Categories

### 1. Unit Tests

Located in `*_test.go` files throughout the codebase.

**Parser Tests** (`internal/parser/`):
- `parser_test.go` — Core parsing tests for all properties constructs
- `parser_fuzz_test.go` — Fuzz tests for the parser
- `grammar_test.go` — Grammar verification tests (ADR 0005)

**Tokenizer Tests** (`internal/tokenizer/`):
- `tokenizer_test.go` — Token recognition and generation tests

**Fast Parser Tests** (`internal/fastparser/`):
- Fuzz tests for the fast parser

**Public API Tests** (`pkg/properties/`):
- `properties_test.go` — Load/Parse/Validate/Render tests
- `properties_bench_test.go` — Benchmark tests

### 2. Fuzzing Tests

Fuzzing ensures the parser handles arbitrary input gracefully without panicking.

#### Running Fuzz Tests

Run seed corpus only (quick, used in CI):
```bash
go test ./internal/parser -run Fuzz
go test ./internal/fastparser -run Fuzz
go test ./internal/tokenizer -run Fuzz
```

Run extended fuzzing:
```bash
# Fuzz the AST parser for 60 seconds
make fuzz-parser

# Fuzz the fast parser for 60 seconds
make fuzz-fast

# Fuzz the tokenizer for 60 seconds
make fuzz-tokenizer

# Run all fuzzers for 30 seconds each
make fuzz
```

### 3. Grammar Verification Tests

Located in `internal/parser/grammar_test.go`.

Implements Shape ADR 0005: Grammar-as-Verification.

- **TestGrammarFileExists** — Verifies `docs/grammar/properties.ebnf` exists and contains expected rules
- **TestGrammarDocumentation** — Verifies grammar header documentation
- **TestGrammarVerification** — Attempts grammar-driven test generation (graceful fallback if parser does not yet support all EBNF constructs)

Run grammar tests:
```bash
# All grammar tests
make grammar-test

# Verify grammar file exists only
make grammar-verify
```

## Running Tests

### Run All Tests
```bash
make test
# or
go test -v -race ./internal/... ./pkg/...
```

### Run Tests with Coverage
```bash
make coverage
# Opens coverage/coverage.html and prints total coverage %
```

### Run Specific Package Tests
```bash
# Parser only
go test ./internal/parser -v

# Tokenizer only
go test ./internal/tokenizer -v

# Fast parser only
go test ./internal/fastparser -v

# Public API only
go test ./pkg/properties -v
```

### Run Grammar Tests
```bash
make grammar-test
make grammar-verify
```

### Run Tests with Race Detection
```bash
go test -race ./...
```

## Benchmark Tests

### Run All Benchmarks
```bash
make bench
```

### Run Benchmarks by Size
```bash
make bench-small    # Small inputs only
make bench-medium   # Medium inputs only
make bench-large    # Large inputs only
```

### Save Results to File
```bash
make bench-report
# Saves to benchmarks/results.txt
```

### Statistical Analysis
```bash
make bench-compare
# Runs 10 times, saves to benchmarks/benchstat.txt
# Analyze with: benchstat benchmarks/benchstat.txt
```

### Generate Performance Report
```bash
make performance-report
# Writes PERFORMANCE_REPORT.md and saves history
```

### View Benchmark History
```bash
make bench-history
make bench-compare-history
```

## Test Best Practices

### Table-Driven Tests
```go
tests := []struct {
    name  string
    input string
    want  map[string]string
    err   bool
}{
    {name: "simple", input: "key=value", want: map[string]string{"key": "value"}},
    {name: "empty key", input: "=value", err: true},
}
for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        // Test logic
    })
}
```

### Coverage Goals

- **Parser code**: 95%+
- **Tokenizer code**: 95%+
- **Fast parser code**: 90%+
- **Public API code**: 90%+

### Debugging Failed Tests

```bash
# Verbose output for a specific test
go test ./internal/parser -run TestParse_DuplicateKey -v

# Coverage for a single package
go test -coverprofile=coverage.out ./internal/parser
go tool cover -html=coverage.out

# Uncovered lines
go tool cover -func=coverage.out | grep -v "100.0%"
```

## Related Documentation

- [Grammar Specification](grammar/properties.ebnf)
- [Properties Format](../properties-format.md)
- [ADR 0005: Grammar-as-Verification](https://github.com/shapestone/shape-core/blob/main/docs/adr/0005-grammar-as-verification.md)
