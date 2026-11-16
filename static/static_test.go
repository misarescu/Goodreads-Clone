package static_test

import (
	"io/fs"
	"testing"

	"goodreads-clone.com/static"
)

// TestFilesEmbedsDist checks that the //go:embed directive is in effect: the
// embedded file system must contain the dist directory. With a malformed
// directive the FS is empty and the frontend is never served.
func TestFilesEmbedsDist(t *testing.T) {
	info, err := fs.Stat(static.Files, "dist")
	if err != nil {
		t.Fatalf("static.Files has no dist directory (is the //go:embed directive missing?): %v", err)
	}
	if !info.IsDir() {
		t.Fatalf("dist is not a directory")
	}
}
