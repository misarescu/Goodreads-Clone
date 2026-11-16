// Package static provides embedded static files for the frontend application.
package static

import "embed"

// Files contains the embedded frontend static files from the dist directory.
//
// The tracked static/dist/.gitkeep placeholder keeps the pattern matching on a
// fresh clone where the frontend has not been built.
//
//go:embed all:dist
var Files embed.FS
