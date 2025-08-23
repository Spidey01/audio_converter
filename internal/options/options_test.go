// SPDX-License-Identifier: Zlib
// Copyright 2025, Terry M. Poulin.
package options

import (
	"context"
	"flag"
	"math"
	"os"
	"path"
	"strconv"
	"testing"
)

// Type for factory functions.
type factoryFunc func([]string) *flag.FlagSet

// Helper class for unit testing options. Uses a factory that creates the
// options of whatever type and returns its flag set, because the structures
// don't follow an interface, they're just varied sets of fields.
type FlagTest struct {
	factory      factoryFunc
	name         string
	goodValues   []string
	badValues    []string
	defaultValue string
}

// Look up the flag object.
func (test *FlagTest) lookup(t *testing.T, fs *flag.FlagSet) *flag.Flag {
	f := fs.Lookup(test.name)
	if f == nil {
		t.Fatalf("Failed to lookup flag %q", test.name)
	}
	return f
}

// Assert the flag's value is the expected.
func (test FlagTest) assert(t *testing.T, flag *flag.Flag, expected, msg string) {
	if actual := flag.Value.String(); actual != expected {
		t.Errorf("%s: -%s: actual: %q expected: %q\nflag: %+v", msg, test.name, actual, expected, flag)
	}
}

// Handles testing a binary (on/off) flag, like -v for verbose.
func (test *FlagTest) BoolFlag(t *testing.T) {
	if len(test.goodValues) != 0 && len(test.badValues) != 0 {
		t.Fatalf("%s test data does not define a boolean flag", test.name)
	}
	prog, input, output := setup(t)

	// Default case, not provided should be default value.
	args := []string{prog, input, output}
	if fs := test.factory(args); fs == nil {
		t.Errorf("Failed on default args")
	} else {
		test.assert(t, test.lookup(t, fs), test.defaultValue, "Default args did not yield default value")
	}

	// Provided, should invert the default value.
	args = []string{prog, "-" + test.name, input, output}
	if fs := test.factory(args); fs == nil {
		t.Errorf("Failed with -%s", test.name)
	} else {
		expected := "true"
		if test.defaultValue == expected {
			expected = "false"
		}
		test.assert(t, test.lookup(t, fs), expected, "flag did not invert default value")
	}
}

// Handles testing a flag taking a string, like "-s str"
func (test *FlagTest) StringFlag(t *testing.T) {
	prog, input, output := setup(t)

	args := []string{prog, input, output}
	if fs := test.factory(args); fs == nil {
		t.Errorf("Failed on default args")
	} else {
		test.assert(t, test.lookup(t, fs), test.defaultValue, "Default args did not yield default value")
	}

	for _, value := range test.goodValues {
		args = []string{prog, "-" + test.name, value, input, output}
		t.Logf("Testing args %+v", args)
		fs := test.factory(args)
		if fs == nil {
			t.Errorf("Failed on -%s %s", test.name, value)
			continue
		}
		f := fs.Lookup(test.name)
		if f == nil {
			t.Errorf("Failed to lookup %q", test.name)
			continue
		}
		if s := f.Value.String(); s != value {
			t.Errorf("-%s actual: %q expected: %s", test.name, s, value)
		}
	}
	for _, value := range test.badValues {
		args = []string{prog, "-" + test.name, value, input, output}
		t.Logf("Testing args %+v", args)
		fs := test.factory(args)
		if fs != nil {
			t.Errorf("Failed on -%s %s - value was parsed: %+v", test.name, value, fs.Lookup(test.name))
			continue
		}
	}
}

// Handles testing a flag taking an int, like "-n number"
func (test *FlagTest) IntFlag(t *testing.T) {
	// Technically the only difference is 'and make sure it can be converted to
	// an integer' which is handled by the flag library presuming the type was
	// registered as such. So, for now, let's just call this an alias of the
	// string test.
	test.StringFlag(t)
}

