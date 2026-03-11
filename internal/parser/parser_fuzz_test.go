package parser

import (
	"testing"
)

// FuzzParser tests the AST parser with arbitrary input.
// It should never panic regardless of input.
func FuzzParser(f *testing.F) {
	// Seed corpus with valid inputs
	f.Add("host=localhost")
	f.Add("port=8080")
	f.Add("key.with.dots=value")
	f.Add("key-with-dashes=value")
	f.Add("key_with_underscores=value")
	f.Add("# comment line\nhost=localhost")
	f.Add("empty=")
	f.Add("   key = value with spaces")
	f.Add("key=value=with=equals")
	f.Add("key=value # not a comment")
	f.Add("a=1\nb=2\nc=3")
	f.Add("a=1\r\nb=2\r\nc=3")

	// Edge cases
	f.Add("")
	f.Add("   ")
	f.Add("\n\n\n")
	f.Add("# only comments")
	f.Add("=")
	f.Add("=value")
	f.Add("no-equals-sign")
	f.Add("123invalidkey=value")
	f.Add("dup=1\ndup=2")

	// Unicode
	f.Add("greeting=こんにちは")
	f.Add("emoji=🎉")

	// Control characters (should error, not panic)
	f.Add("key=value\x00more")
	f.Add("key=value\x01more")

	f.Fuzz(func(t *testing.T, input string) {
		p := NewParser(input)

		// Parse should not panic, even on invalid input
		_, _ = p.Parse()
	})
}
