package sos

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hashicorp/go-multierror"
)

const (
	maxSingleCopyObjectSize         int64 = 5 * 1024 * 1024 * 1024
	defaultMultipartCopyConcurrency       = 5
)

type MovedObject struct {
	SourceBucket      string
	SourceKey         string
	DestinationBucket string
	DestinationKey    string
}

type StorageMoveConfig struct {
	Recursive                bool
	DryRun                   bool
	MultipartCopyConcurrency int
}

type moveSourceObject struct {
	CopySource              string
	ContentLength           int64
	Metadata                map[string]string
	ETag                    *string
	CacheControl            *string
	ContentDisposition      *string
	ContentEncoding         *string
	ContentLanguage         *string
	ContentType             *string
	Expires                 *time.Time
	WebsiteRedirectLocation *string
	StorageClass            s3types.StorageClass
	ServerSideEncryption    s3types.ServerSideEncryption
	SSEKMSKeyID             *string
	BucketKeyEnabled        bool
	ACL                     *s3.GetObjectAclOutput
}

type multipartCopyPart struct {
	Index      int
	PartNumber int32
	Start      int64
	End        int64
}

func (c *Client) MoveObjects(
	ctx context.Context,
	srcClient *Client,
	srcBucket, srcKey, dstBucket, dstKey string,
	config *StorageMoveConfig,
) ([]MovedObject, error) {
	config = normalizeStorageMoveConfig(config)

	if srcKey == "" {
		return nil, fmt.Errorf("source must include an object or prefix")
	}

	srcIsPrefix := srcKey == "/" || strings.HasSuffix(srcKey, "/")
	if !srcIsPrefix && config.Recursive {
		return nil, fmt.Errorf("source %q is an object, remove flag `-r` or suffix the source with `/` to move a prefix", srcKey)
	}

	if srcIsPrefix {
		if dstKey != "" && dstKey != "/" && !strings.HasSuffix(dstKey, "/") {
			return nil, fmt.Errorf("moving a prefix requires the destination to end with `/`")
		}

		sourceKeys, err := srcClient.listMoveSourceKeys(ctx, srcBucket, srcKey, config.Recursive)
		if err != nil {
			return nil, err
		}

		moved := make([]MovedObject, 0, len(sourceKeys))
		var errs *multierror.Error

		for _, sourceKey := range sourceKeys {
			destinationKey := resolveMovePrefixDestinationKey(srcKey, sourceKey, dstKey)
			mapping := MovedObject{
				SourceBucket:      srcBucket,
				SourceKey:         sourceKey,
				DestinationBucket: dstBucket,
				DestinationKey:    destinationKey,
			}

			if moveLocationsEqual(mapping) {
				errs = multierror.Append(errs, fmt.Errorf("source and destination are identical: %s/%s", srcBucket, sourceKey))
				continue
			}

			if !config.DryRun {
				if err := c.moveObject(ctx, srcClient, mapping, config.MultipartCopyConcurrency); err != nil {
					errs = multierror.Append(errs, err)
					continue
				}
			}

			moved = append(moved, mapping)
		}

		return moved, errs.ErrorOrNil()
	}

	destinationKey := resolveMoveObjectDestinationKey(srcKey, dstKey)
	mapping := MovedObject{
		SourceBucket:      srcBucket,
		SourceKey:         srcKey,
		DestinationBucket: dstBucket,
		DestinationKey:    destinationKey,
	}

	if moveLocationsEqual(mapping) {
		return nil, fmt.Errorf("source and destination are identical: %s/%s", srcBucket, srcKey)
	}

	if config.DryRun {
		if _, err := srcClient.describeMoveSourceObject(ctx, srcBucket, srcKey); err != nil {
			return nil, err
		}
		return []MovedObject{mapping}, nil
	}

	if err := c.moveObject(ctx, srcClient, mapping, config.MultipartCopyConcurrency); err != nil {
		return nil, err
	}

	return []MovedObject{mapping}, nil
}

