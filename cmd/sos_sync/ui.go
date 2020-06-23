package sos_sync

import (
	"context"
	"io"
	"sync"
)

type Ui interface {
	AddTask(task Task) FileUi
}

type UiFactory interface {
	Make(wg *sync.WaitGroup, ctx context.Context) Ui
}

type FileUi struct {
	// Return a reader pipe that allows the UI to track how much data is being transferred
	GetReader func(r io.ReadCloser) io.ReadCloser
	// Callback when an upload error happens.
	OnError func()
	// Callback when the upload is complete.
	OnComplete func()
}