// Returns three strings: a program name for argv[0], the current working
// directory (can be used as input arg), and a temporary directory (can be used
// as output arg) that will be removed when context.Background() is closed.
func setup(t *testing.T) (string, string, string) {
	inroot, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed getting current working directory: %v", err)
		return "", "", ""
	}
	outroot := os.TempDir()
	context.AfterFunc(context.Background(), func() {
		os.Remove(outroot)
	})
	return "go test", inroot, outroot
}

// Handles testing that -C or -N control the copy unknown flag.
func copyUnknownTest(t *testing.T, factory factoryFunc) {
	prog, input, output := setup(t)
	assert := func(args []string, expected string, msg string) {
		fs := factory(args)
		if fs == nil {
			t.Fatalf("Failed to parse %+v", args)
		}
		f := fs.Lookup("C")
		if f == nil {
			t.Fatal("Failed to look up flag status")
		}
		if f.Value.String() != expected {
			t.Error(msg)
		}
	}
	assert([]string{prog, input, output}, "true", "Default is -N but -C was expected")
	assert([]string{prog, "-C", input, output}, "true", "Flag -C did not turn on copy unknown")
	assert([]string{prog, "-N", input, output}, "false", "Flag -N did not turn on copy unknown")
}

// Handles testing options that take an input file and output file as required args.
func inputOutputFileTest(t *testing.T, factory factoryFunc) {
	prog, input, output := setup(t)
	if factory([]string{prog, input, output}) == nil {
		t.Errorf("Failed with input: %q output: %q", input, output)
	}
	if factory([]string{prog, input, input}) != nil {
		t.Errorf("Failed with input == output == %q", input)
	}
	if factory([]string{prog, input}) != nil {
		t.Errorf("Failed when only one dir provided")
	}
	if factory([]string{prog, "", output}) != nil {
		t.Errorf("Failed when input is empty str")
	}
	if factory([]string{prog, input, ""}) != nil {
		t.Errorf("Failed when output is empty str")
	}
}

// Handles testing options that take an input root and output root as required args.
func rootTest(t *testing.T, factory factoryFunc) {
	prog, inroot, outroot := setup(t)
	if opts := factory([]string{prog, inroot, outroot}); opts == nil {
		t.Fatalf("Baseline failed for inroot=%q outroot=%q", inroot, outroot)
	}

	if opts := factory([]string{prog, inroot, inroot}); opts != nil {
		t.Errorf("Failed to catch outroot = inroot (%q)", inroot)
	}

	if opts := factory([]string{prog, "", outroot}); opts != nil {
		t.Errorf("Failed to catch inroot empty")
	}
	if opts := factory([]string{prog, inroot, ""}); opts != nil {
		t.Errorf("Failed to catch outroot empty")
	}

	doesNotExist := "/does/not/exist"
	if opts := factory([]string{prog, path.Join(inroot, doesNotExist), outroot}); opts != nil {
		t.Errorf("Failed to catch inroot does not exist")
	}
	if opts := factory([]string{prog, inroot, path.Join(outroot, doesNotExist)}); opts != nil {
		t.Errorf("Failed to catch outroot does not exist")
	}

	nested := path.Join(outroot, "nested")
	if err := os.Mkdir(nested, 0755); err != nil {
		t.Errorf("Failed creating %s: %v", nested, err)
	}
	defer os.Remove(nested)
	if opts := factory([]string{prog, outroot, nested}); opts != nil {
		t.Errorf("Failed to catch outroot nested within inroot")
	}

	sxs := path.Join(outroot, "sidebyside")
	if err := os.Mkdir(sxs, 0755); err != nil {
		t.Errorf("Failed creating %s: %v", sxs, err)
	}
	defer os.Remove(sxs)
	if opts := factory([]string{prog, sxs, nested}); opts == nil {
		t.Errorf("Failed to allow side by side within the same parent directory")
	}
}

