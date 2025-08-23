// SPDX-License-Identifier: Zlib
// Copyright 2025, Terry M. Poulin.
package options

import (
	"fmt"
	"regexp"
)

type ExtracterOptions struct {
	GlobalOptions
	InputFile  string
	OutputFile string
	Codec      string
	Scale      string
}

func NewExtracterOptions(args []string) *ExtracterOptions {
	opts := &ExtracterOptions{}
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

func (opts *ExtracterOptions) Usage() {
	opts.printf("%s [options] {input} {output}\n", opts.fs.Name())
	opts.printf("\nExtracts cover art from {input} into {output} using ffmpeg.\n")
	opts.printf("The format is detected based on the file extension of {output} unless the codec is specified.\n")
	opts.printf("For best compatibility, consider scaling to 500x500 as a jpg.\n\n")
	opts.fs.PrintDefaults()
}

func (opts *ExtracterOptions) AddOptions(args []string) {
	fs := AddGlobalOptions(args, &opts.GlobalOptions)
	fs.StringVar(&opts.Codec, "c", "", "Override the ffmpeg codec rather than based on {output}.")
	fs.StringVar(&opts.Scale, "s", "", "Alias for -scale `SCALE`")
	fs.StringVar(&opts.Scale, "scale", "", "Scale image to `SCALE`. Format is HEIGHTxWIDTH. E.g., \"500x500\"")
	fs.Usage = opts.Usage
}

func (opts *ExtracterOptions) Parse(args []string) error {
	if opts.Err = opts.parse(args); opts.Err != nil {
		return nil
	}
	opts.InputFile = opts.fs.Arg(0)
	opts.OutputFile = opts.fs.Arg(1)
	return nil
}

func (opts *ExtracterOptions) Validate() error {
	if err := ValidateFileArgs(opts.InputFile, opts.OutputFile); err != nil {
		return err
	}
	if err := ValidateHeightWidth(opts.Scale); err != nil {
		return err
	}
	return nil
}

func ValidateHeightWidth(value string) error {
	if matched, err := regexp.MatchString("[[:digit:]]+x[[:digit:]]+", value); err != nil {
		return err
	} else if !matched && value != "" {
		return fmt.Errorf("bad scale format: %q", value)
	}
	return nil
}
