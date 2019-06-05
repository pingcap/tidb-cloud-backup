// Copyright 2018 The Go Cloud Development Kit Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package fileblob provides a blob implementation that uses the filesystem.
// Use OpenBucket to construct a *blob.Bucket.
//
// Open URLs
//
// For blob.Open URLs, fileblob registers for the scheme "file"; URLs start
// with "file://" like "file:///path/to/directory". For full details, see
// URLOpener.
//
// Escaping
//
// Go CDK supports all UTF-8 strings; to make this work with providers lacking
// full UTF-8 support, strings must be escaped (during writes) and unescaped
// (during reads). The following escapes are performed for fileblob:
//  - Blob keys: ASCII characters 0-31 are escaped to "__0x<hex>__".
//    If os.PathSeparator != "/", it is also escaped.
//    Additionally, the "/" in "../", the trailing "/" in "//", and a trailing
//    "/" is key names are escaped in the same way.
//
// As
//
// fileblob exposes the following types for As:
//  - Error: *os.PathError
package pkg

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"gocloud.dev/blob"
	"gocloud.dev/blob/driver"
	"gocloud.dev/gcerrors"
)

// Options sets options for constructing a *blob.Bucket backed by fileblob.
type Options struct{}

type bucket struct {
	dir string
}

// openBucket creates a driver.Bucket that reads and writes to dir.
// dir must exist.
func openBucket(dir string, _ *Options) (driver.Bucket, error) {
	dir = filepath.Clean(dir)
	info, err := os.Stat(dir)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", dir)
	}
	return &bucket{dir}, nil
}

// OpenBucket creates a *blob.Bucket backed by the filesystem and rooted at
// dir, which must exist. See the package documentation for an example.
func OpenBucket(dir string, opts *Options) (*blob.Bucket, error) {
	drv, err := openBucket(dir, opts)
	if err != nil {
		return nil, err
	}
	return blob.NewBucket(drv), nil
}

var errNotImplemented = errors.New("not implemented")

func (b *bucket) ErrorCode(err error) gcerrors.ErrorCode {
	switch {
	case os.IsNotExist(err):
		return gcerrors.NotFound
	case err == errNotImplemented:
		return gcerrors.Unimplemented
	default:
		return gcerrors.Unknown
	}
}

// path returns the full path for a key
func (b *bucket) path(key string) string {
	path := filepath.Join(b.dir, key)
	return path
}

// ListPaged implements driver.ListPaged.
func (b *bucket) ListPaged(ctx context.Context, opts *driver.ListOptions) (*driver.ListPage, error) {
	return nil, nil
}

// As implements driver.As.
func (b *bucket) As(i interface{}) bool { return false }

// As implements driver.ErrorAs.
func (b *bucket) ErrorAs(err error, i interface{}) bool {
	if perr, ok := err.(*os.PathError); ok {
		if p, ok := i.(**os.PathError); ok {
			*p = perr
			return true
		}
	}
	return false
}

// Attributes implements driver.Attributes.
func (b *bucket) Attributes(ctx context.Context, key string) (driver.Attributes, error) {
	return driver.Attributes{}, nil
}

// NewRangeReader implements driver.NewRangeReader.
func (b *bucket) NewRangeReader(ctx context.Context, key string, offset, length int64, opts *driver.ReaderOptions) (driver.Reader, error) {
	return &reader{}, nil
}

type reader struct {
	r     io.Reader
	c     io.Closer
	attrs driver.ReaderAttributes
}

func (r *reader) Read(p []byte) (int, error) {
	if r.r == nil {
		return 0, io.EOF
	}
	return r.r.Read(p)
}

func (r *reader) Close() error {
	if r.c == nil {
		return nil
	}
	return r.c.Close()
}

func (r *reader) Attributes() driver.ReaderAttributes {
	return r.attrs
}

func (r *reader) As(i interface{}) bool { return false }

// NewTypedWriter implements driver.NewTypedWriter.
func (b *bucket) NewTypedWriter(ctx context.Context, key string, contentType string, opts *driver.WriterOptions) (driver.Writer, error) {
	path := b.path(key)
	if err := os.MkdirAll(filepath.Dir(path), 0777); err != nil {
		return nil, err
	}
	f, err := ioutil.TempFile(filepath.Dir(path), "fileblob")
	if err != nil {
		return nil, err
	}
	if opts.BeforeWrite != nil {
		if err := opts.BeforeWrite(func(interface{}) bool { return false }); err != nil {
			return nil, err
		}
	}
	w := &writer{
		ctx:        ctx,
		f:          f,
		path:       path,
		contentMD5: opts.ContentMD5,
		md5hash:    md5.New(),
	}
	return w, nil
}

type writer struct {
	ctx        context.Context
	f          *os.File
	path       string
	contentMD5 []byte
	// We compute the MD5 hash so that we can store it with the file attributes,
	// not for verification.
	md5hash hash.Hash
}

func (w *writer) Write(p []byte) (n int, err error) {
	if _, err := w.md5hash.Write(p); err != nil {
		return 0, err
	}
	return w.f.Write(p)
}

func (w *writer) Close() error {
	err := w.f.Close()
	if err != nil {
		return err
	}
	// Always delete the temp file. On success, it will have been renamed so
	// the Remove will fail.
	defer func() {
		_ = os.Remove(w.f.Name())
	}()

	// Check if the write was cancelled.
	if err := w.ctx.Err(); err != nil {
		return err
	}

	// Rename the temp file to path.
	if err := os.Rename(w.f.Name(), w.path); err != nil {
		return err
	}
	return nil
}

// Delete implements driver.Delete.
func (b *bucket) Delete(ctx context.Context, key string) error {
	path := b.path(key)
	err := os.Remove(path)
	if err != nil {
		return err
	}
	return nil
}

func (b *bucket) SignedURL(ctx context.Context, key string, opts *driver.SignedURLOptions) (string, error) {
	return "", errNotImplemented
}
