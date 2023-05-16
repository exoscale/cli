package sos_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/stretchr/testify/assert"

	"github.com/exoscale/cli/pkg/storage/sos"
)

func TestDeleteObjectHeaders(t *testing.T) {
	returnEmptyMockCopyObject := func(ctx context.Context, params *s3.CopyObjectInput, optFns ...func(*s3.Options)) (*s3.CopyObjectOutput, error) {
		return &s3.CopyObjectOutput{}, nil
	}

	tests := []struct {
		name           string
		bucket         string
		key            string
		headers        []string
		mockGetObject  func(ctx context.Context, input *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
		mockCopyObject func(ctx context.Context, params *s3.CopyObjectInput, optFns ...func(*s3.Options)) (*s3.CopyObjectOutput, error)
		expectErr      bool
	}{
		{
			name:    "successful deletion of headers",
			bucket:  "test-bucket",
			key:     "test-key",
			headers: []string{sos.ObjectHeaderCacheControl, sos.ObjectHeaderContentDisposition},
			mockGetObject: func(ctx context.Context, input *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
				return &s3.GetObjectOutput{
					CacheControl:       aws.String("private, max-age=0, no-cache"),
					ContentDisposition: aws.String("attachment; filename=\"filename.jpg\""),
				}, nil
			},
			mockCopyObject: func(ctx context.Context, params *s3.CopyObjectInput, optFns ...func(*s3.Options)) (*s3.CopyObjectOutput, error) {
				assert.Nil(t, params.CacheControl)
				assert.Nil(t, params.ContentDisposition)
				return &s3.CopyObjectOutput{}, nil
			},
			expectErr: false,
		},
		{
			name:    "error in GetObject",
			bucket:  "test-bucket",
			key:     "test-key",
			headers: []string{sos.ObjectHeaderCacheControl, sos.ObjectHeaderContentDisposition},
			mockGetObject: func(ctx context.Context, input *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
				return nil, errors.New("error getting object")
			},
			mockCopyObject: returnEmptyMockCopyObject,
			expectErr:      true,
		},
		{
			name:    "error in CopyObject",
			bucket:  "test-bucket",
			key:     "test-key",
			headers: []string{sos.ObjectHeaderCacheControl, sos.ObjectHeaderContentDisposition},
			mockGetObject: func(ctx context.Context, input *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
				return &s3.GetObjectOutput{
					CacheControl:       aws.String("private, max-age=0, no-cache"),
					ContentDisposition: aws.String("attachment; filename=\"filename.jpg\""),
				}, nil
			},
			mockCopyObject: func(ctx context.Context, params *s3.CopyObjectInput, optFns ...func(*s3.Options)) (*s3.CopyObjectOutput, error) {
				return nil, errors.New("error copying object")
			},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockS3 := &MockS3API{
				mockGetObject:  tc.mockGetObject,
				mockCopyObject: tc.mockCopyObject,
				mockGetObjectAcl: func(ctx context.Context, input *s3.GetObjectAclInput, optFns ...func(*s3.Options)) (*s3.GetObjectAclOutput, error) {
					return &s3.GetObjectAclOutput{
						Owner: &types.Owner{
							ID: aws.String("jack"),
						},
					}, nil
				},
			}

			client := &sos.Client{
				S3Client: mockS3,
			}

			err := client.DeleteObjectHeaders(context.Background(), tc.bucket, tc.key, tc.headers)
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUpdateObjectHeaders(t *testing.T) {
	testCases := []struct {
		name           string
		bucket         string
		key            string
		headers        map[string]*string
		mockCopyObject func(ctx context.Context, params *s3.CopyObjectInput, optFns ...func(*s3.Options)) (*s3.CopyObjectOutput, error)
		expectedErr    bool
	}{
		{
			name:   "successful update of headers",
			bucket: "test-bucket",
			key:    "test-key",
			headers: map[string]*string{
				"Cache-Control":       aws.String("max-age=3600"),
				"Content-Disposition": aws.String("attachment"),
			},
			mockCopyObject: func(ctx context.Context, params *s3.CopyObjectInput, optFns ...func(*s3.Options)) (*s3.CopyObjectOutput, error) {
				return &s3.CopyObjectOutput{}, nil
			},
			expectedErr: false,
		},
		{
			name:   "error in CopyObject",
			bucket: "test-bucket",
			key:    "test-key",
			headers: map[string]*string{
				"Cache-Control":       aws.String("max-age=3600"),
				"Content-Disposition": aws.String("attachment"),
			},
			mockCopyObject: func(ctx context.Context, params *s3.CopyObjectInput, optFns ...func(*s3.Options)) (*s3.CopyObjectOutput, error) {
				return nil, fmt.Errorf("mock CopyObject error")
			},
			expectedErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockS3 := &MockS3API{
				mockCopyObject: tc.mockCopyObject,
				mockGetObject: func(ctx context.Context, input *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
					return &s3.GetObjectOutput{}, nil
				},
				mockGetObjectAcl: func(ctx context.Context, input *s3.GetObjectAclInput, optFns ...func(*s3.Options)) (*s3.GetObjectAclOutput, error) {
					return &s3.GetObjectAclOutput{
						Owner: &types.Owner{
							ID: aws.String("mark"),
						},
					}, nil
				},
			}

			client := &sos.Client{
				S3Client: mockS3,
			}

			err := client.UpdateObjectHeaders(context.Background(), tc.bucket, tc.key, tc.headers)
			if tc.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
