package tokenizer

import (
	"testing"
)

// FuzzTokenizer tests the tokenizer with arbitrary input.
// It should never panic regardless of input.
func FuzzTokenizer(f *testing.F) {
	// Seed corpus with valid inputs
	f.Add("host=localhost")
	f.Add("port=8080")
	f.Add("key.with.dots=value")
	f.Add("key-with-dashes=value")
	f.Add("key_with_underscores=value")
	f.Add("# comment line")
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
	f.Add("=")
	f.Add("=value")
	f.Add("no-equals-sign")
	f.Add("123invalidkey=value")

	// Unicode
	f.Add("greeting=こんにちは")
	f.Add("emoji=🎉")

	f.Fuzz(func(t *testing.T, input string) {
		tok := NewTokenizer(input)

		// Consume all tokens - should not panic
		for {
			token := tok.NextToken()
			if token.Kind == TokenEOF {
				break
			}
			// If we see EQUALS, also scan the value
			if token.Kind == TokenEquals {
				_ = tok.ScanValue()
			}
		}
	})
}
