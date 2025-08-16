// SPDX-License-Identifier: Zlib
// Copyright 2025, Terry M. Poulin.
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

func TestIsTrashfile(t *testing.T) {
	if !IsTrashFile(".DS_Store") || !IsTrashFile("/foo/bar/.DS_Store") {
		t.Errorf("Failed to catch finder info file.")
	}
	if !IsTrashFile("._DS_Store") || !IsTrashFile("/foo/bar/.DS_Store") || !IsTrashFile("/foo/bar/._file.ext") {
		t.Errorf("Failed to catch Apple Double file.")
	}
	if IsTrashFile(".hidden") {
		t.Errorf("Failed caught Unix hidden file, which isn't trash")
	}
}
