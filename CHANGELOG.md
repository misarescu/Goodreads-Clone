# Changelog

## Unreleased

- Added `GET /api/books/{id}/similar` (gc-77): the books sharing the most tags with a book (`{books: [{book, sharedTags}], total, limit, offset}`), most shared tags first then by `id`; `sharedTags` is alphabetical. `limit` (max 100) and `offset` as in other lists; 404 for an unknown book. One grouped query plus one count.

- Added `GET /api/tags` (gc-75): the tags in use with the number of books carrying each (`{tags: [{tag, books}], total, limit, offset}`), most books first then alphabetical. `limit` (max 100) and `offset` as in other lists; `prefix` filters (trimmed, lowercased) and answers 400 when no valid tag can start with it. Counted in one grouped query plus one count.

- Added `GET /api/books/{id}/ratings` (gc-74): a book's review count, average rating and the number of reviews at each of 1 to 5 stars (`stars` always has all five keys). Counted in one grouped query.

- The server now applies `db/migrations/*.sql` itself at startup (gc-45), embedded in the binary and recorded in a `schema_migrations` table. A database made by `make migrate` has 001 to 006 recorded as applied. The docker-compose `migrate` service is removed; `make migrate` stays for local use.

- Full-text book search with SQLite FTS5 (gc-29): `GET /api/books/search?q=` now matches words in title or author (prefix match on the last word, ranked by bm25 with title above author) and supports `author:<word>` and `lang:<code>` filters. It takes `limit` (max 100) and `offset`, and **the response is now `{books, total, limit, offset}` instead of a bare array**. A query with no searchable word or author filter gets 400. New migration `006_books_fts.sql`.

- Added yearly reading challenges (gc-19): `PUT /api/users/{id}/challenges/{year}` with `{"goal": 24}` (1 to 1000) creates or updates, `GET` returns `{year, goal, finished, expected_by_now, ahead_by}`; 404 for an unknown user or challenge. New migration `004_reading_challenges.sql`.

- Added reading progress (gc-18): `PUT /api/shelves/{id}/books/{bookID}/progress` takes `pages_read` or `progress_percent`. Shelf books (that response and `GET /api/shelves/{id}/books`) now include `finished_at` (RFC3339 string, `null` until finished), set when progress first reaches 100 and cleared if it drops below. New migration `003_shelf_book_finished_at.sql`.
- Added an optional `rating` (0 to 5 in increments of 0.25, `null` when unrated) to books (gc-12): settable via `PUT /api/books/{id}`, returned by get and list. New migration `002_book_rating.sql`.
- Added `GET /api/books/count` (gc-6), returning the catalog size as `{"book_count": <number>}`.
- Add `GET /api/version`, which returns `{"version": "<version>"}`. The version defaults to `dev` and can be set at build time with `-ldflags "-X goodreads-clone.com/internal/server.Version=<version>"`.
- Add `GET /api/health`.
