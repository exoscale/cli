package cmd

import (
	"errors"
	"testing"
	"time"
)

func sosSyncGetTestConfiguration() sosSyncConfiguration {
	return sosSyncConfiguration{
		RemoveDeleted:   true,
		SourceDirectory: "/",
		TargetBucket:    "test",
		TargetPath:      "",
	}
}

func assertNoErrors(errors chan error, t *testing.T) {
	select {
	case err := <-errors:
		t.Errorf("Unexpected error: %s", err)
	default:
	}

}

//region Tests
func TestSosSyncDiffEmpty(t *testing.T) {
	var objectList []sosSyncObject
	var fileList []sosSyncFile

	done := make(chan bool, 1)
	errorChannel := make(chan error)

	syncDiffChannel := newSosSyncDiff(
		newSosSyncTestListObjects(objectList),
		newSosSyncTestGetFile(fileList),
		newSosSyncTestListFiles(fileList),
	)(sosSyncGetTestConfiguration(), done, errorChannel)

	assertNoErrors(errorChannel, t)
	if len(syncDiffChannel) != 0 {
		t.Errorf("Diffing an empty local directory and empty bucket resulted in %d items in a diff.", len(syncDiffChannel))
	}
}

func TestSosSyncDiffOneRemote(t *testing.T) {
	var objectList []sosSyncObject
	var fileList []sosSyncFile

	done := make(chan bool, 1)
	errorChannel := make(chan error)

	objectList = append(objectList, sosSyncObject{
		Key:          "test.txt",
		LastModified: time.Now(),
		ContentType:  "text/plain",
		Size:         0,
	})

	syncDiffChannel := newSosSyncDiff(
		newSosSyncTestListObjects(objectList),
		newSosSyncTestGetFile(fileList),
		newSosSyncTestListFiles(fileList),
	)(sosSyncGetTestConfiguration(), done, errorChannel)

	syncDiffFile := sosSyncTestTaskChannelToSlice(syncDiffChannel)

	assertNoErrors(errorChannel, t)
	if len(syncDiffFile) != 1 {
		t.Errorf("Diffing an empty local directory and bucket with one file resulted in %d items in a diff.", len(syncDiffFile))
	} else {
		if syncDiffFile[0].Action != sosSyncDeleteAction {
			t.Errorf("Diffing an empty local directory and bucket with one file in a non-delete action.")
		}
		if syncDiffFile[0].File != "test.txt" {
			t.Errorf("Diffing an empty local directory and bucket with one file resulted in the wrong file: %s", syncDiffFile[0].File)
		}
	}
}

func TestSosSyncDiffOneLocal(t *testing.T) {
	var objectList []sosSyncObject
	var fileList []sosSyncFile

	done := make(chan bool, 1)
	errorChannel := make(chan error)

	fileList = append(fileList, sosSyncFile{
		Path:         "test.txt",
		LastModified: time.Now(),
		Size:         0,
	})

	syncDiffChannel := newSosSyncDiff(
		newSosSyncTestListObjects(objectList),
		newSosSyncTestGetFile(fileList),
		newSosSyncTestListFiles(fileList),
	)(sosSyncGetTestConfiguration(), done, errorChannel)

	syncDiffFile := sosSyncTestTaskChannelToSlice(syncDiffChannel)

	assertNoErrors(errorChannel, t)
	if len(syncDiffFile) != 1 {
		t.Errorf("Diffing an empty bucket and a directory with one file resulted in %d items in a diff.", len(syncDiffFile))
	} else {
		if syncDiffFile[0].Action != sosSyncUploadAction {
			t.Errorf("Diffing an empty bucket and a directory with one file resulted in a non-upload action.")
		}
		if syncDiffFile[0].File != "test.txt" {
			t.Errorf("Diffing an empty bucket and a directory with one file resulted in the wrong file: %s", syncDiffFile[0].File)
		}
	}
}

func TestSosSyncDiffSameTimestampsAndSize(t *testing.T) {
	var objectList []sosSyncObject
	var fileList []sosSyncFile

	done := make(chan bool, 1)
	errorChannel := make(chan error)

	modified := time.Now()

	fileList = append(fileList, sosSyncFile{
		Path:         "test.txt",
		LastModified: modified,
		Size:         0,
	})

	objectList = append(objectList, sosSyncObject{
		Key:          "test.txt",
		LastModified: modified,
		ContentType:  "text/plain",
		Size:         0,
	})

	syncDiffChannel := newSosSyncDiff(
		newSosSyncTestListObjects(objectList),
		newSosSyncTestGetFile(fileList),
		newSosSyncTestListFiles(fileList),
	)(sosSyncGetTestConfiguration(), done, errorChannel)

	syncDiffFile := sosSyncTestTaskChannelToSlice(syncDiffChannel)

	assertNoErrors(errorChannel, t)
	if len(syncDiffFile) != 0 {
		t.Errorf("Diffing an local directory and bucket with one file resulted in %d items in a diff.", len(syncDiffFile))
	}
}

