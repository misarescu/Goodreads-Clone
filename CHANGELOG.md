# Changelog

## Unreleased

- Add `GET /api/version`, which returns `{"version": "<version>"}`. The version defaults to `dev` and can be set at build time with `-ldflags "-X goodreads-clone.com/internal/server.Version=<version>"`.
- Add `GET /api/health`.
