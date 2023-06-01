package sos_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/stretchr/testify/assert"

	"github.com/exoscale/cli/pkg/storage/sos"
)

func TestClientBucketVersioningState(t *testing.T) {
	bucket := "example-bucket"

	enabledStatus := types.BucketVersioningStatusEnabled
	suspendedStatus := types.BucketVersioningStatusSuspended
	disabledStatus := "Disabled"

	ctx := context.Background()

	returnStatus := func(status string) func(ctx context.Context, params *s3.GetBucketVersioningInput, optFns ...func(*s3.Options)) (*s3.GetBucketVersioningOutput, error) {
		return func(ctx context.Context, params *s3.GetBucketVersioningInput, optFns ...func(*s3.Options)) (*s3.GetBucketVersioningOutput, error) {
			return &s3.GetBucketVersioningOutput{
				Status: types.BucketVersioningStatus(status),
			}, nil
		}
	}

	t.Run("versioning enabled", func(t *testing.T) {
		c := &sos.Client{
			S3Client: &MockS3API{
				mockGetBucketVersioning: returnStatus(string(enabledStatus)),
			},
		}

		status, err := c.GetBucketVersioning(ctx, bucket)
		assert.NoError(t, err)

		assert.Equal(t, status, enabledStatus)
	})

	t.Run("versioning never enabled before", func(t *testing.T) {
		c := &sos.Client{
			S3Client: &MockS3API{
				mockGetBucketVersioning: returnStatus(""),
			},
		}

		status, err := c.GetBucketVersioning(ctx, bucket)
		assert.NoError(t, err)

		assert.Equal(t, string(status), string(disabledStatus))
	})

	t.Run("versioning suspended", func(t *testing.T) {
		c := &sos.Client{
			S3Client: &MockS3API{
				mockGetBucketVersioning: returnStatus(string(suspendedStatus)),
			},
		}

		status, err := c.GetBucketVersioning(ctx, bucket)
		assert.NoError(t, err)

		assert.Equal(t, status, suspendedStatus)
	})
}
