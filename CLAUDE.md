# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with the Go project at `/home/vinifranco/coding/go/lista_v2` (version `cc_version=2.1.105.834`, entrypoint `cli`).

---

## **Common Development Commands**
1. **Build**
   ```bash
   go build -v ./...
   ```
   - Builds the entire project. Use `go build -p /home/vinifranco/coding/go/lista_v2` for standalone binaries.

2. **Test**
   ```bash
   go test ./... -v -count=1
   ```
   - Runs all tests. To focus on a single package, replace `./...` with the package path, e.g. `go test ./internal/handlers -v -count=1`.

3. **Lint**
   ```bash
   golangci-lint run
   ```
   - Linting is configured via the project's `golangci.yml` (if present).

4. **Run in Development Mode**
   ```bash
   go run .
   ```
   - Starts the server using default environment configuration.

---

## **Code Architecture**
1. **Core Components**
   - **Handlers**: `internal/handlers/*` – HTTP/REST endpoint implementations (e.g., `api_lista.go`).
   - **Models**: `internal/models/*` – Data structures, often mirroring database tables.
   - **Services**: `internal/services/*` – Business‑logic layer invoked by handlers.

2. **Database**
   - Schema migrations live in `migrations/` and are compiled with **sqlc** (see `sqlc.yaml`).
   - Connection pooling and DB helper functions are defined in `internal/database.go`.

3. **Configuration**
   - Central configuration logic resides in `internal/config.go`, pulling values from environment variables (e.g., `DATABASE_URL`).

---

## **Dependencies**
- **Go Modules**: Declared in `go.mod`/`go.sum`.
- **sqlc**: Generates type‑safe Go code from SQL migrations.
- **golangci-lint**: Linting suite used via `make lint` or the direct command above.

---

## **Common Tasks**
1. **Add a New API Endpoint**
   - Create a handler file in `internal/handlers/`.
   - Define any new request/response structs in `internal/models/`.
   - Add business logic in `internal/services/` if needed.
   - Write tests alongside the handler (`*_test.go`).

2. **Database Migration**
   - Add a new `.sql` file under `migrations/`.
   - Run `go run internal/migrate.go` (or the project's migration script) to apply.
   - Regenerate sqlc bindings if necessary: `sqlc generate`.

3. **Debugging**
   - Use Delve: `dlv debug .` to step through the server.
   - Logs are printed to stdout; configure log level via the `LOG_LEVEL` env var.

---

## **File Structure Overview**
```
lista_v2/
├── cmd/                # CLI entry points
├── config/             # Static configuration files
├── db/                 # Database utilities and migrations
├── internal/
│   ├── handlers/      # HTTP handlers / controllers
│   ├── models/        # Data structures & DB models
│   ├── services/      # Business logic layer
│   └── ...
├── static/             # Web assets (CSS, JS)
├── views/              # HTML templates
├── go.mod, go.sum      # Module definition
└── main.go             # Application entry point
```

---

## **Project Context**
- The repository includes a `README.md` that outlines the purpose: a task‑management API with REST and optional GraphQL endpoints.
- No `.cursor` or Copilot rule files were found, so no additional instruction sets apply.

> **Tip:** When making changes, run the lint and test commands above to keep the codebase consistent and pass CI checks.
