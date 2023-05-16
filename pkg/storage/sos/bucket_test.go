package sos_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	"github.com/stretchr/testify/assert"

	"github.com/exoscale/cli/pkg/storage/sos"
)

func TestCreateNewBucket(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name                   string
		bucket                 string
		acl                    string
		expectError            bool
		createBucketFuncErrors bool
		expectedNrOfCalls      int
	}{
		{
			name:                   "Success",
			bucket:                 "test-bucket",
			acl:                    "",
			expectedNrOfCalls:      1,
			expectError:            false,
			createBucketFuncErrors: false,
		},
		{
			name:                   "Invalid ACL",
			bucket:                 "test-bucket",
			acl:                    "invalid-acl",
			expectedNrOfCalls:      0,
			expectError:            true,
			createBucketFuncErrors: false,
		},
		{
			name:                   "S3 Client Error",
			bucket:                 "test-bucket",
			acl:                    "",
			expectedNrOfCalls:      1,
			expectError:            true,
			createBucketFuncErrors: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			nrOfCalls := 0

			client := &sos.Client{
				S3Client: &MockS3API{
					mockCreateBucket: func(ctx context.Context, params *s3.CreateBucketInput, optFns ...func(*s3.Options)) (*s3.CreateBucketOutput, error) {
						nrOfCalls++

						if tc.createBucketFuncErrors {
							return nil, fmt.Errorf("some error")
						}

						return nil, nil
					},
				},
			}

			err := client.CreateNewBucket(ctx, tc.bucket, tc.acl)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tc.expectedNrOfCalls, nrOfCalls)
		})
	}
}

func TestShowBucket(t *testing.T) {
	ctx := context.Background()
	bucket := "test-bucket"

	t.Run("successful_show_bucket", func(t *testing.T) {
		client := &sos.Client{
			S3Client: &MockS3API{
				mockGetBucketAcl: func(ctx context.Context, params *s3.GetBucketAclInput, optFns ...func(*s3.Options)) (*s3.GetBucketAclOutput, error) {
					return &s3.GetBucketAclOutput{
						Grants: []types.Grant{
							{
								Grantee: &types.Grantee{
									Type:        types.TypeCanonicalUser,
									DisplayName: aws.String("CanonicalUser"),
								},
								Permission: types.PermissionRead,
							},
						},
					}, nil
				},
				mockGetBucketCors: func(ctx context.Context, params *s3.GetBucketCorsInput, optFns ...func(*s3.Options)) (*s3.GetBucketCorsOutput, error) {
					return &s3.GetBucketCorsOutput{
						CORSRules: []types.CORSRule{
							{
								AllowedOrigins: []string{"*"},
								AllowedMethods: []string{"GET"},
								AllowedHeaders: []string{"*"},
							},
						},
					}, nil
				},
			},
			Zone: "myzone",
		}

		output, err := client.ShowBucket(ctx, bucket)
		assert.NoError(t, err)

		expectedOutput := &sos.ShowBucketOutput{
			Name: bucket,
			Zone: "myzone",
			ACL: sos.ACL{
				Read:        "CanonicalUser",
				Write:       "-",
				ReadACP:     "-",
				WriteACP:    "-",
				FullControl: "-",
			},
			CORS: []sos.CORSRule{
				{
					AllowedOrigins: []string{"*"},
					AllowedMethods: []string{"GET"},
					AllowedHeaders: []string{"*"},
				},
			},
		}
		assert.Equal(t, expectedOutput, output)
	})

	t.Run("error_get_bucket_acl", func(t *testing.T) {
		client := &sos.Client{
			S3Client: &MockS3API{
				mockGetBucketAcl: func(ctx context.Context, params *s3.GetBucketAclInput, optFns ...func(*s3.Options)) (*s3.GetBucketAclOutput, error) {
					return nil, errors.New("get bucket ACL error")
				},
			},
			Zone: "myzone",
		}
		output, err := client.ShowBucket(ctx, bucket)
		assert.Error(t, err)
		assert.Nil(t, output)
	})

	t.Run("error_get_bucket_cors", func(t *testing.T) {
		client := &sos.Client{
			S3Client: &MockS3API{
				mockGetBucketAcl: func(ctx context.Context, params *s3.GetBucketAclInput, optFns ...func(*s3.Options)) (*s3.GetBucketAclOutput, error) {
					return &s3.GetBucketAclOutput{}, nil
				},
				mockGetBucketCors: func(ctx context.Context, params *s3.GetBucketCorsInput, optFns ...func(*s3.Options)) (*s3.GetBucketCorsOutput, error) {
					return nil, errors.New("get bucket CORS error")
				},
			},
			Zone: "myzone",
		}

		output, err := client.ShowBucket(ctx, bucket)
		assert.Error(t, err)
		assert.Nil(t, output)
	})

	t.Run("error_no_such_cors_configuration", func(t *testing.T) {
		client := &sos.Client{
			S3Client: &MockS3API{
				mockGetBucketAcl: func(ctx context.Context, params *s3.GetBucketAclInput, optFns ...func(*s3.Options)) (*s3.GetBucketAclOutput, error) {
					return &s3.GetBucketAclOutput{}, nil
				},
				mockGetBucketCors: func(ctx context.Context, params *s3.GetBucketCorsInput, optFns ...func(*s3.Options)) (*s3.GetBucketCorsOutput, error) {
					return nil, &smithy.GenericAPIError{
						Code: "NoSuchCORSConfiguration",
					}
				},
			},
			Zone: "myzone",
		}

		output, err := client.ShowBucket(ctx, bucket)
		assert.NoError(t, err)

		expectedOutput := &sos.ShowBucketOutput{
			Name: bucket,
			Zone: "myzone",
			ACL: sos.ACL{
				Read:        "-",
				Write:       "-",
				ReadACP:     "-",
				WriteACP:    "-",
				FullControl: "-",
			},
			CORS: []sos.CORSRule{},
		}
		assert.Equal(t, expectedOutput, output)
	})
}

