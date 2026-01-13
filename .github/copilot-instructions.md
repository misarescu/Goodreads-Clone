# Goodreads Clone - AI Coding Guidelines

## Architecture Overview
- **Backend**: Go application with layered architecture - `business` (logic) → `storage` (repositories) → `db` (sqlc-generated queries)
- **Frontend**: React 19 with Vite (rolldown-vite), React Compiler enabled via babel-plugin-react-compiler
- **Database**: SQLite with migrations in `db/migrations/`, queries in `db/queries/`, code generated to `internal/db/`
- **Deployment**: Multi-stage Docker build combines Go backend serving embedded React static files

## Key Patterns
- **Repository Pattern**: Interfaces in `internal/storage/interfaces.go`, implementations in `internal/storage/`
- **Thin Business Layer**: `internal/business/` delegates to storage, minimal logic
- **SQL Code Generation**: Use sqlc for type-safe database operations (run `make gen-db` after query changes)
- **Testing**: testify framework with mockery-generated mocks (run `make gen-mocks` before testing)

## Development Workflows
- **Build Everything**: `make` or `make build` (builds frontend, generates DB code, compiles Go)
- **Run Locally**: `make run-local` (builds and starts server on :8080)
- **Run with Docker**: `make run-docker` (uses docker-compose with SQLite volume)
- **Frontend Dev**: `make frontend-dev` (starts Vite dev server with HMR)
- **Generate Mocks**: `make gen-mocks` (required before running tests)
- **Coverage**: `make coverage` (generates HTML report excluding generated code)

## Database Schema Conventions
- Users have custom shelves (including defaults like "Want to Read", "Currently Reading")
- Books are global catalog, linked to user shelves via `shelf_books` junction table
- Progress tracking per book-shelf combination (pages read, percentage)
- Foreign key constraints enabled: `PRAGMA foreign_keys = ON`

## Code Examples
- **Add Business Method**: Create in `internal/business/`, delegate to storage repository
- **Database Query**: Write SQL in `db/queries/`, run `sqlc generate`, use generated types
- **Repository Implementation**: Follow pattern in `internal/storage/` - wrap sqlc queries, return pointers to structs
- **Test Structure**: Use testify table-driven tests with mocked repositories

## File Organization
- `cmd/goodreads/main.go`: Simple static file server serving embedded frontend
- `internal/business/`: Business logic layer
- `internal/storage/`: Repository implementations and interfaces
- `internal/db/`: sqlc-generated code (do not edit manually)
- `db/migrations/`: Schema files
- `db/queries/`: SQL query files
- `frontend/src/`: React application (currently template, customize as needed)</content>
<parameter name="filePath">/Users/mi/Projects/Playgrounds/VibeCoded/GoodreadsClone/.github/copilot-instructions.md