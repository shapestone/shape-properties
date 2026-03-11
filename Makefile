.PHONY: test lint build coverage clean all
.PHONY: bench bench-small bench-medium bench-large bench-compare bench-profile
.PHONY: fuzz fuzz-parser fuzz-fast fuzz-tokenizer

# Default target
all: test lint build

# Core targets
test:
	go test -v -race ./internal/... ./pkg/...

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

clean:
	rm -rf coverage/ benchmarks/

# Benchmark targets
bench:
	go test -bench=. -benchmem ./pkg/properties/

bench-small:
	go test -bench=Small -benchmem ./pkg/properties/

bench-medium:
	go test -bench=Medium -benchmem ./pkg/properties/

bench-large:
	go test -bench=Large -benchmem ./pkg/properties/

bench-compare:
	@mkdir -p benchmarks
	go test -bench=. -benchmem -count=10 ./pkg/properties/ | tee benchmarks/results.txt

bench-profile:
	@mkdir -p benchmarks
	go test -bench=Large -benchmem -cpuprofile=benchmarks/cpu.prof ./pkg/properties/
	go test -bench=Large -benchmem -memprofile=benchmarks/mem.prof ./pkg/properties/

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
