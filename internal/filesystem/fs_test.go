// SPDX-License-Identifier: Zlib
// Copyright 2025, Terry M. Poulin.
package filesystem

import (
	"encoding/hex"
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

func TestCleaner(t *testing.T) {
	replacement := "_"
	cleaner := NewCleaner(replacement, ReservedCharacters)
	assert := func(input, expected string) {
		if actual := cleaner.CleanPath(input); actual != expected {
			t.Errorf("bad result\ninput   : %q\nactual  : %q\nexpected: %q\n", input, actual, expected)
			t.Logf("hex dump actual:\n%s", hex.Dump([]byte(actual)))
		}
	}
	dump := func(s string) {
		t.Log("hexdumping string:\n", hex.Dump([]byte(s)))
	}
	t.Run("CleanPath", func(t *testing.T) {
		assert("/ham<>:/spam\"\\|?/eggs/file*name.ext", "/ham___/spam____/eggs/file_name.ext")
		assert("/foo//bar", "/foo/bar")
	})
	t.Run("CleanName", func(t *testing.T) {
		expected := "foo_________bar"
		input := "foo<>:\"/\\|?*bar"
		if actual := cleaner.CleanName(input); actual != expected {
			t.Errorf("input: %q actual: %q expected: %q", input, actual, expected)
		}
		for cp := range 32 {
			if s := cleaner.CleanName(string(rune(cp))); s != replacement {
				t.Errorf("ASCII control character %d:o%o:x%x not scrubbed", cp, cp, cp)
				dump(s)
			}
		}
	})
	t.Run("Empty cleaner", func(t *testing.T) {
		c := NewCleaner("_", []string{})
		input := "/foo<>bar/ham \\ spam/quux.ext"
		if s := c.CleanPath(input); s != input {
			t.Errorf("empy cleaner did not return identity for CleanPath(%q)", input)
			dump(s)
		}
		input = "foo?bar.ext"
		if s := c.CleanName(input); s != input {
			t.Errorf("empy cleaner did not return identity for CleanName(%q)", input)
			dump(s)
		}
	})
}