// Adds tests for global options using t.Run() and the provided factory.
func testGlobalOptions(t *testing.T, factory factoryFunc) {
	// Handles testing the --log-file option.
	t.Run("log file", func(t *testing.T) {
		test := FlagTest{
			factory:    factory,
			name:       "log-file",
			goodValues: []string{"-", "/dev/stdin", "somefile", "/ham/spam"},
		}
		test.StringFlag(t)
	})
	// Handles testing the no clobber (-n) and overwrite flags (-y)
	t.Run("noclobber and overwrite", func(t *testing.T) {
		ft := FlagTest{
			factory:      factory,
			name:         "n",
			defaultValue: "false",
		}
		ft.BoolFlag(t)
		ft.name = "y"
		ft.BoolFlag(t)
	})
	// Handles testing the --version flag.
	t.Run("version", func(t *testing.T) {
		prog, input, output := setup(t)
		opts := factory([]string{prog, "-version", input, output})
		if opts != nil {
			t.Errorf("Failed with --version")
		}
	})
	// Handles testing the -v (verbose) flag.
	t.Run("verbose", func(t *testing.T) {
		ft := FlagTest{
			factory:      factory,
			name:         "v",
			defaultValue: "false",
		}
		ft.BoolFlag(t)
	})
}

// Adds tests for converter options using t.Run() and the provided factory.
func testConverterOptions(t *testing.T, factory factoryFunc) {
	t.Run("bitrate", func(t *testing.T) {
		test := FlagTest{
			factory:      factory,
			name:         "b",
			goodValues:   []string{"12345", "128", "256", "320"},
			defaultValue: DefaulConverterOptions.BitRate,
		}
		test.StringFlag(t)
	})
	t.Run("codec", func(t *testing.T) {
		test := FlagTest{
			factory:      factory,
			name:         "c",
			goodValues:   []string{"some_codec"},
			badValues:    []string{},
			defaultValue: DefaulConverterOptions.Codec,
		}
		test.StringFlag(t)
	})
	t.Run("samplerate", func(t *testing.T) {
		test := FlagTest{
			factory:      factory,
			name:         "r",
			goodValues:   []string{"44100", "48000", "12345"},
			badValues:    []string{"notnumeric", "3.14"},
			defaultValue: strconv.Itoa(DefaulConverterOptions.SampleRate),
		}
		test.StringFlag(t)
	})
	t.Run("cover art", func(t *testing.T) {
		ft := FlagTest{
			factory:      factory,
			name:         "cover",
			goodValues:   []string{"mjpeg", "png"},
			defaultValue: "copy",
			// Bad values are left to FFmpeg to decide.
		}
		ft.StringFlag(t)
	})
	t.Run("scale", func(t *testing.T) { // fails on bad values for exporter opts
		ft := FlagTest{
			factory:    factory,
			name:       "scale",
			goodValues: []string{"500x500", "1x1", "4096x4096"},
			badValues:  []string{"500xWidth", "Heightx500", "HxW"},
		}
		ft.StringFlag(t)
	})
}

// For the purposes of unit testing, these are the defaults. They're
// intentionally not like the actual defaults for various converters.
var DefaulConverterOptions = &ConverterOptions{
	BitRate:          "1024",
	Codec:            "testcodec",
	InputExtensions:  []string{},
	OutputExtensions: []string{},
	Channels:         2,
	SampleRate:       48000,
}

func converterOptionsFactory(args []string) *flag.FlagSet {
	opts := NewConverterOptions(args, DefaulConverterOptions)
	if opts != nil {
		return opts.fs
	}
	return nil
}

// Test cases for ConverterOptions used by various tools, like to_aac.
func TestConverterOptions(t *testing.T) {
	testGlobalOptions(t, converterOptionsFactory)
	testConverterOptions(t, converterOptionsFactory)
	// These flags are used to set an actual field from private values. So it's
	// only meaningful to test them on the actual structure.
	t.Run("stereo and mono", func(t *testing.T) {
		prog, input, output := setup(t)
		opts := NewConverterOptions([]string{prog, input, output}, DefaulConverterOptions)
		if opts.Channels != DefaulConverterOptions.Channels {
			t.Errorf("Failed on default channel config")
		}
		opts = NewConverterOptions([]string{prog, "-s", input, output}, DefaulConverterOptions)
		if opts.Channels != 2 {
			t.Errorf("Failed on -s for stereo: opts.Channels: %d", opts.Channels)
		}
		opts = NewConverterOptions([]string{prog, "-m", input, output}, DefaulConverterOptions)
		if opts.Channels != 1 {
			t.Errorf("Failed on -m for mono: opts.Channels: %d", opts.Channels)
		}
	})
	t.Run("input and output file", func(t *testing.T) {
		inputOutputFileTest(t, converterOptionsFactory)
	})
}

