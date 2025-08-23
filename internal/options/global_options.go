// SPDX-License-Identifier: Zlib
// Copyright 2025, Terry M. Poulin.
package options

import (
	"errors"
	"flag"
	"fmt"
	"path/filepath"
)

// Options that are common to every single tool.
type GlobalOptions struct {
	fs           *flag.FlagSet
	Err          error
	LogFile      string
	NoClobber    bool
	Overwrite    bool
	Verbose      bool
	PrintVersion bool
}

// Populates opts with a new flag set and the global options. Returns opts.fs.
func AddGlobalOptions(args []string, opts *GlobalOptions) *flag.FlagSet {
	fs := flag.NewFlagSet(filepath.Base(args[0]), flag.ContinueOnError)
	fs.BoolVar(&opts.PrintVersion, "version", false, "Print version and exit")
	fs.StringVar(&opts.LogFile, "log-file", "", "Log to a file.")
	fs.BoolVar(&opts.NoClobber, "n", false, "Set the no clobber flag: don't overwrite files.")
	fs.BoolVar(&opts.Overwrite, "y", false, "Overwrite files without prompting.")
	fs.BoolVar(&opts.Verbose, "v", false, "Set verbose mode.")
	opts.fs = fs
	return opts.fs
}

// Calls opts.fs.Parse, returning an error if that fails, or the application
// should exit (e.g., because of --version). Generally, the error should be
// printed via a deferred onError() and nil returned from a constructor.
func (opts *GlobalOptions) parse(args []string) error {
	if opts.fs == nil {
		panic("No flag set")
	}
	// Usage gets called automatically by the opts.fs.Parse after printing the
	// error, or if the error is flag.ErrHelp.
	err := opts.fs.Parse(args)

	if opts.PrintVersion {
		err = fmt.Errorf("%s version %s", opts.fs.Name(), Version)
	}

	return err
}

func (opts *GlobalOptions) printf(format string, a ...any) {
	fmt.Fprintf(opts.fs.Output(), format, a...)
}

func (opts *GlobalOptions) println(a ...any) {
	fmt.Fprintln(opts.fs.Output(), a...)
}

// Prints opts.Err if it is not nil, followed by usage. You can just defer this
// in constructors and then return nil on error, provided you make sure opts.Err
// is set on error :).
func (opts *GlobalOptions) onError() {
	// If ErrHelp, opts.fs already took care of this.
	if opts.Err != nil && !errors.Is(opts.Err, flag.ErrHelp) {
		fmt.Fprintln(opts.fs.Output(), opts.Err)
		opts.fs.Usage()
	}
}

func ValidateFileArgs(input string, output string) error {
	if input == "" {
		return fmt.Errorf("must specify input file")
	} else if output == "" {
		return fmt.Errorf("must specify output file")
	} else if input == output {
		return fmt.Errorf("cowardly refusing to output the input %q to itself", input)
	}
	return nil
}
