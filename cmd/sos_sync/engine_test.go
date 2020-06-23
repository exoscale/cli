package sos_sync_test

import (
	"bytes"
	"context"
	"github.com/exoscale/cli/cmd/sos_sync"
	"log"
	"testing"
)

// region Helpers
func assertNoErrors(t *testing.T, errors chan error) {
	select {
	case err := <-errors:
		t.Errorf("Unexpected onError: %s", err)
	default:
	}
}

func getFileList(t *testing.T, fileStorage sos_sync.FileStorage) []sos_sync.File {
	errorChannel := make(chan error)
	files := fileStorage.List(errorChannel)
	assertNoErrors(t, errorChannel)

	var fileList []sos_sync.File
	for file := range files {
		fileList = append(fileList, file)
	}
	return fileList
}

func getObjectList(t *testing.T, objectStorage sos_sync.ObjectStorage) []sos_sync.ObjectStorageObject {
	errorChannel := make(chan error)
	objects := objectStorage.List(errorChannel)

	var objectList []sos_sync.ObjectStorageObject
	for object := range objects {
		objectList = append(objectList, object)
	}
	assertNoErrors(t, errorChannel)
	return objectList
}

func createLocalFile(t *testing.T, fileStorage sos_sync.FileStorage, filename string, data []byte) {
	err := fileStorage.Write(context.Background(), filename, bytes.NewReader(data))
	if err != nil {
		t.Fatalf("an error happened while loading test data (%v)", err)
	}
}

func createObject(t *testing.T, objectStorage sos_sync.ObjectStorage, object string, data []byte) {
	err := objectStorage.Upload(context.Background(), object, bytes.NewReader(data), 0, "text/plain")
	if err != nil {
		t.Fatalf("an error happened while loading test data (%v)", err)
	}
}

func createSyncEngine() (sos_sync.FileStorage, sos_sync.ObjectStorage, *sos_sync.MemoryTaskLogger, sos_sync.SyncEngine) {
	fileStorage := sos_sync.NewMemoryFileStorage()
	objectStorage := sos_sync.NewMemoryObjectStorage()
	taskLogger := sos_sync.NewMemoryTaskLogger()
	uiFactory := sos_sync.NewLoggingUiFactory(taskLogger)

	engine := sos_sync.NewSyncEngine(
		uiFactory,
		objectStorage,
		fileStorage,
		1,
	)
	return fileStorage, objectStorage, taskLogger, engine
}

func synchronize(t *testing.T, engine sos_sync.SyncEngine, removeDeleted bool) {
	err := engine.Synchronize(context.Background(), removeDeleted)
	if err != nil {
		t.Fatalf("an error happened during synchronization (%v)", err)
	}
}

// endregion

// region Tests
func TestSosSyncDiffEmpty(t *testing.T) {
	fileStorage, objectStorage, _, engine := createSyncEngine()

	err := engine.Synchronize(context.Background(), false)
	if err != nil {
		t.Fatalf("an error happened during synchronization (%v)", err)
	}

	files := getFileList(t, fileStorage)
	if len(files) != 0 {
		t.Fatalf("after doing an empty sync %d files were found in the local storage", len(files))
	}

	objects := getObjectList(t, objectStorage)
	if len(objects) != 0 {
		t.Fatalf("after doing an empty sync %d objects were found in the object storage", len(files))
	}
}

func TestSosSyncDiffOneRemote(t *testing.T) {
	fileStorage, objectStorage, _, engine := createSyncEngine()

	var data []byte
	createObject(t, objectStorage, "test.txt", data)

	synchronize(t, engine, false)

	files := getFileList(t, fileStorage)
	if len(files) != 0 {
		t.Fatalf("after doing an empty sync to a non-empty object storage %d files were found in the local storage instead of 0", len(files))
	}

	objects := getObjectList(t, objectStorage)
	if len(objects) != 1 {
		t.Fatalf("after doing an empty sync to a non-empty object storage with removeDeleted=false %d objects were found in the object storage instead of 1", len(files))
	}
}

func TestSosSyncDiffOneRemoteRemoveDeleted(t *testing.T) {
	fileStorage, objectStorage, _, engine := createSyncEngine()

	var data []byte

	createObject(t, objectStorage, "test.txt", data)

	synchronize(t, engine, true)

	files := getFileList(t, fileStorage)
	if len(files) != 0 {
		t.Fatalf("after doing an empty sync to a non-empty object storage %d files were found in the local storage instead of 0", len(files))
	}

	objects := getObjectList(t, objectStorage)
	if len(objects) != 0 {
		t.Fatalf("after doing an empty sync to a non-empty object storage %d objects were found in the object storage instead of 0", len(files))
	}
}

