package object

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

const (
	TimestampFormat = "2006-01-02 15:04:05 MST"
)

type Object struct {
	*types.Object
}

type ObjectVersion struct {
	*types.ObjectVersion
	VersionNumber uint64
}

type ObjectInterface interface {
	GetObject() *types.Object // TODO remove
	GetKey() *string
	GetSize() int64
	GetLastModified() *time.Time
	GetListObjectsItemOutput() *ListObjectsItemOutput
}

// Remove
func (o *Object) GetObject() *types.Object {
	return o.Object
}

// Remove
func (o *ObjectVersion) GetObject() *types.Object {
	return nil
}

type ObjectVersionInterface interface {
	ObjectInterface
	GetIsLatest() bool
	GetVersionId() *string
	SetVersionNumber(uint64)
	GetVersionNumber() uint64
}

func (o *Object) GetKey() *string {
	return o.Key
}

func (o *ObjectVersion) GetKey() *string {
	return o.Key
}

func (o *Object) GetSize() int64 {
	return o.Size
}

func (o *ObjectVersion) GetSize() int64 {
	return o.Size
}

func (o *Object) GetLastModified() *time.Time {
	return o.LastModified
}

func (o *ObjectVersion) GetLastModified() *time.Time {
	return o.LastModified
}

func (o *ObjectVersion) GetIsLatest() bool {
	return o.IsLatest
}

func (o *ObjectVersion) GetVersionId() *string {
	return o.VersionId
}

func getListObjectsItemOutput(o ObjectInterface) *ListObjectsItemOutput {
	return &ListObjectsItemOutput{
		Path:         aws.ToString(o.GetKey()),
		Size:         o.GetSize(),
		LastModified: o.GetLastModified().Format(TimestampFormat),
	}
}

func (o *Object) GetListObjectsItemOutput() *ListObjectsItemOutput {
	return getListObjectsItemOutput(o)
}

func (o *ObjectVersion) GetListObjectsItemOutput() *ListObjectsItemOutput {
	out := getListObjectsItemOutput(o)

	out.VersionId = o.GetVersionId()
	out.VersionNumber = &o.VersionNumber

	return out
}

func (o *ObjectVersion) SetVersionNumber(versionNumber uint64) {
	o.VersionNumber = versionNumber
}

func (o *ObjectVersion) GetVersionNumber() uint64 {
	return o.VersionNumber
}
