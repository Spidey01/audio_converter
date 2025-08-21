// SPDX-License-Identifier: Zlib
// Copyright 2025, Terry M. Poulin.
package options

import (
	"fmt"
	"io"
	"strings"
)

type ConverterOptions struct {
	GlobalOptions
	InputFile        string
	OutputFile       string
	BitRate          string
	Codec            string
	CoverArtFormat   string
	Scale            string
	InputExtensions  []string
	OutputExtensions []string
	Channels         int
	SampleRate       int
	stereo           bool
	mono             bool
}

func NewConverterOptions(args []string, defaults *ConverterOptions) *ConverterOptions {
	opts := &ConverterOptions{
		InputExtensions:  defaults.InputExtensions,
		OutputExtensions: defaults.OutputExtensions,
	}
	defer opts.onError()
	fs := AddGlobalOptions(args, &opts.GlobalOptions)
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

	fs.StringVar(&opts.BitRate, "b", defaults.BitRate, "Sets the output bitrate.")
	fs.StringVar(&opts.Codec, "c", defaults.Codec, "Sets the ffmpeg codec.")
	sampleRate := 44100
	if defaults.SampleRate > 0 {
		sampleRate = defaults.SampleRate
	}
	fs.IntVar(&opts.SampleRate, "r", sampleRate, "Sets sample rate.")
	fs.BoolVar(&opts.stereo, "s", false, "Sets 2.0/stereo mode.")
	fs.BoolVar(&opts.mono, "m", false, "Sets 1.0/mono mode.")
	fs.StringVar(&opts.CoverArtFormat, "cover", "copy", "Sets whether cover art is copied or converted to `FMT`.\nValues may be mjpeg, png, or copy.")
	fs.StringVar(&opts.Scale, "scale", "", "When converting cover art, scale it to `SCALE`. Format is HEIGHTxWIDTH. E.g., \"500x500\"\nNote: only takes affect when -cover is not set to copy")

	if opts.Err = opts.parse(args[1:]); opts.Err != nil {
		return nil
	}
	if opts.stereo {
		opts.Channels = 2
	} else if opts.mono {
		opts.Channels = 1
	} else {
		opts.Channels = defaults.Channels
	}
	if opts.Err = ValidateHeightWidth(opts.Scale); opts.Err != nil {
		return nil
	}
	if opts.Err = opts.parseFiles(); opts.Err != nil {
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

	// opts.MetadataFile is optional. The most we could do is stat, but since we
	// don't stat opts.InputFile, there is no point.

	return nil
}
