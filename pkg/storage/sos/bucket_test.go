package sos_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/exoscale/cli/pkg/storage/sos"
	"github.com/stretchr/testify/assert"
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
