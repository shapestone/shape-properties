# Security Policy

## Supported Versions

We actively support shape-properties with security updates:

- **Latest Release**: Fully supported with security updates
- **Previous Minor Version**: Supported for critical security fixes (30 days)
- **Older Versions**: No longer supported

We strongly recommend always using the latest release for the best security posture.

## Reporting a Vulnerability

We take security vulnerabilities seriously. If you discover a security issue in
shape-properties, please report it responsibly.

### How to Report

**Please do NOT report security vulnerabilities through public GitHub issues.**

Instead, please use one of the following methods:

1. **GitHub Private Vulnerability Reporting** (Preferred)
   - Navigate to the [Security tab](https://github.com/shapestone/shape-properties/security)
     of this repository
   - Click "Report a vulnerability"
   - Fill out the vulnerability report form

2. **Email**
   - Send details to: security@shapestone.com
   - Include "shape-properties Security" in the subject line

### What to Include

Please provide the following information in your report:

- **Description**: Clear description of the vulnerability
- **Impact**: Potential impact and attack scenario
- **Reproduction**: Step-by-step instructions to reproduce the issue
- **Affected Versions**: Which versions are affected
- **Proof of Concept**: Code snippets or test cases (if applicable)
- **Suggested Fix**: If you have ideas for remediation

### Response Timeline

We are committed to responding promptly to security reports:

- **Acknowledgment**: Within 48 hours of receiving your report
- **Initial Assessment**: Within 5 business days
- **Regular Updates**: Every 7 days until resolved
- **Patch Release**: Within 30 days for high/critical severity issues

### Disclosure Policy

- We follow a **coordinated disclosure** process
- We will work with you to understand and validate the issue
- Once a fix is ready, we will:
  1. Release a patch version
  2. Publish a security advisory (GitHub Security Advisory)
  3. Credit you in the advisory (unless you prefer to remain anonymous)
- We request that you do not publicly disclose the vulnerability until we have
  released a fix

## Security Considerations

### Known Attack Surfaces

shape-properties is a parser library that processes untrusted input. Be aware of these
potential security considerations:

1. **Denial of Service (DoS)**
   - Very large input files may consume significant memory and CPU
   - The fast path reads the entire input into memory before parsing
   - There is no built-in limit on input size

2. **Resource Exhaustion**
   - A file with a very large number of properties will allocate a proportionally
     large `map[string]string`
   - The AST path additionally allocates `ast.ObjectNode` and `ast.LiteralNode` objects
     for every property

3. **Input Validation**
   - shape-properties validates key format, rejects NUL bytes, and rejects
     control characters — but does not validate value semantics
   - Type conversion (e.g., parsing a value as an integer) is the caller's responsibility
   - Always validate parsed values against expected ranges and formats

4. **No Code Execution**
   - The format does not support variable expansion, includes, or executable expressions
   - A properties file cannot cause code execution through parsing alone

### Best Practices

When using shape-properties in production with untrusted input:

- **Limit Input Size**: Enforce a maximum byte size before calling `Load` or `Validate`

  ```go
  const maxConfigSize = 1 * 1024 * 1024 // 1 MB
  if len(input) > maxConfigSize {
      return fmt.Errorf("config file too large: %d bytes", len(input))
  }
  props, err := properties.Load(input)
  ```

- **Use ValidateReader with a LimitedReader**: When reading from untrusted sources

  ```go
  limited := io.LimitedReader{R: untrustedReader, N: maxConfigSize}
  if err := properties.ValidateReader(&limited); err != nil {
      return fmt.Errorf("invalid config: %w", err)
  }
  ```

- **Validate Values After Parsing**: Do not trust that a value is a valid integer,
  URL, or path just because it parsed successfully

  ```go
  props, _ := properties.Load(input)
  port, err := strconv.Atoi(props["port"])
  if err != nil || port < 1 || port > 65535 {
      return fmt.Errorf("invalid port: %q", props["port"])
  }
  ```

- **Keep Updated**: Regularly update to the latest version

- **Monitor Resources**: Track memory usage if parsing config files from untrusted sources

## Security Updates

Security updates are released as patch versions and announced via:

- GitHub Security Advisories
- Release notes in CHANGELOG.md
- GitHub Releases page

Subscribe to releases or watch this repository to stay informed.

## Recognition

We appreciate the security research community's efforts. Researchers who responsibly
disclose vulnerabilities will be credited in:

- The security advisory (with their permission)
- A `SECURITY_ACKNOWLEDGMENTS.md` file (if applicable)
- Release notes for the security fix

## Questions?

If you have questions about this security policy or shape-properties' security posture,
please open a public discussion in GitHub Discussions or contact security@shapestone.com.

---

Thank you for helping keep shape-properties and its users safe!
