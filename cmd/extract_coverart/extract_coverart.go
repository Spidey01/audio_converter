package main

import (
	"audio_converter/internal/ffmpeg"
	"audio_converter/internal/logging"
	"audio_converter/internal/options"
	"context"
	"os"
	"os/signal"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()
	opts := options.NewExtracterOptions(os.Args)
	if opts == nil {
		// Arg parsing error. Usage, etc is handled by the constructor.
		os.Exit(1)
	}
	if err := logging.Initialize(ctx, "-", opts.Verbose); err != nil {
		logging.Fatalln(err)
	}
	if err := ffmpeg.ExtractCoverArt(ctx, opts); err != nil {
		logging.Fatalln(err)
	}
}
