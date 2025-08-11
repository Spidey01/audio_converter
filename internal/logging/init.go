package logging

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
)

var logger *log.Logger = log.New(io.Discard, "", 0)
var verbose *log.Logger = log.New(io.Discard, "", 0)

// Initializes loggers based on the provided settings.
//
// The top level logger will point to file specified by name, to stdout
// if name is "-", or to the bit bucket if name is "".
//
// The verbose logger will either go to stdout or the bitbucket depending on
// verboseMode.
func Initialize(ctx context.Context, name string, verboseMode bool) error {
	if name != "" {
		flags := log.Ldate | log.Ltime | log.Lshortfile
		if name == "-" {
			logger = log.New(os.Stdout, "", flags)
		} else if fp, err := os.Create(name); err != nil {
			return fmt.Errorf("failed creating log file %s: %w", name, err)
		} else {
			logger = log.New(fp, "", flags)
			context.AfterFunc(ctx, func() { fp.Close() })
		}
	}
	if verboseMode {
		verbose = log.New(os.Stdout, verbose.Prefix(), verbose.Flags())
	}
	return nil
}
