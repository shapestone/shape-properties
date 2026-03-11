package properties

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

var (
	smallData  []byte
	mediumData []byte
	largeData  []byte
)

func init() {
	// Load benchmark data files
	base := "../../testdata/benchmarks"

	var err error
	smallData, err = os.ReadFile(filepath.Join(base, "small.properties"))
	if err != nil {
		// Generate inline if file doesn't exist
		smallData = []byte(`host=localhost
port=8080
debug=true
log.level=info
db.host=localhost
db.port=5432
db.name=myapp
timeout=30
max.connections=100
enabled=true`)
	}

	mediumData, err = os.ReadFile(filepath.Join(base, "medium.properties"))
	if err != nil {
		// Generate inline if file doesn't exist
		mediumData = generateProperties(500)
	}

	largeData, err = os.ReadFile(filepath.Join(base, "large.properties"))
	if err != nil {
		// Generate inline if file doesn't exist
		largeData = generateProperties(10000)
	}
}

func generateProperties(count int) []byte {
	var data []byte
	for i := 0; i < count; i++ {
		line := fmt.Sprintf("property%d=value%d\n", i, i)
		data = append(data, line...)
	}
	return data
}

// ============================================================================
// Fast Path Benchmarks
// ============================================================================

func BenchmarkLoad_Small(b *testing.B) {
	input := string(smallData)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Load(input)
	}
}

func BenchmarkLoad_Medium(b *testing.B) {
	input := string(mediumData)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Load(input)
	}
}

func BenchmarkLoad_Large(b *testing.B) {
	input := string(largeData)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Load(input)
	}
}

func BenchmarkValidate_Small(b *testing.B) {
	input := string(smallData)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Validate(input)
	}
}

func BenchmarkValidate_Medium(b *testing.B) {
	input := string(mediumData)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Validate(input)
	}
}

func BenchmarkValidate_Large(b *testing.B) {
	input := string(largeData)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Validate(input)
	}
}

// ============================================================================
// AST Path Benchmarks
// ============================================================================

func BenchmarkParse_Small(b *testing.B) {
	input := string(smallData)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Parse(input)
	}
}

func BenchmarkParse_Medium(b *testing.B) {
	input := string(mediumData)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Parse(input)
	}
}

func BenchmarkParse_Large(b *testing.B) {
	input := string(largeData)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Parse(input)
	}
}

// ============================================================================
// Comparison Benchmarks
// ============================================================================

func BenchmarkLoad_vs_Parse_Small(b *testing.B) {
	input := string(smallData)

	b.Run("Load", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = Load(input)
		}
	})

	b.Run("Parse", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = Parse(input)
		}
	})
}

func BenchmarkLoad_vs_Parse_Medium(b *testing.B) {
	input := string(mediumData)

	b.Run("Load", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = Load(input)
		}
	})

	b.Run("Parse", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = Parse(input)
		}
	})
}

func BenchmarkLoad_vs_Parse_Large(b *testing.B) {
	input := string(largeData)

	b.Run("Load", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = Load(input)
		}
	})

	b.Run("Parse", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = Parse(input)
		}
	})
}

// ============================================================================
// Render Benchmarks
// ============================================================================

func BenchmarkRender_Small(b *testing.B) {
	input := string(smallData)
	node, _ := Parse(input)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Render(node)
	}
}

func BenchmarkRender_Large(b *testing.B) {
	input := string(largeData)
	node, _ := Parse(input)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Render(node)
	}
}

func BenchmarkRenderMap_Small(b *testing.B) {
	input := string(smallData)
	m, _ := Load(input)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = RenderMap(m)
	}
}

func BenchmarkRenderMap_Large(b *testing.B) {
	input := string(largeData)
	m, _ := Load(input)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = RenderMap(m)
	}
}
