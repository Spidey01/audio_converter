// SPDX-License-Identifier: Zlib
// Copyright 2025, Terry M. Poulin.
package ffmpeg

import (
	"audio_converter/internal/logging"
	"audio_converter/internal/options"
	"context"
	"os"
	"os/exec"
	"strings"
)

// Extract cover art from input to output. If provided, scale is used as the
// value for the -s flag. If clobbering is less than, equal, or greater than
// zero then the ffmpeg will be told to overwrite, prompt, or no clobber
// existing files.
//
// The clobbering flag is kinda hacky, but there's only one tool that relies on this function.
func ExtractCoverArt(ctx context.Context, opts *options.ExtracterOptions) error {
	args := []string{
		// Set the input file.
		"-i", opts.InputFile,
		// Set the necessary options.
		"-map", "0:v",
		"-map", "-0:V",
	}
	if opts.Codec != "" {
		args = append(args, "-c", opts.Codec)
	}
	// Scale if ya got it!
	if opts.Scale != "" {
		args = append(args, "-s", opts.Scale)
	}
	if opts.NoClobber {
		args = append(args, "-n")
	} else if opts.Overwrite {
		args = append(args, "-y")
	}
	// Set the output file.
	args = append(args, opts.OutputFile)

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	logging.Println("Running:", strings.Join(cmd.Args, " "))
	return cmd.Run()
}
