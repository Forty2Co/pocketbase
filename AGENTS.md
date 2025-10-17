# AGENTS.md - AI Agent Development Guide

## Project Overview

This is a **Go SDK for PocketBase** - a community-maintained client library that provides type-safe, idiomatic Go interfaces for interacting with PocketBase APIs. The project focuses on developer experience, comprehensive testing, and production-ready reliability.

**Key Technologies:**
- Go 1.24+ with generics support
- PocketBase v0.30.x compatibility
- HTTP client with Resty v2
- Server-Sent Events (SSE) for real-time features
- Comprehensive test suite with integration testing

## Architecture & Design Patterns

### Core Components

1. **Client (`client.go`)** - Main HTTP client with authentication
2. **Collections (`collection.go`)** - Type-safe collection operations with generics
3. **Records (`record.go`)** - User authentication and record management
4. **Backups (`backup.go`)** - Backup operations and file management
5. **Subscriptions (`subscribe.go`)** - Real-time SSE connections
6. **Authorization (`authorize.go`, `token_authorize.go`)** - Auth strategies

### Design Patterns Used

- **Builder Pattern**: Client configuration with functional options
- **Generic Types**: Type-safe collection operations (`CollectionSet[T]`)
- **Strategy Pattern**: Multiple authentication methods
- **Observer Pattern**: Real-time subscriptions with event channels
- **Repository Pattern**: Collection-based data access

## Minor Scripting

- For any minor scripting work, example- generating script for github workflow, use python.

## Development Workflow

### Quick Start for Agents
```bash
# 1. Build and start server
make serve-bg

# 2. Run tests
make test-integration

# 3. Stop server
make serve-stop
```

### Available Make Targets

**Server Management:**
- `make serve` - Run server in foreground
- `make serve-bg` - Start server in background
- `make serve-stop` - Stop background server
- `make serve-status` - Check server status
- `make serve-restart` - Restart server

**Testing:**
- `make test-integration` - Full tests with automatic server management (recommended)
- `make test-unit` - Unit tests only (fast)
- `make test` - Manual testing (requires running server)

**Development:**
- `make build` - Build binaries with version info
- `make check` - Linting and security checks
- `make clean` - Clean artifacts and stop servers
- `make format` - Format code with goimports

## Code Structure & Conventions

### File Organization
```
├── client.go          # Main HTTP client
├── collection.go      # Generic collection operations
├── record.go          # User/record authentication
├── backup.go          # Backup management
├── subscribe.go       # Real-time subscriptions
├── authorize.go       # Auth interfaces
├── params.go          # Query parameters
├── response.go        # Response types
├── cmd/pocketbase/    # Server binary
├── example/           # Usage examples
├── migrations/        # Test data setup
└── testressources/    # Test fixtures
```

### Naming Conventions
- **Interfaces**: `authStore`, `Authorizer`
- **Structs**: `Client`, `Collection[T]`, `ParamsList`
- **Methods**: `NewClient()`, `List()`, `Create()`, `AuthWithPassword()`
- **Constants**: `ErrInvalidResponse`

### Go-Specific Patterns
- **Generics**: `Collection[T]` for type-safe operations
- **Functional Options**: `WithAdminEmailPassword()`, `WithDebug()`
- **Context Support**: All HTTP operations support context
- **Error Wrapping**: Consistent error handling with wrapped errors## Testin
g Strategy

### Test Types
1. **Unit Tests** (`*_test.go`) - Fast, isolated tests
2. **Integration Tests** - Full API testing with live server
3. **Real-time Tests** - SSE subscription testing

### Test Data Management
- **Migrations** (`migrations/`) - Consistent test data setup
- **Test Resources** (`testressources/`) - Backup files and fixtures
- **Dynamic Data** - Time-based unique values for isolation

### Testing Best Practices
```go
// Use table-driven tests
tests := []struct {
    name    string
    input   ParamsList
    wantErr bool
}{
    {"valid params", ParamsList{Page: 1}, false},
    {"invalid params", ParamsList{Page: -1}, true},
}

// Clean up test data
defer func() {
    client.Delete("collection", recordID)
}()
```

## Common Development Tasks

