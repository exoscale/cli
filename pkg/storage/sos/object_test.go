package sos_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	s3manager "github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hashicorp/go-multierror"
	"github.com/stretchr/testify/assert"

	"github.com/exoscale/cli/pkg/storage/sos"
)

func TestShowObject(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name           string
		getObjectFn    func(ctx context.Context, input *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
		getObjectACLFn func(ctx context.Context, input *s3.GetObjectAclInput, optFns ...func(*s3.Options)) (*s3.GetObjectAclOutput, error) //nolint:revive
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
			getObjectACLFn: func(ctx context.Context, input *s3.GetObjectAclInput, optFns ...func(*s3.Options)) (*s3.GetObjectAclOutput, error) {
				return &s3.GetObjectAclOutput{}, nil
			},
			expectErr: false,
		},
		{
			name: "failed to get object",
			getObjectFn: func(ctx context.Context, input *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
				return nil, errors.New("failed to get object")
			},
			getObjectACLFn: func(ctx context.Context, input *s3.GetObjectAclInput, optFns ...func(*s3.Options)) (*s3.GetObjectAclOutput, error) {
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
			getObjectACLFn: func(ctx context.Context, input *s3.GetObjectAclInput, optFns ...func(*s3.Options)) (*s3.GetObjectAclOutput, error) {
				return nil, errors.New("failed to get object ACL")
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockS3API := &MockS3API{
				mockGetObject:    tt.getObjectFn,
				mockGetObjectAcl: tt.getObjectACLFn,
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

	// Happy path
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

	// General error
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

	// Individual error in batch delete
	client = sos.Client{
		S3Client: &MockS3API{
			mockDeleteObjects: func(ctx context.Context, params *s3.DeleteObjectsInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error) {
				nCalls++
				assert.Equal(t, expectedDeleteInput, params)
				return &s3.DeleteObjectsOutput{
					Deleted: []types.DeletedObject{
						{Key: aws.String(objectKeys[0])},
						{Key: aws.String(objectKeys[2])},
					},
					Errors: []types.Error{
						{
							Code:      aws.String("AccessDenied"),
							Key:       aws.String("1"),
							Message:   aws.String("Access Denied"),
							VersionId: aws.String("1"),
						},
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
	deleted, err = client.DeleteObjects(context.Background(), bucket, commonPrefix, false)

	assert.Equal(t, 2, len(deleted))
	assert.Error(t, err)
	if merr, ok := err.(*multierror.Error); ok {
		assert.Equal(t, 1, len(merr.Errors))
		assert.Equal(t, "Access Denied", merr.Errors[0].Error())
	} else {
		assert.NoError(t, err)
	}
}

type MockUploader struct {
	tc *testCase

	t *testing.T
}

func (u MockUploader) Upload(ctx context.Context, input *s3.PutObjectInput, opts ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error) {
	uploadedContent, err := io.ReadAll(input.Body)
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

			tempDir, err := os.MkdirTemp("", "exo-cli-test")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(tempDir)

			fileToUpload := tempDir + "/" + tc.file

			err = os.WriteFile(fileToUpload, []byte(tc.content), fs.ModePerm)
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

type expectedUpload struct {
	done    bool
	content string
}

type MultiUploaderTestCase struct {
	toUpload             []string
	nExpectedUploadCalls int
	shouldErr            bool
	uploaderChecklist    map[string]expectedUpload
	dryRun               bool
	recursive            bool
}

type MockMultiUploader struct {
	t            *testing.T
	tc           *MultiUploaderTestCase
	nUploadCalls int
}

func (u *MockMultiUploader) Upload(ctx context.Context, input *s3.PutObjectInput, opts ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error) {
	u.nUploadCalls++

	uploadedContent, err := io.ReadAll(input.Body)
	assert.NoError(u.t, err)

	item, ok := u.tc.uploaderChecklist[*input.Key]
	assert.True(u.t, ok)
	if string(uploadedContent) == item.content {
		item.done = true
		u.tc.uploaderChecklist[*input.Key] = item
	}

	if u.tc.shouldErr {
		return nil, fmt.Errorf("should error")
	}
	return nil, nil
}

func TestUploadFiles(t *testing.T) {
	ctx := context.Background()

	tempDir, err := os.MkdirTemp("", "exo-cli-uploads-test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		os.RemoveAll(tempDir)
	})

	tempFile1 := filepath.Join(tempDir, "file1.txt")
	if err := os.WriteFile(tempFile1, []byte("file1 content"), 0600); err != nil {
		t.Fatal(err)
	}

	tempFile2 := filepath.Join(tempDir, "file2.txt")
	if err := os.WriteFile(tempFile2, []byte("file2 content"), 0600); err != nil {
		t.Fatal(err)
	}

	tempSubdir := filepath.Join(tempDir, "subdir")
	if err := os.Mkdir(tempSubdir, 0755); err != nil {
		t.Fatal(err)
	}

	tempFile3 := filepath.Join(tempSubdir, "file3.txt")
	if err := os.WriteFile(tempFile3, []byte("file3 content"), 0600); err != nil {
		t.Fatal(err)
	}

	mockS3API := &MockS3API{
		mockPutObject: func(ctx context.Context, input *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
			return &s3.PutObjectOutput{}, nil
		},
	}

	testCases := []struct {
		name string
		tc   *MultiUploaderTestCase
	}{
		{
			name: "single file upload",
			tc: &MultiUploaderTestCase{
				toUpload: []string{
					tempFile1,
				},
				uploaderChecklist: map[string]expectedUpload{
					"test-prefix/file1.txt": {
						content: "file1 content",
					},
				},
				nExpectedUploadCalls: 1,
			},
		},
		{
			name: "upload two files",
			tc: &MultiUploaderTestCase{
				toUpload: []string{
					tempFile1,
					tempFile2,
				},
				uploaderChecklist: map[string]expectedUpload{
					"test-prefix/file1.txt": {
						content: "file1 content",
					},
					"test-prefix/file2.txt": {
						content: "file2 content",
					},
				},
				nExpectedUploadCalls: 2,
			},
		},
		{
			name: "directory upload without recursive flag",
			tc: &MultiUploaderTestCase{
				toUpload: []string{
					tempDir,
				},
				dryRun:    false,
				recursive: false,
				shouldErr: true,
			},
		},
		{
			name: "directory upload with recursive flag",
			tc: &MultiUploaderTestCase{
				toUpload: []string{
					tempDir,
				},
				dryRun:    false,
				recursive: true,
				shouldErr: false,
				uploaderChecklist: map[string]expectedUpload{
					"test-prefix" + tempDir + "/file1.txt": {
						content: "file1 content",
					},
					"test-prefix" + tempDir + "/file2.txt": {
						content: "file2 content",
					},
					"test-prefix" + tempDir + "/subdir/file3.txt": {
						content: "file3 content",
					},
				},
				nExpectedUploadCalls: 3,
			},
		},
		{
			name: "dry run",
			tc: &MultiUploaderTestCase{
				toUpload: []string{
					tempFile1,
					tempFile2,
				},
				nExpectedUploadCalls: 0,
				dryRun:               true,
			},
		},
		{
			name: "error handling for non-existent file",
			tc: &MultiUploaderTestCase{
				toUpload: []string{
					"non-existent-file.txt",
				},
				shouldErr: true,
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			uploader := MockMultiUploader{
				t:  t,
				tc: tt.tc,
			}

			client := &sos.Client{
				S3Client: mockS3API,
				NewUploaderFunc: func(client s3manager.UploadAPIClient, options ...func(*s3manager.Uploader)) sos.Uploader {
					return &uploader
				},
			}

			err := client.UploadFiles(ctx, tt.tc.toUpload, &sos.StorageUploadConfig{
				Bucket:    "test-bucket",
				Prefix:    "test-prefix/",
				ACL:       "private",
				DryRun:    tt.tc.dryRun,
				Recursive: tt.tc.recursive,
			})
			if tt.tc.shouldErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.tc.nExpectedUploadCalls, uploader.nUploadCalls)

			for _, item := range tt.tc.uploaderChecklist {
				assert.True(t, item.done)
			}
		})
	}
}

func Test_IsTraversalPath(t *testing.T) {
	tests := []struct {
		path   string
		expect bool
	}{
		{
			path:   "test.txt",
			expect: false,
		},
		{
			path:   "../test.txt",
			expect: true,
		},
		{
			path:   "a/b/../../../test.text",
			expect: true,
		},
		{
			path:   "a/b/../../test.txt",
			expect: false,
		},
		{
			path:   "../a/b/test.txt",
			expect: true,
		},
	}

	for _, ut := range tests {
		t.Run(ut.path, func(t *testing.T) {
			assert.Equal(t, ut.expect, sos.IsTraversalPath(ut.path))
		})
	}
}

// drainDeleteObjectVersions collects all results from the channels returned by
// DeleteObjectVersions and returns the deleted objects and errors.
func drainDeleteObjectVersions(deletedChan <-chan types.DeletedObject, errChan <-chan error) ([]types.DeletedObject, []error) {
	var deleted []types.DeletedObject
	var errs []error
	for deletedChan != nil || errChan != nil {
		select {
		case d, ok := <-deletedChan:
			if !ok {
				deletedChan = nil
				continue
			}
			deleted = append(deleted, d)
		case e, ok := <-errChan:
			if !ok {
				errChan = nil
				continue
			}
			errs = append(errs, e)
		}
	}
	return deleted, errs
}

func TestDeleteObjectVersions_HappyPath(t *testing.T) {
	bucket := "test-bucket"
	listCalls := 0

	client := sos.Client{
		S3Client: &MockS3API{
			mockListObjectVersions: func(_ context.Context, params *s3.ListObjectVersionsInput, _ ...func(*s3.Options)) (*s3.ListObjectVersionsOutput, error) {
				listCalls++
				if listCalls > 1 {
					// Second call: nothing left.
					return &s3.ListObjectVersionsOutput{}, nil
				}
				return &s3.ListObjectVersionsOutput{
					IsTruncated: false,
					Versions: []types.ObjectVersion{
						{Key: aws.String("obj1"), VersionId: aws.String("v1")},
						{Key: aws.String("obj2"), VersionId: aws.String("v2")},
					},
				}, nil
			},
			mockDeleteObjects: func(_ context.Context, params *s3.DeleteObjectsInput, _ ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error) {
				return &s3.DeleteObjectsOutput{
					Deleted: []types.DeletedObject{
						{Key: aws.String("obj1"), VersionId: aws.String("v1")},
						{Key: aws.String("obj2"), VersionId: aws.String("v2")},
					},
				}, nil
			},
		},
	}

	deleted, errs := drainDeleteObjectVersions(client.DeleteObjectVersions(context.Background(), bucket, "/"))
	assert.Empty(t, errs)
	assert.Len(t, deleted, 2)
	assert.Equal(t, 2, listCalls, "should list twice: once for the batch, once to confirm empty")
}

// TestDeleteObjectVersions_ComplianceLock reproduces the customer bug: objects
// under compliance retention cause DeleteObjects to return per-object errors
// with an empty Deleted slice. Before the fix the goroutine looped forever
// re-listing the same objects.
func TestDeleteObjectVersions_ComplianceLock(t *testing.T) {
	bucket := "locked-bucket"
	listCalls := 0
	deleteCalls := 0

	client := sos.Client{
		S3Client: &MockS3API{
			mockListObjectVersions: func(_ context.Context, _ *s3.ListObjectVersionsInput, _ ...func(*s3.Options)) (*s3.ListObjectVersionsOutput, error) {
				listCalls++
				return &s3.ListObjectVersionsOutput{
					IsTruncated: false,
					Versions: []types.ObjectVersion{
						{Key: aws.String("locked-obj"), VersionId: aws.String("v1")},
					},
				}, nil
			},
			mockDeleteObjects: func(_ context.Context, _ *s3.DeleteObjectsInput, _ ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error) {
				deleteCalls++
				// S3 returns HTTP 200 but with a per-object error; Deleted is empty.
				return &s3.DeleteObjectsOutput{
					Errors: []types.Error{
						{
							Key:       aws.String("locked-obj"),
							VersionId: aws.String("v1"),
							Code:      aws.String("ObjectLocked"),
							Message:   aws.String("Object is under compliance retention"),
						},
					},
				}, nil
			},
		},
	}

	deleted, errs := drainDeleteObjectVersions(client.DeleteObjectVersions(context.Background(), bucket, "/"))
	assert.Empty(t, deleted)
	assert.Len(t, errs, 1)
	assert.Contains(t, errs[0].Error(), "delete error:")
	// Without the fix these would keep climbing; with the fix exactly one
	// list+delete attempt is made before giving up.
	assert.Equal(t, 1, listCalls, "should stop after first failed batch")
	assert.Equal(t, 1, deleteCalls, "should stop after first failed batch")
}

// TestDeleteObjectVersions_Pagination verifies that a truncated
// ListObjectVersions response is followed up with the correct KeyMarker and
// VersionIdMarker on the next call. Before the fix pagination markers were
// never forwarded, so only the first page was ever processed.
func TestDeleteObjectVersions_Pagination(t *testing.T) {
	bucket := "big-bucket"
	listCalls := 0

	client := sos.Client{
		S3Client: &MockS3API{
			mockListObjectVersions: func(_ context.Context, params *s3.ListObjectVersionsInput, _ ...func(*s3.Options)) (*s3.ListObjectVersionsOutput, error) {
				listCalls++
				switch listCalls {
				case 1:
					// First page: truncated.
					assert.Empty(t, aws.ToString(params.KeyMarker))
					assert.Empty(t, aws.ToString(params.VersionIdMarker))
					return &s3.ListObjectVersionsOutput{
						IsTruncated:         true,
						NextKeyMarker:       aws.String("obj1"),
						NextVersionIdMarker: aws.String("v1"),
						Versions: []types.ObjectVersion{
							{Key: aws.String("obj1"), VersionId: aws.String("v1")},
						},
					}, nil
				case 2:
					// Second page: markers must be set.
					assert.Equal(t, "obj1", aws.ToString(params.KeyMarker))
					assert.Equal(t, "v1", aws.ToString(params.VersionIdMarker))
					return &s3.ListObjectVersionsOutput{
						IsTruncated: false,
						Versions: []types.ObjectVersion{
							{Key: aws.String("obj2"), VersionId: aws.String("v2")},
						},
					}, nil
				default:
					// Third call: empty, signals end.
					return &s3.ListObjectVersionsOutput{}, nil
				}
			},
			mockDeleteObjects: func(_ context.Context, params *s3.DeleteObjectsInput, _ ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error) {
				deleted := make([]types.DeletedObject, 0, len(params.Delete.Objects))
				for _, o := range params.Delete.Objects {
					deleted = append(deleted, types.DeletedObject{Key: o.Key, VersionId: o.VersionId})
				}
				return &s3.DeleteObjectsOutput{Deleted: deleted}, nil
			},
		},
	}

	deleted, errs := drainDeleteObjectVersions(client.DeleteObjectVersions(context.Background(), bucket, "/"))
	assert.Empty(t, errs)
	assert.Len(t, deleted, 2, "both pages should be processed")
	assert.Equal(t, 3, listCalls, "should call list three times: page1, page2, confirm empty")
}
