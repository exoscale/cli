package sos_sync

import (
	"context"
	"io"
	"time"
)

type File struct {
	Path         string
	LastModified time.Time
	Size         int64
}

type FileStorage interface {
	List(errors chan<- error) <-chan File
	Get(file string) (File, error)
	SetModified(file string, timestamp time.Time) error
	Read(ctx context.Context, file string) (io.ReadCloser, error)
	Write(ctx context.Context, file string, reader io.Reader) error
	Delete(file string) error
}
