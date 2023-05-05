package sos_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/exoscale/cli/pkg/storage/sos"
	"github.com/stretchr/testify/assert"
)

func TestSetObjectACL(t *testing.T) {
	testBucket := "test-bucket"
	testKey := "test-key"
	testACL := &sos.ACL{
		Canned: "public-read",
	}

	testGranteeID := "a1b2c3"
	testGranteeDisplayName := "testDisplayName"

	mockGetObjectACL := func(ctx context.Context, input *s3.GetObjectAclInput, optFns ...func(*s3.Options)) (*s3.GetObjectAclOutput, error) {
		return &s3.GetObjectAclOutput{
			Owner: &s3types.Owner{
				ID:          aws.String(testGranteeID),
				DisplayName: aws.String(testGranteeDisplayName),
			},
		}, nil
	}

	t.Run("canned ACL", func(t *testing.T) {
		mockS3 := &MockS3API{}
		client := &sos.Client{
			S3Client: mockS3,
		}

		mockS3.mockGetObjectAcl = mockGetObjectACL

		putObjectACLCount := 0
		expectedACL2 := s3types.ObjectCannedACL(testACL.Canned)
		mockS3.mockPutObjectAcl = func(ctx context.Context, params *s3.PutObjectAclInput, optFns ...func(*s3.Options)) (*s3.PutObjectAclOutput, error) {
			putObjectACLCount++

			expectedParams := &s3.PutObjectAclInput{
				Bucket: &testBucket,
				Key:    &testKey,
				ACL:    expectedACL2,
			}
			assert.Equal(t, expectedParams, params)

			return &s3.PutObjectAclOutput{}, nil
		}

		err := client.SetObjectACL(context.Background(), testBucket, testKey, testACL)

		assert.NoError(t, err)
		assert.Equal(t, 1, putObjectACLCount, "PutObjectAcl should be called once")
	})
	t.Run("no canned ACL", func(t *testing.T) {
		mockS3 := &MockS3API{}
		client := &sos.Client{
			S3Client: mockS3,
		}

		mockS3.mockGetObjectAcl = mockGetObjectACL

		granteeID := "id=example-cid"
		acl := &sos.ACL{
			Canned:      "",
			Read:        granteeID,
			Write:       "",
			ReadACP:     "",
			WriteACP:    "",
			FullControl: "",
		}

		putObjectACLCount := 0
		mockS3.mockPutObjectAcl = func(ctx context.Context, params *s3.PutObjectAclInput, optFns ...func(*s3.Options)) (*s3.PutObjectAclOutput, error) {
			putObjectACLCount++

			expectedParams := &s3.PutObjectAclInput{Bucket: &testBucket,
				Key: &testKey,
				ACL: "",
				AccessControlPolicy: &s3types.AccessControlPolicy{
					Grants: []s3types.Grant{
						{
							Grantee: &s3types.Grantee{
								Type: s3types.TypeCanonicalUser,
								ID:   aws.String(granteeID),
							},
							Permission: s3types.PermissionRead,
						},
						{
							Grantee: &s3types.Grantee{
								Type: s3types.TypeCanonicalUser,
								ID:   aws.String(testGranteeID),
							},
							Permission: s3types.PermissionFullControl,
						},
					},
				},
			}

			assert.Equal(t, expectedParams, params)

			return &s3.PutObjectAclOutput{}, nil
		}

		err := client.SetObjectACL(context.Background(), testBucket, testKey, acl)

		assert.NoError(t, err)
		assert.Equal(t, 1, putObjectACLCount, "PutObjectAcl should be called once")
	})
}
