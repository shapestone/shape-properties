# Simple Properties Configuration Format

This document defines a **minimal, deterministic “properties-style” configuration format**
based on `key=value` pairs.

The goals of this format are:

- Human readability
- Trivial parsing
- Predictable behavior
- Safe ingestion (no surprising edge cases)
- Forward compatibility

This is **not** intended to be compatible with Java `.properties`, dotenv, or shell syntax,
though optional compatibility modes are discussed at the end.

---

## 1. File Structure

A configuration file consists of zero or more **property assignments**, one per line.

```
host=localhost
port=1234
```

The file is interpreted as UTF-8 text.

---

## 2. Lexical Rules

### 2.1 Line Endings

- Lines are terminated by `\n` or `\r\n`
- The final line MAY end without a newline

---

### 2.2 Whitespace

- Whitespace is defined as ASCII space (`0x20`) or horizontal tab (`0x09`)
- Leading and trailing whitespace around keys and values is ignored
- Whitespace **inside values** is preserved

---

### 2.3 Comments

- A line whose first non-whitespace character is `#` is a comment
- Comments extend to the end of the line

```
# this is a comment
port=1234   # this is NOT a comment
```

Inline comments are **not supported**.

---

## 3. Property Assignments

Each assignment has the form:

```
key = value
```

The first `=` character separates the key from the value.

---

### 3.1 Keys

Keys MUST match the following pattern:

```
[A-Za-z_][A-Za-z0-9_.-]*
```

Rules:
- Keys are **case-sensitive**
- Keys MUST be unique within a file
- Duplicate keys are an error

Examples:
```
host
db.port
log-level
SERVICE_NAME
```

---

### 3.2 Values

- Values are UTF-8 strings
- Values MAY be empty
- Values MAY contain any characters except line terminators
- No escaping is performed
- No quoting is performed

Examples:
```
path=/var/log/app
empty=
message=hello world
```

---

## 4. Invalid Input

The following conditions MUST cause a parse error:

- Missing `=` separator
- Empty key
- Invalid key characters
- Duplicate keys
- Control characters other than TAB in keys or values
- NUL (`0x00`) anywhere

---

## 5. Semantics

- The configuration represents a **map of string keys to string values**
- Ordering has no semantic meaning
- All values are opaque strings
- Type conversion (e.g. integer parsing) is the responsibility of the consumer

---

## 6. Formal Grammar (EBNF)

The following grammar uses ISO-style EBNF.

```
file        = { line } ;

line        = ws
            | comment
            | assignment ;

assignment  = ws key ws "=" ws value ws ;

comment     = ws "#" { character } ;

key         = key-start { key-char } ;

key-start   = letter | "_" ;
key-char    = letter | digit | "_" | "-" | "." ;

value       = { value-char } ;

value-char  = character - line-terminator ;

ws          = { " " | "\t" } ;

letter      = "A"…"Z" | "a"…"z" ;
digit       = "0"…"9" ;
```

---

## 7. Forward-Compatible Extensions

Parsers SHOULD ignore lines starting with an unknown sigil followed by `:`.

This allows future extensions without breaking older parsers.

Example (ignored by current spec):
```
@include: common.properties
```

Recommended extension rules:
- Extension lines MUST start with a non-alphanumeric character
- Extensions MUST NOT affect core key/value parsing
- Core parsers MAY expose extension lines as metadata

---

## 8. Explicit Non-Goals

This format intentionally does **not** support:

- Multiline values
- Line continuation (`\`)
- Variable expansion
- Quoting or escaping
- Nested structures
- Arrays
- Expressions
- Environment interpolation

If you need those features, use YAML, TOML, or JSON instead.

---

## 9. Compatibility Notes (Non-Normative)

### 9.1 Java `.properties`
Not compatible:
- Java allows `:` as a separator
- Java supports escapes and line continuation
- Java allows duplicate keys (last-wins)

Supporting Java compatibility significantly increases parser complexity.

### 9.2 dotenv
Not compatible:
- dotenv allows `export`
- dotenv semantics vary across implementations
- Quoting and interpolation rules differ

### 9.3 Shell
Not compatible:
- No variable expansion
- No command substitution
- No quoting rules

---

## 10. Design Rationale

This format optimizes for:
- Safety over convenience
- Explicitness over cleverness
- Configuration as data, not code

It is deliberately boring.

---

## 11. Summary

- `key=value`
- One assignment per line
- No magic
- No surprises
- Easy to parse correctly

That is the entire point.
