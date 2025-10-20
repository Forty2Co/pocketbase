# Testing Guide

This project has two types of tests with convenient Makefile targets for easy execution:

## Quick Start

```bash
# Run unit tests only (fast, no server required)
make test-unit

# Run all tests with automatic server management (recommended)
make test-integration

# Manual test execution (requires running server on :8090)
make test
```

## Unit Tests

Unit tests are located in each sub-package and test individual components in isolation using mocks. They don't require external dependencies and run quickly.

**Run unit tests:**
```bash
# Using Makefile (recommended)
make test-unit

# Direct go command
go test -short -shuffle=on -race ./...
```

**Unit test locations:**
- `auth/auth_test.go` - Tests for authentication functionality
- `realtime/subscribe_test.go` - Tests for real-time subscription functionality  
- `admin/backup_test.go` - Tests for admin backup functionality
- `collections/collection_test.go` - Tests for collection operations

## Integration Tests

Integration tests are located in `integration_test.go` and test the complete functionality end-to-end with a real PocketBase server.

**Run integration tests:**
```bash
# Automatic server management (recommended)
make test-integration
# This will:
# 1. Build the PocketBase binary
# 2. Start server in background on 127.0.0.1:8090
# 3. Run all tests (unit + integration)
# 4. Stop the server automatically
# 5. Show detailed test summary

# Manual execution (requires server already running)
make test
# Expects PocketBase server running on :8090

# Run specific integration test
go test -v -run TestCollection_List_Integration
```

**Integration test categories:**
- **Auth Integration Tests** - Authentication, password reset, email verification
- **Realtime Integration Tests** - Real-time subscriptions, reconnection handling
- **Collections Integration Tests** - CRUD operations, querying
- **Admin Integration Tests** - Backup operations, file management

## Server Management

The Makefile provides convenient server management commands:

```bash
# Start server in foreground (development)
make serve

# Start server in background
make serve-bg

# Stop background server
make serve-stop

# Check server status
make serve-status

# Restart server
make serve-restart
```

## Test Organization

The tests are organized to avoid import cycles:

- **Unit tests** in sub-packages test their specific functionality without importing the main package
- **Integration tests** use the `pocketbase_test` package to import and test the main package functionality
- **Original test files** in the root directory are preserved for reference but may contain import cycle issues

## Test Configuration

Tests use the following configuration:
- **Server URL**: `http://127.0.0.1:8090` (configurable via `SERVER_HOST` and `SERVER_PORT`)
- **Test flags**: `-shuffle=on -race` for better test reliability
- **Short mode**: `-short` flag skips integration tests

## Running Tests in CI/CD

For automated testing environments:

```bash
# Unit tests only (no external dependencies)
make test-unit

# Full test suite with automatic server management
make test-integration

# Custom configuration
SERVER_HOST=0.0.0.0 SERVER_PORT=9090 make test-integration
```

## Test Coverage

To check test coverage:

```bash
# Unit test coverage
go test -cover -short ./...

# Integration test coverage (with server)
make serve-bg
go test -cover ./...
make serve-stop

# Combined coverage with automatic server management
make test-integration 2>&1 | grep coverage
```

## Troubleshooting

**Server connection issues:**
- Ensure no other process is using port 8090
- Check server status with `make serve-status`
- View server logs when running `make serve`

**Test failures:**
- Run `make test-unit` first to isolate unit test issues
- Check if server is properly started with `make serve-bg && make serve-status`
- Use `go test -v -run SpecificTest` to debug individual tests