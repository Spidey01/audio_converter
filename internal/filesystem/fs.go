// SPDX-License-Identifier: Zlib
// Copyright 2025, Terry M. Poulin.
package filesystem

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

type FS interface {
	fs.FS
	fs.ReadDirFS
	fs.ReadFileFS
	fs.StatFS
	// fs.GlobFS
	// Resolves to the legit path. This is perhaps evil based on the _concept_ of Go's fs.FS, but implements

	// Create a file in the FS. Returns the handle as per os.Create().
	Create(name string) (fs.File, error)

	// Create a directory in the FS.
	MkDir(name string, mode fs.FileMode) error
	// Create a directory in the FS, recursively.
	MkDirAll(name string, mode fs.FileMode) error
}

// Implements our extended FS for the target OS.
type FileSystem struct {
	root string
}

func NewFileSystem(root string) *FileSystem {
	return &FileSystem{root: root}
}

// Open opens the named file.
// [File.Close] must be called to release any associated resources.
//
// When Open returns an error, it should be of type *PathError
// with the Op field set to "open", the Path field set to name,
// and the Err field describing the problem.
//
// Open should reject attempts to open names that do not satisfy
// ValidPath(name), returning a *PathError with Err set to
// ErrInvalid or ErrNotExist.
func (fsys *FileSystem) Open(name string) (fs.File, error) {
	if path, err := fsys.resolve(name); err != nil {
		return nil, err
	} else {
		return os.Open(path)
	}
}

// ReadDir reads the named directory
// and returns a list of directory entries sorted by filename.
func (fsys *FileSystem) ReadDir(name string) ([]fs.DirEntry, error) {
	if path, err := fsys.resolve(name); err != nil {
		return nil, err
	} else {
		return os.ReadDir(path)
	}
}

// ReadFile reads the named file and returns its contents.
// A successful call returns a nil error, not io.EOF.
// (Because ReadFile reads the whole file, the expected EOF
// from the final Read is not treated as an error to be reported.)
//
// The caller is permitted to modify the returned byte slice.
// This method should return a copy of the underlying data.
func (fsys *FileSystem) ReadFile(name string) ([]byte, error) {
	if path, err := fsys.resolve(name); err != nil {
		return nil, err
	} else {
		return os.ReadFile(path)
	}
}

// Stat returns a FileInfo describing the file.
// If there is an error, it should be of type *PathError.
func (fsys *FileSystem) Stat(name string) (fs.FileInfo, error) {
	if path, err := fsys.resolve(name); err != nil {
		return nil, err
	} else {
		return os.Stat(path)
	}
}

func (fsys *FileSystem) resolve(name string) (string, error) {
	if !fs.ValidPath(name) {
		return "", fs.ErrInvalid
	}
	return filepath.Join(fsys.root, name), nil
}

func (fsys *FileSystem) Create(name string) (fs.File, error) {
	if path, err := fsys.resolve(name); err != nil {
		return nil, err
	} else {
		return os.Create(path)
	}
}

func (fsys *FileSystem) MkDir(name string, mode fs.FileMode) error {
	if path, err := fsys.resolve(name); err != nil {
		return err
	} else {
		return os.Mkdir(path, mode)
	}
}

func (fsys *FileSystem) MkDirAll(name string, mode fs.FileMode) error {
	if path, err := fsys.resolve(name); err != nil {
		return err
	} else {
		return os.MkdirAll(path, mode)
	}
}

// Helper function that performs a copy between to filesystem.FS instances.
func CopyFile(srcFS FS, source string, dstFS FS, destination string) (int64, error) {
	src, err := srcFS.Open(source)
	if err != nil {
		return 0, err
	}
	defer src.Close()

	dst, err := dstFS.Create(destination)
	if err != nil {
		return 0, err
	}
	defer dst.Close()

	fp := dst.(*os.File)
	if fp == nil {
		return 0, fmt.Errorf("dstFS.Create did not return a pointer to an os.File")
	}
	return io.Copy(fp, src)
}
