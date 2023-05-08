package sos_test

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
	"github.com/exoscale/cli/pkg/storage/sos"
	"github.com/stretchr/testify/assert"
)

func TestAddBucketCORSRule(t *testing.T) {
	ctx := context.Background()

	returnEmptyGetBucketCors := func(ctx context.Context, params *s3.GetBucketCorsInput, optFns ...func(*s3.Options)) (*s3.GetBucketCorsOutput, error) {
		return &s3.GetBucketCorsOutput{}, nil
	}

	returnEmptyPutBucketCors := func(ctx context.Context, params *s3.PutBucketCorsInput, optFns ...func(*s3.Options)) (*s3.PutBucketCorsOutput, error) {
		return &s3.PutBucketCorsOutput{}, nil
	}

	t.Run("no_such_cors_configuration", func(t *testing.T) {
		c := &sos.Client{
			S3Client: &MockS3API{
				mockGetBucketCors: func(ctx context.Context, params *s3.GetBucketCorsInput, optFns ...func(*s3.Options)) (*s3.GetBucketCorsOutput, error) {
					return &s3.GetBucketCorsOutput{}, &smithy.GenericAPIError{
						Code: "NoSuchCORSConfiguration",
					}
				},
				mockPutBucketCors: returnEmptyPutBucketCors,
			},
		}

		cors := &sos.CORSRule{}

		err := c.AddBucketCORSRule(ctx, "test-bucket", cors)
		assert.NoError(t, err)
	})

	t.Run("get_bucket_cors_error", func(t *testing.T) {
		c := &sos.Client{
			S3Client: &MockS3API{
				mockGetBucketCors: func(ctx context.Context, params *s3.GetBucketCorsInput, optFns ...func(*s3.Options)) (*s3.GetBucketCorsOutput, error) {
					return &s3.GetBucketCorsOutput{}, errors.New("get bucket CORS error")
				},
				mockPutBucketCors: returnEmptyPutBucketCors,
			},
		}

		err := c.AddBucketCORSRule(ctx, "test-bucket", nil)
		assert.Error(t, err)
	})

	t.Run("put_bucket_cors_error", func(t *testing.T) {
		c := &sos.Client{
			S3Client: &MockS3API{
				mockGetBucketCors: returnEmptyGetBucketCors,
				mockPutBucketCors: func(ctx context.Context, params *s3.PutBucketCorsInput, optFns ...func(*s3.Options)) (*s3.PutBucketCorsOutput, error) {
					return nil, errors.New("put bucket CORS error")
				},
			},
		}

		cors := &sos.CORSRule{}

		err := c.AddBucketCORSRule(ctx, "test-bucket", cors)
		assert.Error(t, err)
	})

	t.Run("success", func(t *testing.T) {
		c := &sos.Client{
			S3Client: &MockS3API{
				mockGetBucketCors: returnEmptyGetBucketCors,
				mockPutBucketCors: returnEmptyPutBucketCors,
			},
		}

		cors := &sos.CORSRule{}

		err := c.AddBucketCORSRule(ctx, "test-bucket", cors)
		assert.NoError(t, err)
	})
}
