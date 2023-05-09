package sos_test

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	s3manager "github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
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
				Zone:     "bern",
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
				assert.Equal(t, bucket, output.Bucket)
				assert.Equal(t, key, output.Path)
				assert.Equal(t, int64(100), output.Size)
				assert.Equal(t, fmt.Sprintf("https://sos-%s.exo.io/%s/%s", client.Zone, bucket, key), output.URL)
			}
		})
	}
}

func TestDeleteObjects(t *testing.T) {
	bucket := "test-bucket"
	commonPrefix := "myobjects/"
	objectKeys := []string{commonPrefix + "object1", commonPrefix + "object2", commonPrefix + "object3"}

	nCalls := 0
	expectedDeleteInput := &s3.DeleteObjectsInput{
		Bucket: &bucket,
		Delete: &types.Delete{
			Objects: []types.ObjectIdentifier{
				{Key: aws.String(objectKeys[0])},
				{Key: aws.String(objectKeys[1])},
				{Key: aws.String(objectKeys[2])},
			},
		},
	}
	client := sos.Client{
		S3Client: &MockS3API{
			mockDeleteObjects: func(ctx context.Context, params *s3.DeleteObjectsInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error) {
				nCalls++
				assert.Equal(t, expectedDeleteInput, params)
				return &s3.DeleteObjectsOutput{
					Deleted: []types.DeletedObject{
						{Key: aws.String(objectKeys[0])},
						{Key: aws.String(objectKeys[1])},
						{Key: aws.String(objectKeys[2])},
					},
				}, nil
			},
			mockListObjectsV2: func(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
				return &s3.ListObjectsV2Output{
					IsTruncated: false,
					Contents: []types.Object{
						{Key: aws.String(objectKeys[0])},
						{Key: aws.String(objectKeys[1])},
						{Key: aws.String(objectKeys[2])},
					},
				}, nil
			},
		},
	}

	deleted, err := client.DeleteObjects(context.Background(), bucket, commonPrefix, false)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(deleted))
	assert.Equal(t, 1, nCalls)

	for i, key := range deleted {
		assert.Equal(t, objectKeys[i], *key.Key)
	}

	client = sos.Client{
		S3Client: &MockS3API{
			mockDeleteObjects: func(ctx context.Context, params *s3.DeleteObjectsInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error) {
				return nil, errors.New("delete error")
			},
			mockListObjectsV2: func(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
				return &s3.ListObjectsV2Output{
					IsTruncated: false,
					Contents: []types.Object{
						{
							Key: aws.String("some-file"),
						},
					},
				}, nil
			},
		},
	}

	_, err = client.DeleteObjects(context.Background(), bucket, commonPrefix, false)
	assert.Error(t, err)
}

func TestListObjects(t *testing.T) {
	mockS3API := &MockS3API{
		mockListObjectsV2: func(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
			return &s3.ListObjectsV2Output{
				Contents: []s3types.Object{
					{
						Key:          aws.String("file1.txt"),
						Size:         100,
						LastModified: aws.Time(time.Now()),
					},
					{
						Key:          aws.String("file2.txt"),
						Size:         200,
						LastModified: aws.Time(time.Now()),
					},
				},
				CommonPrefixes: []s3types.CommonPrefix{
					{
						Prefix: aws.String("folder1/"),
					},
					{
						Prefix: aws.String("folder2/"),
					},
				},
				IsTruncated: false,
			}, nil
		},
	}

	// Create a new sos.Client instance with the mocked S3API
	client := sos.Client{
		S3Client: mockS3API,
		Zone:     "testzone",
	}

	// Call the ListObjects function
	bucket := "testbucket"
	prefix := ""
	recursive := false
	stream := false
	ctx := context.Background()
	output, err := client.ListObjects(ctx, bucket, prefix, recursive, stream)
	assert.NoError(t, err)

	// Define the expected output
	expectedOutput := &sos.ListObjectsOutput{
		{
			Path: "folder1/",
			Dir:  true,
		},
		{
			Path: "folder2/",
			Dir:  true,
		},
		{
			Path:         "file1.txt",
			Size:         100,
			LastModified: time.Now().Format(sos.TimestampFormat),
		},
		{
			Path:         "file2.txt",
			Size:         200,
			LastModified: time.Now().Format(sos.TimestampFormat),
		},
	}

	// Compare the output with the expected output
	assert.Equal(t, expectedOutput, output)
}

