# Contributing to MCPGate

Thank you for your interest in contributing to MCPGate! This document provides guidelines and instructions for contributing.

## Getting Started

### Prerequisites

- Go 1.25 or later
- git
- golangci-lint (for linting)
- goreleaser (for building releases)
- UPX (optional, for binary compression)

### Setup Development Environment

```bash
# Clone the repository
git clone https://github.com/j4ng5y/mcpgate.git
cd mcpgate

# Install dependencies
go mod download

# Verify setup
make test
```

## Development Workflow

### 1. Create a Branch

```bash
# Create a feature branch
git checkout -b feature/your-feature-name

# Or for bug fixes
git checkout -b fix/your-bug-name
```

### 2. Make Your Changes

- Follow Go conventions and idioms
- Keep commits focused and logical
- Write clear commit messages

### 3. Write or Update Tests

- Add tests for new functionality
- Update tests for modified functionality
- Ensure all tests pass: `make test`
- Aim for >70% code coverage

### 4. Code Quality

Ensure your code passes all quality checks:

```bash
# Format code
make fmt

# Run linter
make lint

# Run all tests
make test

# Generate coverage report
make test-coverage-html
```

### 5. Commit and Push

```bash
# Stage changes
git add .

# Commit with meaningful message
git commit -m "feat: add new feature" -m "Detailed description of changes"

# Push to your fork
git push origin feature/your-feature-name
```

### 6. Create a Pull Request

- Use the provided PR template
- Reference any related issues
- Ensure all CI checks pass
- Request review from maintainers

## Code Style Guidelines

### Go Conventions

- Follow [Effective Go](https://golang.org/doc/effective_go)
- Use `gofmt` for formatting (auto-run via `make fmt`)
- Use `golint` recommendations
- Keep functions focused and small

### Naming Conventions

- Use descriptive variable and function names
- Use `CamelCase` for exported items
- Use `camelCase` for unexported items
- Use full words, not abbreviations (except standard ones like `ctx`, `err`)

### Comments

- Add package-level documentation
- Document exported functions and types
- Use clear, concise language
- Explain the "why", not just the "what"

### Error Handling

- Always handle errors appropriately
- Use error wrapping with `fmt.Errorf("%w", err)`
- Don't ignore errors unless intentional (use `_ = value`)
- Provide helpful error messages

## Testing Guidelines

### Writing Tests

```go
// Test function naming: TestFunctionName_Scenario_Expected
func TestAgent_InjectStdio_MemoryConfig(t *testing.T) {
    // Arrange
    agent := NewTestAgent()

    // Act
    err := agent.InjectStdio("cmd", []string{"arg"}, "name", map[string]interface{}{})

    // Assert
    if err != nil {
        t.Fatalf("expected no error, got %v", err)
    }
}
```

### Test Coverage

- Unit tests: Test individual functions in isolation
- Integration tests: Test interactions between components
- Aim for >70% coverage
- Use table-driven tests for multiple scenarios

### Running Tests

```bash
# Run all tests
make test

# Run specific test
go test -v -run TestName ./...

# Run with race detector
go test -race ./...

# Generate HTML coverage report
make test-coverage-html
```

## Documentation

### Code Documentation

- Add godoc comments to all exported items
- Keep documentation up-to-date with code changes
- Include examples where helpful

### README Updates

- Update README.md if adding new features
- Document configuration options
- Add examples for new functionality

### Commit Messages

Format commit messages as:

```
<type>(<scope>): <subject>

<body>

<footer>
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, missing semicolons, etc.)
- `refactor`: Code refactoring
- `perf`: Performance improvements
- `test`: Test additions or updates
- `chore`: Build, dependency, or tooling changes

Examples:
```
feat(inject): add support for new agent
fix(config): handle missing config file gracefully
docs: update installation instructions
```

## Pull Request Process

1. **Before Submitting**
   - All tests pass: `make test`
   - Linting passes: `make lint`
   - Code is formatted: `make fmt`
   - Coverage hasn't decreased
   - No security issues

2. **PR Description**
   - Use the PR template
   - Clearly describe changes
   - Link related issues
   - Include testing instructions

3. **Review Process**
   - Address feedback promptly
   - Engage in constructive discussion
   - Push updates to the same branch
   - No need to recreate PR

4. **Merging**
   - All checks must pass
   - At least one approval required
   - Squash commits when appropriate
   - Delete branch after merging

## Issue Reporting

### Bug Reports

Include:
- Go version
- Operating system
- MCPGate version
- Reproducible steps
- Expected vs actual behavior
- Relevant logs

### Feature Requests

Include:
- Clear use case
- Why it's needed
- Proposed implementation (optional)
- Alternatives considered

## Release Process

Releases are automated via GitHub Actions:

1. Create and push a git tag:
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```

2. GitHub Actions automatically:
   - Runs all tests
   - Builds multi-platform binaries
   - Creates GitHub release
   - Uploads artifacts

## Development Tips

### Local Build

```bash
# Quick snapshot build
make build

# Run tests during development
make test

# Watch for changes and rebuild
go build -o bin/mcpgate ./main.go
```

### Debugging

```bash
# Build with debug symbols
go build -gcflags="all=-N -l" -o bin/mcpgate ./main.go

# Use dlv debugger
dlv debug ./main.go
```

### Performance Profiling

```bash
# CPU profiling
go test -cpuprofile=cpu.prof -bench=. ./...

# Memory profiling
go test -memprofile=mem.prof -bench=. ./...

# View profiles
go tool pprof cpu.prof
```

## Community

- Discussions: Use GitHub Discussions for questions
- Issues: Use GitHub Issues for bugs and features
- Security: See SECURITY.md for security concerns

## Code of Conduct

- Be respectful and inclusive
- Assume good intentions
- Focus on the code, not the person
- Provide constructive feedback
- Welcome different perspectives

## Resources

- [Go Documentation](https://golang.org/doc/)
- [Effective Go](https://golang.org/doc/effective_go)
- [Go Code Review Comments](https://golang.org/wiki/CodeReviewComments)
- [golangci-lint](https://golangci-lint.run/)

## Questions?

- Check existing issues and discussions
- Open a new discussion
- Ask in a pull request
- Email the maintainers

Thank you for contributing to MCPGate!
