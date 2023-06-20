package object

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type Object struct {
	*types.Object
}

type ObjectVersion struct {
	*types.ObjectVersion
}

type ObjectInterface interface {
	GetKey() *string
	GetSize() int64
	GetLastModified() *time.Time
}

type ObjectVersionInterface interface {
	ObjectInterface
	GetIsLatest() bool
	GetVersionId() *string
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
