// SPDX-License-Identifier: Zlib
// Copyright 2025, Terry M. Poulin.
package main

import (
	"audio_converter/internal/ffmpeg"
	"audio_converter/internal/filesystem"
	"audio_converter/internal/logging"
	"audio_converter/internal/options"
	"context"
	"log"
	"os"
	"os/signal"
)

var opts *options.ExporterOptions
var InRoot filesystem.FS
var OutRoot filesystem.FS

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	if opts = options.NewExporterOptions(os.Args, nil); opts == nil {
		// Arg parsing error. Usage, etc is handled by the constructor.
		os.Exit(1)
	}

	if err := logging.Initialize(ctx, opts.LogFile, opts.Verbose); err != nil {
		log.Fatalln(err)
	}

	// Look up the default options for the current format.  This is done here,
	// because the format specific defaults live in the ffmpeg package, which
	// imports options to provide the same data type.
	defs := ffmpeg.GetDefaultOptions("." + opts.Format)
	opts.Merge(defs)

	done := logging.When("export", logging.Verbose)
	defer done()

	exporter := newExporter(ctx, opts)
	if err := exporter.Run(); err != nil {
		log.Fatalln(err)
	}
}
