package logging

// Wrapper that calls Printf on both the standard and verbose loggers.
func Verbosef(format string, args ...any) {
	logger.Printf(format, args...)
	verbose.Printf(format, args...)
}

// Like Verbosef, but uses Println rather than Printf.
func Verbose(args ...any) {
	logger.Println(args...)
	verbose.Println(args...)
}
