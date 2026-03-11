package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// BenchmarkResult represents a single benchmark result
type BenchmarkResult struct {
	Name        string
	Iterations  int
	NsPerOp     float64
	MBPerSec    float64
	BytesPerOp  int64
	AllocsPerOp int64
}

// BenchmarkGroup groups related benchmarks for comparison (Load vs Parse)
type BenchmarkGroup struct {
	Name    string
	Load    *BenchmarkResult // fast path (Load)
	Parse   *BenchmarkResult // AST path (Parse)
	Size    string
	InputSize int64

	// Dual-path comparison
	LoadVsParseSpeed  float64
	LoadVsParseMemory float64
	LoadVsParseAllocs float64
}

// BenchmarkMetadata contains information about a benchmark run
type BenchmarkMetadata struct {
	Timestamp   string `json:"timestamp"`
	GitCommit   string `json:"commit"`
	Platform    string `json:"platform"`
	OS          string `json:"os"`
	Arch        string `json:"arch"`
	GoVersion   string `json:"go_version"`
	BenchTime   string `json:"bench_time"`
	Description string `json:"description"`
}

func main() {
	saveHistory := flag.Bool("save-history", true, "Save benchmark results to history directory")
	description := flag.String("description", "", "Optional description for this benchmark run")
	flag.Parse()

	fmt.Println("Shape-Properties Performance Report Generator")
	fmt.Println("=============================================")
	fmt.Println()

	cwd, err := os.Getwd()
	if err != nil {
		fatal("Failed to get working directory: %v", err)
	}

	projectRoot := findProjectRoot(cwd)
	if projectRoot == "" {
		fatal("Could not find project root (looking for go.mod)")
	}

	fmt.Printf("Project root: %s\n", projectRoot)
	fmt.Println()

	fmt.Println("Running benchmarks (this may take a few minutes)...")
	benchmarkOutput, err := runBenchmarks(projectRoot)
	if err != nil {
		fatal("Failed to run benchmarks: %v", err)
	}

	fmt.Println("Benchmarks completed successfully!")
	fmt.Println()

	fmt.Println("Parsing benchmark results...")
	results, err := parseBenchmarkOutput(benchmarkOutput)
	if err != nil {
		fatal("Failed to parse benchmark results: %v", err)
	}

	fmt.Printf("Parsed %d benchmark results\n", len(results))
	fmt.Println()

	groups := groupBenchmarks(results)
	fmt.Printf("Created %d comparison groups\n", len(groups))
	fmt.Println()

	fmt.Println("Generating performance report...")
	report := generateReport(groups)

	reportPath := filepath.Join(projectRoot, "PERFORMANCE_REPORT.md")
	err = os.WriteFile(reportPath, []byte(report), 0644)
	if err != nil {
		fatal("Failed to write report: %v", err)
	}

	fmt.Printf("Performance report written to: %s\n", reportPath)
	fmt.Println()

	if *saveHistory {
		fmt.Println("Saving benchmark history...")
		err = saveToHistory(projectRoot, benchmarkOutput, report, *description)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to save history: %v\n", err)
		} else {
			fmt.Println("Benchmark history saved!")
		}
		fmt.Println()
	}

	fmt.Println("Done!")
}

func findProjectRoot(startDir string) string {
	dir := startDir
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

func runBenchmarks(projectRoot string) (string, error) {
	cmd := exec.Command("go", "test", "-bench=.", "-benchmem", "-benchtime=3s", "./pkg/properties/")
	cmd.Dir = projectRoot

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("benchmark execution failed: %v\nStderr: %s", err, stderr.String())
	}

	return stdout.String(), nil
}

func parseBenchmarkOutput(output string) (map[string]*BenchmarkResult, error) {
	results := make(map[string]*BenchmarkResult)

	// BenchmarkName-N    123456    7890 ns/op    12.34 MB/s    5678 B/op    90 allocs/op
	pattern := regexp.MustCompile(`^(Benchmark\S+)-\d+\s+(\d+)\s+(\d+(?:\.\d+)?)\s+ns/op(?:\s+(\d+(?:\.\d+)?)\s+MB/s)?\s+(\d+)\s+B/op\s+(\d+)\s+allocs/op`)

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		matches := pattern.FindStringSubmatch(line)
		if matches == nil {
			continue
		}

		name := matches[1]
		iterations, _ := strconv.Atoi(matches[2])
		nsPerOp, _ := strconv.ParseFloat(matches[3], 64)
		bytesPerOp, _ := strconv.ParseInt(matches[5], 10, 64)
		allocsPerOp, _ := strconv.ParseInt(matches[6], 10, 64)

		var mbPerSec float64
		if matches[4] != "" {
			mbPerSec, _ = strconv.ParseFloat(matches[4], 64)
		}

		results[name] = &BenchmarkResult{
			Name:        name,
			Iterations:  iterations,
			NsPerOp:     nsPerOp,
			MBPerSec:    mbPerSec,
			BytesPerOp:  bytesPerOp,
			AllocsPerOp: allocsPerOp,
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no benchmark results found in output")
	}

	return results, nil
}

