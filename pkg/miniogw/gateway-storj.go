// Copyright (C) 2018 Storj Labs, Inc.
// See LICENSE for copying information.

package storj

import (
	"context"
	"io"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/minio/cli"
	minio "github.com/minio/minio/cmd"
	"github.com/minio/minio/pkg/auth"
	"github.com/minio/minio/pkg/hash"
	"github.com/zeebo/errs"
	monkit "gopkg.in/spacemonkeygo/monkit.v2"

	"storj.io/storj/pkg/paths"
	"storj.io/storj/pkg/storage/objects"
	"storj.io/storj/pkg/storage/streams"
	mpb "storj.io/storj/protos/objects"
)

var (
	mon = monkit.Package()
	//Error is the errs class of standard End User Client errors
	Error = errs.Class("Storj Gateway error")
)

func init() {
	minio.RegisterGatewayCommand(cli.Command{
		Name:            "storj",
		Usage:           "Storj",
		Action:          storjGatewayMain,
		HideHelpCommand: true,
	})
}

func storjGatewayMain(ctx *cli.Context) {
	s := &Storj{os: mockObjectStore()}
	minio.StartGateway(ctx, s)
}

func mockObjectStore() objects.ObjStore {
	//	return &objects.Objects{}
	//streamStore := streams.Streams{}
	//return &objects.ObjStore{s: streams.New}
	return objects.NewStore(&streams.StreamStore{})
}

// Storj is the implementation of a minio cmd.Gateway
type Storj struct {
	os objects.ObjStore
}

// Name implements cmd.Gateway
func (s *Storj) Name() string {
	return "storj"
}

// NewGatewayLayer implements cmd.Gateway
func (s *Storj) NewGatewayLayer(creds auth.Credentials) (
	minio.ObjectLayer, error) {
	return &storjObjects{storj: s}, nil
}

// Production implements cmd.Gateway
func (s *Storj) Production() bool {
	return false
}

type storjObjects struct {
	minio.GatewayUnsupported
	storj *Storj
}

func (s *storjObjects) DeleteBucket(ctx context.Context, bucket string) (err error) {
	defer mon.Task()(&ctx)(&err)
	panic("TODO")
}

func (s *storjObjects) DeleteObject(ctx context.Context, bucket, object string) (err error) {
	defer mon.Task()(&ctx)(&err)
	objpath := paths.New(bucket, object)
	return s.storj.os.Delete(ctx, objpath)
}

func (s *storjObjects) GetBucketInfo(ctx context.Context, bucket string) (
	bucketInfo minio.BucketInfo, err error) {
	defer mon.Task()(&ctx)(&err)
	panic("TODO")
}

func (s *storjObjects) GetObject(ctx context.Context, bucket, object string,
	startOffset int64, length int64, writer io.Writer, etag string) (err error) {
	defer mon.Task()(&ctx)(&err)
	objpath := paths.New(bucket, object)
	rr, _, err := s.storj.os.Get(ctx, objpath)
	if err != nil {
		return err
	}
	defer rr.Close()
	r, err := rr.Range(ctx, startOffset, length)
	if err != nil {
		return err
	}
	defer r.Close()
	_, err = io.Copy(writer, r)
	return err
}

func (s *storjObjects) GetObjectInfo(ctx context.Context, bucket,
	object string) (objInfo minio.ObjectInfo, err error) {
	defer mon.Task()(&ctx)(&err)
	objPath := paths.New(bucket, object)
	rr, m, err := s.storj.os.Get(ctx, objPath)
	if err != nil {
		return objInfo, err
	}
	defer rr.Close()
	newmetainfo := &mpb.StorjMetaInfo{}
	err = proto.Unmarshal(m.Data, newmetainfo)
	if err != nil {
		return objInfo, err
	}
	return minio.ObjectInfo{
		Name:    newmetainfo.GetName(),
		Bucket:  newmetainfo.GetBucket(),
		ModTime: m.Modified,
		Size:    newmetainfo.GetSize(),
		ETag:    newmetainfo.GetETag(),
	}, err
}

func (s *storjObjects) ListBuckets(ctx context.Context) (
	buckets []minio.BucketInfo, err error) {
	defer mon.Task()(&ctx)(&err)
	buckets = nil
	err = nil
	return buckets, err
}

func (s *storjObjects) ListObjects(ctx context.Context, bucket, prefix, marker,
	delimiter string, maxKeys int) (result minio.ListObjectsInfo, err error) {
	defer mon.Task()(&ctx)(&err)

	//TODO: Fix parameters
	_, _, err = s.storj.os.List(ctx, paths.New(prefix, ""), paths.New(marker, ""), paths.New(delimiter, ""), true, int(100), uint64(0))
	if err != nil {
		return result, err
	}
	// TODO: Fill the result from the return of the List()
	result = minio.ListObjectsInfo{}

	err = nil
	return result, err
}

func (s *storjObjects) MakeBucketWithLocation(ctx context.Context,
	bucket string, location string) (err error) {
	defer mon.Task()(&ctx)(&err)
	return nil
}

func (s *storjObjects) PutObject(ctx context.Context, bucket, object string,
	data *hash.Reader, metadata map[string]string) (objInfo minio.ObjectInfo,
	err error) {
	defer mon.Task()(&ctx)(&err)
	//metadata serialized
	serMetaInfo := &mpb.StorjMetaInfo{
		ContentType: metadata["content-type"],
		Bucket:      bucket,
		Name:        object,
	}
	metainfo, err := proto.Marshal(serMetaInfo)
	if err != nil {
		return objInfo, err
	}
	objPath := paths.New(bucket, object)
	// setting zero value means the object never expires
	expTime := time.Time{}
	m, err := s.storj.os.Put(ctx, objPath, data, metainfo, expTime)
	return minio.ObjectInfo{
		Name:   object,
		Bucket: bucket,
		// TODO create a followup ticket in JIRA to fix ModTime
		ModTime: m.Modified,
		Size:    data.Size(),
		ETag:    minio.GenETag(),
	}, err
}

func (s *storjObjects) Shutdown(ctx context.Context) (err error) {
	defer mon.Task()(&ctx)(&err)
	panic("TODO")
}

func (s *storjObjects) StorageInfo(context.Context) minio.StorageInfo {
	return minio.StorageInfo{}
}
