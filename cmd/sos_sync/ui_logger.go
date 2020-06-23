package sos_sync

import (
	"context"
	"io"
	"sync"
)

// region Ui

// This UI does nothing apart from logging the files and actions that were performed
type LoggingUi struct {
	taskLogger TaskLogger
}

func (ui *LoggingUi) AddTask(task Task) FileUi {
	ui.taskLogger.Log(task)

	return FileUi{
		GetReader: func(r io.ReadCloser) io.ReadCloser {
			return r
		},
		OnError:    func() {},
		OnComplete: func() {},
	}
}

// endregion

// region Factory
func NewLoggingUiFactory(taskLogger TaskLogger) *LoggingUiFactory {
	return &LoggingUiFactory{
		taskLogger: taskLogger,
	}
}

type LoggingUiFactory struct {
	taskLogger TaskLogger
}

func (factory *LoggingUiFactory) Make(wg *sync.WaitGroup, ctx context.Context) Ui {
	return &LoggingUi{
		taskLogger: factory.taskLogger,
	}
}

// endregion
