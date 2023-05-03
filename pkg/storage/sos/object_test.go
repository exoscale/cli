package sos_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/exoscale/cli/pkg/storage/sos"
	"github.com/stretchr/testify/assert"
)

func TestShowObject(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name           string
		getObjectFn    func(ctx context.Context, input *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
		getObjectAclFn func(ctx context.Context, input *s3.GetObjectAclInput, optFns ...func(*s3.Options)) (*s3.GetObjectAclOutput, error) //nolint:revive
		expectErr      bool
	}{
		{
			name: "successful retrieval",
			getObjectFn: func(ctx context.Context, input *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
				return &s3.GetObjectOutput{
					ContentLength: 100,
					LastModified:  &now,
				}, nil
			},
			getObjectAclFn: func(ctx context.Context, input *s3.GetObjectAclInput, optFns ...func(*s3.Options)) (*s3.GetObjectAclOutput, error) {
				return &s3.GetObjectAclOutput{}, nil
			},
			expectErr: false,
		},
		{
			name: "failed to get object",
			getObjectFn: func(ctx context.Context, input *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
				return nil, errors.New("failed to get object")
			},
			getObjectAclFn: func(ctx context.Context, input *s3.GetObjectAclInput, optFns ...func(*s3.Options)) (*s3.GetObjectAclOutput, error) {
				return &s3.GetObjectAclOutput{}, nil
			},
			expectErr: true,
		},
		{
			name: "failed to get object ACL",
			getObjectFn: func(ctx context.Context, input *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
				return &s3.GetObjectOutput{
					ContentLength: 100,
					LastModified:  &now,
				}, nil
			},
			getObjectAclFn: func(ctx context.Context, input *s3.GetObjectAclInput, optFns ...func(*s3.Options)) (*s3.GetObjectAclOutput, error) {
				return nil, errors.New("failed to get object ACL")
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockS3API := &MockS3API{
				mockGetObject:    tt.getObjectFn,
				mockGetObjectAcl: tt.getObjectAclFn,
			}

			client := &sos.Client{
				S3Client: mockS3API,
				Zone:     "bern",
			}

			ctx := context.Background()
			bucket := "test-bucket"
			key := "test-key"

			output, err := client.ShowObject(ctx, bucket, key)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, output)
				assert.Equal(t, bucket, output.Bucket)
				assert.Equal(t, key, output.Path)
				assert.Equal(t, int64(100), output.Size)
				assert.Equal(t, fmt.Sprintf("https://sos-%s.exo.io/%s/%s", client.Zone, bucket, key), output.URL)
			}
		})
	}
}

func TestClient_DeleteObjects(t *testing.T) {
	bucket := "test-bucket"
	commonPrefix := "myobjects/"
	objectKeys := []string{commonPrefix + "object1", commonPrefix + "object2", commonPrefix + "object3"}

	nCalls := 0
	expectedDeleteInput := &s3.DeleteObjectsInput{
		Bucket: &bucket,
		Delete: &types.Delete{
			Objects: []types.ObjectIdentifier{
				{Key: aws.String(objectKeys[0])},
				{Key: aws.String(objectKeys[1])},
				{Key: aws.String(objectKeys[2])},
			},
		},
	}
	client := sos.Client{
		S3Client: &MockS3API{
			mockDeleteObjects: func(ctx context.Context, params *s3.DeleteObjectsInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error) {
				nCalls++
				assert.Equal(t, expectedDeleteInput, params)
				return &s3.DeleteObjectsOutput{
					Deleted: []types.DeletedObject{
						{Key: aws.String(objectKeys[0])},
						{Key: aws.String(objectKeys[1])},
						{Key: aws.String(objectKeys[2])},
					},
				}, nil
			},
			mockListObjectsV2: func(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
				return &s3.ListObjectsV2Output{
					IsTruncated: false,
					Contents: []types.Object{
						{Key: aws.String(objectKeys[0])},
						{Key: aws.String(objectKeys[1])},
						{Key: aws.String(objectKeys[2])},
					},
				}, nil
			},
		},
	}

	deleted, err := client.DeleteObjects(context.Background(), bucket, commonPrefix, false)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(deleted))
	assert.Equal(t, 1, nCalls)

	for i, key := range deleted {
		assert.Equal(t, objectKeys[i], *key.Key)
	}

	client = sos.Client{
		S3Client: &MockS3API{
			mockDeleteObjects: func(ctx context.Context, params *s3.DeleteObjectsInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error) {
				return nil, errors.New("delete error")
			},
			mockListObjectsV2: func(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
				return &s3.ListObjectsV2Output{
					IsTruncated: false,
					Contents: []types.Object{
						{
							Key: aws.String("some-file"),
						},
					},
				}, nil
			},
		},
	}

	_, err = client.DeleteObjects(context.Background(), bucket, commonPrefix, false)
	assert.Error(t, err)
}
