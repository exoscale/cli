package sos_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
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
			}
		})
	}
}
