package sos_test

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/exoscale/cli/pkg/storage/sos"
	"github.com/exoscale/cli/pkg/storage/sos/object"
	"github.com/stretchr/testify/assert"
)

type testObject struct {
	Key          string
	Size         int64
	LastModified time.Time
}

func TestListObjects(t *testing.T) {
	now := time.Now()
	timeNow := func() time.Time { return now }
	bucket := "testbucket"

	testData := []struct {
		name           string
		prefix         string
		recursive      bool
		filters        []object.ObjectFilterFunc
		objects        []testObject
		commonPrefixes []string
		expected       *sos.ListObjectsOutput
	}{
		{
			name: "simple list",
			objects: []testObject{
				{
					Key:          "file1.txt",
					Size:         100,
					LastModified: timeNow(),
				},
				{
					Key:          "file2.txt",
					Size:         200,
					LastModified: timeNow(),
				},
			},
			commonPrefixes: []string{"folder1/", "folder2/"},
			expected: &sos.ListObjectsOutput{
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
					LastModified: timeNow().Format(sos.TimestampFormat),
				},
				{
					Path:         "file2.txt",
					Size:         200,
					LastModified: timeNow().Format(sos.TimestampFormat),
				},
			},
		},
		{
			name: "with timestamp filters",
			objects: []testObject{
				{
					Key:          "file1.txt",
					Size:         100,
					LastModified: timeNow(),
				},
				{
					Key:          "file2.txt",
					Size:         200,
					LastModified: timeNow().Add(-2 * time.Hour),
				},
				{
					Key:          "file3.txt",
					Size:         200,
					LastModified: timeNow().Add(-4 * time.Hour),
				},
			},
			commonPrefixes: []string{"folder1/", "folder2/"},
			filters: []object.ObjectFilterFunc{
				object.OlderThanFilterFunc(timeNow().Add(-time.Hour)),
				object.NewerThanFilterFunc(timeNow().Add(-3 * time.Hour)),
			},
			expected: &sos.ListObjectsOutput{
				{
					Path: "folder1/",
					Dir:  true,
				},
				{
					Path: "folder2/",
					Dir:  true,
				},
				{
					Path:         "file2.txt",
					Size:         200,
					LastModified: timeNow().Add(-2 * time.Hour).Format(sos.TimestampFormat),
				},
			},
		},
		{
			name:      "recursive",
			recursive: true,
			objects: []testObject{
				{
					Key:          "folder1/file1.txt",
					Size:         100,
					LastModified: timeNow(),
				},
				{
					Key:          "folder1/file2.txt",
					Size:         200,
					LastModified: timeNow(),
				},
				{
					Key:          "folder2/file3.txt",
					Size:         200,
					LastModified: timeNow(),
				},
			},
			commonPrefixes: []string{"folder1/", "folder2/"},
			expected: &sos.ListObjectsOutput{
				{
					Path:         "folder1/file1.txt",
					Size:         100,
					LastModified: timeNow().Format(sos.TimestampFormat),
				},
				{
					Path:         "folder1/file2.txt",
					Size:         200,
					LastModified: timeNow().Format(sos.TimestampFormat),
				},
				{
					Path:         "folder2/file3.txt",
					Size:         200,
					LastModified: timeNow().Format(sos.TimestampFormat),
				},
			},
		},
		{
			name:   "prefix",
			prefix: "folder1/",
			objects: []testObject{
				{
					Key:          "folder1/file1.txt",
					Size:         100,
					LastModified: timeNow(),
				},
				{
					Key:          "folder1/file2.txt",
					Size:         200,
					LastModified: timeNow(),
				},
			},
			commonPrefixes: []string{"folder1/"},
			expected: &sos.ListObjectsOutput{
				{
					Path: "folder1/",
					Dir:  true,
				},
				{
					Path:         "folder1/file1.txt",
					Size:         100,
					LastModified: timeNow().Format(sos.TimestampFormat),
				},
				{
					Path:         "folder1/file2.txt",
					Size:         200,
					LastModified: timeNow().Format(sos.TimestampFormat),
				},
			},
		},
	}

	for _, testCase := range testData {
		t.Run(testCase.name, func(t *testing.T) {
			mockS3API := &MockS3API{
				mockListObjectsV2: func(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
					contents := make([]s3types.Object, len(testCase.objects))
					for i, object := range testCase.objects {
						contents[i] = s3types.Object{
							Key:          aws.String(object.Key),
							Size:         object.Size,
							LastModified: aws.Time(object.LastModified),
						}
					}

					awsCommonPrefixes := make([]s3types.CommonPrefix, len(testCase.commonPrefixes))
					for i, prefix := range testCase.commonPrefixes {
						awsCommonPrefixes[i] = s3types.CommonPrefix{
							Prefix: aws.String(prefix),
						}
					}

					return &s3.ListObjectsV2Output{
						Contents:       contents,
						CommonPrefixes: awsCommonPrefixes,
						IsTruncated:    false,
					}, nil
				},
				mockListObjectVersions: func(ctx context.Context, params *s3.ListObjectVersionsInput, optFns ...func(*s3.Options)) (*s3.ListObjectVersionsOutput, error) {
					versions := make([]s3types.ObjectVersion, len(testCase.objects))
					for i, object := range testCase.objects {
						versions[i] = s3types.ObjectVersion{
							Key:          aws.String(object.Key),
							Size:         object.Size,
							LastModified: aws.Time(object.LastModified),
						}
					}

					awsCommonPrefixes := make([]s3types.CommonPrefix, len(testCase.commonPrefixes))
					for i, prefix := range testCase.commonPrefixes {
						awsCommonPrefixes[i] = s3types.CommonPrefix{
							Prefix: aws.String(prefix),
						}
					}

					return &s3.ListObjectVersionsOutput{
						Versions:       versions,
						CommonPrefixes: awsCommonPrefixes,
						IsTruncated:    false,
					}, nil
				},
			}

			client := sos.Client{
				S3Client: mockS3API,
				Zone:     "testzone",
			}

			prefix := ""
			stream := false
			ctx := context.Background()

			list := client.ListObjectsFunc(bucket, prefix, testCase.recursive, stream, testCase.filters)
			output, err := client.ListObjects(ctx, list, testCase.recursive, stream)
			assert.NoError(t, err)

			assert.Equal(t, testCase.expected, output)

			list = client.ListVersionedObjectsFunc(bucket, prefix, testCase.recursive, stream, testCase.filters, nil)
			versionedOutput, err := client.ListObjects(ctx, list, testCase.recursive, stream)
			assert.NoError(t, err)

			assert.Equal(t, testCase.expected, versionedOutput)
		})
	}
}
