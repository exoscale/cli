package cmd

import (
	"errors"
	"testing"
	"time"
)

func SosSyncGetTestConfiguration() SosSyncConfiguration {
	return SosSyncConfiguration{
		RemoveDeleted: true,
		SourceDirectory: "/",
		TargetBucket: "test",
		TargetPath: "",
	}
}

//region Tests
func TestSosSyncDiffEmpty(t *testing.T) {
	var objectList []SosSyncObject
	var fileList []SosSyncFile

	syncDiffChannel := NewSosSyncDiff(
		NewSosSyncTestListObjects(objectList),
		NewSosSyncTestGetFile(fileList),
		NewSosSyncTestListFiles(fileList),
	)(SosSyncGetTestConfiguration())

	if len(syncDiffChannel) != 0 {
		t.Errorf("Diffing an empty local directory and empty bucket resulted in %d items in a diff.", len(syncDiffChannel))
	}
}

func TestSosSyncDiffOneRemote(t *testing.T) {
	var objectList []SosSyncObject
	var fileList []SosSyncFile

	objectList = append(objectList, SosSyncObject{
		Key:          "test.txt",
		LastModified: time.Now(),
		ContentType:  "text/plain",
		Size:         0,
	})

	syncDiffChannel := NewSosSyncDiff(
		NewSosSyncTestListObjects(objectList),
		NewSosSyncTestGetFile(fileList),
		NewSosSyncTestListFiles(fileList),
	)(SosSyncGetTestConfiguration())

	syncDiffFile := SosSyncTestTaskChannelToSlice(syncDiffChannel)

	if len(syncDiffFile) != 1 {
		t.Errorf("Diffing an empty local directory and bucket with one file resulted in %d items in a diff.", len(syncDiffFile))
	} else {
		if syncDiffFile[0].Action != SosSyncDeleteAction {
			t.Errorf("Diffing an empty local directory and bucket with one file in a non-delete action.")
		}
		if syncDiffFile[0].File != "test.txt" {
			t.Errorf("Diffing an empty local directory and bucket with one file resulted in the wrong file: %s", syncDiffFile[0].File)
		}
	}
}


func TestSosSyncDiffOneLocal(t *testing.T) {
	var objectList []SosSyncObject
	var fileList []SosSyncFile

	fileList = append(fileList, SosSyncFile{
		Path:          "test.txt",
		LastModified: time.Now(),
		ContentType:  "text/plain",
		Size:         0,
	})

	syncDiffChannel := NewSosSyncDiff(
		NewSosSyncTestListObjects(objectList),
		NewSosSyncTestGetFile(fileList),
		NewSosSyncTestListFiles(fileList),
	)(SosSyncGetTestConfiguration())

	syncDiffFile := SosSyncTestTaskChannelToSlice(syncDiffChannel)

	if len(syncDiffFile) != 1 {
		t.Errorf("Diffing an empty bucket and a directory with one file resulted in %d items in a diff.", len(syncDiffFile))
	} else {
		if syncDiffFile[0].Action != SosSyncUploadAction {
			t.Errorf("Diffing an empty bucket and a directory with one file resulted in a non-upload action.")
		}
		if syncDiffFile[0].File != "test.txt" {
			t.Errorf("Diffing an empty bucket and a directory with one file resulted in the wrong file: %s", syncDiffFile[0].File)
		}
	}
}

func TestSosSyncDiffSameTimestampsAndSize(t *testing.T) {
	var objectList []SosSyncObject
	var fileList []SosSyncFile

	modified := time.Now()

	fileList = append(fileList, SosSyncFile{
		Path:          "test.txt",
		LastModified: modified,
		ContentType:  "text/plain",
		Size:         0,
	})

	objectList = append(objectList, SosSyncObject{
		Key:          "test.txt",
		LastModified: modified,
		ContentType:  "text/plain",
		Size:         0,
	})

	syncDiffChannel := NewSosSyncDiff(
		NewSosSyncTestListObjects(objectList),
		NewSosSyncTestGetFile(fileList),
		NewSosSyncTestListFiles(fileList),
	)(SosSyncGetTestConfiguration())

	syncDiffFile := SosSyncTestTaskChannelToSlice(syncDiffChannel)

	if len(syncDiffFile) != 0 {
		t.Errorf("Diffing an local directory and bucket with one file resulted in %d items in a diff.", len(syncDiffFile))
	}
}

