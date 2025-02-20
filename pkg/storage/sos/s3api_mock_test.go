package sos_test

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type MockS3API struct {
	mockGetObject                  func(ctx context.Context, input *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	mockGetObjectAcl               func(ctx context.Context, input *s3.GetObjectAclInput, optFns ...func(*s3.Options)) (*s3.GetObjectAclOutput, error) //nolint:stylecheck
	mockDeleteObjects              func(ctx context.Context, params *s3.DeleteObjectsInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error)
	mockListObjectsV2              func(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error)
	mockListObjectVersions         func(ctx context.Context, params *s3.ListObjectVersionsInput, optFns ...func(*s3.Options)) (*s3.ListObjectVersionsOutput, error)
	mockCopyObject                 func(ctx context.Context, params *s3.CopyObjectInput, optFns ...func(*s3.Options)) (*s3.CopyObjectOutput, error)
	mockDeleteBucketCors           func(ctx context.Context, params *s3.DeleteBucketCorsInput, optFns ...func(*s3.Options)) (*s3.DeleteBucketCorsOutput, error)
	mockGetBucketCors              func(ctx context.Context, params *s3.GetBucketCorsInput, optFns ...func(*s3.Options)) (*s3.GetBucketCorsOutput, error)
	mockPutBucketCors              func(ctx context.Context, params *s3.PutBucketCorsInput, optFns ...func(*s3.Options)) (*s3.PutBucketCorsOutput, error)
	mockPutBucketAcl               func(ctx context.Context, params *s3.PutBucketAclInput, optFns ...func(*s3.Options)) (*s3.PutBucketAclOutput, error) //nolint:stylecheck
	mockPutObjectAcl               func(ctx context.Context, params *s3.PutObjectAclInput, optFns ...func(*s3.Options)) (*s3.PutObjectAclOutput, error) //nolint:stylecheck
	mockCreateBucket               func(ctx context.Context, params *s3.CreateBucketInput, optFns ...func(*s3.Options)) (*s3.CreateBucketOutput, error)
	mockGetBucketAcl               func(ctx context.Context, params *s3.GetBucketAclInput, optFns ...func(*s3.Options)) (*s3.GetBucketAclOutput, error) //nolint:stylecheck
	mockGetBucketOwnershipControls func(ctx context.Context, params *s3.GetBucketOwnershipControlsInput, optFns ...func(*s3.Options)) (*s3.GetBucketOwnershipControlsOutput, error)
	mockPutBucketOwnershipControls func(ctx context.Context, params *s3.PutBucketOwnershipControlsInput, optFns ...func(*s3.Options)) (*s3.PutBucketOwnershipControlsOutput, error)
	mockGetBucketVersioning        func(ctx context.Context, params *s3.GetBucketVersioningInput, optFns ...func(*s3.Options)) (*s3.GetBucketVersioningOutput, error)
	mockPutBucketVersioning        func(ctx context.Context, params *s3.PutBucketVersioningInput, optFns ...func(*s3.Options)) (*s3.PutBucketVersioningOutput, error)
	mockListMultipartUploads       func(ctx context.Context, params *s3.ListMultipartUploadsInput, optFns ...func(*s3.Options)) (*s3.ListMultipartUploadsOutput, error)
	mockDeleteBucket               func(ctx context.Context, params *s3.DeleteBucketInput, optFns ...func(*s3.Options)) (*s3.DeleteBucketOutput, error)
	mockGetBucketReplication       func(ctx context.Context, params *s3.GetBucketReplicationInput, optFns ...func(*s3.Options)) (*s3.GetBucketReplicationOutput, error)
	mockPutBucketReplication       func(ctx context.Context, params *s3.PutBucketReplicationInput, optFns ...func(*s3.Options)) (*s3.PutBucketReplicationOutput, error)
	mockDeleteBucketReplication    func(ctx context.Context, params *s3.DeleteBucketReplicationInput, optFns ...func(*s3.Options)) (*s3.DeleteBucketReplicationOutput, error)

	// s3manager.UploadAPIClient
	mockPutObject               func(context.Context, *s3.PutObjectInput, ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	mockUploadPart              func(context.Context, *s3.UploadPartInput, ...func(*s3.Options)) (*s3.UploadPartOutput, error)
	mockCreateMultipartUpload   func(context.Context, *s3.CreateMultipartUploadInput, ...func(*s3.Options)) (*s3.CreateMultipartUploadOutput, error)
	mockCompleteMultipartUpload func(context.Context, *s3.CompleteMultipartUploadInput, ...func(*s3.Options)) (*s3.CompleteMultipartUploadOutput, error)
	mockAbortMultipartUpload    func(ctx context.Context, params *s3.AbortMultipartUploadInput, optFns ...func(*s3.Options)) (*s3.AbortMultipartUploadOutput, error)
}

func (m *MockS3API) GetObject(ctx context.Context, input *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	return m.mockGetObject(ctx, input, optFns...)
}

func (m *MockS3API) GetObjectAcl(ctx context.Context, input *s3.GetObjectAclInput, optFns ...func(*s3.Options)) (*s3.GetObjectAclOutput, error) { //nolint:stylecheck
	return m.mockGetObjectAcl(ctx, input, optFns...)
}

func (m *MockS3API) DeleteObjects(ctx context.Context, params *s3.DeleteObjectsInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error) {
	return m.mockDeleteObjects(ctx, params, optFns...)

}

func (m *MockS3API) ListObjectsV2(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
	return m.mockListObjectsV2(ctx, params, optFns...)
}

func (m *MockS3API) ListObjectVersions(ctx context.Context, params *s3.ListObjectVersionsInput, optFns ...func(*s3.Options)) (*s3.ListObjectVersionsOutput, error) {
	return m.mockListObjectVersions(ctx, params, optFns...)
}

func (m *MockS3API) CopyObject(ctx context.Context, params *s3.CopyObjectInput, optFns ...func(*s3.Options)) (*s3.CopyObjectOutput, error) {
	return m.mockCopyObject(ctx, params, optFns...)
}

func (m *MockS3API) DeleteBucketCors(ctx context.Context, params *s3.DeleteBucketCorsInput, optFns ...func(*s3.Options)) (*s3.DeleteBucketCorsOutput, error) {
	return m.mockDeleteBucketCors(ctx, params, optFns...)
}

func (m *MockS3API) GetBucketCors(ctx context.Context, params *s3.GetBucketCorsInput, optFns ...func(*s3.Options)) (*s3.GetBucketCorsOutput, error) {
	return m.mockGetBucketCors(ctx, params, optFns...)
}

func (m *MockS3API) PutBucketCors(ctx context.Context, params *s3.PutBucketCorsInput, optFns ...func(*s3.Options)) (*s3.PutBucketCorsOutput, error) {
	return m.mockPutBucketCors(ctx, params, optFns...)
}

func (m *MockS3API) PutBucketAcl(ctx context.Context, params *s3.PutBucketAclInput, optFns ...func(*s3.Options)) (*s3.PutBucketAclOutput, error) { //nolint:stylecheck
	return m.mockPutBucketAcl(ctx, params, optFns...)
}

func (m *MockS3API) PutObjectAcl(ctx context.Context, params *s3.PutObjectAclInput, optFns ...func(*s3.Options)) (*s3.PutObjectAclOutput, error) { //nolint:stylecheck
	return m.mockPutObjectAcl(ctx, params, optFns...)
}

func (m *MockS3API) CreateBucket(ctx context.Context, params *s3.CreateBucketInput, optFns ...func(*s3.Options)) (*s3.CreateBucketOutput, error) {
	return m.mockCreateBucket(ctx, params, optFns...)
}

func (m *MockS3API) GetBucketAcl(ctx context.Context, params *s3.GetBucketAclInput, optFns ...func(*s3.Options)) (*s3.GetBucketAclOutput, error) { //nolint:stylecheck
	return m.mockGetBucketAcl(ctx, params, optFns...)
}

func (m *MockS3API) GetBucketOwnershipControls(ctx context.Context, params *s3.GetBucketOwnershipControlsInput, optFns ...func(*s3.Options)) (*s3.GetBucketOwnershipControlsOutput, error) {
	return m.mockGetBucketOwnershipControls(ctx, params, optFns...)
}

func (m *MockS3API) PutBucketOwnershipControls(ctx context.Context, params *s3.PutBucketOwnershipControlsInput, optFns ...func(*s3.Options)) (*s3.PutBucketOwnershipControlsOutput, error) {
	return m.mockPutBucketOwnershipControls(ctx, params, optFns...)
}

func (m *MockS3API) GetBucketVersioning(ctx context.Context, params *s3.GetBucketVersioningInput, optFns ...func(*s3.Options)) (*s3.GetBucketVersioningOutput, error) {
	return m.mockGetBucketVersioning(ctx, params, optFns...)
}

func (m *MockS3API) PutBucketVersioning(ctx context.Context, params *s3.PutBucketVersioningInput, optFns ...func(*s3.Options)) (*s3.PutBucketVersioningOutput, error) {
	return m.mockPutBucketVersioning(ctx, params, optFns...)
}

func (m *MockS3API) ListMultipartUploads(ctx context.Context, params *s3.ListMultipartUploadsInput, optFns ...func(*s3.Options)) (*s3.ListMultipartUploadsOutput, error) {
	return m.mockListMultipartUploads(ctx, params, optFns...)
}

func (m *MockS3API) DeleteBucket(ctx context.Context, params *s3.DeleteBucketInput, optFns ...func(*s3.Options)) (*s3.DeleteBucketOutput, error) {
	return m.mockDeleteBucket(ctx, params, optFns...)
}

func (m *MockS3API) PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	return m.mockPutObject(ctx, params, optFns...)
}

func (m *MockS3API) UploadPart(ctx context.Context, params *s3.UploadPartInput, optFns ...func(*s3.Options)) (*s3.UploadPartOutput, error) {
	return m.mockUploadPart(ctx, params, optFns...)
}

func (m *MockS3API) CreateMultipartUpload(ctx context.Context, params *s3.CreateMultipartUploadInput, optFns ...func(*s3.Options)) (*s3.CreateMultipartUploadOutput, error) {
	return m.mockCreateMultipartUpload(ctx, params, optFns...)
}

func (m *MockS3API) CompleteMultipartUpload(ctx context.Context, params *s3.CompleteMultipartUploadInput, optFns ...func(*s3.Options)) (*s3.CompleteMultipartUploadOutput, error) {
	return m.mockCompleteMultipartUpload(ctx, params, optFns...)
}

func (m *MockS3API) AbortMultipartUpload(ctx context.Context, params *s3.AbortMultipartUploadInput, optFns ...func(*s3.Options)) (*s3.AbortMultipartUploadOutput, error) {
	return m.mockAbortMultipartUpload(ctx, params, optFns...)
}

func (m *MockS3API) GetBucketReplication(ctx context.Context, params *s3.GetBucketReplicationInput, optFns ...func(*s3.Options)) (*s3.GetBucketReplicationOutput, error) {
	return m.mockGetBucketReplication(ctx, params, optFns...)
}

func (m *MockS3API) PutBucketReplication(ctx context.Context, params *s3.PutBucketReplicationInput, optFns ...func(*s3.Options)) (*s3.PutBucketReplicationOutput, error) {
	return m.mockPutBucketReplication(ctx, params, optFns...)
}

func (m *MockS3API) DeleteBucketReplication(ctx context.Context, params *s3.DeleteBucketReplicationInput, optFns ...func(*s3.Options)) (*s3.DeleteBucketReplicationOutput, error) {
	return m.mockDeleteBucketReplication(ctx, params, optFns...)
}
