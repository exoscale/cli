package sos_test

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/exoscale/cli/pkg/storage/sos"
)

func TestGetBucketLifecycle(t *testing.T) {
	ctx := context.Background()
	bucket := "test-bucket"

	t.Run("success with prefix filter", func(t *testing.T) {
		prefix := "logs/"
		days := int32(30)

		c := &sos.Client{
			S3Client: &MockS3API{
				mockGetBucketLifecycleConfiguration: func(_ context.Context, params *s3.GetBucketLifecycleConfigurationInput, _ ...func(*s3.Options)) (*s3.GetBucketLifecycleConfigurationOutput, error) {
					assert.Equal(t, aws.String(bucket), params.Bucket)
					return &s3.GetBucketLifecycleConfigurationOutput{
						Rules: []s3types.LifecycleRule{
							{
								ID:     aws.String("rule-1"),
								Status: s3types.ExpirationStatusEnabled,
								Filter: &s3types.LifecycleRuleFilterMemberPrefix{Value: prefix},
								Expiration: &s3types.LifecycleExpiration{
									Days: aws.Int32(days),
								},
							},
						},
					}, nil
				},
			},
		}

		lc, err := c.GetBucketLifecycle(ctx, bucket)
		require.NoError(t, err)
		assert.Equal(t, bucket, lc.Bucket)
		require.Len(t, lc.Rules, 1)
		assert.Equal(t, aws.String("rule-1"), lc.Rules[0].ID)
		assert.Equal(t, s3types.ExpirationStatusEnabled, lc.Rules[0].Status)
		require.NotNil(t, lc.Rules[0].Filter)
		assert.Equal(t, aws.String(prefix), lc.Rules[0].Filter.Prefix)
		require.NotNil(t, lc.Rules[0].Expiration)
		assert.Equal(t, aws.Int32(days), lc.Rules[0].Expiration.Days)
	})

	t.Run("success with and filter", func(t *testing.T) {
		prefix := "data/"
		sizeGT := int64(1024)
		sizeLT := int64(1048576)

		c := &sos.Client{
			S3Client: &MockS3API{
				mockGetBucketLifecycleConfiguration: func(_ context.Context, _ *s3.GetBucketLifecycleConfigurationInput, _ ...func(*s3.Options)) (*s3.GetBucketLifecycleConfigurationOutput, error) {
					return &s3.GetBucketLifecycleConfigurationOutput{
						Rules: []s3types.LifecycleRule{
							{
								ID:     aws.String("rule-and"),
								Status: s3types.ExpirationStatusEnabled,
								Filter: &s3types.LifecycleRuleFilterMemberAnd{
									Value: s3types.LifecycleRuleAndOperator{
										Prefix:                aws.String(prefix),
										ObjectSizeGreaterThan: aws.Int64(sizeGT),
										ObjectSizeLessThan:    aws.Int64(sizeLT),
									},
								},
							},
						},
					}, nil
				},
			},
		}

		lc, err := c.GetBucketLifecycle(ctx, bucket)
		require.NoError(t, err)
		require.Len(t, lc.Rules, 1)
		require.NotNil(t, lc.Rules[0].Filter)
		require.NotNil(t, lc.Rules[0].Filter.And)
		assert.Equal(t, aws.String(prefix), lc.Rules[0].Filter.And.Prefix)
		assert.Equal(t, aws.Int64(sizeGT), lc.Rules[0].Filter.And.ObjectSizeGreaterThan)
		assert.Equal(t, aws.Int64(sizeLT), lc.Rules[0].Filter.And.ObjectSizeLessThan)
	})

	t.Run("success with size filter", func(t *testing.T) {
		sizeGT := int64(512)

		c := &sos.Client{
			S3Client: &MockS3API{
				mockGetBucketLifecycleConfiguration: func(_ context.Context, _ *s3.GetBucketLifecycleConfigurationInput, _ ...func(*s3.Options)) (*s3.GetBucketLifecycleConfigurationOutput, error) {
					return &s3.GetBucketLifecycleConfigurationOutput{
						Rules: []s3types.LifecycleRule{
							{
								ID:     aws.String("rule-size"),
								Status: s3types.ExpirationStatusDisabled,
								Filter: &s3types.LifecycleRuleFilterMemberObjectSizeGreaterThan{Value: sizeGT},
								NoncurrentVersionExpiration: &s3types.NoncurrentVersionExpiration{
									NoncurrentDays: aws.Int32(7),
								},
							},
						},
					}, nil
				},
			},
		}

		lc, err := c.GetBucketLifecycle(ctx, bucket)
		require.NoError(t, err)
		require.Len(t, lc.Rules, 1)
		require.NotNil(t, lc.Rules[0].Filter)
		assert.Equal(t, aws.Int64(sizeGT), lc.Rules[0].Filter.ObjectSizeGreaterThan)
		assert.Nil(t, lc.Rules[0].Filter.Prefix)
		require.NotNil(t, lc.Rules[0].NoncurrentVersionExpiration)
		assert.Equal(t, aws.Int32(7), lc.Rules[0].NoncurrentVersionExpiration.NoncurrentDays)
	})

	t.Run("success with noncurrent version expiration", func(t *testing.T) {
		c := &sos.Client{
			S3Client: &MockS3API{
				mockGetBucketLifecycleConfiguration: func(_ context.Context, _ *s3.GetBucketLifecycleConfigurationInput, _ ...func(*s3.Options)) (*s3.GetBucketLifecycleConfigurationOutput, error) {
					return &s3.GetBucketLifecycleConfigurationOutput{
						Rules: []s3types.LifecycleRule{
							{
								ID:     aws.String("noncurrent-rule"),
								Status: s3types.ExpirationStatusEnabled,
								NoncurrentVersionExpiration: &s3types.NoncurrentVersionExpiration{
									NoncurrentDays: aws.Int32(14),
								},
							},
						},
					}, nil
				},
			},
		}

		lc, err := c.GetBucketLifecycle(ctx, bucket)
		require.NoError(t, err)
		require.Len(t, lc.Rules, 1)
		assert.Nil(t, lc.Rules[0].Filter)
		require.NotNil(t, lc.Rules[0].NoncurrentVersionExpiration)
		assert.Equal(t, aws.Int32(14), lc.Rules[0].NoncurrentVersionExpiration.NoncurrentDays)
	})

	t.Run("success with expired object delete marker", func(t *testing.T) {
		c := &sos.Client{
			S3Client: &MockS3API{
				mockGetBucketLifecycleConfiguration: func(_ context.Context, _ *s3.GetBucketLifecycleConfigurationInput, _ ...func(*s3.Options)) (*s3.GetBucketLifecycleConfigurationOutput, error) {
					return &s3.GetBucketLifecycleConfigurationOutput{
						Rules: []s3types.LifecycleRule{
							{
								ID:     aws.String("delete-marker-rule"),
								Status: s3types.ExpirationStatusEnabled,
								Expiration: &s3types.LifecycleExpiration{
									ExpiredObjectDeleteMarker: aws.Bool(true),
								},
							},
						},
					}, nil
				},
			},
		}

		lc, err := c.GetBucketLifecycle(ctx, bucket)
		require.NoError(t, err)
		require.Len(t, lc.Rules, 1)
		require.NotNil(t, lc.Rules[0].Expiration)
		assert.Nil(t, lc.Rules[0].Expiration.Days)
		assert.Nil(t, lc.Rules[0].Expiration.Date)
		require.NotNil(t, lc.Rules[0].Expiration.ExpiredObjectDeleteMarker)
		assert.True(t, *lc.Rules[0].Expiration.ExpiredObjectDeleteMarker)
	})

	t.Run("api error", func(t *testing.T) {
		c := &sos.Client{
			S3Client: &MockS3API{
				mockGetBucketLifecycleConfiguration: func(_ context.Context, _ *s3.GetBucketLifecycleConfigurationInput, _ ...func(*s3.Options)) (*s3.GetBucketLifecycleConfigurationOutput, error) {
					return nil, errors.New("get lifecycle error")
				},
			},
		}

		_, err := c.GetBucketLifecycle(ctx, bucket)
		assert.Error(t, err)
	})
}

