package fastparser

import (
	"testing"
)

// FuzzFastParser tests the fast parser with arbitrary input.
// It should never panic regardless of input.
func FuzzFastParser(f *testing.F) {
	// Seed corpus with valid inputs
	f.Add([]byte("host=localhost"))
	f.Add([]byte("port=8080"))
	f.Add([]byte("key.with.dots=value"))
	f.Add([]byte("key-with-dashes=value"))
	f.Add([]byte("key_with_underscores=value"))
	f.Add([]byte("# comment line\nhost=localhost"))
	f.Add([]byte("empty="))
	f.Add([]byte("   key = value with spaces"))
	f.Add([]byte("key=value=with=equals"))
	f.Add([]byte("key=value # not a comment"))
	f.Add([]byte("a=1\nb=2\nc=3"))
	f.Add([]byte("a=1\r\nb=2\r\nc=3"))

	// Edge cases
	f.Add([]byte(""))
	f.Add([]byte("   "))
	f.Add([]byte("\n\n\n"))
	f.Add([]byte("# only comments"))
	f.Add([]byte("="))
	f.Add([]byte("=value"))
	f.Add([]byte("no-equals-sign"))
	f.Add([]byte("123invalidkey=value"))
	f.Add([]byte("dup=1\ndup=2"))

	// Unicode
	f.Add([]byte("greeting=こんにちは"))
	f.Add([]byte("emoji=🎉"))

	// Control characters (should error, not panic)
	f.Add([]byte("key=value\x00more"))
	f.Add([]byte("key=value\x01more"))

	f.Fuzz(func(t *testing.T, input []byte) {
		p := NewParser(input)

		// Parse should not panic, even on invalid input
		_, _ = p.Parse()

		// Also test Validate
		p2 := NewParser(input)
		_ = p2.Validate()
	})
}
