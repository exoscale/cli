package sos_sync

import (
	"context"
	"io"
	"time"
)

type ObjectStorageObject struct {
	Key          string
	LastModified time.Time
	ContentType  string
	Size         int64
}

// TODO this may need to be genericized and extracted into a general purpose overlay so we can switch
//      S3 libraries
type ObjectStorage interface {
	List(errors chan<- error) <-chan ObjectStorageObject
	Upload(ctx context.Context, objectName string, reader io.Reader, objectSize int64, contentType string) error
	Download(ctx context.Context, objectName string, writer io.Writer) error
	Delete(objectName string) error
}
