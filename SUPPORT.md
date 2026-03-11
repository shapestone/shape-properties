# Support

Thank you for using shape-properties! This document will help you find the right
resources for getting help.

## Getting Help

### Documentation

Before opening an issue, please check our documentation:

- **[README.md](README.md)** - Quick start, installation, API reference, and examples
- **[ARCHITECTURE.md](ARCHITECTURE.md)** - How shape-properties works internally
- **[properties-format.md](properties-format.md)** - Complete format specification
- **[docs/TESTING.md](docs/TESTING.md)** - Testing strategy and coverage guide
- **[API Reference](https://pkg.go.dev/github.com/shapestone/shape-properties)** - Generated Go package documentation
- **[CHANGELOG.md](CHANGELOG.md)** - Recent changes

**Most questions are answered in these resources. Please check them first!**

---

## Questions and Discussions

**Have a question about using shape-properties?**

### Where to Ask

- **GitHub Discussions** - [Ask questions here](https://github.com/shapestone/shape-properties/discussions)
  - Best for: General questions, usage help, design discussions
  - Community-driven with maintainer participation
  - Searchable by others with similar questions

- **Search Closed Issues** - [Check if it has been asked before](https://github.com/shapestone/shape-properties/issues?q=is%3Aissue+is%3Aclosed)
  - Many questions have already been answered

### Common Questions

Before asking, check if your question is covered in the guides:

- **"How do I load a config file?"** - See [README Quick Start](README.md#quick-start) and
  [README Examples](README.md#examples)
- **"Which API should I use: Load or Parse?"** - See
  [README Dual-Path Architecture](README.md#dual-path-architecture)
- **"Why is Load faster than Parse?"** - See [ARCHITECTURE.md](ARCHITECTURE.md#dual-path-design)
- **"How do I validate user-supplied config?"** - See
  [README Examples](README.md#examples)
- **"Can I use shape-properties with concurrent goroutines?"** - See
  [README Thread Safety](README.md#thread-safety)
- **"What characters are allowed in keys?"** - See
  [properties-format.md](properties-format.md#keys)
- **"Is this compatible with Java .properties files?"** - See
  [properties-format.md](properties-format.md#compatibility-notes-non-normative)

### Please Do NOT Open Issues for Questions

**GitHub Issues are for bug reports and feature requests only.**

Opening issues for usage questions:
- Clutters the issue tracker
- Makes it harder to track real bugs
- Takes longer to get answered

Use GitHub Discussions instead — you will get faster, better answers!

---

## Reporting Bugs

Found a bug? Please open an issue using our bug report template.

### Before Reporting

1. **Search existing issues** - Check if it is already reported:
   [Open Issues](https://github.com/shapestone/shape-properties/issues)
2. **Verify your version** - Make sure you are using the latest release:
   ```bash
   go list -m github.com/shapestone/shape-properties
   ```
3. **Check the CHANGELOG** - It may be a known issue that is already fixed
4. **Create minimal reproduction** - Reduce to the smallest input that shows the problem

### What to Include

Your bug report should include:

- **shape-properties version** - From `go list -m github.com/shapestone/shape-properties`
- **Go version** - From `go version`
- **Operating system** - macOS, Linux, Windows, etc.
- **Minimal input** - The smallest properties string that reproduces the issue
- **Expected behavior** - What should happen (e.g., "should parse successfully")
- **Actual behavior** - What actually happens (e.g., "returns error: ...")
- **Error messages** - Full error output if applicable

**Good bug reports save everyone time!**

### What Happens Next

- We aim to acknowledge bugs within **2-3 business days**
- Parsing correctness bugs are prioritized
- You may be asked for additional information or a smaller reproduction case
- Once confirmed, we will add appropriate labels

---

## Feature Requests

Have an idea for an improvement? Please open an issue with your proposal.

### Before Requesting

1. **Check existing requests** - Search
   [enhancement issues](https://github.com/shapestone/shape-properties/issues?q=is%3Aissue+label%3Aenhancement)
2. **Review our scope** - See [Scope Policy](CONTRIBUTING.md#scope-policy)
3. **Consider fit** - Does it align with shape-properties' mission (parsing the Simple
   Properties Configuration Format)?

### What We Accept

- Parsing correctness improvements — better error messages, edge case handling
- Performance improvements — speed, memory reduction, allocation reduction
- API ergonomics — new convenience functions that do not change existing behavior
- Documentation improvements — examples, guides, clarifications
- Test coverage — additional tests for uncovered edge cases
- Tooling improvements — CI, Makefile, benchmark infrastructure

### What We Generally Do Not Accept

- Support for Java `.properties` escaping or `:` separators
- dotenv-style variable expansion, quoting, or `export` support
- Multiline values or line continuation (`\`)
- Nested structures, arrays, or type annotations
- Breaking changes to the public API

See our [Contributing Guide](CONTRIBUTING.md) for details on scope.

### Response Timeline

- Feature requests are reviewed during planning cycles
- We may ask clarifying questions about use cases
- Not all requests will be accepted (scope, complexity, maintenance burden)
- Rejected requests will receive a clear explanation

---

## Security Vulnerabilities

**Do NOT open public issues for security vulnerabilities.**

Security issues require private disclosure to protect users.

### How to Report

**Preferred method:**
1. Go to the [Security tab](https://github.com/shapestone/shape-properties/security)
2. Click "Report a vulnerability"
3. Fill out the private vulnerability report form

**Alternative:**
- Email: security@shapestone.com
- Subject: "shape-properties Security Issue"

### Our Commitment

- **Acknowledgment**: Within 48 hours
- **Initial assessment**: Within 5 business days
- **Regular updates**: Every 7 days until resolved
- **Patch release**: Within 30 days for high/critical issues

See our complete [Security Policy](SECURITY.md) for details.

---

## Response Times

shape-properties is an open source project maintained by Shapestone. We aim to respond
within these timeframes:

| Type | Response Time | Notes |
|------|---------------|-------|
| **Security issues** | 48 hours | See [SECURITY.md](SECURITY.md) for full policy |
| **Bug reports** | 2-3 business days | Parsing correctness bugs prioritized |
| **Feature requests** | Reviewed during planning | May take longer for complex proposals |
| **Questions on Discussions** | Best effort | Community-driven; maintainers participate when available |
| **Pull requests** | 3-5 business days | Initial review; may require iterations |

**Note:** These are goals, not guarantees. Response times may vary based on maintainer
availability, holidays, and issue complexity.

---

## Contributing

Want to contribute code, documentation, or tests?

See our **[Contributing Guide](CONTRIBUTING.md)** for:
- Development setup instructions
- Code style and testing requirements
- Pull request process
- What kinds of contributions we are looking for

Quick links:
- **[Development Setup](CONTRIBUTING.md#development-setup)**
- **[Testing Guidelines](CONTRIBUTING.md#testing-guidelines)**
- **[Pull Request Process](CONTRIBUTING.md#pull-request-process)**

---

## Community Guidelines

All interactions in the shape-properties community (issues, discussions, PRs) are
governed by community standards of respectful and constructive engagement.

Expected behavior:
- Be respectful and inclusive
- Provide constructive feedback
- Focus on what is best for the project and community
- Show empathy toward other community members

To report violations, contact: conduct@shapestone.com

---

## Additional Resources

### Learning Resources

- **[Go Documentation](https://go.dev/doc/)** - General Go programming help
- **[Shape Core](https://github.com/shapestone/shape-core)** - Universal AST and
  tokenizer framework
- **[Shape Ecosystem](https://github.com/shapestone/shape)** - Multi-format parser
  ecosystem

### Related Projects

- **[shape-core](https://github.com/shapestone/shape-core)** - Core infrastructure
  (AST types used by shape-properties)
- **[shape-json](https://github.com/shapestone/shape-json)** - JSON parser in the
  same ecosystem
- **[shape](https://github.com/shapestone/shape)** - Multi-format parser library

---

## What We Do Not Support

To keep the project focused and maintainable, we generally do not provide:

- Support for EOL Go versions (we support Go 1.25+, see [go.mod](go.mod))
- Custom implementation help — we cannot debug your specific application code
- Third-party integration debugging — issues with other libraries that use
  shape-properties
- Format extensions — see our [scope policy](CONTRIBUTING.md#scope-policy)
- Java `.properties`, dotenv, or shell compatibility — shape-properties is the
  Simple Properties Configuration Format only

---

## Quick Reference

**I want to...**

- Ask a question - [GitHub Discussions](https://github.com/shapestone/shape-properties/discussions)
- Report a bug - [Open an Issue](https://github.com/shapestone/shape-properties/issues/new)
- Request a feature - [Open an Issue](https://github.com/shapestone/shape-properties/issues/new)
- Report a security issue - [Private Vulnerability Reporting](https://github.com/shapestone/shape-properties/security)
- Learn how to use shape-properties - [README.md](README.md)
- Contribute code - [CONTRIBUTING.md](CONTRIBUTING.md)
- Understand the architecture - [ARCHITECTURE.md](ARCHITECTURE.md)
- Read the format spec - [properties-format.md](properties-format.md)
- See recent changes - [CHANGELOG.md](CHANGELOG.md)

---

## Still Need Help?

If you have:
- Checked the documentation
- Searched existing issues and discussions
- Asked on GitHub Discussions
- Still cannot find an answer

Then please open a **[discussion](https://github.com/shapestone/shape-properties/discussions)** with:
- What you are trying to accomplish
- What you have already tried
- Specific error messages or unexpected behavior
- A minimal properties string that demonstrates the issue

The community and maintainers will do their best to help!

---

Thank you for being part of the shape-properties community!

---

*Last Updated: March 10, 2026*