### Adding New API Endpoints
1. Add method to `Client` struct
2. Define request/response types
3. Implement HTTP call with proper error handling
4. Add comprehensive tests
5. Update documentation

### Adding Collection Operations
1. Add method to `Collection[T]` struct
2. Use generics for type safety
3. Support query parameters via `ParamsList`
4. Handle pagination and filtering

### Authentication Methods
```go
// Admin authentication
client := pocketbase.NewClient(url, 
    pocketbase.WithAdminEmailPassword(email, password))

// User authentication
client := pocketbase.NewClient(url,
    pocketbase.WithUserEmailPassword(email, password))

// Token-based auth
client := pocketbase.NewClient(url,
    pocketbase.WithUserToken(token))
```## Real
-time Features

### Server-Sent Events (SSE)
```go
// Subscribe to collection changes
stream, err := collection.Subscribe()
if err != nil {
    return err
}
defer stream.Unsubscribe()

// Wait for connection
<-stream.Ready()

// Handle events
for event := range stream.Events() {
    switch event.Action {
    case "create":
        // Handle new record
    case "update":
        // Handle updated record
    case "delete":
        // Handle deleted record
    }
}
```

### Connection Management
- Automatic reconnection on connection loss
- Graceful shutdown with `Unsubscribe()`
- Event buffering during reconnection

## Error Handling Patterns

### Standard Error Types
```go
// Custom errors
var ErrInvalidResponse = errors.New("invalid response")

// HTTP errors with context
if err != nil {
    return fmt.Errorf("failed to create record: %w", err)
}

// Validation errors from PocketBase
if strings.Contains(err.Error(), "validation_") {
    // Handle validation error
}
```

### Retry Logic
- Built-in retry with exponential backoff
- Configurable retry attempts
- Circuit breaker for failing endpoints

## Performance Considerations

### HTTP Client Optimization
- Connection pooling via Resty
- Request/response compression
- Timeout configuration
- Keep-alive connections

### Memory Management
- Streaming for large responses
- Proper cleanup of SSE connections
- Efficient JSON marshaling/unmarshaling

### Caching Strategies
- Client-side token caching
- Response caching for read-heavy operations
- ETags support for conditional requests## Se
curity Best Practices

### Authentication Security
- Secure token storage
- Token refresh handling
- Admin vs user permission separation
- Environment variable configuration

### Input Validation
- Query parameter sanitization
- File upload validation
- SQL injection prevention via parameterized queries

### Network Security
- HTTPS enforcement
- Request signing
- Rate limiting compliance

## Debugging & Troubleshooting

### Debug Mode
```go
client := pocketbase.NewClient(url, pocketbase.WithDebug())
// Enables HTTP request/response logging
```

### Common Issues
1. **Connection Refused**: Server not running - use `make serve-bg`
2. **Authentication Failed**: Check credentials and permissions
3. **SSE Disconnects**: Network issues - automatic reconnection handles this
4. **Test Failures**: Ensure clean test data - use `make clean`

### Logging
- HTTP request/response logging with `WithDebug()`
- SSE connection logging
- Error context preservation

## Contributing Guidelines

### Code Quality
- Run `make check` before commits (linting + security)
- Maintain test coverage above 80%
- Follow Go naming conventions
- Use `gofmt` and `goimports` for formatting

### Pull Request Process
1. Create feature branch
2. Add tests for new functionality
3. Update documentation
4. Run full test suite: `make test-integration`
5. Submit PR with clear description

### Version Compatibility
- Maintain backward compatibility
- Follow semantic versioning
- Test against multiple PocketBase versions
- Update compatibility matrix in README

## AI Agent Specific Notes

### When Working on This Project
1. **Always run tests**: Use `make test-integration` for comprehensive testing
2. **Check server status**: Use `make serve-status` before debugging
3. **Clean state**: Use `make clean` to reset environment
4. **Version awareness**: Check `go.mod` for Go version and dependencies
5. **Type safety**: Leverage generics for collection operations
6. **Error context**: Always wrap errors with meaningful context

### Common Agent Tasks
- Adding new API endpoints following existing patterns
- Improving test coverage and reliability
- Optimizing performance for high-throughput scenarios
- Enhancing real-time features and connection stability
- Updating documentation and examples