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

func TestListObjects(t *testing.T) {
	bucket := "testbucket"

	testData := []struct {
		name              string
		prefix            string
		recursive         bool
		filters           []object.ObjectFilterFunc
		mockListObjectsV2 func(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error)
		expected          *sos.ListObjectsOutput
	}{
		{
			name: "simple list",
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
					LastModified: time.Now().Format(sos.TimestampFormat),
				},
				{
					Path:         "file2.txt",
					Size:         200,
					LastModified: time.Now().Format(sos.TimestampFormat),
				},
			},
		},
		{
			name: "with timestamp filters",
			mockListObjectsV2: func(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
				return &s3.ListObjectsV2Output{
					Contents: []s3types.Object{
						{
							Key:          aws.String("file1.txt"),
							Size:         100,
							LastModified: aws.Time(time.Now().Add(-time.Hour)),
						},
						{
							Key:          aws.String("file2.txt"),
							Size:         200,
							LastModified: aws.Time(time.Now().Add(-2 * time.Hour)),
						},
						{
							Key:          aws.String("file3.txt"),
							Size:         200,
							LastModified: aws.Time(time.Now().Add(-4 * time.Hour)),
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
			filters: []object.ObjectFilterFunc{
				object.OlderThanFilterFunc(time.Now().Add(-time.Hour)),
				object.NewerThanFilterFunc(time.Now().Add(-3 * time.Hour)),
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
					LastModified: time.Now().Add(-2 * time.Hour).Format(sos.TimestampFormat),
				},
			},
		},
		{
			name:      "recursive",
			recursive: true,
			mockListObjectsV2: func(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
				return &s3.ListObjectsV2Output{
					Contents: []s3types.Object{
						{
							Key:          aws.String("folder1/file1.txt"),
							Size:         100,
							LastModified: aws.Time(time.Now().Add(-time.Hour)),
						},
						{
							Key:          aws.String("folder1/file2.txt"),
							Size:         200,
							LastModified: aws.Time(time.Now().Add(-2 * time.Hour)),
						},
						{
							Key:          aws.String("folder2/file3.txt"),
							Size:         200,
							LastModified: aws.Time(time.Now().Add(-4 * time.Hour)),
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
			expected: &sos.ListObjectsOutput{
				{
					Path:         "folder1/file1.txt",
					Size:         100,
					LastModified: time.Now().Add(-time.Hour).Format(sos.TimestampFormat),
				},
				{
					Path:         "folder1/file2.txt",
					Size:         200,
					LastModified: time.Now().Add(-2 * time.Hour).Format(sos.TimestampFormat),
				},
				{
					Path:         "folder2/file3.txt",
					Size:         200,
					LastModified: time.Now().Add(-4 * time.Hour).Format(sos.TimestampFormat),
				},
			},
		},
		{
			name:   "prefix",
			prefix: "folder1/",
			mockListObjectsV2: func(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
				return &s3.ListObjectsV2Output{
					Contents: []s3types.Object{
						{
							Key:          aws.String("folder1/file1.txt"),
							Size:         100,
							LastModified: aws.Time(time.Now().Add(-time.Hour)),
						},
						{
							Key:          aws.String("folder1/file2.txt"),
							Size:         200,
							LastModified: aws.Time(time.Now().Add(-2 * time.Hour)),
						},
					},
					CommonPrefixes: []s3types.CommonPrefix{
						{
							Prefix: aws.String("folder1/"),
						},
					},
					IsTruncated: false,
				}, nil
			},
			expected: &sos.ListObjectsOutput{
				{
					Path: "folder1/",
					Dir:  true,
				},
				{
					Path:         "folder1/file1.txt",
					Size:         100,
					LastModified: time.Now().Add(-time.Hour).Format(sos.TimestampFormat),
				},
				{
					Path:         "folder1/file2.txt",
					Size:         200,
					LastModified: time.Now().Add(-2 * time.Hour).Format(sos.TimestampFormat),
				},
			},
		},
	}

	for _, testCase := range testData {
		t.Run(testCase.name, func(t *testing.T) {
			mockS3API := &MockS3API{
				mockListObjectsV2: testCase.mockListObjectsV2,
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
		})
	}
}
