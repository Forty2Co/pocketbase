[![Check & test & build](https://github.com/Forty2Co/pocketbase/actions/workflows/main.yml/badge.svg)](https://github.com/Forty2Co/pocketbase/actions/workflows/main.yml)
[![PocketBase](https://pocketbase.io/images/logo.svg)](https://pocketbase.io)

### Project

I forked this to run some of my personal projects. I'm open for any contribution.

#### Roadmap

> Note
> This will be updated as we go.****

- [] Add observability support.
- [] Improve pocketbase version compatibility

### Compatibility

- `v0.22.0` version of SDK is compatible with Pocketbase v0.22.x
- `v0.21.0` version of SDK is compatible with Pocketbase v0.21.x
- `v0.20.0` version of SDK is compatible with Pocketbase v0.20.x
- `v0.19.0` version of SDK is compatible with Pocketbase v0.19.x
- `v0.13.0` version of SDK is compatible with Pocketbase v0.13.x and higher
- `v0.12.0` version of SDK is compatible with Pocketbase v0.12.x
- `v0.11.0` version of SDK is compatible with Pocketbase v0.11.x
- `v0.10.1` version of SDK is compatible with Pocketbase v0.10.x
- `v0.9.2` version of SDK is compatible with Pocketbase v0.9.x (SSE & generics support introduced)
- `v0.8.0` version of SDK is compatible with Pocketbase v0.8.x

### PocketBase

[Pocketbase](https://pocketbase.io) is a simple, self-hosted, open-source, no-code, database for your personal data.
It's a great alternative to Airtable, Notion, and Google Sheets. Source code is available on [GitHub](https://github.com/pocketbase/pocketbase)

### Currently supported operations

This SDK doesn't have feature parity with official SDKs and supports the following operations:

- **Authentication** - anonymous, admin and user via email/password
- **Create**
- **Update**
- **Delete**
- **List** - with pagination, filtering, sorting
- **Backups** - with create, restore, delete, upload, download and list all available downloads
- **Real-time subscriptions** - Server-Sent Events (SSE) for live updates
- **Type-safe collections** - Generic collection wrappers for better type safety
- **Other** - feel free to create an issue or contribute

### Package Organization

The SDK has been restructured into focused sub-packages for better maintainability, clearer APIs, and improved code organization while maintaining full backward compatibility.

#### Package Structure

- **Root package** (`pocketbase`) - Main client, common types (`ResponseList`, `ResponseCreate`, `ParamsList`), and unified API
- **Admin package** (`admin`) - Administrative operations (backups, files)
- **Collections package** (`collections`) - Collection-specific operations and type-safe wrappers  
- **Realtime package** (`realtime`) - Server-Sent Events and live subscriptions
- **Auth package** (`auth`) - Authentication strategies and token management

#### Key Benefits

- **Backward Compatibility** - All existing code continues to work without changes
- **Modular Design** - Import only the packages you need for focused functionality
- **Shared Resources** - All sub-packages share the same HTTP client, authentication, and configuration
- **Better Organization** - Clear separation of concerns with logical grouping of functionality
- **Type Safety** - Enhanced type safety with dedicated collection wrappers
- **Consistent APIs** - Unified patterns across all packages
- **Maintainability** - Easier to locate, modify, and test specific functionality
- **Performance** - No overhead from restructuring - same efficient resource usage

#### Migration Guide

**No migration required!** Your existing code will continue to work exactly as before:

```go
// This continues to work unchanged
client := pocketbase.NewClient("http://localhost:8090")
records, err := client.List("posts", pocketbase.ParamsList{})
```

**Optional: Adopt modular approach** for new code or when refactoring:

```go
// New modular approach (optional)
client := pocketbase.NewClient("http://localhost:8090")
collection := collections.CollectionSet[Post](client.Collections, "posts")
records, err := collection.List(pocketbase.ParamsList{})
```

### Usage & examples

#### Client Instantiation

The client instantiation remains the same as before - all existing code continues to work:

```go
client := pocketbase.NewClient("http://localhost:8090")

// With authentication options:
client := pocketbase.NewClient("http://localhost:8090", 
    pocketbase.WithAdminEmailPassword("admin@admin.com", "password"))

// With additional options:
client := pocketbase.NewClient("http://localhost:8090",
    pocketbase.WithTimeout(30*time.Second),
    pocketbase.WithRestDebug(),
    pocketbase.WithSseDebug())
```

#### Usage Approaches

The SDK supports two usage patterns - choose the one that best fits your needs:

**1. Unified Client Approach (Recommended for most users)**

The traditional approach where all functionality is accessed through the main client:

```go
client := pocketbase.NewClient("http://localhost:8090", 
    pocketbase.WithAdminEmailPassword("admin@admin.com", "password"))

// All existing methods work exactly as before
err := client.Backup().Create("backup.zip")
files := client.Files()
collection := pocketbase.CollectionSet[MyStruct](client, "my_collection")
stream, err := collection.Subscribe()
```

**2. Modular Sub-Package Approach (For focused functionality)**

Import and use specific sub-packages for more targeted functionality:

```go
import (
    "github.com/Forty2Co/pocketbase"
    "github.com/Forty2Co/pocketbase/admin"
    "github.com/Forty2Co/pocketbase/collections"
)

client := pocketbase.NewClient("http://localhost:8090")

// Use collections sub-package directly
collection := collections.CollectionSet[MyStruct](client.Collections, "my_collection")
records, err := collection.List(pocketbase.ParamsList{})

// Use admin sub-package directly  
backup := admin.Backup{Client: client.Admin}
err := backup.Create("backup.zip")
```

Both approaches share the same underlying HTTP client, authentication, and configuration for efficient resource usage.

**When to use each approach:**

- **Unified Client**: Best for most applications, simpler imports, backward compatibility
- **Modular Sub-Packages**: Best for large applications, when you need only specific functionality, or when building libraries that depend on PocketBase

#### Basic Operations

Simple list example without authentication (assuming your collections are public):

```go
package main

import (
 "log"

 "github.com/Forty2Co/pocketbase"
)

func main() {
 client := pocketbase.NewClient("http://localhost:8090")

 // You can list with pagination:
 response, err := client.List("posts_public", pocketbase.ParamsList{
  Page: 1, Size: 10, Sort: "-created", Filters: "field~'test'",
 })
 if err != nil {
  log.Fatal(err)
 }
 log.Print(response.TotalItems)

 // Or you can use the FullList method (v0.0.7)
 response, err := client.FullList("posts_public", pocketbase.ParamsList{
  Sort: "-created", Filters: "field~'test'",
 })
 if err != nil {
  log.Fatal(err)
 }

 log.Print(response.TotalItems)
}
```

Creating an item with admin user (auth via email/pass).
Please note that you can pass `map[string]any` or `struct with JSON tags` as a payload:

```go
package main

import (
 "log"

 "github.com/Forty2Co/pocketbase"
)

func main() {
 client := pocketbase.NewClient("http://localhost:8090", 
  pocketbase.WithAdminEmailPassword("admin@admin.com", "admin@admin.com"))
 response, err := client.Create("posts_admin", map[string]any{
  "field": "test",
 })
 if err != nil {
  log.Fatal(err)
 }
 log.Print(response.ID)
}
```

For even easier interaction with collection results as user-defined types, you can go with `CollectionSet`:

```go
package main

import (
 "log"

 "github.com/Forty2Co/pocketbase"
)

type post struct {
 ID      string
 Field   string
 Created string
}

func main() {
 client := pocketbase.NewClient("http://localhost:8090")
 collection := pocketbase.CollectionSet[post](client, "posts_public")

 // List with pagination
 response, err := collection.List(pocketbase.ParamsList{
  Page: 1, Size: 10, Sort: "-created", Filters: "field~'test'",
 })
 if err != nil {
  log.Fatal(err)
 }

 // FullList also available for collections:
 response, err := collection.FullList(pocketbase.ParamsList{
  Sort: "-created", Filters: "field~'test'",
 })
 if err != nil {
  log.Fatal(err)
 }
 
    log.Printf("%+v", response.Items)
}
```

Realtime API via Server-Sent Events (SSE) is also supported:

```go
package main

import (
 "log"

 "github.com/Forty2Co/pocketbase"
)

type post struct {
 ID      string
 Field   string
 Created string
}

func main() {
 client := pocketbase.NewClient("http://localhost:8090")
 collection := pocketbase.CollectionSet[post](client, "posts_public")
 response, err := collection.List(pocketbase.ParamsList{
  Page: 1, Size: 10, Sort: "-created", Filters: "field~'test'",
 })
 if err != nil {
  log.Fatal(err)
 }
 
 stream, err := collection.Subscribe()
 if err != nil {
  log.Fatal(err)
 }
 defer stream.Unsubscribe()
 <-stream.Ready()
 for ev := range stream.Events() {
  log.Print(ev.Action, ev.Record)
 }
}
```

You can fetch a single record by its ID using the `One` method to get the raw map, or the `OneTo` method to unmarshal directly into a custom struct.

Here's an example of fetching a single record as a map:

```go
package main

import (
 "log"

 "github.com/Forty2Co/pocketbase"
)

func main() {
 client := pocketbase.NewClient("http://localhost:8090")

 // Fetch a single record by ID
 record, err := client.One("posts_public", "record_id")
 if err != nil {
  log.Fatal(err)
 }

 // Access the record fields
 log.Print(record["field"])
}
```

You can fetch and unmarshal a single record directly into your custom struct using `OneTo`:

```go
package main

import (
 "log"

 "github.com/Forty2Co/pocketbase"
)

type Post struct {
 ID    string `json:"id"`
 Field string `json:"field"`
}

func main() {
 client := pocketbase.NewClient("http://localhost:8090")

 // Fetch a single record by ID and unmarshal into struct
 var post Post
 err := client.OneTo("posts", "post_id", &post)
 if err != nil {
  log.Fatal(err)
 }

 // Access the struct fields
 log.Printf("Fetched Post: %+v\n", post)
}
```

Trigger to create a new backup.

```go
package main

import (
 "log"

 "github.com/Forty2Co/pocketbase"
)

func main() {
 client := pocketbase.NewClient("http://localhost:8090", 
  pocketbase.WithAdminEmailPassword("admin@admin.com", "admin@admin.com"))
 err := client.Backup().Create("foobar.zip")
 if err != nil {
     log.Println("create new backup failed")
  log.Fatal(err)
 }
}
```

Authenticate user from collection

```go
package main

import (
 "log"

 "github.com/Forty2Co/pocketbase"
)

type User struct {
 AuthProviders    []interface{} `json:"authProviders"`
 UsernamePassword bool          `json:"usernamePassword"`
 EmailPassword    bool          `json:"emailPassword"`
 OnlyVerified     bool          `json:"onlyVerified"`
}

func main() {
 client := pocketbase.NewClient("http://localhost:8090")
 response, err := pocketbase.CollectionSet[User](client, "users").AuthWithPassword("user", "user@user.com")
 if err != nil {
  log.Println("user-authentication failed")
  log.Fatal(err)
  return
 }
 log.Println("authentication successful")
 log.Printf("JWT-token: %s\n", response.Token)
}
```

### Package-Specific Usage

#### Collections Package

For type-safe collection operations with enhanced functionality:

```go
import "github.com/Forty2Co/pocketbase/collections"

// Create type-safe collection wrapper
users := collections.CollectionSet[User](client.Collections, "users")

// All collection operations with type safety
response, err := users.List(pocketbase.ParamsList{Size: 10})
user, err := users.Create(User{Name: "John", Email: "john@example.com"})
err = users.Update(user.ID, User{Name: "John Doe"})
err = users.Delete(user.ID)

// Authentication methods for user collections
authResponse, err := users.AuthWithPassword("john@example.com", "password")
methods, err := users.AuthMethods()
```

#### Admin Package

For administrative operations (requires admin authentication):

```go
import "github.com/Forty2Co/pocketbase/admin"

// Backup operations
backup := admin.Backup{Client: client.Admin}
err := backup.Create("my_backup.zip")
backups, err := backup.FullList()
err = backup.Delete("old_backup.zip")

// File operations
files := admin.Files{Client: client.Admin}
token, err := files.GetToken()
```

#### Realtime Package

For Server-Sent Events and live subscriptions:

```go
import "github.com/Forty2Co/pocketbase/realtime"

// Subscribe to collection changes
collection := collections.CollectionSet[Post](client.Collections, "posts")
stream, err := collection.Subscribe()
if err != nil {
    log.Fatal(err)
}
defer stream.Unsubscribe()

// Wait for connection and handle events
<-stream.Ready()
for event := range stream.Events() {
    log.Printf("Event: %s, Record: %+v", event.Action, event.Record)
}
```

#### Auth Package

For custom authentication strategies:

```go
import "github.com/Forty2Co/pocketbase/auth"

// Create custom authenticators
emailAuth := auth.NewEmailPasswordAuth(client.GetClient(), "http://localhost:8090/api/collections/users/auth-with-password", "user@example.com", "password")
tokenAuth := auth.NewTokenAuth(client.GetClient(), "http://localhost:8090/api/collections/users/auth-refresh", "your-token")

// Use with client
client := pocketbase.NewClient("http://localhost:8090")
// Set custom auth store if needed
```

### Examples and Testing

More examples can be found in:

- [Unified usage example](./example/main.go) - Traditional approach with both unified and modular examples
- [Modular usage example](./example/modular_usage.go) - Focused sub-package usage examples
- [Client tests](./client_test.go) - Comprehensive client functionality tests
- [Collection tests](./collections/collection_test.go) - Type-safe collection operation tests
- Remember to start PocketBase before running examples with `make serve` command
- For integration tests, use `make test-integration` which automatically manages the server

## Development

### Makefile targets

**Server Management:**

- `make serve` - builds all binaries and runs local PocketBase server in foreground
- `make serve-bg` - starts PocketBase server in background (saves PID for management)
- `make serve-stop` - stops the background PocketBase server
- `make serve-status` - checks if the server is running
- `make serve-restart` - restarts the background server

**Testing:**

- `make test-integration` - runs all tests with automatic server management (recommended)
- `make test-unit` - runs only unit tests (fast, no server required)
- `make test` - runs tests (requires PocketBase server running manually)

**Development:**

- `make build` - builds all binaries (examples and PocketBase server)
- `make check` - runs linters and security checks (run this before commit)
- `make clean` - removes build artifacts and stops any running servers
- `make help` - shows help and other targets

## Contributing

> **⚠️ IMPORTANT: VERSION File Requirement**
> 
> **All pull requests MUST update the VERSION file or they will be automatically rejected.**
> 
> This project enforces semantic versioning - every change must be properly versioned:
> - **Bug fixes**: Increment patch version (`0.2.1` → `0.2.2`)
> - **New features**: Increment minor version (`0.2.1` → `0.3.0`)
> - **Breaking changes**: Increment major version (`0.2.1` → `1.0.0`)
> 
> PRs without VERSION updates will fail CI checks and cannot be merged.

### Development Requirements

- Go 1.24+ (for making changes in the Go code)
- While developing use `WithDebug()` client option to see HTTP requests and responses
- **Update VERSION file** in every PR (see warning above)
- Make sure that all checks are green (run `make check` before commit)
- Make sure that all tests pass (run `make test-integration` before commit)
- Create a PR with your changes and wait for review

### Running Tests

**Recommended approach:**

```bash
make test-integration  # Automatically starts server, runs tests, stops server
```

**Manual approach:**

```bash
# Terminal 1: Start server
make serve

# Terminal 2: Run tests
make test

# Terminal 1: Stop server (Ctrl+C)
```

**Unit tests only:**

```bash
make test-unit  # Fast tests that don't require a server
```

## Release Process

This project uses semantic versioning and automated releases via GitHub Actions.

### Automated Release Workflow

Releases are automatically triggered when pull requests are merged to the main branch with VERSION file changes:

1. **Update VERSION file** in your pull request (e.g., `0.2.1`)
2. **Merge PR to main** - This automatically:
   - Detects VERSION file changes
   - Creates git tag (e.g., `v0.2.1`)
   - Triggers release build workflow
   - Creates GitHub release with auto-generated notes
   - Builds and uploads release artifacts

### Version Management

- **VERSION file** - Contains the current version (e.g., `0.2.0`)
- **Git tags** - Automatically created (e.g., `v0.2.0`)
- **Release notes** - Auto-generated from PR titles and commit messages
- **Automatic builds** - Version info is injected into binaries

### Creating a Release

> **📋 Note:** VERSION file updates are **mandatory** for all PRs. The CI system will automatically reject any pull request that doesn't include a VERSION change.

1. **Update VERSION file in a pull request:**
   ```bash
   # Create feature branch
   git checkout -b release/0.2.1
   
   # Update VERSION file
   echo "0.2.1" > VERSION
   
   # Commit and push
   git add VERSION
   git commit -m "chore: bump version to 0.2.1"
   git push origin release/0.2.1
   ```

2. **Create and merge pull request:**
   - Create PR from your branch to main
   - Include release notes in PR description
   - Merge PR to main

3. **Automatic release process:**
   - GitHub Actions detects VERSION file change
   - Creates git tag `v0.2.1`
   - Builds binaries and creates GitHub release
   - Release notes are auto-generated from PR and commit history

### Version Strategy

- **Patch** (0.2.1) - Bug fixes, documentation updates
- **Minor** (0.3.0) - New features, API additions
- **Major** (1.0.0) - Breaking changes

### Manual Release (if needed)

For manual releases, you can still use the traditional approach:

```bash
# Update VERSION file
echo "0.2.1" > VERSION

# Commit changes
git add VERSION
git commit -m "chore: bump version to 0.2.1"
git push origin main

# The automated workflow will handle the rest
```

### Checking Version

```bash
make version  # Shows current version, commit, build time
```
