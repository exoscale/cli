package sos_sync

import (
	"context"
	"fmt"
	"io"
	"time"
)

type memoryObjectStorageObject struct {
	ObjectStorageObject
	data []byte
}

type MemoryObjectStorage struct {
	objects map[string]memoryObjectStorageObject
}

func NewMemoryObjectStorage() ObjectStorage {
	return &MemoryObjectStorage{
		objects: make(map[string]memoryObjectStorageObject),
	}
}

func (memoryObjectStorgeOverlay *MemoryObjectStorage) List(errors chan<- error) <-chan ObjectStorageObject {
	result := make(chan ObjectStorageObject)
	go func() {
		defer close(result)
		for _, entry := range memoryObjectStorgeOverlay.objects {
			result <- entry.ObjectStorageObject
		}
	}()
	return result
}

func (memoryObjectStorgeOverlay *MemoryObjectStorage) Upload(ctx context.Context, objectName string, reader io.Reader, objectSize int64, contentType string) error {
	data := make([]byte, objectSize)
	if objectSize > 0 {
		readBytes, err := reader.Read(data)
		if err != nil {
			return err
		}
		if int64(readBytes) != objectSize {
			return fmt.Errorf("invalid number of bytes read: %d instead of %d", readBytes, objectSize)
		}
	}
	memoryObjectStorgeOverlay.objects[objectName] = memoryObjectStorageObject{
		ObjectStorageObject: ObjectStorageObject{
			Key:          objectName,
			LastModified: time.Time{},
			ContentType:  contentType,
			Size:         objectSize,
		},
		data: data,
	}
	return nil
}

func (memoryObjectStorgeOverlay *MemoryObjectStorage) Download(ctx context.Context, objectName string, writer io.Writer) error {
	if object, ok := memoryObjectStorgeOverlay.objects[objectName]; ok {
		bytesWritten, err := writer.Write(object.data)
		if err != nil {
			return err
		}
		if int64(bytesWritten) != object.Size {
			return fmt.Errorf("incorrect number of bytes written: %d instead of %d", bytesWritten, object.Size)
		}
		return nil
	}
	return fmt.Errorf("object not found (%s)", objectName)
}

func (memoryObjectStorgeOverlay *MemoryObjectStorage) Delete(objectName string) error {
	delete(memoryObjectStorgeOverlay.objects, objectName)
	return nil
}
