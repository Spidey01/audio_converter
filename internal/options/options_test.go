package options

import (
	"context"
	"flag"
	"os"
	"path"
	"testing"
)

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

func testBoolFlag(t *testing.T, name string) {
	prog, inroot, outroot := setup(t)
	var opts *Options
	var flag *flag.Flag

	// Default case.
	opts = NewOptions([]string{prog, inroot, outroot})
	if opts == nil {
		t.Errorf("NewOptions failed on minimum args!")
		return
	}
	flag = opts.fs.Lookup(name)
	if flag.Value.String() != flag.DefValue {
		t.Errorf("Flag -%s was set %q by default but %q is the default", name, flag.Value, flag.DefValue)
	}

	// Explicit case.
	opts = NewOptions([]string{prog, "-" + name, inroot, outroot})
	if opts == nil {
		t.Errorf("NewOptions failed!")
	}
	flag = opts.fs.Lookup(name)
	if flag.Value.String() != "true" {
		t.Errorf("flag -%s was not set", name)
	}
}

func testStringFlag(t *testing.T, name string, get func(*Options) string, goodValues []string, badValues []string) {
	prog, inroot, outroot := setup(t)
	for _, value := range goodValues {
		args := []string{prog, "-" + name, value, inroot, outroot}
		opts := NewOptions(args)
		if opts == nil {
			// details to stderr
			t.Errorf("NewOptions failed")
		} else if actual := get(opts); actual != value {
			t.Errorf("option -%s got %q but expected %q", name, actual, value)
		}
	}

	for _, value := range badValues {
		args := []string{prog, "-" + name, value, inroot, outroot}
		if opts := NewOptions(args); opts != nil {
			t.Errorf("option -%s got bad value %q but didn't fail!", name, get(opts))
		}
	}

	if opts := NewOptions([]string{prog, inroot, outroot}); opts == nil {
		t.Fatalf("NewOptions failed on minimum args")
	} else if def := opts.fs.Lookup(name).DefValue; get(opts) != def {
		t.Logf("flag: %+v", opts.fs.Lookup(name))
		t.Logf("get(opts): %+v", get(opts))
		t.Errorf("option %q got %q but default %q expected", name, get(opts), def)
	}
}

func testRoots(t *testing.T) {
	prog, inroot, outroot := setup(t)
	if opts := NewOptions([]string{prog, inroot, outroot}); opts == nil {
		t.Fatalf("Baseline failed for inroot=%q outroot=%q", inroot, outroot)
	}

	if opts := NewOptions([]string{prog, inroot, inroot}); opts != nil {
		t.Errorf("Failed to catch outroot = inroot (%q)", inroot)
	}

	if opts := NewOptions([]string{prog, "", outroot}); opts != nil {
		t.Errorf("Failed to catch inroot empty")
	}
	if opts := NewOptions([]string{prog, inroot, ""}); opts != nil {
		t.Errorf("Failed to catch outroot empty")
	}

	doesNotExist := "/does/not/exist"
	if opts := NewOptions([]string{prog, path.Join(inroot, doesNotExist), outroot}); opts != nil {
		t.Errorf("Failed to catch inroot does not exist")
	}
	if opts := NewOptions([]string{prog, inroot, path.Join(outroot, doesNotExist)}); opts != nil {
		t.Errorf("Failed to catch outroot does not exist")
	}

	nested := path.Join(outroot, "nested")
	if err := os.Mkdir(nested, 0755); err != nil {
		t.Errorf("Failed creating %s: %v", nested, err)
	}
	defer os.Remove(nested)
	if opts := NewOptions([]string{prog, outroot, nested}); opts != nil {
		t.Errorf("Failed to catch outroot nested within inroot")
	}

	sxs := path.Join(outroot, "sidebyside")
	if err := os.Mkdir(sxs, 0755); err != nil {
		t.Errorf("Failed creating %s: %v", sxs, err)
	}
	defer os.Remove(sxs)
	if opts := NewOptions([]string{prog, sxs, nested}); opts == nil {
		t.Errorf("Failed to allow side by side within the same parent directory")
	}
}

func TestOptions(t *testing.T) {
	t.Run("roots", testRoots)
	t.Run("verbose", func(t *testing.T) {
		testBoolFlag(t, "v")
	})
	t.Run("no clobber", func(t *testing.T) {
		testBoolFlag(t, "n")
	})
	t.Run("copy unknown", func(t *testing.T) {
		testBoolFlag(t, "C")
	})
	t.Run("format", func(t *testing.T) {
		get := func(o *Options) string { return o.Format }
		goodValues := []string{"flac", "m4a", "m4r", "mp3"}
		badValues := []string{"bogon"}
		testStringFlag(t, "f", get, goodValues, badValues)
	})
	t.Run("log file", func(t *testing.T) {
		get := func(o *Options) string { return o.LogFile }
		goodValues := []string{"file.log"}
		testStringFlag(t, "log-file", get, goodValues, []string{})
	})
}

/*
func TestConverterOptions(t *testing.T) {
	t.Run("verbose", func(t *testing.T) {
		testBoolFlag(t, "v")
	})
	t.Run("no clobber", func(t *testing.T) {
		testBoolFlag(t, "n")
	})
	t.Run("copy unknown", func(t *testing.T) {
		testBoolFlag(t, "C")
	})
	t.Run("format", func(t *testing.T) {
		get := func(o *Options) string { return o.Format }
		goodValues := []string{"flac", "m4a", "m4r", "mp3"}
		badValues := []string{"bogon"}
		testStringFlag(t, "f", get, goodValues, badValues)
	})
	t.Run("log file", func(t *testing.T) {
		get := func(o *Options) string { return o.LogFile }
		goodValues := []string{"file.log"}
		testStringFlag(t, "log-file", get, goodValues, []string{})
	})
}
*/
