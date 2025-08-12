// SPDX-License-Identifier: Zlib
// Copyright 2025, Terry M. Poulin.
package ffmpeg

import (
	"audio_converter/internal/logging"
	"audio_converter/internal/options"
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

// List of supported input format extensions.
var InputExtensions = []string{
	".flac",
	".m4a", ".m4r",
	".mp3",
	".wav",
}

var DefaultOptions = []options.ConverterOptions{
	FlacOptions,
	AacOptions,
	Mp3Options,
}

func GetDefaultOptions(ext string) options.ConverterOptions {
	for _, opts := range DefaultOptions {
		if slices.Contains(opts.OutputExtensions, ext) {
			return opts
		}
	}
	return options.ConverterOptions{Err: fmt.Errorf("no defaults for extension %q", ext)}
}

// Returns true if name has one of InputExtensions.
func IsMediaFile(name string) bool {
	return slices.Contains(InputExtensions, filepath.Ext(name))
}

// Implements the main() for various to_<format>. Just provide the default
// options for the format. Suitable defaults are exposed as package level
// variables. E.g., FlacOptions.
func ConvertMain(defaults options.ConverterOptions) {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()
	opts := options.NewConverterOptions(os.Args, defaults)
	if opts == nil {
		// Arg parsing error. Usage, etc is handled by the constructor.
		os.Exit(1)
	}
	if err := logging.Initialize(ctx, opts.LogFile, opts.Verbose); err != nil {
		logging.Fatalln(err)
	}
	Convert(ctx, opts)
}

func makeCmd(ctx context.Context, opts *options.ConverterOptions) *exec.Cmd {
	args := []string{
		// Set the input file.
		"-i", opts.InputFile,
		// Wrangle the metadata.
		"-map_metadata", "0",
		// Copy the cover art if it exists.
		"-c:v", "copy",
	}

	if opts.NoClobber {
		args = append(args, "-n")
	} else if opts.Overwrite {
		args = append(args, "-y")
	}
	if opts.Codec != "" {
		// Set the audio codec
		args = append(args, "-c:a", opts.Codec)
	}
	// Set the audio parameters.
	if opts.BitRate != "" {
		args = append(args, "-b:a", opts.BitRate)
	}
	if opts.SampleRate > 0 {
		args = append(args, "-ar", strconv.Itoa(opts.SampleRate))
	}
	if opts.Channels > 0 {
		args = append(args, "-ac", strconv.Itoa(opts.Channels))
	}

	// Set the output file.
	args = append(args, opts.OutputFile)
	return exec.CommandContext(ctx, "ffmpeg", args...)
}

// Runs ffmpeg using the current process's standard I/O for output.
func Convert(ctx context.Context, opts *options.ConverterOptions) error {
	cmd := makeCmd(ctx, opts)
	logging.Println("Running:", strings.Join(cmd.Args, " "))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Runs ffmpeg in a background process, returning its combined standard output
// and error.
func ConvertInBackground(ctx context.Context, opts *options.ConverterOptions) ([]byte, error) {
	cmd := makeCmd(ctx, opts)
	logging.Println("Running in background:", strings.Join(cmd.Args, " "))
	return cmd.CombinedOutput()
}
