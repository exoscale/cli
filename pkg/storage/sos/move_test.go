package sos_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/stretchr/testify/assert"

	"github.com/exoscale/cli/pkg/storage/sos"
)

func TestMoveObject(t *testing.T) {
	ctx := context.Background()
	sourceBucket := "source-bucket"
	destinationBucket := "destination-bucket"
	sourceKey := "folder/source file.txt"
	destinationKey := "archive/source file.txt"

	copyCalls := 0
	deleteCalls := 0

	srcClient := &sos.Client{S3Client: &MockS3API{
		mockHeadObject: func(ctx context.Context, input *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error) {
			assert.Equal(t, sourceBucket, aws.ToString(input.Bucket))
			assert.Equal(t, sourceKey, aws.ToString(input.Key))
			return &s3.HeadObjectOutput{
				ContentLength: 1024,
				ContentType:   aws.String("text/plain"),
				Metadata:      map[string]string{"key": "value"},
				ETag:          aws.String("\"etag\""),
			}, nil
		},
		mockGetObjectAcl: func(ctx context.Context, input *s3.GetObjectAclInput, optFns ...func(*s3.Options)) (*s3.GetObjectAclOutput, error) {
			assert.Equal(t, sourceBucket, aws.ToString(input.Bucket))
			assert.Equal(t, sourceKey, aws.ToString(input.Key))
			return &s3.GetObjectAclOutput{
				Owner: &types.Owner{ID: aws.String("owner-id")},
			}, nil
		},
		mockDeleteObjects: func(ctx context.Context, params *s3.DeleteObjectsInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error) {
			deleteCalls++
			assert.Equal(t, sourceBucket, aws.ToString(params.Bucket))
			assert.Len(t, params.Delete.Objects, 1)
			assert.Equal(t, sourceKey, aws.ToString(params.Delete.Objects[0].Key))
			return &s3.DeleteObjectsOutput{
				Deleted: []types.DeletedObject{{Key: aws.String(sourceKey)}},
			}, nil
		},
	}}

	dstClient := &sos.Client{S3Client: &MockS3API{
		mockCopyObject: func(ctx context.Context, params *s3.CopyObjectInput, optFns ...func(*s3.Options)) (*s3.CopyObjectOutput, error) {
			copyCalls++
			assert.Equal(t, destinationBucket, aws.ToString(params.Bucket))
			assert.Equal(t, destinationKey, aws.ToString(params.Key))
			assert.Equal(t, "source-bucket/folder/source%20file.txt", aws.ToString(params.CopySource))
			assert.Equal(t, types.MetadataDirectiveReplace, params.MetadataDirective)
			assert.Equal(t, map[string]string{"key": "value"}, params.Metadata)
			assert.Equal(t, "text/plain", aws.ToString(params.ContentType))
			assert.Equal(t, "\"etag\"", aws.ToString(params.CopySourceIfMatch))
			assert.Equal(t, "id=owner-id", aws.ToString(params.GrantFullControl))
			return &s3.CopyObjectOutput{}, nil
		},
	}}

	moved, err := dstClient.MoveObjects(ctx, srcClient, sourceBucket, sourceKey, destinationBucket, destinationKey, nil)
	assert.NoError(t, err)
	assert.Equal(t, []sos.MovedObject{{
		SourceBucket:      sourceBucket,
		SourceKey:         sourceKey,
		DestinationBucket: destinationBucket,
		DestinationKey:    destinationKey,
	}}, moved)
	assert.Equal(t, 1, copyCalls)
	assert.Equal(t, 1, deleteCalls)
}

