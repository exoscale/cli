package sos_sync

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
)

type SyncEngine interface {
	Synchronize(
		ctx context.Context,
		removeDeleted bool,
	) error
}

type syncEngine struct {
	uiFactory     UiFactory
	objectStorage ObjectStorage
	fileStorage   FileStorage
	concurrency   int
}

func NewSyncEngine(
	uiFactory UiFactory,
	objectStorage ObjectStorage,
	fileStorage FileStorage,
	concurrency int,
) SyncEngine {
	return &syncEngine{
		uiFactory:     uiFactory,
		objectStorage: objectStorage,
		fileStorage:   fileStorage,
		concurrency:   concurrency,
	}
}

func (engine *syncEngine) Synchronize(
	ctx context.Context,
	removeDeleted bool,
) error {
	done := make(chan bool, 1)
	errorChannel := make(chan error)
	taskList := engine.getDiff(removeDeleted, done, errorChannel)
	engine.process(ctx, taskList, done, errorChannel)
	var err error = nil
	for {
		select {
		case errChannelResult := <-errorChannel:
			fmt.Fprintf(os.Stderr, "Error while processing: %s", errChannelResult)
			err = errChannelResult
		default:
			err = nil
		}
		if err == nil {
			break
		}
	}
	//todo logic error here
	if err != nil {
		return fmt.Errorf("one or more errors happened while processing your sync (%v)", err)
	} else {
		return nil
	}
}

func (engine *syncEngine) process(ctx context.Context, tasks <-chan Task, done chan bool, errors chan<- error) {
	var taskWG sync.WaitGroup
	workerSem := make(chan int, engine.concurrency)
	uploadUi := engine.uiFactory.Make(&taskWG, ctx)
	for task := range tasks {
		taskWG.Add(1)
		task := task
		go func() {
			defer taskWG.Done()
			workerSem <- 1
			defer func() {
				<-workerSem
			}()

			taskUi := uploadUi.AddTask(task)
			if task.Action == DeleteAction {
				err := engine.objectStorage.Delete(task.File)
				if err != nil {
					errors <- err
				}
			} else {
				file, err := engine.fileStorage.Get(task.File)
				if err != nil {
					errors <- err
				}
				var contentType string
				var bufferSize int64

				if task.Size >= 512 {
					bufferSize = 512
				} else {
					bufferSize = task.Size
				}
				reader, err := engine.fileStorage.Read(ctx, task.File)
				if err != nil {
					taskUi.OnError()
					return
				}
				defer reader.Close()

				buffer := make([]byte, bufferSize)
				_, err = reader.Read(buffer)
				contentType = http.DetectContentType(buffer)

				reader, err = engine.fileStorage.Read(ctx, task.File)
				if err != nil {
					taskUi.OnError()
					return
				}
				r := taskUi.GetReader(reader)
				err = engine.objectStorage.Upload(ctx, file.Path, r, file.Size, contentType)
				if err != nil {
					taskUi.OnError()
					return
				}
				taskUi.OnComplete()
			}

		}()
	}

	<-done
	close(done)
	taskWG.Wait()
}

func (engine *syncEngine) getDiff(removeDeleted bool, done chan bool, errors chan<- error) <-chan Task {
	result := make(chan Task)

	go func() {
		defer close(result)

		remoteObjectsIndexed := map[string]ObjectStorageObject{}
		remoteObjects := engine.objectStorage.List(errors)
		for remoteObject := range remoteObjects {
			_, err := engine.fileStorage.Get(remoteObject.Key)
			if err != nil {
				if removeDeleted {
					result <- Task{
						File:   remoteObject.Key,
						Action: DeleteAction,
					}
				}
			} else {
				remoteObjectsIndexed[remoteObject.Key] = remoteObject
			}
		}

		localFiles := engine.fileStorage.List(errors)
		for localFile := range localFiles {
			if _, ok := remoteObjectsIndexed[localFile.Path]; ok != true {
				result <- Task{
					File:   localFile.Path,
					Action: UploadAction,
					Size:   0,
				}
			} else {
				if remoteObjectsIndexed[localFile.Path].LastModified.Before(localFile.LastModified) ||
					remoteObjectsIndexed[localFile.Path].Size != localFile.Size {
					result <- Task{
						File:   localFile.Path,
						Action: UploadAction,
						Size:   localFile.Size,
					}
				}
			}
		}

		if done != nil {
			done <- true
		}
	}()
	return result
}
