package sos_sync

import (
	"context"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

type LocalFileStorage struct {
	sourceDirectory string
	dryRun          bool
}

func NewLocalFileStorage(sourceDirectory string, dryRun bool) FileStorage {
	return &LocalFileStorage{
		sourceDirectory: sourceDirectory,
		dryRun:          dryRun,
	}
}

func (localFileStorage *LocalFileStorage) List(errors chan<- error) <-chan File {
	result := make(chan File)

	go func() {
		defer close(result)
		walkErr := filepath.Walk(localFileStorage.sourceDirectory,
			func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if !info.IsDir() {
					result <- File{
						Path:         filepath.ToSlash(path[len(localFileStorage.sourceDirectory)+1:]),
						LastModified: info.ModTime(),
						Size:         info.Size(),
					}
				}
				return nil
			})
		if walkErr != nil {
			errors <- walkErr
		}
	}()
	return result
}

func (localFileStorage *LocalFileStorage) getLocalPath(file string) string {
	trimmedFile := strings.TrimLeft(file, "/")
	return path.Join(localFileStorage.sourceDirectory, trimmedFile)
}

func (localFileStorage *LocalFileStorage) Get(file string) (File, error) {
	trimmedFile := strings.TrimLeft(file, "/")
	stat, err := os.Stat(localFileStorage.getLocalPath(trimmedFile))
	if err != nil {
		return File{}, err
	}

	return File{
		Path:         trimmedFile,
		Size:         stat.Size(),
		LastModified: stat.ModTime(),
	}, nil
}

func (localFileStorage *LocalFileStorage) SetModified(file string, timestamp time.Time) error {
	localPath := localFileStorage.getLocalPath(file)
	_, err := os.Stat(localPath)
	if err != nil {
		return err
	}

	return os.Chtimes(localPath, timestamp, timestamp)
}

func (localFileStorage *LocalFileStorage) Read(ctx context.Context, fileName string) (io.ReadCloser, error) {
	file, err := localFileStorage.Get(fileName)
	if err != nil {
		return nil, err
	}
	localPath := localFileStorage.getLocalPath(file.Path)
	fp, err := os.Open(localPath)
	if err != nil {
		return nil, err
	}
	return fp, nil
}

func (localFileStorage *LocalFileStorage) Write(ctx context.Context, fileName string, reader io.Reader) error {
	file, err := localFileStorage.Get(fileName)
	if err != nil {
		return err
	}
	localPath := localFileStorage.getLocalPath(file.Path)

	if localFileStorage.dryRun {
		_, err := ioutil.ReadAll(reader)
		return err
	}

	fp, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer fp.Close()
	_, err = io.Copy(fp, reader)
	if err != nil {
		return err
	}
	return nil
}

func (localFileStorage *LocalFileStorage) Delete(fileName string) error {
	if localFileStorage.dryRun {
		return nil
	}
	return os.Remove(fileName)
}
