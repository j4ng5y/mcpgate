# MCPGate Testing Guide

This document describes the testing strategy and how to run tests for MCPGate.

## Test Architecture

MCPGate uses comprehensive testing across multiple levels:

### Unit Tests
- **Location**: `*_test.go` files in each package
- **Coverage**: Individual components and functions
- **Run**: `make test`

### Integration Tests
- **Location**: `integration_test.go` (root package)
- **Coverage**: Multi-component workflows and end-to-end scenarios
- **Run**: `make test-integration`
- **Tag**: `+build integration`

## Test Files

### Config Package (`config/config_test.go`)
Tests for TOML configuration loading and validation:
- Valid configuration parsing
- Default values
- Multiple servers
- All transport types
- Error handling (missing files, invalid TOML)
- Metadata and environment variables

**Test Count**: 10 tests

### Transport Package (`transport/transport_test.go`)
Tests for transport layer abstraction:
- Factory pattern implementation
- All transport type creation (stdio, HTTP, WebSocket, Unix)
- Connection state management
- Error handling for missing configuration
- Concurrency safety

**Test Count**: 14 tests

### Server Package
#### Registry (`server/registry_test.go`)
- Server registration and unregistration
- Duplicate registration prevention
- Server retrieval and listing
- Capability-based filtering
- Concurrent operations

**Test Count**: 9 tests

#### Managed Server (`server/managed_test.go`)
- Server creation and initialization
- Connection state tracking
- Capability management
- Metadata handling
- Concurrency safety

**Test Count**: 10 tests

#### Manager (`server/manager_test.go`)
- Server lifecycle management
- Enabled/disabled server handling
- Server registry operations
- Reconnection logic
- Multi-server scenarios
- Error handling

**Test Count**: 13 tests

### MCP Package (`mcp/router_test.go`)
Tests for request routing and JSON-RPC protocol:
- Gateway-level methods (list_servers, get_server, etc.)
- Request routing to upstream servers
- Parameter validation
- Error responses
- Request ID propagation
- Capability extraction from method names

**Test Count**: 13 tests

### Pool Package (`pool/connection_pool_test.go`)
Tests for connection pooling:
- Pool creation and configuration
- Statistics and health monitoring
- Connection lifecycle
- Concurrency safety
- Idle connection cleanup
- Pool closure

**Test Count**: 13 tests

### Integration Tests (`integration_test.go`)
End-to-end workflow tests:
- Full configuration load and server management
- Multiple server scenarios
- Router operations with multiple servers
- Registry operations
- Response structure validation
- Error handling workflows
- Concurrent request processing
- Context timeout handling

**Test Count**: 10 tests

## Running Tests

### Unit Tests Only
```bash
make test
```

Displays:
- Test execution with verbose output
- Race condition detection
- Coverage percentage
- Summary of test results

### Unit Tests with Coverage HTML Report
```bash
make test-coverage-html
```

Generates `coverage.html` in the project root with visual coverage breakdown by file and function.

### Integration Tests Only
```bash
make test-integration
```

Runs tests tagged with `+build integration` to test multi-component workflows.

### All Tests
```bash
make test-all
```

Runs both unit and integration tests with full coverage reporting.

### Quick Test (Without Coverage)
```bash
go test -v ./...
```

### Test Specific Package
```bash
go test -v ./config
go test -v ./transport
go test -v ./server
go test -v ./mcp
go test -v ./pool
```

### Test Specific Function
```bash
go test -v -run TestName ./package
# Example:
go test -v -run TestRegistry_Register ./server
```

## Test Coverage

Target coverage goals:
- **Overall**: ≥ 80% coverage
- **Critical paths**: ≥ 90% coverage
- **Utilities**: ≥ 70% coverage

Current coverage can be viewed by:
1. Running `make test` and checking the summary line
2. Running `make test-coverage-html` for detailed visual report

## Test Patterns

### Error Handling
Tests verify error paths for:
- Missing configuration
- Invalid parameters
- Unavailable resources
- Timeout scenarios

### Concurrency
Critical tests include:
- Concurrent reads on shared resources
- Thread-safe state management
- Race condition detection with `-race` flag

### Mocking/Stubbing
Tests use:
- In-memory configurations
- Temporary files (cleaned up after tests)
- Real subprocess communication for transport tests
- Isolated component testing

## Continuous Integration

Recommended CI workflow:

```bash
# On every commit
make test

# On pull requests
make test-all
make lint
make build

# Before release
make test-all
make build
make release
```

## Common Test Issues

### Transport Tests Hanging
- Some transport tests actually spawn processes
- Add timeout context if needed
- Use `TIMEOUT=30s go test ./transport`

### Integration Test Build Tag Issues
- Ensure Go version ≥ 1.25.5
- Use `go test -tags=integration ./...` if make fails

### Race Condition Failures
- Race detector adds ~20% overhead
- Use `-race` flag in CI but not always locally
- If a race is detected, synchronize access to shared state

## Adding New Tests

When adding new features:

1. **Create `feature_test.go`** in the appropriate package
2. **Follow naming convention**: `TestPackageName_Functionality`
3. **Test both success and failure paths**
4. **Add concurrency tests** for shared resources
5. **Clean up resources** (temp files, goroutines, etc.)
6. **Run locally** before committing:
   ```bash
   make test-all
   ```

## Test Maintenance

Tests should be updated when:
- API changes
- New transport types are added
- Configuration schema changes
- Router logic is modified
- Pool implementation is changed

Regular maintenance:
- Review coverage reports
- Fix flaky tests
- Remove obsolete tests
- Keep dependencies updated

## Performance Testing

For performance-sensitive code:

```bash
go test -bench=. -benchmem ./package
```

Current performance tests:
- Router request throughput
- Pool connection reuse efficiency
- Registry lookup performance

## Debugging Tests

Enable verbose output and tracing:

```bash
go test -v -run TestName ./package -trace trace.out
go tool trace trace.out  # Opens browser with trace analysis
```

Debug output:
```bash
go test -v -run TestName ./package --log.level=debug
```

---

**Last Updated**: 2025-12-16
**Test Framework**: Go `testing` package
**Coverage Tool**: `go tool cover`
