package sos_sync

import (
	"context"
	"github.com/minio/minio-go/v6"
	"io"
	"io/ioutil"
	"strings"
)

type MinioObjectStorageOverlay struct {
	minio        *minio.Client
	dryRun       bool
	targetBucket string
	targetPath   string
}

func NewMinioObjectStorageOverlay(minio *minio.Client, dryRun bool, targetBucket string, targetPath string) ObjectStorage {
	return &MinioObjectStorageOverlay{
		minio:        minio,
		dryRun:       dryRun,
		targetBucket: targetBucket,
		targetPath:   targetPath,
	}
}

func (minioOverlay *MinioObjectStorageOverlay) List(errors chan<- error) <-chan ObjectStorageObject {
	result := make(chan ObjectStorageObject)

	go func() {
		defer close(result)
		doneCh := make(chan struct{})
		defer close(doneCh)
		objects := minioOverlay.minio.ListObjectsV2(minioOverlay.targetBucket, minioOverlay.targetPath, true, doneCh)
		for object := range objects {
			result <- ObjectStorageObject{
				Key:          object.Key[len(minioOverlay.targetPath):],
				Size:         object.Size,
				LastModified: object.LastModified,
				ContentType:  object.ContentType,
			}
		}
	}()

	return result
}

func (minioOverlay *MinioObjectStorageOverlay) getRemotePath(objectName string) string {
	return strings.TrimLeft(minioOverlay.targetPath+"/"+objectName, "/")
}

func (minioOverlay *MinioObjectStorageOverlay) Upload(ctx context.Context, objectName string, reader io.Reader, objectSize int64, contentType string) error {
	if minioOverlay.dryRun {
		_, err := ioutil.ReadAll(reader)
		return err
	}
	remotePath := minioOverlay.getRemotePath(objectName)
	_, err := minioOverlay.minio.PutObjectWithContext(
		ctx,
		minioOverlay.targetBucket,
		remotePath,
		reader,
		objectSize,
		minio.PutObjectOptions{
			ContentType: contentType,
		},
	)
	return err
}

func (minioOverlay *MinioObjectStorageOverlay) Download(ctx context.Context, objectName string, writer io.Writer) error {
	remotePath := minioOverlay.getRemotePath(objectName)
	result, err := minioOverlay.minio.GetObject(minioOverlay.targetBucket, remotePath, minio.GetObjectOptions{})
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, result)
	if err != nil {
		return err
	}
	return nil
}

func (minioOverlay *MinioObjectStorageOverlay) Delete(objectName string) error {
	if minioOverlay.dryRun {
		return nil
	}
	remotePath := minioOverlay.getRemotePath(objectName)
	return minioOverlay.minio.RemoveObjectWithOptions(minioOverlay.targetBucket, remotePath, minio.RemoveObjectOptions{})
}
