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
```

## Architecture

**Backend:** Go application using standard project layout (`cmd/` for entry points, `internal/` for private packages).

**Frontend:** React 19 with Vite (using rolldown-vite), React Compiler enabled via babel-plugin-react-compiler. Entry point at `frontend/src/main.jsx`.

**Deployment:** Multi-stage Dockerfile builds Go backend and Bun-based frontend, outputs static files served by the backend.