func TestSosSyncDiffLocalFileInThePast(t *testing.T) {
	var objectList []sosSyncObject
	var fileList []sosSyncFile

	done := make(chan bool, 1)
	errorChannel := make(chan error)

	modified := time.Now()

	fileList = append(fileList, sosSyncFile{
		Path:         "test.txt",
		LastModified: modified.AddDate(0, 0, -1),
		Size:         0,
	})

	objectList = append(objectList, sosSyncObject{
		Key:          "test.txt",
		LastModified: modified,
		ContentType:  "text/plain",
		Size:         0,
	})

	syncDiffChannel := newSosSyncDiff(
		newSosSyncTestListObjects(objectList),
		newSosSyncTestGetFile(fileList),
		newSosSyncTestListFiles(fileList),
	)(sosSyncGetTestConfiguration(), done, errorChannel)

	syncDiffFile := sosSyncTestTaskChannelToSlice(syncDiffChannel)

	assertNoErrors(errorChannel, t)
	if len(syncDiffFile) != 0 {
		t.Errorf("Diffing an local directory and bucket with one file, local file in the past resulted in %d items in a diff.", len(syncDiffFile))
	}
}

func TestSosSyncDiffRemoteFileInThePast(t *testing.T) {
	var objectList []sosSyncObject
	var fileList []sosSyncFile

	done := make(chan bool, 1)
	errorChannel := make(chan error)

	modified := time.Now()

	fileList = append(fileList, sosSyncFile{
		Path:         "test.txt",
		LastModified: modified,
		Size:         0,
	})

	objectList = append(objectList, sosSyncObject{
		Key:          "test.txt",
		LastModified: modified.AddDate(0, 0, -1),
		ContentType:  "text/plain",
		Size:         0,
	})

	syncDiffChannel := newSosSyncDiff(
		newSosSyncTestListObjects(objectList),
		newSosSyncTestGetFile(fileList),
		newSosSyncTestListFiles(fileList),
	)(sosSyncGetTestConfiguration(), done, errorChannel)

	syncDiffFile := sosSyncTestTaskChannelToSlice(syncDiffChannel)

	assertNoErrors(errorChannel, t)
	if len(syncDiffFile) != 1 {
		t.Errorf("Diffing an local directory and bucket with one file, remote file in the past resulted in %d items in a diff.", len(syncDiffFile))
	} else {
		if syncDiffFile[0].Action != sosSyncUploadAction {
			t.Errorf("Diffing an local directory and bucket with one file, remote file in the past resulted in a non-upload action.")
		}
		if syncDiffFile[0].File != "test.txt" {
			t.Errorf("Diffing an local directory and bucket with one file, remote file in the past resulted in the wrong file: %s", syncDiffFile[0].File)
		}
	}
}

func TestSosSyncDiffDifferingFileSize(t *testing.T) {
	var objectList []sosSyncObject
	var fileList []sosSyncFile

	done := make(chan bool, 1)
	errorChannel := make(chan error)

	modified := time.Now()

	fileList = append(fileList, sosSyncFile{
		Path:         "test.txt",
		LastModified: modified,
		Size:         0,
	})

	objectList = append(objectList, sosSyncObject{
		Key:          "test.txt",
		LastModified: modified,
		ContentType:  "text/plain",
		Size:         1,
	})

	syncDiffChannel := newSosSyncDiff(
		newSosSyncTestListObjects(objectList),
		newSosSyncTestGetFile(fileList),
		newSosSyncTestListFiles(fileList),
	)(sosSyncGetTestConfiguration(), done, errorChannel)

	syncDiffFile := sosSyncTestTaskChannelToSlice(syncDiffChannel)

	assertNoErrors(errorChannel, t)
	if len(syncDiffFile) != 1 {
		t.Errorf("Diffing an local directory and bucket with one file with differing file size resulted in %d items in a diff.", len(syncDiffFile))
	} else {
		if syncDiffFile[0].Action != sosSyncUploadAction {
			t.Errorf("Diffing an local directory and bucket with one file with differing file size resulted in a non-upload action.")
		}
		if syncDiffFile[0].File != "test.txt" {
			t.Errorf("Diffing an local directory and bucket with one file with differing file size resulted in the wrong file: %s", syncDiffFile[0].File)
		}
	}
}

//endregion

//region Mocks
func newSosSyncTestListObjects(objects []sosSyncObject) sosSyncListObjects {
	return func(config sosSyncConfiguration, errors chan<- error) <-chan sosSyncObject {
		result := make(chan sosSyncObject)

		go func() {
			defer close(result)

			for _, object := range objects {
				result <- object
			}
		}()

		return result
	}
}

func newSosSyncTestListFiles(objects []sosSyncFile) sosSyncListFiles {
	return func(config sosSyncConfiguration, errors chan<- error) <-chan sosSyncFile {
		result := make(chan sosSyncFile)

		go func() {
			defer close(result)
			for _, object := range objects {
				result <- object
			}
		}()

		return result
	}
}

func newSosSyncTestGetFile(objects []sosSyncFile) sosSyncGetFile {
	return func(config sosSyncConfiguration, file string) (sosSyncFile, error) {
		for _, object := range objects {
			if object.Path == file {
				return object, nil
			}
		}
		return sosSyncFile{}, errors.New("File not found: " + file)
	}
}

//endregion

//region Utilities
func sosSyncTestTaskChannelToSlice(ch <-chan sosSyncTask) []sosSyncTask {
	var result []sosSyncTask
	for object := range ch {
		result = append(result, object)
	}
	return result
}

//endregion
