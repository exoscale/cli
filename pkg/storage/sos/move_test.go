package sos_test

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/stretchr/testify/assert"

	"github.com/exoscale/cli/pkg/storage/sos"
)

func TestMoveObject_SingleObject(t *testing.T) {
	tests := []struct {
		name        string
		srcBucket   string
		srcKey      string
		dstBucket   string
		dstKey      string
		setupMocks  func(*MockS3API)
		expectError bool
	}{
		{
			name:      "successful move within same bucket",
			srcBucket: "test-bucket",
			srcKey:    "source-key",
			dstBucket: "test-bucket",
			dstKey:    "dest-key",
			setupMocks: func(m *MockS3API) {
				m.mockHeadObject = func(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error) {
					return &s3.HeadObjectOutput{
						ContentLength: aws.Int64(1024),
						Metadata:      map[string]string{"key": "value"},
						ContentType:   aws.String("text/plain"),
					}, nil
				}
				m.mockGetObjectAcl = func(ctx context.Context, params *s3.GetObjectAclInput, optFns ...func(*s3.Options)) (*s3.GetObjectAclOutput, error) {
					return &s3.GetObjectAclOutput{}, nil
				}
				m.mockCopyObject = func(ctx context.Context, params *s3.CopyObjectInput, optFns ...func(*s3.Options)) (*s3.CopyObjectOutput, error) {
					assert.Equal(t, "test-bucket", *params.Bucket)
					assert.Equal(t, "dest-key", *params.Key)
					assert.Equal(t, "test-bucket/source-key", *params.CopySource)
					return &s3.CopyObjectOutput{}, nil
				}
				m.mockDeleteObject = func(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
					assert.Equal(t, "test-bucket", *params.Bucket)
					assert.Equal(t, "source-key", *params.Key)
					return &s3.DeleteObjectOutput{}, nil
				}
			},
			expectError: false,
		},
		{
			name:      "move fails when copy fails",
			srcBucket: "test-bucket",
			srcKey:    "source-key",
			dstBucket: "test-bucket",
			dstKey:    "dest-key",
			setupMocks: func(m *MockS3API) {
				m.mockHeadObject = func(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error) {
					return &s3.HeadObjectOutput{ContentLength: aws.Int64(1024)}, nil
				}
				m.mockGetObjectAcl = func(ctx context.Context, params *s3.GetObjectAclInput, optFns ...func(*s3.Options)) (*s3.GetObjectAclOutput, error) {
					return &s3.GetObjectAclOutput{}, nil
				}
				m.mockCopyObject = func(ctx context.Context, params *s3.CopyObjectInput, optFns ...func(*s3.Options)) (*s3.CopyObjectOutput, error) {
					return nil, errors.New("copy failed")
				}
			},
			expectError: true,
		},
		{
			name:      "move fails when delete fails after copy",
			srcBucket: "test-bucket",
			srcKey:    "source-key",
			dstBucket: "test-bucket",
			dstKey:    "dest-key",
			setupMocks: func(m *MockS3API) {
				m.mockHeadObject = func(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error) {
					return &s3.HeadObjectOutput{ContentLength: aws.Int64(1024)}, nil
				}
				m.mockGetObjectAcl = func(ctx context.Context, params *s3.GetObjectAclInput, optFns ...func(*s3.Options)) (*s3.GetObjectAclOutput, error) {
					return &s3.GetObjectAclOutput{}, nil
				}
				m.mockCopyObject = func(ctx context.Context, params *s3.CopyObjectInput, optFns ...func(*s3.Options)) (*s3.CopyObjectOutput, error) {
					return &s3.CopyObjectOutput{}, nil
				}
				m.mockDeleteObject = func(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
					return nil, errors.New("delete failed")
				}
			},
			expectError: true,
		},
		{
			name:      "head object fails",
			srcBucket: "test-bucket",
			srcKey:    "source-key",
			dstBucket: "test-bucket",
			dstKey:    "dest-key",
			setupMocks: func(m *MockS3API) {
				m.mockHeadObject = func(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error) {
					return nil, errors.New("head object failed")
				}
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockS3API := &MockS3API{}
			tt.setupMocks(mockS3API)

			client := &sos.Client{
				S3Client: mockS3API,
				Zone:     "test-zone",
			}

			err := client.MoveObject(context.Background(), tt.srcBucket, tt.srcKey, tt.dstBucket, tt.dstKey, 1, false)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMoveObject_Multipart(t *testing.T) {
	t.Run("successful multipart move", func(t *testing.T) {
		mockS3API := &MockS3API{
			mockHeadObject: func(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error) {
				return &s3.HeadObjectOutput{
					ContentLength: aws.Int64(5*1024*1024*1024 + 1),
					Metadata:      map[string]string{"key": "value"},
				}, nil
			},
			mockGetObjectAcl: func(ctx context.Context, params *s3.GetObjectAclInput, optFns ...func(*s3.Options)) (*s3.GetObjectAclOutput, error) {
				return &s3.GetObjectAclOutput{}, nil
			},
			mockCreateMultipartUpload: func(ctx context.Context, params *s3.CreateMultipartUploadInput, optFns ...func(*s3.Options)) (*s3.CreateMultipartUploadOutput, error) {
				assert.Equal(t, "value", params.Metadata["key"], "metadata must be preserved on multipart uploads")
				return &s3.CreateMultipartUploadOutput{UploadId: aws.String("test-upload-id")}, nil
			},
			mockUploadPartCopy: func(ctx context.Context, params *s3.UploadPartCopyInput, optFns ...func(*s3.Options)) (*s3.UploadPartCopyOutput, error) {
				return &s3.UploadPartCopyOutput{
					CopyPartResult: &types.CopyPartResult{
						ETag: aws.String("test-etag"),
					},
				}, nil
			},
			mockCompleteMultipartUpload: func(ctx context.Context, params *s3.CompleteMultipartUploadInput, optFns ...func(*s3.Options)) (*s3.CompleteMultipartUploadOutput, error) {
				for i := 1; i < len(params.MultipartUpload.Parts); i++ {
					assert.Less(t, aws.ToInt32(params.MultipartUpload.Parts[i-1].PartNumber), aws.ToInt32(params.MultipartUpload.Parts[i].PartNumber),
						"parts must be sorted by PartNumber")
				}
				return &s3.CompleteMultipartUploadOutput{}, nil
			},
			mockDeleteObject: func(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
				return &s3.DeleteObjectOutput{}, nil
			},
		}

		client := &sos.Client{
			S3Client: mockS3API,
			Zone:     "test-zone",
		}

		err := client.MoveObject(context.Background(), "src-bucket", "large-file", "dst-bucket", "large-file", 2, false)
		assert.NoError(t, err)
	})

	t.Run("multipart abort on part copy failure", func(t *testing.T) {
		abortCalled := false
		mockS3API := &MockS3API{
			mockHeadObject: func(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error) {
				return &s3.HeadObjectOutput{ContentLength: aws.Int64(5*1024*1024*1024 + 1)}, nil
			},
			mockGetObjectAcl: func(ctx context.Context, params *s3.GetObjectAclInput, optFns ...func(*s3.Options)) (*s3.GetObjectAclOutput, error) {
				return &s3.GetObjectAclOutput{}, nil
			},
			mockCreateMultipartUpload: func(ctx context.Context, params *s3.CreateMultipartUploadInput, optFns ...func(*s3.Options)) (*s3.CreateMultipartUploadOutput, error) {
				return &s3.CreateMultipartUploadOutput{UploadId: aws.String("test-upload-id")}, nil
			},
			mockUploadPartCopy: func(ctx context.Context, params *s3.UploadPartCopyInput, optFns ...func(*s3.Options)) (*s3.UploadPartCopyOutput, error) {
				return nil, errors.New("part copy failed")
			},
			mockAbortMultipartUpload: func(ctx context.Context, params *s3.AbortMultipartUploadInput, optFns ...func(*s3.Options)) (*s3.AbortMultipartUploadOutput, error) {
				abortCalled = true
				return &s3.AbortMultipartUploadOutput{}, nil
			},
		}

		client := &sos.Client{
			S3Client: mockS3API,
			Zone:     "test-zone",
		}

		err := client.MoveObject(context.Background(), "src-bucket", "large-file", "dst-bucket", "large-file", 1, false)
		assert.Error(t, err)
		assert.True(t, abortCalled)
		assert.Contains(t, err.Error(), "upload failed")
	})
}
