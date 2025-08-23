// SPDX-License-Identifier: Zlib
// Copyright 2025, Terry M. Poulin.
package options

import (
	"debug/buildinfo"
	"fmt"
	"os"
)

// This should be something like '<module version>-<date time>-<short
// commit>[+dirty]'.  E.g., vN.N.N-YYYYMMDDHHMMSS-XXXXXXXXXXXX, depending on the
// module directive in go.mod, the clock at compile time, and the current commit
// situation. If updating this in init fails, it will be "unknown".
var Version string = "unknown"

// Used by the various option parsers to indicate --version / print it and exit.
var PrintVersion bool

// This avoids having to use go generate as part of the build to generate a
// value from git and then use go embed to include it. Which would require some
// conditional compilation based on the shell. It's less flexible, but also 90%
// of what I want anyway.
func init() {
	if prog, err := os.Executable(); err != nil {
		fmt.Fprintf(os.Stderr, "os.Executable: err: %v", err)
	} else if info, err := buildinfo.ReadFile(prog); err != nil {
		fmt.Fprintf(os.Stderr, "buildinfo.ReadFile: err: %v", err)
	} else {
		Version = info.Main.Version
	}
}
