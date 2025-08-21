// SPDX-License-Identifier: Zlib
// Copyright 2025, Terry M. Poulin.
package options

import (
	"flag"
	"fmt"
	"io"
	"path"
	"regexp"
)

type ExtracterOptions struct {
	InputFile  string
	OutputFile string
	Codec      string
	Scale      string
	fs         *flag.FlagSet
	Err        error
	NoClobber  bool
	Overwrite  bool
	Verbose    bool
}

func NewExtracterOptions(args []string) *ExtracterOptions {
	var opts ExtracterOptions

	prog := path.Base(args[0])
	fs := flag.NewFlagSet(prog, flag.ContinueOnError)
	fs.BoolVar(&PrintVersion, "version", false, "Print version and exit")
	fs.StringVar(&opts.Codec, "c", "", "Override the ffmpeg codec rather than based on {output}.")
	fs.StringVar(&opts.Scale, "s", "", "Alias for -scale `SCALE`")
	fs.StringVar(&opts.Scale, "scale", "", "Scale image to `SCALE`. Format is HEIGHTxWIDTH. E.g., \"500x500\"")
	fs.BoolVar(&opts.NoClobber, "n", false, "Set the no clobber flag: don't overwrite files.")
	fs.BoolVar(&opts.Overwrite, "y", false, "Overwrite files without prompting.")
	fs.BoolVar(&opts.Verbose, "v", false, "Set verbose mode.")
	fs.Usage = func() {
		out := opts.fs.Output()
		io.WriteString(out, fmt.Sprintf("%s [options] {input} {output}\n", opts.fs.Name()))
		io.WriteString(out, "\nExtracts cover art from {input} into {output} using ffmpeg.\n")
		io.WriteString(out, "The format is detected based on the file extension of {output} unless the codec is specified.\n")
		io.WriteString(out, "For best compatibility, consider scaling to 500x500 as a jpg.\n\n")
		opts.fs.PrintDefaults()
	}

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

	opts.InputFile = opts.fs.Arg(0)
	opts.OutputFile = opts.fs.Arg(1)
	if opts.InputFile == "" {
		opts.Err = fmt.Errorf("must specify an {input} file")
	} else if opts.OutputFile == "" {
		opts.Err = fmt.Errorf("must specify an {output} file")
	} else if opts.InputFile == opts.OutputFile {
		opts.Err = fmt.Errorf("cowardly refusing to extract %s to itself", opts.InputFile)
	} else if err := ValidateHeightWidth(opts.Scale); err != nil {
		opts.Err = err
	}
	if opts.Err != nil {
		fmt.Fprintln(opts.fs.Output(), opts.Err)
		opts.fs.Usage()
		return nil
	}

	return &opts
}

func ValidateHeightWidth(value string) error {
	if matched, err := regexp.MatchString("[[:digit:]]+x[[:digit:]]+", value); err != nil {
		return err
	} else if !matched && value != "" {
		return fmt.Errorf("bad scale format: %q", value)
	}
	return nil
}