func TestSetBucketObjectOwnership(t *testing.T) {
	ctx := context.Background()
	bucket := "test-bucket"

	t.Run("successful_set_bucket_object_ownership", func(t *testing.T) {
		client := &sos.Client{
			S3Client: &MockS3API{
				mockPutBucketOwnershipControls: func(ctx context.Context, params *s3.PutBucketOwnershipControlsInput, optFns ...func(*s3.Options)) (*s3.PutBucketOwnershipControlsOutput, error) {
					return &s3.PutBucketOwnershipControlsOutput{}, nil
				},
			},
		}

		ownership := sos.BucketObjectOwnership("BucketOwnerPreferred")
		err := client.SetBucketObjectOwnership(ctx, bucket, ownership)
		assert.NoError(t, err)
	})

	t.Run("error_set_bucket_object_ownership", func(t *testing.T) {
		client := &sos.Client{
			S3Client: &MockS3API{
				mockPutBucketOwnershipControls: func(ctx context.Context, params *s3.PutBucketOwnershipControlsInput, optFns ...func(*s3.Options)) (*s3.PutBucketOwnershipControlsOutput, error) {
					return nil, errors.New("set bucket object ownership error")
				},
			},
		}

		ownership := sos.BucketObjectOwnership("BucketOwnerPreferred")
		err := client.SetBucketObjectOwnership(ctx, bucket, ownership)
		assert.Error(t, err)
	})
}
func TestDeleteBucket(t *testing.T) {
	ctx := context.Background()
	bucket := "test-bucket"

	t.Run("successful_delete_bucket", func(t *testing.T) {
		client := &sos.Client{
			S3Client: &MockS3API{
				mockListMultipartUploads: func(ctx context.Context, params *s3.ListMultipartUploadsInput, optFns ...func(*s3.Options)) (*s3.ListMultipartUploadsOutput, error) {
					return &s3.ListMultipartUploadsOutput{}, nil
				},
				mockDeleteBucket: func(ctx context.Context, params *s3.DeleteBucketInput, optFns ...func(*s3.Options)) (*s3.DeleteBucketOutput, error) {
					return &s3.DeleteBucketOutput{}, nil
				},
			},
		}

		err := client.DeleteBucket(ctx, bucket, false)
		assert.NoError(t, err)
	})

	t.Run("error_list_multipart_uploads", func(t *testing.T) {
		client := &sos.Client{
			S3Client: &MockS3API{
				mockListMultipartUploads: func(ctx context.Context, params *s3.ListMultipartUploadsInput, optFns ...func(*s3.Options)) (*s3.ListMultipartUploadsOutput, error) {
					return nil, errors.New("list multipart uploads error")
				},
			},
		}

		err := client.DeleteBucket(ctx, bucket, false)
		assert.Error(t, err)
	})

	t.Run("error_abort_multipart_upload", func(t *testing.T) {
		client := &sos.Client{
			S3Client: &MockS3API{
				mockListMultipartUploads: func(ctx context.Context, params *s3.ListMultipartUploadsInput, optFns ...func(*s3.Options)) (*s3.ListMultipartUploadsOutput, error) {
					return &s3.ListMultipartUploadsOutput{
						Uploads: []types.MultipartUpload{
							{
								Key:      aws.String("test-key"),
								UploadId: aws.String("test-upload-id"),
							},
						},
					}, nil
				},
				mockAbortMultipartUpload: func(ctx context.Context, params *s3.AbortMultipartUploadInput, optFns ...func(*s3.Options)) (*s3.AbortMultipartUploadOutput, error) {
					return nil, errors.New("abort multipart upload error")
				},
			},
		}
		err := client.DeleteBucket(ctx, bucket, false)
		assert.Error(t, err)
	})

	t.Run("error_delete_bucket", func(t *testing.T) {
		client := &sos.Client{
			S3Client: &MockS3API{
				mockListMultipartUploads: func(ctx context.Context, params *s3.ListMultipartUploadsInput, optFns ...func(*s3.Options)) (*s3.ListMultipartUploadsOutput, error) {
					return &s3.ListMultipartUploadsOutput{}, nil
				},
				mockDeleteBucket: func(ctx context.Context, params *s3.DeleteBucketInput, optFns ...func(*s3.Options)) (*s3.DeleteBucketOutput, error) {
					return nil, errors.New("delete bucket error")
				},
			},
		}

		err := client.DeleteBucket(ctx, bucket, false)
		assert.Error(t, err)
	})

	t.Run("error_bucket_not_empty", func(t *testing.T) {
		client := &sos.Client{
			S3Client: &MockS3API{
				mockListMultipartUploads: func(ctx context.Context, params *s3.ListMultipartUploadsInput, optFns ...func(*s3.Options)) (*s3.ListMultipartUploadsOutput, error) {
					return &s3.ListMultipartUploadsOutput{}, nil
				},
				mockDeleteBucket: func(ctx context.Context, params *s3.DeleteBucketInput, optFns ...func(*s3.Options)) (*s3.DeleteBucketOutput, error) {
					return nil, &smithy.GenericAPIError{
						Code: "BucketNotEmpty",
					}
				},
			},
		}

		err := client.DeleteBucket(ctx, bucket, false)
		assert.Error(t, err)
		assert.Equal(t, "bucket is not empty, either delete files before or use flag `-r`", err.Error())
	})
}
