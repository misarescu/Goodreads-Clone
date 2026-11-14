# Changelog

## Unreleased

- Added an optional `rating` (0 to 5 in increments of 0.25, `null` when unrated) to books (gc-12): settable via `PUT /api/books/{id}`, returned by get and list. New migration `002_book_rating.sql`.
- Added `GET /api/books/count` (gc-6), returning the catalog size as `{"book_count": <number>}`.
- Add `GET /api/version`, which returns `{"version": "<version>"}`. The version defaults to `dev` and can be set at build time with `-ldflags "-X goodreads-clone.com/internal/server.Version=<version>"`.
- Add `GET /api/health`.
