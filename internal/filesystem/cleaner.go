// SPDX-License-Identifier: Zlib
// Copyright 2025, Terry M. Poulin.
package filesystem

import (
	"os"
	"path/filepath"
	"strings"
)

// Reserved characters are defined in terms of common platforms. The resulting
// file name after cleaning with these may contain characters that are not
// recommended, but should at least avoid most 'export from os foo to os bar'
// kind of grumbles.
var ReservedCharacters []string

func init() {
	ReservedCharacters = []string{
		// UNIX systems and most non-unix platforms forbid slash.
		"/",
		// Multiple platforms consider ":" reserved.
		//
		// - Windows uses it as a volume separator, and basically every FAT or IBM
		//   PC related file system will take some level of offense.
		//
		// - Macintosh used it as the path separator in the classic system software
		//   and early OS. The modern OS still considers it a reserved character,
		//   but anything considering it a path separator is either dead by now, as
		//   pendantic as I am, or still legacy aware.
		":",
		// Remaining characters that Windows / FAT / PC file systems consider
		// reserved. Since NUL, slash, colon, and ascii control are already added
		// above, we skip those here.
		"<", ">", "\"", "\\", "|", "?", "*",
	}

	// Virtually the entire world agrees that NUL and ASCII control characters
	// are either verboten or just a damn bad idea. That means 0 - 31.
	for codePoint := range 32 {
		ReservedCharacters = append(ReservedCharacters, string(rune(codePoint)))
	}
}

// A string replacer for cleaning paths.
type Cleaner struct {
	*strings.Replacer
}

// Creates a new cleaner that will replace all occurances of strings in
// `reserved` with `replacement` text when encountered during a clean.
func NewCleaner(replacement string, reserved []string) *Cleaner {
	var r []string
	for _, s := range reserved {
		r = append(r, s, replacement)
	}
	return &Cleaner{Replacer: strings.NewReplacer(r...)}
}

// Replaces reserved characters in `name` with the replacement character.
func (c *Cleaner) CleanName(name string) string {
	return c.Replace(name)
}

// Returns `path` with each element cleaned. E.g., "/foo>bar/file" will become
// "/foo_bar/file".
func (c *Cleaner) CleanPath(path string) string {
	var nodes []string
	sep := string(os.PathSeparator)
	for i, s := range strings.Split(filepath.Clean(path), sep) {
		if s == "" && i == 0 {
			nodes = append(nodes, sep)
		}
		nodes = append(nodes, c.CleanName(s))
	}
	return filepath.Join(nodes...)
}