func (c *Client) listMoveSourceKeys(ctx context.Context, bucket, prefix string, recursive bool) ([]string, error) {
	keys := make([]string, 0)
	if err := c.ForEachObject(ctx, bucket, prefix, recursive, func(o *s3types.Object) error {
		keys = append(keys, aws.ToString(o.Key))
		return nil
	}); err != nil {
		return nil, fmt.Errorf("error listing objects to move: %w", err)
	}

	return keys, nil
}

func (c *Client) moveObject(ctx context.Context, srcClient *Client, move MovedObject, multipartCopyConcurrency int) error {
	sourceObject, err := srcClient.describeMoveSourceObject(ctx, move.SourceBucket, move.SourceKey)
	if err != nil {
		return err
	}

	if sourceObject.ContentLength <= maxSingleCopyObjectSize {
		if err := c.copyObjectToDestination(ctx, move, sourceObject); err != nil {
			return err
		}
	} else {
		if err := c.copyMultipartObjectToDestination(ctx, move, sourceObject, multipartCopyConcurrency); err != nil {
			return err
		}
	}

	if err := srcClient.deleteMovedObject(ctx, move.SourceBucket, move.SourceKey); err != nil {
		return fmt.Errorf("delete source object %s/%s: %w", move.SourceBucket, move.SourceKey, err)
	}

	return nil
}

func (c *Client) describeMoveSourceObject(ctx context.Context, bucket, key string) (*moveSourceObject, error) {
	head, err := c.S3Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve object information: %w", err)
	}

	acl, err := c.S3Client.GetObjectAcl(ctx, &s3.GetObjectAclInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve object ACL: %w", err)
	}

	return &moveSourceObject{
		CopySource:              encodeCopySource(bucket, key, head.VersionId),
		ContentLength:           head.ContentLength,
		Metadata:                head.Metadata,
		ETag:                    head.ETag,
		CacheControl:            head.CacheControl,
		ContentDisposition:      head.ContentDisposition,
		ContentEncoding:         head.ContentEncoding,
		ContentLanguage:         head.ContentLanguage,
		ContentType:             head.ContentType,
		Expires:                 head.Expires,
		WebsiteRedirectLocation: head.WebsiteRedirectLocation,
		StorageClass:            head.StorageClass,
		ServerSideEncryption:    head.ServerSideEncryption,
		SSEKMSKeyID:             head.SSEKMSKeyId,
		BucketKeyEnabled:        head.BucketKeyEnabled,
		ACL:                     acl,
	}, nil
}

func (c *Client) copyObjectToDestination(ctx context.Context, move MovedObject, source *moveSourceObject) error {
	copyInput := &s3.CopyObjectInput{
		Bucket:                  aws.String(move.DestinationBucket),
		Key:                     aws.String(move.DestinationKey),
		CopySource:              aws.String(source.CopySource),
		Metadata:                source.Metadata,
		MetadataDirective:       s3types.MetadataDirectiveReplace,
		CacheControl:            source.CacheControl,
		ContentDisposition:      source.ContentDisposition,
		ContentEncoding:         source.ContentEncoding,
		ContentLanguage:         source.ContentLanguage,
		ContentType:             source.ContentType,
		Expires:                 source.Expires,
		WebsiteRedirectLocation: source.WebsiteRedirectLocation,
		BucketKeyEnabled:        source.BucketKeyEnabled,
		SSEKMSKeyId:             source.SSEKMSKeyID,
		ServerSideEncryption:    source.ServerSideEncryption,
	}

	if source.ETag != nil {
		copyInput.CopySourceIfMatch = source.ETag
	}
	if source.StorageClass != "" {
		copyInput.StorageClass = source.StorageClass
	}

	storageACLToCopyObject(source.ACL, copyInput)

	if _, err := c.S3Client.CopyObject(ctx, copyInput); err != nil {
		return fmt.Errorf("copy object to %s/%s: %w", move.DestinationBucket, move.DestinationKey, err)
	}

	return nil
}

