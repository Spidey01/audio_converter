// SPDX-License-Identifier: Zlib
// Copyright 2025, Terry M. Poulin.
package options

import (
	"flag"
	"fmt"
	"io"
	"path"
	"strings"
)

type ConverterOptions struct {
	fs               *flag.FlagSet
	Err              error
	LogFile          string
	InputFile        string
	OutputFile       string
	BitRate          string
	Codec            string
	InputExtensions  []string
	OutputExtensions []string
	Channels         int
	SampleRate       int
	stereo           bool
	mono             bool
	NoClobber        bool
	Overwrite        bool
	Verbose          bool
}

func NewConverterOptions(args []string, defaults ConverterOptions) *ConverterOptions {
	fs := flag.NewFlagSet(path.Base(args[0]), flag.ContinueOnError)
	opts := &ConverterOptions{
		fs:               fs,
		InputExtensions:  defaults.InputExtensions,
		OutputExtensions: defaults.OutputExtensions,
	}
	fs.Usage = func() {
		out := opts.fs.Output()
		io.WriteString(out, fmt.Sprintf("%s [options] {input} {output}\n", opts.fs.Name()))
		io.WriteString(out, "\nConverts the {input} file into {output} using ffmpeg\n\n")
		if len(opts.InputExtensions) > 0 {
			io.WriteString(out, fmt.Sprintf("Supported input extensions: %s\n\n", strings.Join(opts.InputExtensions, " ")))
		}
		if len(opts.OutputExtensions) > 0 {
			io.WriteString(out, fmt.Sprintf("Be sure to include an %s extension in {output}\n\n", opts.OutputExtensions[0]))
		}
		opts.fs.PrintDefaults()
	}

	fs.BoolVar(&PrintVersion, "version", false, "Print version and exit")
	fs.BoolVar(&opts.Verbose, "v", false, "Display verbose output.")
	fs.BoolVar(&opts.NoClobber, "n", false, "Set the no clobber flag: don't overwrite files.")
	fs.BoolVar(&opts.Overwrite, "y", false, "Overwrite files without prompting. (default: prompt)")
	fs.StringVar(&opts.LogFile, "log-file", "", "Log to a file.")

	fs.StringVar(&opts.BitRate, "b", defaults.BitRate, "Sets the output bitrate.")
	fs.StringVar(&opts.Codec, "c", defaults.Codec, "Sets the ffmpeg codec.")

	sampleRate := 44100
	if defaults.SampleRate > 0 {
		sampleRate = defaults.SampleRate
	}
	fs.IntVar(&opts.SampleRate, "r", sampleRate, "Sets sample rate.")

	fs.BoolVar(&opts.stereo, "s", false, "Sets 2.0/stereo mode.")
	fs.BoolVar(&opts.mono, "m", false, "Sets 1.0/mono mode.")

	if opts.Err = fs.Parse(args[1:]); opts.Err != nil {
		// Usage gets called automatically by the Parse after printing the
		// error, or if the error is flag.ErrHelp.
		return nil
	}
	if opts.stereo {
		opts.Channels = 2
	} else if opts.mono {
		opts.Channels = 1
	} else {
		opts.Channels = defaults.Channels
	}

	if PrintVersion {
		fmt.Printf("%s version %s\n", fs.Name(), Version)
		return nil
	}

	if opts.Err = opts.parseFiles(); opts.Err != nil {
		fmt.Fprintln(opts.fs.Output(), opts.Err)
		return nil
	}

	return opts
}

func (opts *ConverterOptions) parseFiles() error {
	opts.InputFile = opts.fs.Arg(0)
	opts.OutputFile = opts.fs.Arg(1)

	if opts.InputFile == "" {
		return fmt.Errorf("must specify input directory")
	} else if opts.OutputFile == "" {
		return fmt.Errorf("must specify output directory")
	} else if opts.InputFile == opts.OutputFile {
		return fmt.Errorf("cowardly refusing to convert %q into itself", opts.InputFile)
	}

	return nil
}
