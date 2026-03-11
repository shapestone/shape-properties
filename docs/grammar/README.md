# Properties Grammar Specification

This directory contains the EBNF grammar specification for the Simple Properties Configuration Format.

## Files

### `properties.ebnf`

Complete grammar specification covering:

- File structure (zero or more lines)
- Property assignments (`key = value`)
- Comment lines (leading `#`)
- Key rules (must start with letter or underscore; may contain letters, digits, `_`, `-`, `.`)
- Value rules (any character except line terminators; no quoting or escaping)
- Whitespace rules (space and horizontal tab only; stripped around keys and values)
- Line terminator rules (`\n` and `\r\n`)
- Token type catalog
- Error conditions enumerated per specification section 4

This grammar is the authoritative formal specification for the parser implementations in
`internal/parser/` and `internal/fastparser/`.

---

## EBNF Notation

The grammar uses ISO-style EBNF notation:

| Notation | Meaning |
|----------|---------|
| `=` | Rule definition |
| `;` | End of rule |
| `{ ... }` | Zero or more repetitions |
| `[ ... ]` | Optional (zero or one) |
| `( ... )` | Grouping |
| `\|` | Alternative |
| `"..."` | Literal terminal string |
| `? ... ?` | Special sequence (prose description) |
| `A - B` | Difference: matches A but not B |
| `(* ... *)` | Comment |

---

## Grammar Rules

### File

```ebnf
file = { line } ;
```

A file is zero or more lines. An empty file is valid.

### Line

```ebnf
line = ws | comment | assignment ;
```

Each line is exactly one of: blank (whitespace only), a comment, or a property assignment.

### Assignment

```ebnf
assignment = ws key ws "=" ws value ws ;
```

The first `=` on the line is the separator. Whitespace around the key and value is stripped.
Anything after `=` through end of line is the value — there is no inline comment support.

### Comment

```ebnf
comment = ws "#" { character } ;
```

A comment begins when the first non-whitespace character on a line is `#`. The entire line
is ignored. Inline comments (after a value) are not supported.

### Key

```ebnf
key       = key-start { key-char } ;
key-start = letter | "_" ;
key-char  = letter | digit | "_" | "-" | "." ;
```

Keys must begin with a letter or underscore. Subsequent characters may also include digits,
hyphens, and dots. Examples of valid keys:

```
host
db.port
log-level
SERVICE_NAME
_internal
```

Keys are case-sensitive. Duplicate keys within a file are a parse error.

### Value

```ebnf
value      = { value-char } ;
value-char = character - line-terminator ;
```

A value is any sequence of non-line-terminator characters. Values may be empty (`key=` is
valid). No quoting or escape sequences — what you write is what you get. Leading and
trailing whitespace is stripped.

### Whitespace

```ebnf
ws = { " " | "\t" } ;
```

Whitespace is ASCII space (`0x20`) and horizontal tab (`0x09`) only.

### Line Terminator

```ebnf
line-terminator = "\n" | "\r\n" ;
```

Both Unix (`LF`) and Windows (`CRLF`) line endings are accepted. Mixed styles within a
single file are permitted.

### Character

```ebnf
character = ? UTF-8 character except control characters (except TAB) and NUL ? ;
```

Any valid UTF-8 character is permitted, subject to:
- NUL (`0x00`) is never allowed anywhere in input
- Control characters below `0x20` are not allowed, except horizontal tab (`0x09`)

---

## Token Types

| Token | Pattern | Description |
|-------|---------|-------------|
| `KEY` | `key-start { key-char }` | Property key |
| `EQUALS` | `"="` | Assignment separator |
| `VALUE` | `{ value-char }` | Property value (trimmed) |
| `COMMENT` | `"#" { character }` | Comment from `#` to end of line |
| `NEWLINE` | `"\n" \| "\r\n"` | Line terminator |
| `EOF` | special | End of input |

---

## Error Conditions

Six parse error conditions are specified (per specification section 4):

| # | Condition | Example |
|---|-----------|---------|
| 1 | Missing `=` separator | `hostname` (no equals sign) |
| 2 | Empty key | `=value` |
| 3 | Invalid key start character | `123key=val` |
| 4 | Duplicate keys | `host=a` then `host=b` |
| 5 | Control character (other than TAB) | key or value containing `\x01` |
| 6 | NUL byte | any embedded `\x00` |

All conditions produce an error with a line number.

---

## Implementation Mapping

Both parsers implement the same grammar but produce different output:

| Parser | Location | Output | Used By |
|--------|----------|--------|---------|
| Fast parser | `internal/fastparser/` | `map[string]string` | `Load`, `Validate` |
| AST parser | `internal/parser/` | `ast.SchemaNode` | `Parse` |

Each grammar rule maps to a parse function in both implementations:

| Grammar Rule | Function |
|-------------|----------|
| `file` | `Parse()` / `Validate()` |
| `line` | `parseLine()` |
| `assignment` | `parseAssignment()` |
| `comment` | `parseComment()` |
| `key` | `parseKey()` |
| `value` | `parseValue()` |

---

## Related

- [properties-format.md](../../properties-format.md) — Human-readable format specification
- [internal/parser/grammar_test.go](../../internal/parser/grammar_test.go) — Grammar verification tests
- [Shape ADR 0005](https://github.com/shapestone/shape-core/blob/main/docs/adr/0005-grammar-as-verification.md) — Grammar-as-Verification policy