func TestSosSyncDiffLocalFileInThePast(t *testing.T) {
	var objectList []SosSyncObject
	var fileList []SosSyncFile

	modified := time.Now()

	fileList = append(fileList, SosSyncFile{
		Path:          "test.txt",
		LastModified: modified.AddDate(0, 0, -1),
		ContentType:  "text/plain",
		Size:         0,
	})

	objectList = append(objectList, SosSyncObject{
		Key:          "test.txt",
		LastModified: modified,
		ContentType:  "text/plain",
		Size:         0,
	})

	syncDiffChannel := NewSosSyncDiff(
		NewSosSyncTestListObjects(objectList),
		NewSosSyncTestGetFile(fileList),
		NewSosSyncTestListFiles(fileList),
	)(SosSyncGetTestConfiguration())

	syncDiffFile := SosSyncTestTaskChannelToSlice(syncDiffChannel)

	if len(syncDiffFile) != 0 {
		t.Errorf("Diffing an local directory and bucket with one file, local file in the past resulted in %d items in a diff.", len(syncDiffFile))
	}
}

func TestSosSyncDiffRemoteFileInThePast(t *testing.T) {
	var objectList []SosSyncObject
	var fileList []SosSyncFile

	modified := time.Now()

	fileList = append(fileList, SosSyncFile{
		Path:          "test.txt",
		LastModified: modified,
		ContentType:  "text/plain",
		Size:         0,
	})

	objectList = append(objectList, SosSyncObject{
		Key:          "test.txt",
		LastModified: modified.AddDate(0, 0, -1),
		ContentType:  "text/plain",
		Size:         0,
	})

	syncDiffChannel := NewSosSyncDiff(
		NewSosSyncTestListObjects(objectList),
		NewSosSyncTestGetFile(fileList),
		NewSosSyncTestListFiles(fileList),
	)(SosSyncGetTestConfiguration())

	syncDiffFile := SosSyncTestTaskChannelToSlice(syncDiffChannel)

	if len(syncDiffFile) != 1 {
		t.Errorf("Diffing an local directory and bucket with one file, remote file in the past resulted in %d items in a diff.", len(syncDiffFile))
	} else {
		if syncDiffFile[0].Action != SosSyncUploadAction {
			t.Errorf("Diffing an local directory and bucket with one file, remote file in the past resulted in a non-upload action.")
		}
		if syncDiffFile[0].File != "test.txt" {
			t.Errorf("Diffing an local directory and bucket with one file, remote file in the past resulted in the wrong file: %s", syncDiffFile[0].File)
		}
	}
}


func TestSosSyncDiffDifferingFileSize(t *testing.T) {
	var objectList []SosSyncObject
	var fileList []SosSyncFile

	modified := time.Now()

	fileList = append(fileList, SosSyncFile{
		Path:          "test.txt",
		LastModified: modified,
		Size:         0,
	})

	objectList = append(objectList, SosSyncObject{
		Key:          "test.txt",
		LastModified: modified,
		ContentType:  "text/plain",
		Size:         1,
	})

	syncDiffChannel := NewSosSyncDiff(
		NewSosSyncTestListObjects(objectList),
		NewSosSyncTestGetFile(fileList),
		NewSosSyncTestListFiles(fileList),
	)(SosSyncGetTestConfiguration())

	syncDiffFile := SosSyncTestTaskChannelToSlice(syncDiffChannel)

	if len(syncDiffFile) != 1 {
		t.Errorf("Diffing an local directory and bucket with one file with differing file size resulted in %d items in a diff.", len(syncDiffFile))
	} else {
		if syncDiffFile[0].Action != SosSyncUploadAction {
			t.Errorf("Diffing an local directory and bucket with one file with differing file size resulted in a non-upload action.")
		}
		if syncDiffFile[0].File != "test.txt" {
			t.Errorf("Diffing an local directory and bucket with one file with differing file size resulted in the wrong file: %s", syncDiffFile[0].File)
		}
	}
}
//endregion

//region Mocks
func NewSosSyncTestListObjects(objects []SosSyncObject) SosSyncListObjects {
	return func(config SosSyncConfiguration) <-chan SosSyncObject {
		result := make(chan SosSyncObject)

		go func() {
			for _, object := range objects {
				result <- object
			}
			close(result)
		}()

		return result
	}
}

func NewSosSyncTestListFiles(objects []SosSyncFile) SosSyncListFiles {
	return func(config SosSyncConfiguration) <-chan SosSyncFile {
		result := make(chan SosSyncFile)

		go func() {
			for _, object := range objects {
				result <- object
			}
			close(result)
		}()

		return result
	}
}

func NewSosSyncTestGetFile(objects []SosSyncFile) SosSyncGetFile {
	return func(config SosSyncConfiguration, file string) (SosSyncFile, error) {
		for _, object := range objects {
			if object.Path == file {
				return object, nil
			}
		}
		return SosSyncFile{}, errors.New("File not found: " + file)
	}
}
//endregion

//region Utilities
func SosSyncTestTaskChannelToSlice(ch <- chan SosSyncTask) []SosSyncTask {
	var result []SosSyncTask
	for object := range ch {
		result = append(result, object)
	}
	return result
}

//endregion