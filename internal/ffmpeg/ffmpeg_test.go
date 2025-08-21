// SPDX-License-Identifier: Zlib
// Copyright 2025, Terry M. Poulin.
package ffmpeg

import (
	"audio_converter/internal/options"
	"slices"
	"strconv"
	"testing"
)

func TestIsMediaFile(t *testing.T) {
	for _, ext := range InputExtensions {
		if !IsMediaFile("foo." + ext) {
			t.Errorf("Failed to detect %q", ext)
		}
	}
	if IsMediaFile(".yada") {
		t.Errorf("Detected random junk as media file")
	}
}

func TestMakeCmd(t *testing.T) {
	assert := func(t *testing.T, flag, arg string, opts *options.ConverterOptions) {
		if cmd := makeCmd(t.Context(), opts); cmd == nil {
			t.Errorf("makeCmd failed on ConverterOptions: %+v", opts)
		} else if i := slices.Index(cmd.Args, flag); i == -1 && flag != "" {
			t.Errorf("makeCmd didn't add flag %q", flag)
		} else if arg != "" && flag != "" {
			t.Logf("args: %+v expected: %s %s", cmd.Args, flag, arg)
			if len(cmd.Args) < i+1 {
				t.Errorf("makeCmd didn't add a value for %q", flag)
			} else if cmd.Args[i+1] != arg {
				t.Errorf("makeCmd used %s %s instead of %s %s", flag, cmd.Args[i+1], flag, arg)
			}
		}
	}
	for _, value := range []string{"128k", "256k", "320k", "foo", "bar"} {
		assert(t, "-c:a", value, &options.ConverterOptions{Codec: value})
		assert(t, "-b:a", value, &options.ConverterOptions{BitRate: value})
		// While the value of these flags matter, it's left to FFmpeg to decide
		// if they're good or bad, so no need for a separate test.
		assert(t, "-c:v", value, &options.ConverterOptions{CoverArtFormat: value})
		assert(t, "-s", value, &options.ConverterOptions{Scale: value})
		assert(t, "-i", value, &options.ConverterOptions{InputFile: value})
		assert(t, "", value, &options.ConverterOptions{OutputFile: value})
	}
	for i := 1; i < 10; i++ {
		assert(t, "-ac", strconv.Itoa(i), &options.ConverterOptions{Channels: i})
		assert(t, "-ar", strconv.Itoa(i), &options.ConverterOptions{SampleRate: i})
	}
	assert(t, "-y", "", &options.ConverterOptions{GlobalOptions: options.GlobalOptions{NoClobber: false, Overwrite: true}})
	assert(t, "-n", "", &options.ConverterOptions{GlobalOptions: options.GlobalOptions{NoClobber: true, Overwrite: false}})
}

func TestGetDefaultOptions(t *testing.T) {
	assert := func(expected *options.ConverterOptions) {
		// The first is used as the
		for _, ext := range expected.OutputExtensions {
			actual := GetDefaultOptions(ext)
			if actual.Err != nil {
				t.Errorf("%s -> err: %v", ext, actual.Err)
			} else if actual.Codec != expected.Codec {
				t.Errorf("%s -> bad options:\nactual  : %+v\nexpected: %+v", ext, actual, expected)
			}
		}
	}
	assert(AacOptions)
	assert(FlacOptions)
	assert(Mp3Options)
}
