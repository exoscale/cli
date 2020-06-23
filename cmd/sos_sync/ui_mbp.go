package sos_sync

import (
	"context"
	"fmt"
	"github.com/vbauerster/mpb/v4"
	"github.com/vbauerster/mpb/v4/decor"
	"io"
	"sync"
	"time"
)

// region Ui
type MbpUi struct {
	wg       *sync.WaitGroup
	progress *mpb.Progress
}

func (ui *MbpUi) AddTask(task Task) FileUi {
	var filename string
	if task.Action == DeleteAction {
		filename = fmt.Sprintf("[D] %s", task.File)
	} else {
		filename = fmt.Sprintf("[U] %s", task.File)
	}
	bar := ui.progress.AddBar(task.Size,
		mpb.AppendDecorators(
			decor.Name(filename, decor.WC{W: len(filename) + 1, C: decor.DidentRight}),
		),
		mpb.PrependDecorators(
			decor.OnComplete(decor.AverageETA(decor.ET_STYLE_GO), "done!"),
			decor.Percentage(decor.WCSyncSpace),
		),
	)
	return FileUi{
		GetReader: func(r io.ReadCloser) io.ReadCloser {
			return bar.ProxyReader(r)
		},
		OnComplete: func() {
			bar.Completed()
			if task.Size == 0 {
				bar.SetTotal(100, true)
			}
		},
		OnError: func() {
			bar.Abort(false)
		},
	}
}

// endregion

// region Factory

// Create a new MBP UI factory
// @link https://github.com/vbauerster/mpb
func NewMbpUiFactory(quiet bool) UiFactory {
	return &MbpUiFactory{
		quiet: quiet,
	}
}

type MbpUiFactory struct {
	quiet bool
}

func (factory *MbpUiFactory) Make(wg *sync.WaitGroup, ctx context.Context) Ui {
	return &MbpUi{
		wg: wg,
		progress: mpb.NewWithContext(ctx,
			mpb.WithWaitGroup(wg),
			mpb.WithWidth(64),
			mpb.WithRefreshRate(180*time.Millisecond),
			mpb.ContainerOptOnCond(mpb.WithOutput(nil), func() bool { return factory.quiet }),
		),
	}
}

// endregion
