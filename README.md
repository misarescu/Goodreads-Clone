# Goodreads Clone

A book tracking application built with Go and React, inspired by Goodreads. Track your reading progress, organize books into shelves, and search for new books.

## Features

- User management
- Book catalog with search functionality
- Custom bookshelves
- Reading status tracking
- RESTful API backend
- Modern React frontend

## Tech Stack

### Backend
- **Go 1.25** - Backend server
- **SQLite** - Database with [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite) (pure Go)
- **sqlc** - Type-safe SQL query generation
- **net/http** - Standard library HTTP (no external web framework)
- **Testify** - Testing framework
- **Mockery** - Mock generation for tests

### Frontend
- **React 19** - Latest React with concurrent features
- **Vite** (rolldown-vite) - Fast build tool and dev server
- **React Compiler** - Automatic optimization via babel-plugin-react-compiler
- **Bun** - Fast JavaScript runtime and package manager

## Prerequisites

- Go 1.25+
- Bun
- Make

Development tools (installed automatically via `make install-deps`):
- [sqlc](https://sqlc.dev/) - Type-safe SQL code generation
- [mockery](https://github.com/vektra/mockery) - Mock generation

## Quick Start

```bash
# Install dependencies
make install-deps

# Run database migrations
make migrate

# Build and run the application
make run-local
```

The server will start on `http://localhost:8080`

### Frontend Development with Hot Reload

```bash
make frontend-dev
```

Vite dev server runs on `http://localhost:5173` by default.

### Docker

```bash
make run-docker
```

## Project Structure

```
.
├── cmd/
│   └── goodreads/          # Application entry point
├── internal/
│   ├── server/             # HTTP handlers, routes, middleware
│   ├── business/           # Business logic layer
│   ├── storage/            # Data access layer with repository pattern
│   └── db/                 # Auto-generated sqlc code (don't edit manually)
├── db/
│   ├── migrations/         # SQL migration files
│   ├── queries/            # SQL queries for sqlc
│   └── sqlc.yaml           # sqlc configuration
├── frontend/               # React 19 application
│   └── src/
│       └── main.jsx        # Entry point
├── static/                 # Embedded frontend dist files
└── bin/                    # Built binaries (gitignored)
```

## Development

### Available Commands

```bash
# Building
make                       # Build everything (backend + frontend)
make backend-build         # Build Go binary only
make frontend-build        # Production build frontend
make frontend-dev          # Start Vite dev server with hot reload
make frontend-lint         # Run ESLint

# Running
make run-local             # Build and run locally
make run-docker            # Run with Docker Compose

# Database
make migrate               # Run database migrations
make gen-db                # Generate Go code from SQL (after query changes)

# Testing
make test                  # Run all tests
make coverage              # Generate coverage report
make gen-mocks             # Generate test mocks (after interface changes)

# Dependencies
make install-deps          # Install all dependencies
```

### Backend Development Workflow

1. Modify code in `internal/`
2. Run tests: `make test`
3. Build and run: `make run-local`

### Database Changes

1. Add migration file to `db/migrations/` (e.g., `002_add_feature.sql`)
2. Add/update queries in `db/queries/`
3. Generate Go code: `make gen-db`
4. Apply migrations: `make migrate`

### Interface Changes

After modifying interfaces in `internal/storage/interfaces.go`:
```bash
make gen-mocks
```

## API Endpoints

### Users
- `POST /api/users` - Create user
- `GET /api/users/{id}` - Get user by ID
- `GET /api/users` - List users
- `PUT /api/users/{id}` - Update user
- `DELETE /api/users/{id}` - Delete user

### Books
- `POST /api/books` - Create book
- `GET /api/books/{id}` - Get book by ID
- `GET /api/books` - List books
- `GET /api/books/search` - Search books
- `PUT /api/books/{id}` - Update book
- `DELETE /api/books/{id}` - Delete book

### Shelves
- `POST /api/shelves` - Create shelf
- `GET /api/shelves/{id}` - Get shelf by ID
- `GET /api/users/{userID}/shelves` - Get user's shelves
- `PUT /api/shelves/{id}` - Update shelf
- `DELETE /api/shelves/{id}` - Delete shelf

## Environment Variables

- `DATABASE_PATH` - Path to SQLite database file (default: `goodreads.db`)
- `SERVER_ADDR` - Server address (default: `:8080`)

## Architecture

The application follows a clean layered architecture:

**Server → Business → Storage → DB**

- **Server Layer** (`internal/server`): HTTP handlers, routing, middleware
- **Business Layer** (`internal/business`): Business logic and orchestration
- **Storage Layer** (`internal/storage`): Repository interfaces for data access
- **DB Layer** (`internal/db`): Auto-generated type-safe SQL code via sqlc

The frontend builds to `static/dist/` and is embedded in the Go binary using `//go:embed`, allowing the backend to serve the complete application as a single executable.