func TestMovePrefixRecursively(t *testing.T) {
	ctx := context.Background()
	sourceBucket := "source-bucket"
	destinationBucket := "destination-bucket"
	sourcePrefix := "public/"
	destinationPrefix := "archive/"
	sourceKeys := []string{"public/a.txt", "public/sub/b.txt"}
	expectedDestinationKeys := []string{"archive/a.txt", "archive/sub/b.txt"}

	headCalls := 0
	deleteCalls := 0
	copyCalls := 0
	deletedKeys := make([]string, 0, len(sourceKeys))
	copiedKeys := make([]string, 0, len(sourceKeys))

	srcClient := &sos.Client{S3Client: &MockS3API{
		mockListObjectsV2: func(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
			assert.Equal(t, sourceBucket, aws.ToString(params.Bucket))
			assert.Equal(t, sourcePrefix, aws.ToString(params.Prefix))
			return &s3.ListObjectsV2Output{
				Contents: []types.Object{
					{Key: aws.String(sourceKeys[0])},
					{Key: aws.String(sourceKeys[1])},
				},
				IsTruncated: false,
			}, nil
		},
		mockHeadObject: func(ctx context.Context, input *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error) {
			headCalls++
			assert.Contains(t, sourceKeys, aws.ToString(input.Key))
			return &s3.HeadObjectOutput{ContentLength: 1024, ETag: aws.String("\"etag\"")}, nil
		},
		mockGetObjectAcl: func(ctx context.Context, input *s3.GetObjectAclInput, optFns ...func(*s3.Options)) (*s3.GetObjectAclOutput, error) {
			return &s3.GetObjectAclOutput{Owner: &types.Owner{ID: aws.String("owner-id")}}, nil
		},
		mockDeleteObjects: func(ctx context.Context, params *s3.DeleteObjectsInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error) {
			deleteCalls++
			deletedKey := aws.ToString(params.Delete.Objects[0].Key)
			deletedKeys = append(deletedKeys, deletedKey)
			return &s3.DeleteObjectsOutput{Deleted: []types.DeletedObject{{Key: aws.String(deletedKey)}}}, nil
		},
	}}

	dstClient := &sos.Client{S3Client: &MockS3API{
		mockCopyObject: func(ctx context.Context, params *s3.CopyObjectInput, optFns ...func(*s3.Options)) (*s3.CopyObjectOutput, error) {
			copyCalls++
			copiedKeys = append(copiedKeys, aws.ToString(params.Key))
			return &s3.CopyObjectOutput{}, nil
		},
	}}

	moved, err := dstClient.MoveObjects(ctx, srcClient, sourceBucket, sourcePrefix, destinationBucket, destinationPrefix, &sos.StorageMoveConfig{Recursive: true})
	assert.NoError(t, err)
	assert.Len(t, moved, 2)
	assert.Equal(t, sourceKeys, deletedKeys)
	assert.Equal(t, expectedDestinationKeys, copiedKeys)
	assert.Equal(t, len(sourceKeys), headCalls)
	assert.Equal(t, len(sourceKeys), deleteCalls)
	assert.Equal(t, len(sourceKeys), copyCalls)
}

func TestMoveObjectMultipartCopy(t *testing.T) {
	testMoveObjectMultipartCopy(t, nil, 5)
}

func TestMoveObjectMultipartCopyCustomConcurrency(t *testing.T) {
	testMoveObjectMultipartCopy(t, &sos.StorageMoveConfig{MultipartCopyConcurrency: 2}, 2)
}

func testMoveObjectMultipartCopy(t *testing.T, config *sos.StorageMoveConfig, expectedConcurrency int) {
	t.Helper()

	ctx := context.Background()
	sourceBucket := "source-bucket"
	destinationBucket := "destination-bucket"
	sourceKey := "large-object.bin"
	destinationKey := "archive/large-object.bin"
	size := int64(5)*1024*1024*1024 + 1

	partCalls := 0
	deleteCalls := 0
	completeCalls := 0
	maxInFlight := 0
	inFlight := 0
	partRanges := make(map[int32]string)
	var partMu sync.Mutex

	srcClient := &sos.Client{S3Client: &MockS3API{
		mockHeadObject: func(ctx context.Context, input *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error) {
			return &s3.HeadObjectOutput{
				ContentLength: size,
				ContentType:   aws.String("application/octet-stream"),
				ETag:          aws.String("\"etag\""),
			}, nil
		},
		mockGetObjectAcl: func(ctx context.Context, input *s3.GetObjectAclInput, optFns ...func(*s3.Options)) (*s3.GetObjectAclOutput, error) {
			return &s3.GetObjectAclOutput{Owner: &types.Owner{ID: aws.String("owner-id")}}, nil
		},
		mockDeleteObjects: func(ctx context.Context, params *s3.DeleteObjectsInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error) {
			deleteCalls++
			return &s3.DeleteObjectsOutput{Deleted: []types.DeletedObject{{Key: aws.String(sourceKey)}}}, nil
		},
	}}

	dstClient := &sos.Client{S3Client: &MockS3API{
		mockCreateMultipartUpload: func(ctx context.Context, params *s3.CreateMultipartUploadInput, optFns ...func(*s3.Options)) (*s3.CreateMultipartUploadOutput, error) {
			assert.Equal(t, destinationBucket, aws.ToString(params.Bucket))
			assert.Equal(t, destinationKey, aws.ToString(params.Key))
			assert.Equal(t, "application/octet-stream", aws.ToString(params.ContentType))
			return &s3.CreateMultipartUploadOutput{UploadId: aws.String("upload-id")}, nil
		},
		mockUploadPartCopy: func(ctx context.Context, params *s3.UploadPartCopyInput, optFns ...func(*s3.Options)) (*s3.UploadPartCopyOutput, error) {
			partMu.Lock()
			partCalls++
			inFlight++
			if inFlight > maxInFlight {
				maxInFlight = inFlight
			}
			partRanges[params.PartNumber] = aws.ToString(params.CopySourceRange)
			partMu.Unlock()

			time.Sleep(2 * time.Millisecond)

			partMu.Lock()
			inFlight--
			partMu.Unlock()

			assert.Equal(t, "source-bucket/large-object.bin", aws.ToString(params.CopySource))
			assert.Equal(t, "\"etag\"", aws.ToString(params.CopySourceIfMatch))
			return &s3.UploadPartCopyOutput{
				CopyPartResult: &types.CopyPartResult{ETag: aws.String(fmt.Sprintf("etag-%d", params.PartNumber))},
			}, nil
		},
		mockCompleteMultipartUpload: func(ctx context.Context, params *s3.CompleteMultipartUploadInput, optFns ...func(*s3.Options)) (*s3.CompleteMultipartUploadOutput, error) {
			completeCalls++
			assert.Equal(t, 1025, len(params.MultipartUpload.Parts))
			return &s3.CompleteMultipartUploadOutput{}, nil
		},
		mockAbortMultipartUpload: func(ctx context.Context, params *s3.AbortMultipartUploadInput, optFns ...func(*s3.Options)) (*s3.AbortMultipartUploadOutput, error) {
			assert.Fail(t, "abort should not be called on successful multipart copy")
			return nil, nil
		},
	}}

	moved, err := dstClient.MoveObjects(ctx, srcClient, sourceBucket, sourceKey, destinationBucket, destinationKey, config)
	assert.NoError(t, err)
	assert.Len(t, moved, 1)
	assert.Equal(t, 1025, partCalls)
	assert.Equal(t, 1, completeCalls)
	assert.Equal(t, 1, deleteCalls)
	assert.Equal(t, "bytes=0-5242879", partRanges[1])
	assert.Equal(t, "bytes=5368709120-5368709120", partRanges[1025])
	assert.Equal(t, expectedConcurrency, maxInFlight)
}