func (c *Client) copyMultipartObjectToDestination(ctx context.Context, move MovedObject, source *moveSourceObject, multipartCopyConcurrency int) error {
	createInput := &s3.CreateMultipartUploadInput{
		Bucket:                  aws.String(move.DestinationBucket),
		Key:                     aws.String(move.DestinationKey),
		Metadata:                source.Metadata,
		CacheControl:            source.CacheControl,
		ContentDisposition:      source.ContentDisposition,
		ContentEncoding:         source.ContentEncoding,
		ContentLanguage:         source.ContentLanguage,
		ContentType:             source.ContentType,
		Expires:                 source.Expires,
		WebsiteRedirectLocation: source.WebsiteRedirectLocation,
		BucketKeyEnabled:        source.BucketKeyEnabled,
		SSEKMSKeyId:             source.SSEKMSKeyID,
		ServerSideEncryption:    source.ServerSideEncryption,
	}

	if source.StorageClass != "" {
		createInput.StorageClass = source.StorageClass
	}

	storageACLToCreateMultipartUpload(source.ACL, createInput)

	createOutput, err := c.S3Client.CreateMultipartUpload(ctx, createInput)
	if err != nil {
		return fmt.Errorf("create multipart upload for %s/%s: %w", move.DestinationBucket, move.DestinationKey, err)
	}

	uploadID := createOutput.UploadId
	partSize := estimateMultipartCopyPartSize(source.ContentLength)
	parts := buildMultipartCopyParts(source.ContentLength, partSize)
	completedParts := make([]s3types.CompletedPart, len(parts))
	copyCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	jobs := make(chan multipartCopyPart)
	var (
		completedPartsMu sync.Mutex
		copyErrMu        sync.Mutex
		firstCopyErr     error
		wg               sync.WaitGroup
	)

	workerCount := multipartCopyConcurrency
	if workerCount > len(parts) {
		workerCount = len(parts)
	}

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for {
				select {
				case <-copyCtx.Done():
					return

				case part, ok := <-jobs:
					if !ok {
						return
					}

					uploadPartInput := &s3.UploadPartCopyInput{
						Bucket:          aws.String(move.DestinationBucket),
						Key:             aws.String(move.DestinationKey),
						UploadId:        uploadID,
						PartNumber:      part.PartNumber,
						CopySource:      aws.String(source.CopySource),
						CopySourceRange: aws.String(fmt.Sprintf("bytes=%d-%d", part.Start, part.End)),
					}

					if source.ETag != nil {
						uploadPartInput.CopySourceIfMatch = source.ETag
					}

					uploadPartOutput, err := c.S3Client.UploadPartCopy(copyCtx, uploadPartInput)
					if err != nil {
						setMultipartCopyError(&copyErrMu, &firstCopyErr, fmt.Errorf("copy object part %d to %s/%s: %w", part.PartNumber, move.DestinationBucket, move.DestinationKey, err))
						cancel()
						return
					}
					if uploadPartOutput.CopyPartResult == nil {
						setMultipartCopyError(&copyErrMu, &firstCopyErr, fmt.Errorf("copy object part %d to %s/%s: empty copy result", part.PartNumber, move.DestinationBucket, move.DestinationKey))
						cancel()
						return
					}

					completedPart := s3types.CompletedPart{
						ETag:       uploadPartOutput.CopyPartResult.ETag,
						PartNumber: part.PartNumber,
					}

					completedPartsMu.Lock()
					completedParts[part.Index] = completedPart
					completedPartsMu.Unlock()
				}
			}
		}()
	}

	go func() {
		defer close(jobs)

		for _, part := range parts {
			select {
			case jobs <- part:
			case <-copyCtx.Done():
				return
			}
		}
	}()

	wg.Wait()
	if firstCopyErr != nil {
		return c.abortMultipartCopy(ctx, move, uploadID, firstCopyErr)
	}

	_, err = c.S3Client.CompleteMultipartUpload(ctx, &s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(move.DestinationBucket),
		Key:      aws.String(move.DestinationKey),
		UploadId: uploadID,
		MultipartUpload: &s3types.CompletedMultipartUpload{
			Parts: completedParts,
		},
	})
	if err != nil {
		return c.abortMultipartCopy(ctx, move, uploadID, fmt.Errorf("complete multipart upload for %s/%s: %w", move.DestinationBucket, move.DestinationKey, err))
	}

	return nil
}

