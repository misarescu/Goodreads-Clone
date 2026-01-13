# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Goodreads Clone - A book tracking application with a Go backend and React frontend.

## Commands

```bash
make                   # Build both backend and frontend
make build             # Same as above

make backend-build     # Build Go binaries to bin/
make frontend-build    # Install deps + production build
make frontend-dev      # Install deps + start Vite dev server
make frontend-lint     # Run ESLint

make run-local         # Build everything and run the server
make run-docker        # Build and run with Docker Compose

make migrate           # Run database migrations
make gen-db            # Generate Go code from SQL using sqlc
make gen-mocks         # Generate test mocks using mockery

make test              # Run all tests
make coverage          # Generate test coverage report
```

## Architecture

**Backend:** Go application using standard project layout (`cmd/` for entry points, `internal/` for private packages).

- **Layers:** Server → Business → Storage → DB
  - `internal/server`: HTTP handlers, routes, and middleware
  - `internal/business`: Business logic layer (currently minimal)
  - `internal/storage`: Data access layer with repository interfaces
  - `internal/db`: Auto-generated sqlc code from SQL queries

**Database:** SQLite with sqlc for type-safe SQL queries. SQL queries live in `db/queries/`, migrations in `db/migrations/`. Generated Go code outputs to `internal/db/`. Foreign keys are enabled via PRAGMA.

**Testing:** Mockery generates mocks for storage repository interfaces (`.mockery.yaml`). Tests use these mocks to isolate business logic. Coverage excludes mocks, db, cmd, and static packages.

**Frontend:** React 19 with Vite (using rolldown-vite), React Compiler enabled via babel-plugin-react-compiler. Entry point at `frontend/src/main.jsx`.

**Deployment:** Multi-stage Dockerfile builds Go backend and Bun-based frontend, outputs static files served by the backend. Frontend builds to `static/dist/`, which is embedded in the Go binary via `//go:embed`.

## Development Workflow

1. **Database changes:** Add migrations to `db/migrations/`, update queries in `db/queries/`, run `make gen-db`
2. **Interface changes:** Update interfaces in `internal/storage/interfaces.go`, run `make gen-mocks`
3. **Backend changes:** Code changes automatically included on next build
4. **Frontend changes:** Use `make frontend-dev` for hot reload during development