func groupBenchmarks(results map[string]*BenchmarkResult) []*BenchmarkGroup {
	var groups []*BenchmarkGroup

	// Approximate input sizes based on testdata
	sizes := map[string]int64{
		"Small":  500,
		"Medium": 25000,
		"Large":  500000,
	}

	for _, size := range []string{"Small", "Medium", "Large"} {
		loadKeys := []string{
			"BenchmarkLoad_" + size,
			"BenchmarkLoad_FastPath_" + size,
		}
		parseKeys := []string{
			"BenchmarkParse_" + size,
			"BenchmarkParse_ASTPath_" + size,
		}

		load := findFirstResult(results, loadKeys)
		parse := findFirstResult(results, parseKeys)

		if load != nil || parse != nil {
			group := &BenchmarkGroup{
				Name:      "Properties_" + size,
				Load:      load,
				Parse:     parse,
				Size:      size,
				InputSize: sizes[size],
			}
			calculateRatios(group)
			groups = append(groups, group)
		}
	}

	return groups
}

func findFirstResult(results map[string]*BenchmarkResult, keys []string) *BenchmarkResult {
	for _, key := range keys {
		if result, ok := results[key]; ok {
			return result
		}
	}
	return nil
}

func calculateRatios(group *BenchmarkGroup) {
	if group.Load != nil && group.Parse != nil {
		group.LoadVsParseSpeed = group.Parse.NsPerOp / group.Load.NsPerOp
		if group.Load.BytesPerOp > 0 {
			group.LoadVsParseMemory = float64(group.Parse.BytesPerOp) / float64(group.Load.BytesPerOp)
		}
		if group.Load.AllocsPerOp > 0 {
			group.LoadVsParseAllocs = float64(group.Parse.AllocsPerOp) / float64(group.Load.AllocsPerOp)
		}
	}
}

func generateReport(groups []*BenchmarkGroup) string {
	var buf bytes.Buffer

	buf.WriteString("# Performance Benchmark Report: shape-properties\n\n")
	fmt.Fprintf(&buf, "**Date:** %s\n", time.Now().Format("2006-01-02"))
	fmt.Fprintf(&buf, "**Platform:** %s (%s/%s)\n", getPlatformName(), runtime.GOOS, runtime.GOARCH)
	fmt.Fprintf(&buf, "**Go Version:** %s\n", getGoVersion())
	buf.WriteString("**Benchmark Time:** 3 seconds per test\n")
	buf.WriteString("**Generated:** Automatically by `make performance-report`\n\n")

	buf.WriteString("## Executive Summary\n\n")
	buf.WriteString("shape-properties provides a dual-path architecture:\n\n")
	buf.WriteString("- **Fast Path (`Load`)**: Direct `map[string]string` — optimized for config loading\n")
	buf.WriteString("- **AST Path (`Parse`)**: Full AST — for tree manipulation and format conversion\n\n")

	// Key findings: average speed ratio
	if len(groups) > 0 {
		var speedRatioSum float64
		count := 0
		for _, g := range groups {
			if g.Load != nil && g.Parse != nil {
				speedRatioSum += g.LoadVsParseSpeed
				count++
			}
		}
		if count > 0 {
			avg := speedRatioSum / float64(count)
			buf.WriteString("### Key Findings\n\n")
			fmt.Fprintf(&buf, "- **Fast path is %.1fx faster** than AST path on average\n", avg)
			buf.WriteString("- Use `Load()` for configuration loading (common case)\n")
			buf.WriteString("- Use `Parse()` only when tree manipulation is needed\n\n")
		}
	}

	buf.WriteString("---\n\n")

	// Per-size tables
	buf.WriteString("## Performance Comparison: Load vs Parse\n\n")
	buf.WriteString("| Size | Load (fast path) | Parse (AST path) | Speed Ratio | Memory Ratio | Alloc Ratio |\n")
	buf.WriteString("|------|-----------------|-----------------|-------------|--------------|-------------|\n")
	for _, g := range groups {
		loadStr := "N/A"
		parseStr := "N/A"
		speedStr := "N/A"
		memStr := "N/A"
		allocStr := "N/A"

		if g.Load != nil {
			loadStr = formatDuration(g.Load.NsPerOp)
		}
		if g.Parse != nil {
			parseStr = formatDuration(g.Parse.NsPerOp)
		}
		if g.Load != nil && g.Parse != nil {
			speedStr = fmt.Sprintf("%.1fx", g.LoadVsParseSpeed)
			memStr = fmt.Sprintf("%.1fx", g.LoadVsParseMemory)
			allocStr = fmt.Sprintf("%.1fx", g.LoadVsParseAllocs)
		}
		fmt.Fprintf(&buf, "| %s | %s | %s | %s | %s | %s |\n",
			g.Size, loadStr, parseStr, speedStr, memStr, allocStr)
	}
	buf.WriteString("\n")

	// Detailed sections
	for _, g := range groups {
		writeGroupSection(&buf, g)
	}

	buf.WriteString("---\n\n")
	writeAnalysisSection(&buf)

	buf.WriteString("---\n\n")
	writeMethodologySection(&buf)

	buf.WriteString("---\n\n")
	writeUsageSection(&buf)

	return buf.String()
}

