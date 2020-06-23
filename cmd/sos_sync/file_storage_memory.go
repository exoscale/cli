package sos_sync

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"time"
)

type memoryFileStorageFile struct {
	File

	data []byte
}

type MemoryFileStorage struct {
	files map[string]memoryFileStorageFile
}

func NewMemoryFileStorage() FileStorage {
	return &MemoryFileStorage{
		files: make(map[string]memoryFileStorageFile),
	}
}

func (memoryFileStorage *MemoryFileStorage) List(errors chan<- error) <-chan File {
	result := make(chan File)

	go func() {
		defer close(result)
		for _, file := range memoryFileStorage.files {
			result <- file.File
		}
	}()

	return result
}

func (memoryFileStorage *MemoryFileStorage) Get(file string) (File, error) {
	if val, ok := memoryFileStorage.files[file]; ok {
		return val.File, nil
	}
	return File{}, fmt.Errorf("file not found (%s)", file)
}

func (memoryFileStorage *MemoryFileStorage) SetModified(file string, timestamp time.Time) error {
	if val, ok := memoryFileStorage.files[file]; ok {
		val.LastModified = timestamp
		memoryFileStorage.files[file] = val
		return nil
	}
	return fmt.Errorf("file not found (%s)", file)

}

func (memoryFileStorage *MemoryFileStorage) Read(ctx context.Context, file string) (io.ReadCloser, error) {
	if val, ok := memoryFileStorage.files[file]; ok {
		return ioutil.NopCloser(bytes.NewReader(val.data)), nil
	}
	return nil, fmt.Errorf("file not found (%s)", file)
}

func (memoryFileStorage *MemoryFileStorage) Write(ctx context.Context, file string, reader io.Reader) error {
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}
	memoryFileStorage.files[file] = memoryFileStorageFile{
		File: File{
			Path:         file,
			LastModified: time.Time{},
			Size:         int64(len(data)),
		},
		data: data,
	}
	return nil
}

func (memoryFileStorage *MemoryFileStorage) Delete(file string) error {
	delete(memoryFileStorage.files, file)
	return nil
}
