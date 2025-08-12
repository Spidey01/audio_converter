// SPDX-License-Identifier: Zlib
// Copyright 2025, Terry M. Poulin.
package logging

import (
	"fmt"
	"os"
)

// Printf formats according to a format specifier and writes to standard logger.
func Printf(format string, args ...any) {
	logger.Printf(format, args...)
}

// Println formats using the default formats for its operands and writes to
// standard logger. Spaces are always added between operands and a newline is
// appended.
func Println(args ...any) {
	logger.Println(args...)
}

// Wrapper that ensures the message goes to stderr as well as the log file.
func Fatalf(format string, args ...any) {
	if w := logger.Writer(); w != os.Stdout && w != os.Stderr {
		fmt.Fprintf(os.Stderr, format, args...)
	}
	logger.Fatalf(format, args...)
}

// Wrapper that ensures the message goes to stderr as well as the log file.
func Fatalln(args ...any) {
	if w := logger.Writer(); w != os.Stdout && w != os.Stderr {
		fmt.Fprintln(os.Stderr, args...)
	}
	logger.Fatalln(args...)
}
