.PHONY: test lint build coverage clean all grammar-test grammar-verify
.PHONY: bench bench-small bench-medium bench-large bench-compare bench-profile
.PHONY: bench-report performance-report bench-history bench-compare-history
.PHONY: fuzz fuzz-parser fuzz-fast fuzz-tokenizer
.PHONY: check fmt test-all

# Run all checks (grammar, test, lint, build, coverage)
all: grammar-verify test lint build coverage

# Core targets
test:
	go test -v -race ./internal/... ./pkg/...

# Run grammar verification tests only
grammar-test:
	@echo "Running grammar verification tests..."
	go test -v ./internal/parser -run TestGrammar

# Verify grammar file exists and is valid
grammar-verify:
	@echo "Verifying grammar files..."
	@if [ ! -f docs/grammar/properties.ebnf ]; then \
		echo "Error: Grammar file missing at docs/grammar/properties.ebnf"; \
		exit 1; \
	fi
	@echo "✓ Grammar file exists (properties.ebnf)"
	@go test ./internal/parser -run TestGrammarFileExists

lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed, skipping"; \
	fi

build:
	go build ./...

coverage:
	@mkdir -p coverage
	go test -coverprofile=coverage/coverage.out ./internal/... ./pkg/...
	go tool cover -html=coverage/coverage.out -o coverage/coverage.html
	@echo "Coverage report: coverage/coverage.html"
	@go tool cover -func=coverage/coverage.out | grep total

clean:
	rm -rf coverage/ benchmarks/
	go clean

# ================================
# Benchmark Targets
# ================================

bench:
	go test -bench=. -benchmem ./pkg/properties/

bench-small:
	go test -bench=Small -benchmem ./pkg/properties/

bench-medium:
	go test -bench=Medium -benchmem ./pkg/properties/

bench-large:
	go test -bench=Large -benchmem ./pkg/properties/

# Run benchmarks and save output to a file
bench-report:
	@mkdir -p benchmarks
	@echo "Running benchmarks and saving to benchmarks/results.txt..."
	go test -bench=. -benchmem ./pkg/properties/ | tee benchmarks/results.txt
	@echo "Benchmark results saved to benchmarks/results.txt"

# Run benchmarks multiple times with benchstat for statistical analysis
bench-compare:
	@mkdir -p benchmarks
	@echo "Running benchmarks 10 times for statistical analysis..."
	@echo "This will take several minutes..."
	@for i in 1 2 3 4 5 6 7 8 9 10; do \
		echo "Run $$i/10..."; \
		go test -bench=. -benchmem ./pkg/properties/ >> benchmarks/benchstat.txt; \
	done
	@echo "Results saved to benchmarks/benchstat.txt"
	@echo "Install benchstat with: go install golang.org/x/perf/cmd/benchstat@latest"
	@echo "Analyze with: benchstat benchmarks/benchstat.txt"

bench-profile:
	@mkdir -p benchmarks
	go test -bench=Large -benchmem -cpuprofile=benchmarks/cpu.prof ./pkg/properties/
	go test -bench=Large -benchmem -memprofile=benchmarks/mem.prof ./pkg/properties/

# Generate performance report from benchmark results
performance-report:
	@echo "Generating performance report..."
	@go run scripts/generate_benchmark_report/main.go
	@echo "Performance report updated: PERFORMANCE_REPORT.md"

# List available benchmark history runs
bench-history:
	@if [ -d "benchmarks/history" ]; then \
		has_benchmarks=false; \
		for dir in benchmarks/history/*/; do \
			if [ -d "$$dir" ] && [ -f "$${dir}benchmark_output.txt" ]; then \
				has_benchmarks=true; \
				break; \
			fi; \
		done; \
		if [ "$$has_benchmarks" = "true" ]; then \
			echo "Available benchmark history:"; \
			echo ""; \
			for dir in benchmarks/history/*/; do \
				if [ -d "$$dir" ] && [ -f "$${dir}benchmark_output.txt" ]; then \
					timestamp=$$(basename "$$dir"); \
					echo "  $$timestamp"; \
					if [ -f "$${dir}metadata.json" ]; then \
						grep -E '"(commit|platform)"' "$${dir}metadata.json" | sed 's/^/    /'; \
					fi; \
					echo ""; \
				fi; \
			done; \
		else \
			echo "No benchmark history found."; \
			echo "Run 'make performance-report' to create your first benchmark."; \
		fi; \
	else \
		echo "No benchmark history found."; \
		echo "Run 'make performance-report' to create your first benchmark."; \
	fi

# Compare current benchmarks vs most recent historical run
bench-compare-history:
	@if ! command -v benchstat >/dev/null 2>&1; then \
		echo "Error: benchstat not found. Install with:"; \
		echo "  go install golang.org/x/perf/cmd/benchstat@latest"; \
		exit 1; \
	fi
	@if [ ! -d "benchmarks/history" ] || [ -z "$$(ls -A benchmarks/history 2>/dev/null)" ]; then \
		echo "Error: No benchmark history found."; \
		echo "Run 'make performance-report' to create benchmark history."; \
		exit 1; \
	fi
	@echo "Comparing benchmarks..."
	@go run scripts/compare_benchmarks/main.go latest previous

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

# Quick check - runs tests without verbose output
check:
	go test ./...

# Format code
fmt:
	go fmt ./...

# Run all tests including benchmarks
test-all: test bench