func TestMoveObjectMultipartCopyAbortOnFailure(t *testing.T) {
	ctx := context.Background()
	sourceBucket := "source-bucket"
	destinationBucket := "destination-bucket"
	sourceKey := "large-object.bin"
	destinationKey := "archive/large-object.bin"
	size := int64(5)*1024*1024*1024 + 1

	abortCalls := 0
	deleteCalls := 0
	completeCalls := 0

	srcClient := &sos.Client{S3Client: &MockS3API{
		mockHeadObject: func(ctx context.Context, input *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error) {
			return &s3.HeadObjectOutput{ContentLength: size, ETag: aws.String("\"etag\"")}, nil
		},
		mockGetObjectAcl: func(ctx context.Context, input *s3.GetObjectAclInput, optFns ...func(*s3.Options)) (*s3.GetObjectAclOutput, error) {
			return &s3.GetObjectAclOutput{Owner: &types.Owner{ID: aws.String("owner-id")}}, nil
		},
		mockDeleteObjects: func(ctx context.Context, params *s3.DeleteObjectsInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error) {
			deleteCalls++
			return &s3.DeleteObjectsOutput{}, nil
		},
	}}

	dstClient := &sos.Client{S3Client: &MockS3API{
		mockCreateMultipartUpload: func(ctx context.Context, params *s3.CreateMultipartUploadInput, optFns ...func(*s3.Options)) (*s3.CreateMultipartUploadOutput, error) {
			return &s3.CreateMultipartUploadOutput{UploadId: aws.String("upload-id")}, nil
		},
		mockUploadPartCopy: func(ctx context.Context, params *s3.UploadPartCopyInput, optFns ...func(*s3.Options)) (*s3.UploadPartCopyOutput, error) {
			return nil, errors.New("copy failed")
		},
		mockAbortMultipartUpload: func(ctx context.Context, params *s3.AbortMultipartUploadInput, optFns ...func(*s3.Options)) (*s3.AbortMultipartUploadOutput, error) {
			abortCalls++
			assert.Equal(t, "upload-id", aws.ToString(params.UploadId))
			return &s3.AbortMultipartUploadOutput{}, nil
		},
		mockCompleteMultipartUpload: func(ctx context.Context, params *s3.CompleteMultipartUploadInput, optFns ...func(*s3.Options)) (*s3.CompleteMultipartUploadOutput, error) {
			completeCalls++
			return &s3.CompleteMultipartUploadOutput{}, nil
		},
	}}

	moved, err := dstClient.MoveObjects(ctx, srcClient, sourceBucket, sourceKey, destinationBucket, destinationKey, nil)
	assert.Error(t, err)
	assert.Nil(t, moved)
	assert.Equal(t, 1, abortCalls)
	assert.Equal(t, 0, deleteCalls)
	assert.Equal(t, 0, completeCalls)
}