func writeGroupSection(buf *bytes.Buffer, g *BenchmarkGroup) {
	fmt.Fprintf(buf, "### %s\n\n", g.Size)
	buf.WriteString("```\n")
	if g.Load != nil {
		buf.WriteString(formatBenchmarkLine(g.Load))
	}
	if g.Parse != nil {
		buf.WriteString(formatBenchmarkLine(g.Parse))
	}
	buf.WriteString("```\n\n")

	if g.Load != nil && g.Parse != nil {
		buf.WriteString("**Analysis:**\n")
		fmt.Fprintf(buf, "- **Speed**: Load() is **%.1fx faster** (%s vs %s)\n",
			g.LoadVsParseSpeed,
			formatDuration(g.Load.NsPerOp),
			formatDuration(g.Parse.NsPerOp))
		fmt.Fprintf(buf, "- **Memory**: Parse() uses **%.1fx more memory** (%s vs %s)\n",
			g.LoadVsParseMemory,
			formatBytes(g.Parse.BytesPerOp),
			formatBytes(g.Load.BytesPerOp))
		fmt.Fprintf(buf, "- **Allocations**: Parse() makes **%.1fx more allocations** (%d vs %d)\n",
			g.LoadVsParseAllocs,
			g.Parse.AllocsPerOp,
			g.Load.AllocsPerOp)
		buf.WriteString("\n")
	}
}

func writeAnalysisSection(buf *bytes.Buffer) {
	buf.WriteString("## Analysis and Recommendations\n\n")

	buf.WriteString("### Why is Load() faster?\n\n")
	buf.WriteString("The fast path (`Load`) avoids AST construction entirely:\n\n")
	buf.WriteString("1. **No AST nodes** — values go directly into a `map[string]string`\n")
	buf.WriteString("2. **Fewer allocations** — no ObjectNode/LiteralNode heap objects\n")
	buf.WriteString("3. **Single pass** — tokenize and collect in one sweep\n\n")

	buf.WriteString("### When to use each path\n\n")
	buf.WriteString("| Scenario | Recommended API |\n")
	buf.WriteString("|----------|-----------------|\n")
	buf.WriteString("| Loading config at startup | `Load()` / `LoadReader()` |\n")
	buf.WriteString("| Validating user-supplied config | `Validate()` / `ValidateReader()` |\n")
	buf.WriteString("| Format conversion / tree manipulation | `Parse()` / `ParseReader()` |\n")
	buf.WriteString("| Generating properties text | `RenderMap()` / `Render()` |\n\n")
}

func writeMethodologySection(buf *bytes.Buffer) {
	buf.WriteString("## Benchmark Methodology\n\n")
	buf.WriteString("### Test Data\n\n")
	buf.WriteString("- **Small**: ~10 properties\n")
	buf.WriteString("- **Medium**: ~500 properties\n")
	buf.WriteString("- **Large**: ~10,000 properties\n\n")
	buf.WriteString("### Configuration\n\n")
	buf.WriteString("- **Benchmark time**: 3 seconds per test (`-benchtime=3s`)\n")
	fmt.Fprintf(buf, "- **Platform**: %s (%s/%s)\n", getPlatformName(), runtime.GOOS, runtime.GOARCH)
	fmt.Fprintf(buf, "- **Go Version**: %s\n", getGoVersion())
	buf.WriteString("- **Memory**: measured with `-benchmem`\n\n")
}

