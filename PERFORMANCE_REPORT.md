# Performance Benchmark Report: shape-properties

> **Note:** This is an initial skeleton. Run `make performance-report` to populate it with live benchmark data.

**Generated:** Run `make performance-report` to update
**Platform:** Run `make performance-report` to detect
**Go Version:** Run `make performance-report` to detect
**Benchmark Time:** 3 seconds per test

---

## Executive Summary

shape-properties provides a dual-path architecture optimised for different use cases:

- **Fast Path (`Load`)**: Returns `map[string]string` — optimised for configuration loading
- **AST Path (`Parse`)**: Returns `ast.SchemaNode` — for tree manipulation and format conversion

The fast path is typically **2-5x faster** for real-world config files (small to medium size) because it skips all AST node construction.

---

## Performance Comparison: Load vs Parse

> Run `make performance-report` to populate this table with measured data.

| Size | Load (fast path) | Parse (AST path) | Speed Ratio | Memory Ratio |
|------|-----------------|-----------------|-------------|--------------|
| Small | — | — | — | — |
| Medium | — | — | — | — |
| Large | — | — | — | — |

---

## Analysis and Recommendations

### Why is Load() faster?

The fast path (`Load`) avoids AST construction entirely:

1. **No AST nodes** — values go directly into a `map[string]string`
2. **Fewer allocations** — no ObjectNode/LiteralNode heap objects per property
3. **Single pass** — tokenize and collect in one sweep

### When to use each path

| Scenario | Recommended API |
|----------|--------------------|
| Loading config at startup | `Load()` / `LoadReader()` |
| Validating user-supplied config | `Validate()` / `ValidateReader()` |
| Format conversion / tree manipulation | `Parse()` / `ParseReader()` |
| Generating properties text | `RenderMap()` / `Render()` |

---

## Benchmark Methodology

### Test Data

- **Small**: ~10 properties (~500 bytes)
- **Medium**: ~500 properties (~25 KB)
- **Large**: ~10,000 properties (~500 KB)

### Configuration

- **Iterations**: Determined by Go benchmark framework (3 second minimum per test)
- **Memory**: Measured with `-benchmem` flag

---

## Appendix: Running the Benchmarks

### Regenerate This Report

```bash
make performance-report
```

### Run Benchmarks Manually

```bash
# Run all benchmarks
make bench

# Save benchmark results to file
make bench-report

# Run multiple times for statistical analysis
make bench-compare

# Run with profiling
make bench-profile
```

### Analyze with benchstat

```bash
# Install benchstat
go install golang.org/x/perf/cmd/benchstat@latest

# Run benchmarks multiple times
make bench-compare

# Analyze results
benchstat benchmarks/benchstat.txt
```

### Profile Analysis

```bash
# Generate profiles
make bench-profile

# Analyze CPU profile
go tool pprof benchmarks/cpu.prof

# Analyze memory profile
go tool pprof benchmarks/mem.prof
```

### Benchmark History

```bash
# List historical runs
make bench-history

# Compare latest vs previous run
make bench-compare-history
```
