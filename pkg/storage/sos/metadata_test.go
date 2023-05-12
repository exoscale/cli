package sos_test

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/exoscale/cli/pkg/storage/sos"
	"github.com/stretchr/testify/assert"
)

func TestAddObjectMetadata(t *testing.T) {
	ctx := context.Background()
	bucket := "test-bucket"
	key := "test-key"
	metadata := map[string]string{"key": "value"}

	returnEmptyMockGetObject := func(ctx context.Context, input *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
		return &s3.GetObjectOutput{}, nil
	}

	returnEmptyMockGetObjectACL := func(ctx context.Context, input *s3.GetObjectAclInput, optFns ...func(*s3.Options)) (*s3.GetObjectAclOutput, error) {
		return &s3.GetObjectAclOutput{
			Owner: &types.Owner{
				ID: aws.String("sarah"),
			},
		}, nil
	}

	mockS3API := &MockS3API{
		mockCopyObject: func(ctx context.Context, params *s3.CopyObjectInput, optFns ...func(*s3.Options)) (*s3.CopyObjectOutput, error) {
			return &s3.CopyObjectOutput{}, nil
		},
		mockGetObject:    returnEmptyMockGetObject,
		mockGetObjectAcl: returnEmptyMockGetObjectACL,
	}

	client := &sos.Client{S3Client: mockS3API}

	t.Run("Successful metadata addition", func(t *testing.T) {
		err := client.AddObjectMetadata(ctx, bucket, key, metadata)
		assert.NoError(t, err)
	})

	t.Run("Error due to invalid metadata key", func(t *testing.T) {
		invalidMetadata := map[string]string{"invalid@key": "value"}
		err := client.AddObjectMetadata(ctx, bucket, key, invalidMetadata)
		assert.Error(t, err)
	})

	mockS3APIWithError := &MockS3API{
		mockCopyObject: func(ctx context.Context, params *s3.CopyObjectInput, optFns ...func(*s3.Options)) (*s3.CopyObjectOutput, error) {
			return nil, errors.New("copy object error")
		},
		mockGetObject:    returnEmptyMockGetObject,
		mockGetObjectAcl: returnEmptyMockGetObjectACL,
	}

	clientWithError := &sos.Client{S3Client: mockS3APIWithError}

	t.Run("Error from CopyObject", func(t *testing.T) {
		err := clientWithError.AddObjectMetadata(ctx, bucket, key, metadata)
		assert.Error(t, err)
	})
}
