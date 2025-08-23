// SPDX-License-Identifier: Zlib
// Copyright 2025, Terry M. Poulin.
package main

import (
	"audio_converter/internal/ffmpeg"
	"audio_converter/internal/filesystem"
	"audio_converter/internal/logging"
	"audio_converter/internal/options"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

type Exporter struct {
	ctx     context.Context
	opts    *options.ExporterOptions
	pool    *WorkPool
	InRoot  filesystem.FS
	OutRoot filesystem.FS
}

func newExporter(ctx context.Context, opts *options.ExporterOptions) *Exporter {
	return &Exporter{
		ctx:     ctx,
		opts:    opts,
		pool:    NewWorkPool(ctx, opts.MaxJobs, opts.MaxQueue),
		InRoot:  filesystem.NewFileSystem(opts.InRoot),
		OutRoot: filesystem.NewFileSystem(opts.OutRoot),
	}
}

// Make the magic happen, or return the error code.
func (p *Exporter) Run() error {
	// First execute WalkDir to ensure that all directories are created. This
	// will allow us to run the remaining tasks asyncronously without having
	// data races over "hey, I was just about to create that directory."
	if err := fs.WalkDir(p.InRoot, ".", p.visitDir); err != nil {
		return err
	}

	// Spin up the work pool.
	p.pool.Start()

	// Periodically log the status of the pool.
	go func() {
		for {
			time.Sleep(time.Second * 30)
			logging.Printf("WorkPool %p: size: %d limit: %d buffer: %d (%f %%)",
				p.pool, p.pool.Size(), p.pool.Limit(), p.pool.Remaining(), p.pool.PercentFull())
		}
	}()

	// Now execute WalkDir to feed the beast. This will block until all items are in the queue, which may require blocking until
	if err := fs.WalkDir(p.InRoot, ".", p.visitFile); err != nil {
		return err
	}

	// Now wait for everyone to finish.
	p.pool.Wait()

	return nil
}

// Walk function for creating directories in the output root.
//
// Called with <path> <base name of dir if its a dir> <err>
//
// E.g., if there is ${root}/artist/album/song, this would be called the
// following ways:
//
// - path=.                 dir=.
// - path=artist            dir=artest
// - path=artist/album      dir=album
// - path=artist/album/song dir=.
//
// return fs.SkipDir to skip the current directory (parent if called for a file)
//
// return fs.SkipAll to abort the entire walk.
//
// return any other non-nil error to abort and return the error code.
func (p *Exporter) visitDir(path string, d fs.DirEntry, err error) error {
	logging.Printf("Visiting path: %q d.Name: %q err: %v", path, d, err)

	if !d.IsDir() || path == "." {
		return nil
	}

	// We can't count on d.Type().Perm() to be populated by fs.WalkDir, we
	// need to do a stat of our own.
	st, err := p.InRoot.Stat(path)
	if err != nil {
		return fmt.Errorf("stat failed: %w", err)
	}
	logging.Printf("Mkdirs %q", path)
	return p.OutRoot.MkDirAll(path, st.Mode().Perm())
}

// Walk function for converting files.
//
// See walkDir for a description of how the parameters will be populated by fs.WalkDir().
//
// Since directories are created by the initial walk, we only need concern
// ourselves with valid files. These are either copied or converted as
// appropriate.
func (p *Exporter) visitFile(path string, d fs.DirEntry, err error) error {
	logging.Printf("Visiting path: %q d.Name: %q err: %v", path, d, err)

	// Handle exclusions.
	if d.IsDir() {
		// Created by the initial walk using visitDir().
		return nil
	} else if path == "." {
		// We don't care about the root itself.
		return nil
	} else if filesystem.IsTrashFile(path) {
		logging.Verbosef("Skipping %q", path)
		return nil
	}

	if ffmpeg.IsMediaFile(path) {
		// Add the conversion to the queue.
		p.pool.Add(func() {
			if output, err := p.Convert(path); err != nil {
				logging.Fatalf("!!! FATAL: %v !!!\n=== Start Output %q ===\n%s\n=== End Output %q ===\n", err, path, output, path)
			} else {
				logging.Printf("=== Start Output %q ===\n%s\n=== End Output %q ===\n", path, output, path)
			}
		})
	} else if p.opts.CopyUnknown && !d.IsDir() {
		// Add copying the file to the queue.
		p.pool.Add(func() {
			if err := p.Copy(path); err != nil {
				logging.Fatalln(err)
			}
		})
	}
	return nil
}

// Handle copying path between roots. If no clobber is set, we silently ignore
// the operation when it looks like the file exists.
func (p *Exporter) Copy(path string) error {
	if p.opts.NoClobber {
		if _, err := p.OutRoot.Stat(path); !errors.Is(err, os.ErrNotExist) {
			logging.Verbosef("Not clobbering %q", path)
			return nil
		}
	}
	logging.Verbosef("Copying %q to %q",
		filepath.Join(p.opts.InRoot, path),
		filepath.Join(p.opts.OutRoot, path))
	nb, err := filesystem.CopyFile(p.InRoot, path, p.OutRoot, path)
	logging.Printf("Copied %d bytes of %s", nb, path)
	return err
}

func (p *Exporter) Convert(path string) (string, error) {
	oldExt := filepath.Ext(path)
	newExt := "." + p.opts.Format

	if oldExt == newExt {
		logging.Println(path, "already in target format")
		return "", p.Copy(path)
	}

	// A shallow copy is sufficent for our purposes. We just need to update the input/output fields.
	copts := p.opts.ConverterOptions
	if copts.Err != nil {
		return "", copts.Err
	}
	copts.InputFile = filepath.Join(p.opts.InRoot, path)
	copts.OutputFile = filepath.Join(p.opts.OutRoot, path[:len(path)-len(oldExt)]) + newExt

	logging.Verbosef("Converting %q -> %q", copts.InputFile, copts.OutputFile)
	output, err := ffmpeg.ConvertInBackground(p.ctx, &copts)
	if err != nil {
		if output == nil {
			output = []byte{}
		}
		return string(output), fmt.Errorf("converting %q failed with error: %v", copts.InputFile, err)
	}
	return string(output), err
}
