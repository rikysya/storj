// Copyright (C) 2018 Storj Labs, Inc.
// See LICENSE for copying information.

package streams

import (
	"context"
	"io"
	"time"

	monkit "gopkg.in/spacemonkeygo/monkit.v2"
	"storj.io/storj/pkg/paths"
	"storj.io/storj/pkg/ranger"
	"storj.io/storj/pkg/storage"
)

var mon = monkit.Package()

//StreamStore structure
type StreamStore struct {
}

// Store for streams
type Store interface {
	Meta(ctx context.Context, path paths.Path) (storage.Meta, error)
	Get(ctx context.Context, path paths.Path) (ranger.RangeCloser,
		storage.Meta, error)
	Put(ctx context.Context, path paths.Path, data io.Reader, metadata []byte,
		expiration time.Time) (storage.Meta, error)
	Delete(ctx context.Context, path paths.Path) error
	List(ctx context.Context, prefix, startAfter, endBefore paths.Path,
		recursive bool, limit int, metaFlags uint64) (items []paths.Path,
		more bool, err error)
}

func (s *StreamStore) Meta(ctx context.Context, path paths.Path) (meta storage.Meta, err error) {
	defer mon.Task()(&ctx)(&err)
	return meta, err
}

func (s *StreamStore) Get(ctx context.Context, path paths.Path) (
	rr ranger.RangeCloser, meta storage.Meta, err error) {
	defer mon.Task()(&ctx)(&err)
	meta = storage.Meta{}
	rr = nil
	err = nil
	return rr, meta, err
}

func (s *StreamStore) Put(ctx context.Context, path paths.Path,
	data io.Reader, metadata []byte, expiration time.Time) (meta storage.Meta, err error) {
	defer mon.Task()(&ctx)(&err)
	meta = storage.Meta{}
	return meta, err
}

func (s *StreamStore) Delete(ctx context.Context, path paths.Path) (err error) {
	defer mon.Task()(&ctx)(&err)
	err = nil
	return err
}

func (s *StreamStore) List(ctx context.Context, prefix, startAfter, endBefore paths.Path,
	recursive bool, limit int, metaFlags uint64) (items []paths.Path, more bool, err error) {
	defer mon.Task()(&ctx)(&err)
	return items, more, err
}
