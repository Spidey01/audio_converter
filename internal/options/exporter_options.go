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

type ExporterOptions struct {
	InRoot        string
	OutRoot       string
	Format        string
	LogFile       string
	fs            *flag.FlagSet
	Err           error
	NoClobber     bool
	Overwrite     bool
	CopyUnknown   bool
	Verbose       bool
	noCopyUnknown bool
}

func NewExporterOptions(args []string) *ExporterOptions {
	var opts ExporterOptions

	prog := path.Base(args[0])
	fs := flag.NewFlagSet(prog, flag.ContinueOnError)

	fs.BoolVar(&PrintVersion, "version", false, "Print version and exit")
	fs.BoolVar(&opts.Verbose, "v", false, "Display verbose output.")
	fs.BoolVar(&opts.NoClobber, "n", false, "Set the no clobber flag: don't overwrite files.")
	fs.BoolVar(&opts.Overwrite, "y", false, "Overwrite files without prompting. (default: prompt)")
	fs.BoolVar(&opts.CopyUnknown, "C", true, "Copy unknown files, like album art and booklets. (default)")
	fs.BoolVar(&opts.noCopyUnknown, "N", false, "Do not copy unknown files.")
	fs.StringVar(&opts.LogFile, "log-file", "", "Log to a file.")
	fs.Usage = opts.usage

	// Since we can't just look up the flag and set its DefValue, we can't use
	// Func to bind a parse function to the flag and have working unit tests,
	// since those expect the DefValue and Value to actually work. So instead,
	// we need to make this a normal flag and validate after parse.
	fs.StringVar(&opts.Format, "f", "m4a", "Set the output extension/format.")

	opts.fs = fs
	if opts.Err = fs.Parse(args[1:]); opts.Err != nil {
		// Usage gets called automatically by the Parse after printing the
		// error, or if the error is flag.ErrHelp.
		return nil
	}

	if opts.Err = opts.validateFormat(opts.Format); opts.Err != nil {
		fmt.Fprintln(opts.fs.Output(), opts.Err)
		return nil
	}
	if opts.noCopyUnknown {
		opts.CopyUnknown = false
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

func (opts *ExporterOptions) parseRoots() error {
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

func (opts *ExporterOptions) validateFormat(s string) error {
	opts.Format = strings.ToLower(s)
	switch opts.Format {
	case "flac", "m4a", "m4r", "mp3":
		return nil
	default:
		return fmt.Errorf("unsupported format: %q", s)
	}
}

func (opts *ExporterOptions) usage() {
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
