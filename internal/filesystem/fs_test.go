package filesystem

import (
	"testing"
	"testing/fstest"
)

func TestFileSystem(t *testing.T) {
	codeFS := NewFileSystem("../..")
	if err := fstest.TestFS(codeFS, "internal/filesystem/fs_test.go"); err != nil {
		t.Fatal(err)
	}
}
