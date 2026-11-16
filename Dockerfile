FROM oven/bun:1.3-alpine AS frontend-builder

WORKDIR /app/frontend
COPY frontend/ ./
RUN bun install && \
    bun run build

FROM golang:1.25-alpine AS backend-builder

WORKDIR /app
COPY . .
RUN go mod download && \
	go install github.com/vektra/mockery/v2@latest && \
    go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

COPY --from=frontend-builder /app/static/dist ./static/dist

# Generate database code
RUN sqlc -f db/sqlc.yaml generate

ARG VERSION=dev
RUN CGO_ENABLED=0 go build -ldflags "-X goodreads-clone.com/internal/server.Version=${VERSION}" -o bin/ ./cmd/...

FROM alpine:latest AS application

RUN adduser -D appuser

WORKDIR /app

COPY --from=backend-builder /app/bin/ ./

RUN chown -R appuser:appuser /app

USER appuser

EXPOSE 8080

CMD ["./goodreads"]
