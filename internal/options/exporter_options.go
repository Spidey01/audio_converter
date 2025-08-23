// SPDX-License-Identifier: Zlib
// Copyright 2025, Terry M. Poulin.
package options

import (
	"fmt"
	"os"
	"strings"
)

type ExporterOptions struct {
	ConverterOptions
	InRoot        string
	OutRoot       string
	Format        string
	MaxQueue      int
	MaxJobs       int
	CopyUnknown   bool
	noCopyUnknown bool
}

func NewExporterOptions(args []string) *ExporterOptions {
	opts := &ExporterOptions{}
	opts.AddOptions(args)
	defer opts.onError() // handle printing if opts.Err != nil

	if opts.Err = opts.Parse(args[1:]); opts.Err != nil {
		return nil
	}
	if opts.Err = opts.Validate(); opts.Err != nil {
		return nil
	}
	return opts
}

func (opts *ExporterOptions) AddOptions(args []string) {
	// fs := AddGlobalOptions(args, &opts.GlobalOptions)
	opts.ConverterOptions.AddOptions(args, &ConverterOptions{})
	// So, this would work ^, but takes us back to the injecting defaults issue.
	fs := opts.fs

	fs.BoolVar(&opts.CopyUnknown, "C", true, "Copy unknown files, like album art and booklets. (default)")
	fs.BoolVar(&opts.noCopyUnknown, "N", false, "Do not copy unknown files.")
	fs.IntVar(&opts.MaxQueue, "q", 0, "Sets the maximum queue depth.")
	fs.IntVar(&opts.MaxJobs, "j", 0, "Sets the maximum number of concurrent jobs.")
	fs.Usage = opts.Usage

	// Since we can't just look up the flag and set its DefValue, we can't use
	// Func to bind a parse function to the flag and have working unit tests,
	// since those expect the DefValue and Value to actually work. So instead,
	// we need to make this a normal flag and validate after parse.
	fs.StringVar(&opts.Format, "f", "m4a", "Set the output extension/format.")
}

func (opts *ExporterOptions) Parse(args []string) error {
	if opts.Err = opts.parse(args); opts.Err != nil {
		return nil
	}
	if opts.noCopyUnknown {
		opts.CopyUnknown = false
	}
	opts.InRoot = opts.fs.Arg(0)
	opts.OutRoot = opts.fs.Arg(1)

	return nil
}

func (opts *ExporterOptions) Validate() error {
	opts.Format = strings.ToLower(opts.Format)
	switch opts.Format {
	case "flac", "m4a", "m4r", "mp3":
	default:
		return fmt.Errorf("unsupported format: %q", opts.Format)
	}

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

func (opts *ExporterOptions) Usage() {
	opts.printf("usage: %s [options] {indir} {outdir}\n\n", opts.fs.Name())

	opts.printf("Given a tree of source files %q, export them to the output folder\n", "{inroot}")
	opts.printf("{outdir} retaining the same structure. For example if the %q is like\n", "{inroot}")
	opts.printf("%q then %q will end up with the same structure.\n", "Artists/Album/Song.ext", "{outroot}")
	opts.printf("This is useful for say, exporting a library in a different format.\n")
	opts.printf("\n")
	opts.printf("Copies and conversions are executed concurrently. Defaults are based on CPU core count.\n")
	opts.printf("Set max jobs to lower CPU usage from conversions, the default is one per core.\n")
	opts.printf("\n")

	opts.fs.PrintDefaults()
}