func TestSosSyncDiffOneLocal(t *testing.T) {
	fileStorage, objectStorage, _, engine := createSyncEngine()

	var data []byte
	createLocalFile(t, fileStorage, "test.txt", data)

	synchronize(t, engine, false)

	files := getFileList(t, fileStorage)
	if len(files) != 1 {
		t.Fatalf("after doing a single file sync %d files were found in the local storage instead of 1", len(files))
	}
	if files[0].Path != "test.txt" {
		t.Fatalf("after doing a single file sync the file in the local directory was %s instead of test.txt", files[0].Path)
	}
	if files[0].Size != 0 {
		t.Fatalf("after doing a single file sync the file in the local directory was %d bytes instead of 0", files[0].Size)
	}

	objects := getObjectList(t, objectStorage)
	if len(objects) != 1 {
		t.Fatalf("after doing a single file sync %d objects were found in the object storage instead of 1", len(files))
	}
	if objects[0].Key != "test.txt" {
		t.Fatalf("after doing a single file sync the file in the object storage was %s instead of test.txt", files[0].Path)
	}
	if objects[0].Size != 0 {
		t.Fatalf("after doing a single file sync the file in the object storage was %d bytes instead of 0", files[0].Size)
	}
}

func TestSosSyncDiffSameTimestampsAndSize(t *testing.T) {
	fileStorage, objectStorage, taskLogger, engine := createSyncEngine()

	var data []byte
	createLocalFile(t, fileStorage, "test.txt", data)
	createObject(t, objectStorage, "test.txt", data)
	objectList := getObjectList(t, objectStorage)
	err := fileStorage.SetModified("test.txt", objectList[0].LastModified)
	if err != nil {
		log.Fatalf("error while setting modified time (%v)", err)
	}

	synchronize(t, engine, true)

	if len(taskLogger.Tasks) != 0 {
		log.Fatalf("more than 0 tasks despite no difference")
	}
}

func TestSosSyncDiffLocalFileInThePast(t *testing.T) {
	fileStorage, objectStorage, taskLogger, engine := createSyncEngine()

	var data []byte
	createLocalFile(t, fileStorage, "test.txt", data)
	createObject(t, objectStorage, "test.txt", data)
	objectList := getObjectList(t, objectStorage)
	err := fileStorage.SetModified("test.txt", objectList[0].LastModified.AddDate(0, 0, -1))
	if err != nil {
		log.Fatalf("error while setting modified time (%v)", err)
	}

	synchronize(t, engine, true)

	if len(taskLogger.Tasks) != 0 {
		log.Fatalf("more than 0 tasks with local file in the past")
	}
}

func TestSosSyncDiffRemoteFileInThePast(t *testing.T) {
	fileStorage, objectStorage, taskLogger, engine := createSyncEngine()

	var data []byte
	createLocalFile(t, fileStorage, "test.txt", data)
	createObject(t, objectStorage, "test.txt", data)
	objectList := getObjectList(t, objectStorage)
	err := fileStorage.SetModified("test.txt", objectList[0].LastModified.AddDate(0, 0, 1))
	if err != nil {
		log.Fatalf("error while setting modified time (%v)", err)
	}

	synchronize(t, engine, true)

	if len(taskLogger.Tasks) != 1 {
		log.Fatalf("unexpected number of sync tasks (%d) for remote file in the past", len(taskLogger.Tasks))
	}
	if taskLogger.Tasks[0].Action != sos_sync.UploadAction {
		log.Fatalf("incorrect action (delete) for remote file in the past")
	}
}

func TestSosSyncDiffDifferingFileSize(t *testing.T) {
	fileStorage, objectStorage, taskLogger, engine := createSyncEngine()

	var data []byte
	createLocalFile(t, fileStorage, "test.txt", []byte("Hello world!"))
	createObject(t, objectStorage, "test.txt", data)
	objectList := getObjectList(t, objectStorage)
	err := fileStorage.SetModified("test.txt", objectList[0].LastModified)
	if err != nil {
		log.Fatalf("error while setting modified time (%v)", err)
	}

	synchronize(t, engine, true)

	if len(taskLogger.Tasks) != 1 {
		log.Fatalf("unexpected number of sync tasks (%d) for file size difference", len(taskLogger.Tasks))
	}
	if taskLogger.Tasks[0].Action != sos_sync.UploadAction {
		log.Fatalf("incorrect action (delete) for different file size")
	}
}

//endregion
