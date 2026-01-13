# Goodreads Clone API

REST API for the Goodreads Clone application.

## Base URL
```
http://localhost:8080
```

## API Endpoints

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
- `GET /api/books/{id}` - Get book by ID
- `GET /api/books?isbn={isbn}` - Get book by ISBN
- `GET /api/books?limit={limit}&offset={offset}` - List books (paginated)
- `GET /api/books/search?q={query}&limit={limit}` - Search books by title
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
