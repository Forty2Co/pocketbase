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
- **Other** - feel free to create an issue or contribute

### Usage & examples

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

More examples can be found in:

- [example file](./example/main.go)
- [tests for the client](./client_test.go)
- [tests for the collection](./collection_test.go)
- remember to start the Pocketbase before running examples with `make serve` command
- for integration tests, you can use `make test-integration` which automatically manages the server

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