func TestPutBucketLifecycle(t *testing.T) {
	ctx := context.Background()
	bucket := "test-bucket"

	t.Run("success", func(t *testing.T) {
		days := int32(90)
		conf := &s3types.BucketLifecycleConfiguration{
			Rules: []s3types.LifecycleRule{
				{
					ID:     aws.String("expire-old"),
					Status: s3types.ExpirationStatusEnabled,
					Filter: &s3types.LifecycleRuleFilterMemberPrefix{Value: "archive/"},
					Expiration: &s3types.LifecycleExpiration{
						Days: aws.Int32(days),
					},
				},
			},
		}

		putCount := 0
		c := &sos.Client{
			S3Client: &MockS3API{
				mockPutBucketLifecycleConfiguration: func(_ context.Context, params *s3.PutBucketLifecycleConfigurationInput, _ ...func(*s3.Options)) (*s3.PutBucketLifecycleConfigurationOutput, error) {
					putCount++
					assert.Equal(t, aws.String(bucket), params.Bucket)
					assert.Equal(t, conf, params.LifecycleConfiguration)
					return &s3.PutBucketLifecycleConfigurationOutput{}, nil
				},
			},
		}

		err := c.PutBucketLifecycle(ctx, bucket, conf)
		assert.NoError(t, err)
		assert.Equal(t, 1, putCount)
	})

	t.Run("api error", func(t *testing.T) {
		c := &sos.Client{
			S3Client: &MockS3API{
				mockPutBucketLifecycleConfiguration: func(_ context.Context, _ *s3.PutBucketLifecycleConfigurationInput, _ ...func(*s3.Options)) (*s3.PutBucketLifecycleConfigurationOutput, error) {
					return nil, errors.New("put lifecycle error")
				},
			},
		}

		err := c.PutBucketLifecycle(ctx, bucket, &s3types.BucketLifecycleConfiguration{})
		assert.Error(t, err)
	})
}
