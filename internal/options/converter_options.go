// SPDX-License-Identifier: Zlib
// Copyright 2025, Terry M. Poulin.
package options

import (
	"reflect"
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

// Creates a new instance based on defaults.
func NewConverterOptions(args []string, defaults *ConverterOptions) *ConverterOptions {
	opts := &ConverterOptions{}
	opts.AddOptions(args, defaults)
	defer opts.onError()
	if opts.Err = opts.Parse(args[1:]); opts.Err != nil {
		return nil
	}
	if opts.Err = opts.Validate(); opts.Err != nil {
		return nil
	}
	return opts
}

func (opts *ConverterOptions) AddOptions(args []string, defs *ConverterOptions) {
	fs := AddGlobalOptions(args, &opts.GlobalOptions)
	fs.Usage = opts.Usage

	opts.InputExtensions = defs.InputExtensions
	opts.OutputExtensions = defs.OutputExtensions
	opts.Channels = defs.Channels

	fs.StringVar(&opts.BitRate, "b", defs.BitRate, "Sets the output bitrate.")
	fs.StringVar(&opts.Codec, "c", defs.Codec, "Sets the ffmpeg codec.")

	if defs.SampleRate > 0 {
		opts.SampleRate = defs.SampleRate
	} else if opts.SampleRate == 0 {
		opts.SampleRate = 44100
	}
	fs.IntVar(&opts.SampleRate, "r", opts.SampleRate, "Sets sample rate.")
	fs.BoolVar(&opts.stereo, "s", defs.stereo, "Sets 2.0/stereo mode.")
	fs.BoolVar(&opts.mono, "m", defs.mono, "Sets 1.0/mono mode.")

	if defs.CoverArtFormat == "" && opts.CoverArtFormat == "" {
		opts.CoverArtFormat = "copy"
	}
	fs.StringVar(&opts.CoverArtFormat, "cover", opts.CoverArtFormat, "Sets whether cover art is copied or converted to `FMT`.\nValues may be mjpeg, png, or copy.")
	fs.StringVar(&opts.Scale, "scale", defs.Scale, "When converting cover art, scale it to `SCALE`. Format is HEIGHTxWIDTH. E.g., \"500x500\"\nNote: only takes affect when -cover is not set to copy")
}

func (opts *ConverterOptions) Parse(args []string) error {
	if opts.Err = opts.parse(args); opts.Err != nil {
		return nil
	}
	opts.InputFile = opts.fs.Arg(0)
	opts.OutputFile = opts.fs.Arg(1)
	return nil
}

func (opts *ConverterOptions) Validate() error {
	if opts.stereo {
		opts.Channels = 2
	} else if opts.mono {
		opts.Channels = 1
	}
	if err := ValidateHeightWidth(opts.Scale); err != nil {
		return err
	}
	if err := ValidateFileArgs(opts.InputFile, opts.OutputFile); err != nil {
		return err
	}
	return nil
}

func (opts *ConverterOptions) Usage() {
	opts.printf("%s [options] {input} {output}\n", opts.fs.Name())
	opts.printf("\nConverts the {input} file into {output} using ffmpeg\n\n")
	if len(opts.InputExtensions) > 0 {
		opts.printf("Supported input extensions: %s\n\n", strings.Join(opts.InputExtensions, " "))
	}
	if len(opts.OutputExtensions) > 0 {
		opts.printf("Be sure to include an %s extension in {output}\n\n", opts.OutputExtensions[0])
	}
	opts.fs.PrintDefaults()
}

// Merges options from `source` into the current options structure. Any fields
// that are zero initialized are ignored.
//
// Note, this performs a shallow copy. That's suitable for our purposes, but not great.
func (opts *ConverterOptions) Merge(source *ConverterOptions) {
	// Since they're pointers, ValueOf(...).Elem() should never crash.
	self := reflect.ValueOf(opts).Elem()
	other := reflect.ValueOf(source).Elem()
	if !other.IsValid() {
		return
	}
	if self.NumField() != other.NumField() {
		// I don't think this is technically possible when they're the same
		// type. But if it ever is, I want to see it logged (^_^)
		panic("ConverterOptions.Merge: different number of fields for the same type")
	}
	for i := range other.NumField() {
		s := self.Field(i)
		o := other.Field(i)

		// These should never fail, but let's be pendantic.
		if !o.IsValid() || o.IsZero() {
			continue
		}
		// Skip already initialized.
		if !s.IsZero() {
			continue
		}
		if !s.CanSet() {
			panic("ConverterOptions.Merge: can't set destination field")
		}
		s.Set(o)
	}
}
