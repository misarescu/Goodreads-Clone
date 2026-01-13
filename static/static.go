// Package static provides embedded static files for the frontend application.
package static

import "embed"

// Files contains the embedded frontend static files from the dist directory.
//
// nolint go:embed dist/*
var Files embed.FS
