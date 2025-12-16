---
name: Bug Report
about: Report a bug to help us improve
title: "[BUG] "
labels: bug
assignees: ""
---

## Description

A clear and concise description of what the bug is.

## Environment

- **OS**: (e.g., Ubuntu 22.04, macOS 13.0, Windows 11)
- **Go Version**: (output of `go version`)
- **MCPGate Version**: (output of `mcpgate --version` or commit hash)
- **Architecture**: (amd64, arm64, arm)

## Steps to Reproduce

1. First step
2. Second step
3. ...

## Expected Behavior

A clear and concise description of what you expected to happen.

## Actual Behavior

A clear and concise description of what actually happened.

## Configuration

If relevant, please share your configuration (redact any sensitive information):

```toml
[gateway]
log_level = "info"

[[server]]
name = "example"
transport = "stdio"
enabled = true
command = "example-command"
args = []
```

## Logs

If available, please share relevant log output:

```
[paste logs here]
```

## Screenshots

If applicable, add screenshots to help explain your problem.

## Additional Context

Add any other context about the problem here.

## Possible Solution

If you have a suggestion for how to fix this, please include it here.