func extracterOptionsFactory(args []string) *flag.FlagSet {
	opts := NewExtracterOptions(args)
	if opts != nil {
		return opts.fs
	}
	return nil
}

func TestExtracterOptions(t *testing.T) {
	testGlobalOptions(t, extracterOptionsFactory)
	t.Run("codec", func(t *testing.T) {
		ft := FlagTest{
			factory:    extracterOptionsFactory,
			name:       "c",
			goodValues: []string{"mjpeg", "png", "gif"},
		}
		ft.StringFlag(t)
	})
	t.Run("scale", func(t *testing.T) {
		ft := FlagTest{
			factory:    extracterOptionsFactory,
			name:       "scale",
			goodValues: []string{"500x500", "1x1", "4096x4096"},
			badValues:  []string{"500xWidth", "Heightx500", "HxW"},
		}
		ft.StringFlag(t)
	})
	t.Run("input and output file", func(t *testing.T) {
		inputOutputFileTest(t, extracterOptionsFactory)
	})
}

func exporterOptionsFactory(args []string) *flag.FlagSet {
	opts := NewExporterOptions(args, DefaulConverterOptions)
	if opts != nil {
		return opts.fs
	}
	return nil
}

func TestExporterOptions(t *testing.T) {
	testGlobalOptions(t, exporterOptionsFactory)
	testConverterOptions(t, exporterOptionsFactory)
	// These flags are used to set an actual field from private values. So it's
	// only meaningful to test them on the actual structure.
	t.Run("stereo and mono", func(t *testing.T) {
		prog, input, output := setup(t)
		opts := NewExporterOptions([]string{prog, input, output}, DefaulConverterOptions)
		if opts.Channels != DefaulConverterOptions.Channels {
			t.Errorf("Failed on default channel config")
		}
		opts = NewExporterOptions([]string{prog, "-s", input, output}, DefaulConverterOptions)
		if opts.Channels != 2 {
			t.Errorf("Failed on -s for stereo: opts.Channels: %d", opts.Channels)
		}
		opts = NewExporterOptions([]string{prog, "-m", input, output}, DefaulConverterOptions)
		if opts.Channels != 1 {
			t.Errorf("Failed on -m for mono: opts.Channels: %d", opts.Channels)
		}
	})
	t.Run("copy unknown", func(t *testing.T) {
		copyUnknownTest(t, exporterOptionsFactory)
	})
	t.Run("max jobs", func(t *testing.T) {
		ft := FlagTest{
			factory:      exporterOptionsFactory,
			name:         "j",
			goodValues:   []string{"1", "4", "8", "32", strconv.Itoa(math.MaxInt)},
			badValues:    []string{"nan"},
			defaultValue: "0",
		}
		ft.IntFlag(t)
	})
	t.Run("max queue", func(t *testing.T) {
		ft := FlagTest{
			factory:      exporterOptionsFactory,
			name:         "q",
			goodValues:   []string{"1", "200", "1024", strconv.Itoa(math.MaxInt)},
			badValues:    []string{"nan"},
			defaultValue: "0",
		}
		ft.IntFlag(t)
	})
	t.Run("format", func(t *testing.T) {
		ft := FlagTest{
			factory:      exporterOptionsFactory,
			name:         "f",
			goodValues:   []string{"flac", "m4a", "m4r", "mp3"},
			badValues:    []string{"unknown"},
			defaultValue: "m4a",
		}
		ft.StringFlag(t)
	})
	t.Run("input and output root", func(t *testing.T) {
		rootTest(t, exporterOptionsFactory)
	})
}
