// SPDX-License-Identifier: Zlib
// Copyright 2025, Terry M. Poulin.
package options

import (
	"fmt"
	"io"
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
	var opts ExtracterOptions

	fs := AddGlobalOptions(args, &opts.GlobalOptions)
	defer opts.onError() // handle printing if opts.Err != nil

	fs.StringVar(&opts.Codec, "c", "", "Override the ffmpeg codec rather than based on {output}.")
	fs.StringVar(&opts.Scale, "s", "", "Alias for -scale `SCALE`")
	fs.StringVar(&opts.Scale, "scale", "", "Scale image to `SCALE`. Format is HEIGHTxWIDTH. E.g., \"500x500\"")
	fs.Usage = func() {
		out := opts.fs.Output()
		io.WriteString(out, fmt.Sprintf("%s [options] {input} {output}\n", opts.fs.Name()))
		io.WriteString(out, "\nExtracts cover art from {input} into {output} using ffmpeg.\n")
		io.WriteString(out, "The format is detected based on the file extension of {output} unless the codec is specified.\n")
		io.WriteString(out, "For best compatibility, consider scaling to 500x500 as a jpg.\n\n")
		opts.fs.PrintDefaults()
	}

	if opts.Err = opts.parse(args[1:]); opts.Err != nil {
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
