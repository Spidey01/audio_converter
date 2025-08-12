// SPDX-License-Identifier: Zlib
// Copyright 2025, Terry M. Poulin.
package filesystem

import (
	"path/filepath"
	"strings"
)

// Returns true if we think the file is trash for the purposes of this
// application. This mainly exists to exclude dot files that shouldn't be copied
// nor mistaken for content.
func IsTrashFile(name string) bool {
	base := filepath.Base(name)

	// Skip Finder metadata files.
	if base == ".DS_Store" {
		return true
	}
	// Skip Apple Double files, like '._somefile.ext'
	if strings.HasPrefix(base, "._") {
		return true
	}

	return false
}
