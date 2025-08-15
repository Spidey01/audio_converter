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
)

type Exporter struct {
	ctx     context.Context
	opts    *options.ExporterOptions
	InRoot  filesystem.FS
	OutRoot filesystem.FS
}

func newExporter(ctx context.Context, opts *options.ExporterOptions) *Exporter {
	return &Exporter{
		ctx:     ctx,
		opts:    opts,
		InRoot:  filesystem.NewFileSystem(opts.InRoot),
		OutRoot: filesystem.NewFileSystem(opts.OutRoot),
	}
}

// Make the magic happen, or return the error code.
func (p *Exporter) Run() error {
	return fs.WalkDir(p.InRoot, ".", p.walkFunc)
}

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
func (p *Exporter) walkFunc(path string, d fs.DirEntry, err error) error {
	logging.Verbosef("Visiting path: %q d.Name: %q err: %v", path, d, err)

	// Handle exclusions.
	if path == "." {
		// We don't care about the root itself.
		return nil
	} else if filesystem.IsTrashFile(path) {
		logging.Verbosef("Skipping %q", path)
		return nil
	}

	// Handle exports.
	if d.IsDir() {
		// We can't count on d.Type().Perm() to be populated by fs.WalkDir, we
		// need to do a stat. So no point in passing it here.
		return p.Mkdirs(path)
	} else if ffmpeg.IsMediaFile(path) {
		return p.Convert(path)
	} else if p.opts.CopyUnknown && !d.IsDir() {
		return p.Copy(path)
	}
	return nil
}

// Handle creating the directories defined by path.
func (p *Exporter) Mkdirs(path string) error {
	logging.Printf("Mkdirs %q", path)
	st, err := p.InRoot.Stat(path)
	if err != nil {
		return fmt.Errorf("stat failed: %w", err)
	}
	return p.OutRoot.MkDirAll(path, st.Mode().Perm())
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

func (p *Exporter) Convert(path string) error {
	oldExt := filepath.Ext(path)
	newExt := "." + p.opts.Format

	if oldExt == newExt {
		logging.Println(path, "already in target format")
		return p.Copy(path)
	}

	copts := ffmpeg.GetDefaultOptions(newExt)
	if copts.Err != nil {
		return copts.Err
	}
	copts.InputFile = filepath.Join(p.opts.InRoot, path)
	copts.OutputFile = filepath.Join(p.opts.OutRoot, path[:len(path)-len(oldExt)]) + newExt

	logging.Verbosef("Converting %q -> %q", copts.InputFile, copts.OutputFile)
	return ffmpeg.Convert(p.ctx, &copts)
}