func writeUsageSection(buf *bytes.Buffer) {
	buf.WriteString("## Appendix: Running the Benchmarks\n\n")
	buf.WriteString("### Regenerate This Report\n\n")
	buf.WriteString("```bash\nmake performance-report\n```\n\n")
	buf.WriteString("### Run Benchmarks Manually\n\n")
	buf.WriteString("```bash\n# Run all benchmarks\nmake bench\n\n# Save to file\nmake bench-report\n\n# Statistical analysis (10 runs)\nmake bench-compare\n\n# Profiling\nmake bench-profile\n```\n\n")
	buf.WriteString("### Analyze with benchstat\n\n")
	buf.WriteString("```bash\ngo install golang.org/x/perf/cmd/benchstat@latest\nmake bench-compare\nbenchstat benchmarks/benchstat.txt\n```\n\n")
	buf.WriteString("### View history\n\n")
	buf.WriteString("```bash\nmake bench-history\nmake bench-compare-history\n```\n")
}

// Helpers

func formatBenchmarkLine(r *BenchmarkResult) string {
	line := fmt.Sprintf("%-50s %8d %12.0f ns/op", r.Name+"-10", r.Iterations, r.NsPerOp)
	if r.MBPerSec > 0 {
		line += fmt.Sprintf(" %8.2f MB/s", r.MBPerSec)
	}
	line += fmt.Sprintf(" %12d B/op %8d allocs/op\n", r.BytesPerOp, r.AllocsPerOp)
	return line
}

func formatDuration(ns float64) string {
	if ns < 1000 {
		return fmt.Sprintf("%.0fns", ns)
	} else if ns < 1_000_000 {
		return fmt.Sprintf("%.1fµs", ns/1000)
	}
	return fmt.Sprintf("%.1fms", ns/1_000_000)
}

func formatBytes(b int64) string {
	if b < 1024 {
		return fmt.Sprintf("%d B", b)
	} else if b < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(b)/1024)
	}
	return fmt.Sprintf("%.1f MB", float64(b)/(1024*1024))
}

func getPlatformName() string {
	if runtime.GOOS == "darwin" {
		cmd := exec.Command("sysctl", "-n", "machdep.cpu.brand_string")
		output, err := cmd.Output()
		if err == nil {
			cpuName := strings.TrimSpace(string(output))
			if strings.Contains(cpuName, "Apple") {
				return cpuName
			}
		}
		return "macOS"
	}
	return runtime.GOOS
}

func getGoVersion() string {
	return strings.TrimPrefix(runtime.Version(), "go")
}

func fatal(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "ERROR: "+format+"\n", args...)
	os.Exit(1)
}

func saveToHistory(projectRoot, benchmarkOutput, report, description string) error {
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	historyDir := filepath.Join(projectRoot, "benchmarks", "history", timestamp)

	if err := os.MkdirAll(historyDir, 0755); err != nil {
		return fmt.Errorf("failed to create history directory: %v", err)
	}

	if err := os.WriteFile(filepath.Join(historyDir, "benchmark_output.txt"), []byte(benchmarkOutput), 0644); err != nil {
		return fmt.Errorf("failed to write benchmark output: %v", err)
	}

	if err := os.WriteFile(filepath.Join(historyDir, "PERFORMANCE_REPORT.md"), []byte(report), 0644); err != nil {
		return fmt.Errorf("failed to write report: %v", err)
	}

	metadata := BenchmarkMetadata{
		Timestamp:   timestamp,
		GitCommit:   getGitCommit(projectRoot),
		Platform:    getPlatformName(),
		OS:          runtime.GOOS,
		Arch:        runtime.GOARCH,
		GoVersion:   getGoVersion(),
		BenchTime:   "3s",
		Description: description,
	}

	metadataJSON, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %v", err)
	}

	if err := os.WriteFile(filepath.Join(historyDir, "metadata.json"), metadataJSON, 0644); err != nil {
		return fmt.Errorf("failed to write metadata: %v", err)
	}

	// Create .gitignore for history dir
	gitignorePath := filepath.Join(projectRoot, "benchmarks", "history", ".gitignore")
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		gitignoreContent := "# Benchmark history files are large and change frequently\n*\n!.gitignore\n!README.md\n"
		_ = os.WriteFile(gitignorePath, []byte(gitignoreContent), 0644)
	}

	fmt.Printf("  Saved to: %s\n", historyDir)
	return nil
}

func getGitCommit(projectRoot string) string {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = projectRoot
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(output))
}
