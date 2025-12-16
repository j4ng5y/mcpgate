# Security Policy

## Reporting a Vulnerability

**Please do not open a public issue on GitHub for security vulnerabilities.**

Instead, please report security vulnerabilities by emailing security concerns privately.

### Disclosure Process

1. Report the vulnerability with as much detail as possible
2. Allow time for investigation and patch development
3. Coordinate the public disclosure timing

We appreciate your responsible disclosure and will acknowledge your report within 48 hours.

## Security Best Practices

When using MCPGate:

### Configuration Security

- **Sensitive Data**: Avoid storing credentials in configuration files
- **File Permissions**: Ensure config files have appropriate permissions (e.g., `0600`)
- **Environment Variables**: Use environment variables for sensitive values
- **Secrets Management**: For production, use proper secrets management tools (e.g., HashiCorp Vault, AWS Secrets Manager)

### Subprocess Security (stdio mode)

- **Command Validation**: Only execute trusted commands
- **Path Safety**: Use full paths for executables, not relative paths
- **Environment**: Be cautious with environment variables passed to subprocesses

### Network Security (HTTP/WebSocket mode)

- **HTTPS**: Use HTTPS URLs for remote connections
- **TLS Verification**: Ensure proper certificate validation
- **Authentication**: Implement proper authentication mechanisms
- **Firewall Rules**: Restrict network access appropriately

### Unix Socket Security

- **Socket Permissions**: Set appropriate permissions on socket files
- **Socket Location**: Place sockets in secure directories
- **File Ownership**: Ensure proper file ownership

## Dependency Security

MCPGate actively monitors dependencies for security vulnerabilities:

- **Dependabot**: Automatically creates PRs for dependency updates
- **Security Patches**: Applied promptly for all high/critical vulnerabilities
- **Regular Audits**: `go list -json -m all | nancy sleuth` for vulnerability scanning

## Code Security

- **Memory Safety**: Go's memory safety prevents certain classes of vulnerabilities
- **Input Validation**: All inputs are validated and sanitized
- **Error Handling**: Proper error handling without information disclosure
- **Logging**: Secure logging without sensitive data exposure

## Supported Versions

| Version | Supported          |
|---------|-------------------|
| 1.x     | :white_check_mark: |

Security patches are applied to the latest version. We recommend always running the latest release.

## Security Checklist for Deployments

Before deploying MCPGate to production:

- [ ] Review and restrict configuration file permissions
- [ ] Use environment variables for secrets
- [ ] Configure firewall rules appropriately
- [ ] Enable TLS/HTTPS for remote connections
- [ ] Implement proper logging and monitoring
- [ ] Run on a dedicated user account (not root)
- [ ] Keep Go runtime updated
- [ ] Keep all dependencies updated
- [ ] Review upstream MCP servers for security
- [ ] Implement rate limiting if exposed to network

## Acknowledgments

Thank you to all security researchers who have helped improve MCPGate's security.
