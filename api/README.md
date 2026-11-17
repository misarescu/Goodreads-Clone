# Goodreads Clone API

REST API for the Goodreads Clone application.

## Base URL
```
http://localhost:8080
```

Requests use `{{baseUrl}}` from the `ci` environment. `mise run e2e` runs the
suite against `E2E_BASE_URL` (default `http://localhost:8080`) and writes JUnit
results to `E2E_RESULTS` (default `api/e2e-results.xml`); the factory runs it
the same way against a stage.

### Repeatable runs

The suite can run any number of times against a server that keeps its data:

- `collection.bru` sets a per-run token (`runId`, from `Date.now()`) once; every
  unique value (ISBN, email, username, shelf name, tag) includes `{{runId}}`.
- Requests never use fixed IDs: each create request saves the ID it returns
  (`userId`, `bookId`, `shelfId`, `shelf2Id`) with `bru.setVar`, and later
  requests in the folder use it.
- Each folder creates what it needs (the Reviews and Shelves folders start with
  Setup User and Setup Book) and ends by deleting it, in a valid order, so a
  run leaves nothing behind. Requests run in `seq` order within a folder.

### Assertions

Every request asserts what its `docs` block promises:

- `assert { res.status: eq N }` checks the status code.
- A `tests` block checks the body: a create echoes what it sent and returns an
  `id`, a get returns the requested resource (compare against the saved
  `userId`, `bookId`, ...), a list returns an array (the search also
  `total`, `limit`, `offset`).
- Each delete is followed by a get of the same resource that must answer 404.
- Negative cases cover the errors the docs name: 400 for an invalid body or ID,
  404 for a deleted resource, 409 for a duplicate (a book added to a shelf
  twice). Invalid IDs are non-numeric (`abc`); "unknown" resources are ones
  the run just deleted, never a fixed number.

### Issue tags

A request that checks behaviour an issue added carries that issue's key in its
name, for example `gc-65 List books by tag`, so a failure points to the issue.
Requests that predate this convention need no tag.

## API Endpoints

### Health
- `GET /api/health` - Liveness: `{"status":"ok"}`, does not touch the database
- `GET /api/ready` - Readiness: 200 `{"status":"ready"}` when the database answers a ping, else 503 `{"status":"unavailable"}`

### Users
- `POST /api/users` - Create a new user
- `GET /api/users/{id}` - Get user by ID
- `GET /api/users?email={email}` - Get user by email
- `GET /api/users?username={username}` - Get user by username
- `GET /api/users?limit={limit}&offset={offset}` - List users (paginated)
- `PUT /api/users/{id}` - Update user
- `DELETE /api/users/{id}` - Delete user

### Books
- `POST /api/books` - Create a new book
- `GET /api/books/count` - Get the number of books in the catalog. Returns `200 OK` with `Content-Type: application/json` and `{"book_count": <number>}`.
- `GET /api/books/{id}` - Get book by ID
- `GET /api/books?isbn={isbn}` - Get book by ISBN
- `GET /api/books?limit={limit}&offset={offset}` - List books (paginated)
- `GET /api/books/search?q={query}&limit={limit}&offset={offset}` - Full-text search (FTS5) over title and author, ranked by bm25 with title weighted higher; the last word is a prefix match. Filters in `q`: `author:<word>`, `lang:<code>` (e.g. `q=hobbit author:tolkien lang:en`). Returns `{"books": [...], "total": N, "limit": N, "offset": N}`; a query with no searchable word or `author:` filter gets 400
- `PUT /api/books/{id}` - Update book
- `DELETE /api/books/{id}` - Delete book

### Shelves
- `POST /api/shelves` - Create a custom shelf
- `GET /api/shelves/{id}` - Get shelf by ID
- `GET /api/users/{userID}/shelves` - List all shelves for a user
- `GET /api/users/{userID}/shelves?name={name}` - Get specific shelf by name
- `PUT /api/shelves/{id}` - Update shelf name (custom shelves only)
- `DELETE /api/shelves/{id}` - Delete shelf (custom shelves only)

## Response Format

All API responses follow a consistent envelope structure:

### Success Response
```json
{
  "data": { ... }
}
```

### Error Response
```json
{
  "error": "error message"
}
```

## Pagination

List endpoints support pagination via query parameters:
- `limit` - Number of items to return (default: 20, max: 100)
- `offset` - Number of items to skip (default: 0)

Example:
```
GET /api/users?limit=50&offset=100
```

## Default Shelves

When a user is created, they can have 4 default shelves created:
1. To Read
2. Currently Reading
3. Finished
4. DNF (Did Not Finish)

Default shelves cannot be updated or deleted.

## Date Format

All dates use RFC3339 format:
```
2006-01-02T15:04:05Z07:00
```

## Status Codes

- `200 OK` - Successful GET/PUT request
- `201 Created` - Successful POST request
- `204 No Content` - Successful DELETE request
- `400 Bad Request` - Invalid request body or parameters
- `404 Not Found` - Resource not found
- `500 Internal Server Error` - Server error

## Testing with Bruno

This directory contains Bruno collection files for testing all API endpoints.

1. Open Bruno
2. Open this collection (api/ directory)
3. Navigate through the folders: Users, Books, Shelves
4. Run individual requests or use the collection runner

## Environment Variables

Configure the server with these environment variables:
- `DATABASE_PATH` - Path to SQLite database (default: `goodreads.db`)
- `SERVER_ADDR` - Server address (default: `:8080`)
