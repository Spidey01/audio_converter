// SPDX-License-Identifier: Zlib
// Copyright 2025, Terry M. Poulin.
package options

import (
	"flag"
	"fmt"
	"os"
	"path"
	"strings"
)

type Options struct {
	InRoot      string
	OutRoot     string
	Format      string
	LogFile     string
	fs          *flag.FlagSet
	Err         error
	NoClobber   bool
	CopyUnknown bool
	Verbose     bool
}

func NewOptions(args []string) *Options {
	var opts Options

	prog := path.Base(args[0])
	fs := flag.NewFlagSet(prog, flag.ContinueOnError)

	fs.BoolVar(&PrintVersion, "version", false, "Print version and exit")
	fs.BoolVar(&opts.Verbose, "v", false, "Display verbose output.")
	fs.BoolVar(&opts.NoClobber, "n", false, "Set the no clobber flag: don't overwrite files.")
	fs.BoolVar(&opts.CopyUnknown, "C", true, "Copy unknown files, like album art and booklets. (default)")
	fs.StringVar(&opts.LogFile, "log-file", "", "Log to a file.")
	fs.Usage = opts.usage

	// Set this so we _actually_ have the default value if -f not provided.
	opts.Format = "m4a"
	fs.Func("f", "Set the output extension/format. (default: m4a)", opts.parseFormat)
	// Set this so the structure is correct for unit tests.
	fs.Lookup("f").DefValue = opts.Format

	opts.fs = fs
	if opts.Err = fs.Parse(args[1:]); opts.Err != nil {
		// Usage gets called automatically by the Parse after printing the
		// error, or if the error is flag.ErrHelp.
		return nil
	}

	if PrintVersion {
		fmt.Printf("%s version %s\n", fs.Name(), Version)
		return nil
	}

	if opts.Err = opts.parseRoots(); opts.Err != nil {
		fmt.Fprintln(opts.fs.Output(), opts.Err)
		return nil
	}

	return &opts
}

func (opts *Options) parseRoots() error {
	opts.InRoot = opts.fs.Arg(0)
	opts.OutRoot = opts.fs.Arg(1)

	if opts.InRoot == "" {
		return fmt.Errorf("must specify input directory")
	} else if _, err := os.Stat(opts.InRoot); err != nil {
		return fmt.Errorf("input directory: %w", err)
	} else if opts.OutRoot == "" {
		return fmt.Errorf("must specify output directory")
	} else if _, err := os.Stat(opts.OutRoot); err != nil {
		return fmt.Errorf("out directory: %w", err)
	} else if opts.InRoot == opts.OutRoot {
		return fmt.Errorf("cowardly refusing to export %q into itself", opts.InRoot)
	} else if strings.HasPrefix(opts.OutRoot, opts.InRoot) {
		return fmt.Errorf("output directory cannot be nested within input directory")
	}

	return nil
}

func (opts *Options) parseFormat(s string) error {
	opts.Format = strings.ToLower(s)
	switch opts.Format {
	case "flac", "m4a", "m4r", "mp3":
		return nil
	default:
		return fmt.Errorf("unsupported format: %q", s)
	}
}

func (opts *Options) usage() {
	out := func(f string, a ...any) {
		stream := opts.fs.Output()
		fmt.Fprintf(stream, f, a...)
	}

	out("usage: %s [options] {indir} {outdir}\n\n", opts.fs.Name())

	out("Given a tree of source files %q, export them to the output folder\n", "{inroot}")
	out("{outdir} retaining the same structure. For example if the %q is like\n", "{inroot}")
	out("%q then %q will end up with the same structure.\n", "Artists/Album/Song.ext", "{outroot}")
	out("This is useful for say, exporting a library in a different format.\n")
	out("\n")

	opts.fs.PrintDefaults()
}
