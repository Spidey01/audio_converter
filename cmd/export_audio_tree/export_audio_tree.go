// SPDX-License-Identifier: Zlib
// Copyright 2025, Terry M. Poulin.
package main

import (
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

	if opts = options.NewExporterOptions(os.Args); opts == nil {
		// Arg parsing error. Usage, etc is handled by the constructor.
		os.Exit(1)
	}

	if err := logging.Initialize(ctx, opts.LogFile, opts.Verbose); err != nil {
		log.Fatalln(err)
	}

	done := logging.When("export", logging.Verbose)
	defer done()

	exporter := newExporter(ctx, opts)
	if err := exporter.Run(); err != nil {
		log.Fatalln(err)
	}
}