type MockUploader struct {
	tc *testCase

	t *testing.T
}

func (u MockUploader) Upload(ctx context.Context, input *s3.PutObjectInput, opts ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error) {
	uploadedContent, err := ioutil.ReadAll(input.Body)
	assert.NoError(u.t, err)

	assert.Equal(u.t, u.tc.content, string(uploadedContent))

	if u.tc.shouldErr {
		return nil, fmt.Errorf("should error")
	}
	return nil, nil
}

func NewMockUploaderFunc(t *testing.T, tc *testCase) func(client s3manager.UploadAPIClient, options ...func(*s3manager.Uploader)) sos.Uploader {
	return func(client s3manager.UploadAPIClient, options ...func(*s3manager.Uploader)) sos.Uploader {
		return &MockUploader{
			t:  t,
			tc: tc,
		}
	}
}

type testCase struct {
	name      string
	bucket    string
	file      string
	content   string
	key       string
	acl       string
	shouldErr bool
}

func TestUploadFile(t *testing.T) {
	testCases := []testCase{
		{
			name:      "successful upload",
			bucket:    "test-bucket",
			file:      "test-file.txt",
			content:   "test conent 1",
			key:       "test-key",
			acl:       "public-read",
			shouldErr: false,
		},
		{
			name:      "invalid ACL error",
			bucket:    "test-bucket",
			file:      "test-file.txt",
			content:   "test conent 2",
			key:       "test-key",
			acl:       "invalid-acl",
			shouldErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := sos.Client{
				S3Client:        &MockS3API{},
				NewUploaderFunc: NewMockUploaderFunc(t, &tc),
			}

			tempDir, err := ioutil.TempDir("", "exo-cli-test")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(tempDir)

			fileToUpload := tempDir + "/" + tc.file

			err = ioutil.WriteFile(fileToUpload, []byte(tc.content), fs.ModePerm)
			assert.NoError(t, err)

			err = client.UploadFile(context.Background(), tc.bucket, fileToUpload, tc.key, tc.acl)
			if tc.shouldErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCopyObject(t *testing.T) {
	returnEmptyMockGetObjectACL := func(ctx context.Context, input *s3.GetObjectAclInput, optFns ...func(*s3.Options)) (*s3.GetObjectAclOutput, error) {
		return &s3.GetObjectAclOutput{
			Owner: &types.Owner{
				ID: aws.String("christopher"),
			},
		}, nil
	}

	tests := []struct {
		name             string
		mockGetObject    func(ctx context.Context, input *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
		mockGetObjectACL func(ctx context.Context, input *s3.GetObjectAclInput, optFns ...func(*s3.Options)) (*s3.GetObjectAclOutput, error)
		expectError      bool
	}{
		{
			name: "successful copy object",
			mockGetObject: func(ctx context.Context, input *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
				return &s3.GetObjectOutput{}, nil
			},
			mockGetObjectACL: returnEmptyMockGetObjectACL,
			expectError:      false,
		},
		{
			name: "get object error",
			mockGetObject: func(ctx context.Context, input *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
				return nil, errors.New("get object error")
			},
			mockGetObjectACL: func(ctx context.Context, input *s3.GetObjectAclInput, optFns ...func(*s3.Options)) (*s3.GetObjectAclOutput, error) {
				return nil, nil
			},
			expectError: true,
		},
		{
			name: "get object acl error",
			mockGetObject: func(ctx context.Context, input *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
				return &s3.GetObjectOutput{}, nil
			},
			mockGetObjectACL: func(ctx context.Context, input *s3.GetObjectAclInput, optFns ...func(*s3.Options)) (*s3.GetObjectAclOutput, error) {
				return nil, errors.New("get object acl error")
			},
			expectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockS3API := MockS3API{
				mockGetObject:    test.mockGetObject,
				mockGetObjectAcl: test.mockGetObjectACL,
			}

			client := sos.Client{
				S3Client: &mockS3API,
			}

			ctx := context.Background()
			bucket := "test-bucket"
			key := "test-key"

			_, err := client.CopyObject(ctx, bucket, key)
			if test.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
