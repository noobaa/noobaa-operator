package bundle

import (
	"strings"
	"testing"

	"github.com/noobaa/noobaa-operator/v5/deploy"
)

func TestAllEmbeddedFilesAreNonEmpty(t *testing.T) {
	walkDir(t, ".")
}

func TestMustReadPanicsOnMissing(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("MustRead did not panic for a missing file")
		}
	}()
	MustRead("nonexistent/file.yaml")
}

func walkDir(t *testing.T, dir string) {
	t.Helper()
	entries, err := deploy.FS.ReadDir(dir)
	if err != nil {
		t.Fatalf("failed to read embedded dir %q: %v", dir, err)
	}
	for _, entry := range entries {
		path := dir + "/" + entry.Name()
		if dir == "." {
			path = entry.Name()
		}
		if entry.IsDir() {
			walkDir(t, path)
			continue
		}
		if strings.HasSuffix(path, ".go") {
			continue
		}
		t.Run(path, func(t *testing.T) {
			data, err := deploy.FS.ReadFile(path)
			if err != nil {
				t.Fatalf("failed to read embedded file %q: %v", path, err)
			}
			if len(data) == 0 {
				t.Fatalf("embedded file %q is empty", path)
			}
		})
	}
}