func (c *Client) abortMultipartCopy(ctx context.Context, move MovedObject, uploadID *string, copyErr error) error {
	_, abortErr := c.S3Client.AbortMultipartUpload(ctx, &s3.AbortMultipartUploadInput{
		Bucket:   aws.String(move.DestinationBucket),
		Key:      aws.String(move.DestinationKey),
		UploadId: uploadID,
	})
	if abortErr != nil {
		return fmt.Errorf("%w (abort multipart upload: %v)", copyErr, abortErr)
	}

	return copyErr
}

func (c *Client) deleteMovedObject(ctx context.Context, bucket, key string) error {
	res, err := c.S3Client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
		Bucket: aws.String(bucket),
		Delete: &s3types.Delete{Objects: []s3types.ObjectIdentifier{{Key: aws.String(key)}}},
	})
	if err != nil {
		return err
	}
	if len(res.Errors) > 0 {
		deleteErr := res.Errors[0]
		if deleteErr.Message != nil {
			return fmt.Errorf("%s", aws.ToString(deleteErr.Message))
		}
		if deleteErr.Code != nil {
			return fmt.Errorf("%s", aws.ToString(deleteErr.Code))
		}
		return fmt.Errorf("delete failed")
	}

	return nil
}

func encodeCopySource(bucket, key string, versionID *string) string {
	u := url.URL{Path: bucket + "/" + key}
	if versionID != nil && aws.ToString(versionID) != "" && aws.ToString(versionID) != "null" {
		query := url.Values{}
		query.Set("versionId", aws.ToString(versionID))
		u.RawQuery = query.Encode()
	}

	return u.String()
}

func estimateMultipartCopyPartSize(size int64) int64 {
	return estimateMultipartPartSize(size)
}

func buildMultipartCopyParts(size, partSize int64) []multipartCopyPart {
	parts := make([]multipartCopyPart, 0, ((size-1)/partSize)+1)

	for start, index, partNumber := int64(0), 0, int32(1); start < size; start, index, partNumber = start+partSize, index+1, partNumber+1 {
		end := start + partSize - 1
		if end >= size {
			end = size - 1
		}

		parts = append(parts, multipartCopyPart{
			Index:      index,
			PartNumber: partNumber,
			Start:      start,
			End:        end,
		})
	}

	return parts
}

func setMultipartCopyError(mu *sync.Mutex, firstErr *error, err error) {
	mu.Lock()
	defer mu.Unlock()

	if *firstErr == nil {
		*firstErr = err
	}
}

func normalizeStorageMoveConfig(config *StorageMoveConfig) *StorageMoveConfig {
	if config == nil {
		config = &StorageMoveConfig{}
	}

	if config.MultipartCopyConcurrency <= 0 {
		config.MultipartCopyConcurrency = defaultMultipartCopyConcurrency
	}

	return config
}

func resolveMoveObjectDestinationKey(srcKey, dstKey string) string {
	if dstKey == "" || dstKey == "/" {
		return path.Base(srcKey)
	}

	if strings.HasSuffix(dstKey, "/") {
		return path.Join(strings.TrimSuffix(normalizeMovePrefix(dstKey), "/"), path.Base(srcKey))
	}

	return dstKey
}

func resolveMovePrefixDestinationKey(srcPrefix, srcKey, dstPrefix string) string {
	sourceBase := normalizeMovePrefix(srcPrefix)
	relativeKey := strings.TrimPrefix(srcKey, sourceBase)
	destinationBase := normalizeMovePrefix(dstPrefix)
	if destinationBase == "" {
		return relativeKey
	}

	return path.Join(strings.TrimSuffix(destinationBase, "/"), relativeKey)
}

func normalizeMovePrefix(prefix string) string {
	if prefix == "/" {
		return ""
	}

	return prefix
}

func moveLocationsEqual(move MovedObject) bool {
	return move.SourceBucket == move.DestinationBucket && move.SourceKey == move.DestinationKey
}
