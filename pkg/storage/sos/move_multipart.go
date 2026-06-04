package sos

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hashicorp/go-multierror"
)

func (c *Client) moveLargeObject(ctx context.Context, srcBucket, srcKey, dstBucket, dstKey string, headRes *s3.HeadObjectOutput, concurrency int, verbose bool) error {
	srcURL := fmt.Sprintf("sos://%s/%s", srcBucket, srcKey)
	dstURL := fmt.Sprintf("sos://%s/%s", dstBucket, dstKey)

	if verbose {
		fmt.Printf("copying: %s -> %s\n", srcURL, dstURL)
	}

	acl, err := c.S3Client.GetObjectAcl(ctx, &s3.GetObjectAclInput{
		Bucket: aws.String(srcBucket),
		Key:    aws.String(srcKey),
	})
	if err != nil {
		return fmt.Errorf("unable to retrieve object ACL: %w", err)
	}

	createMPInput := &s3.CreateMultipartUploadInput{
		Bucket:   aws.String(dstBucket),
		Key:      aws.String(dstKey),
		Metadata: headRes.Metadata,
		ACL:      getACLFromGrants(acl.Grants),
	}

	if headRes.CacheControl != nil {
		createMPInput.CacheControl = headRes.CacheControl
	}
	if headRes.ContentDisposition != nil {
		createMPInput.ContentDisposition = headRes.ContentDisposition
	}
	if headRes.ContentEncoding != nil {
		createMPInput.ContentEncoding = headRes.ContentEncoding
	}
	if headRes.ContentLanguage != nil {
		createMPInput.ContentLanguage = headRes.ContentLanguage
	}
	if headRes.ContentType != nil {
		createMPInput.ContentType = headRes.ContentType
	}
	if headRes.Expires != nil {
		createMPInput.Expires = headRes.Expires
	}

	createRes, err := c.S3Client.CreateMultipartUpload(ctx, createMPInput)
	if err != nil {
		return fmt.Errorf("create multipart upload: %w", err)
	}
	if createRes.UploadId == nil {
		return fmt.Errorf("no upload id returned")
	}

	size := headRes.ContentLength
	completedParts, err := c.uploadParts(ctx, srcBucket, srcKey, dstBucket, dstKey, aws.ToString(createRes.UploadId), size, concurrency)
	if err != nil {
		_, abortErr := c.S3Client.AbortMultipartUpload(ctx, &s3.AbortMultipartUploadInput{
			Bucket:   aws.String(dstBucket),
			Key:      aws.String(dstKey),
			UploadId: createRes.UploadId,
		})
		if abortErr != nil {
			return fmt.Errorf("upload failed: %w, abort failed: %v", err, abortErr)
		}
		return fmt.Errorf("upload failed: %w", err)
	}

	sort.Slice(completedParts, func(i, j int) bool {
		return completedParts[i].PartNumber < completedParts[j].PartNumber
	})

	_, err = c.S3Client.CompleteMultipartUpload(ctx, &s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(dstBucket),
		Key:      aws.String(dstKey),
		UploadId: createRes.UploadId,
		MultipartUpload: &s3types.CompletedMultipartUpload{
			Parts: completedParts,
		},
	})
	if err != nil {
		return fmt.Errorf("complete multipart upload: %w", err)
	}

	if verbose {
		fmt.Printf("deleting: %s\n", srcURL)
	}

	if err := c.DeleteObject(ctx, srcBucket, srcKey); err != nil {
		return fmt.Errorf("delete source: %w", err)
	}

	return nil
}

func (c *Client) uploadParts(ctx context.Context, srcBucket, srcKey, dstBucket, dstKey, uploadID string, size int64, concurrency int) ([]s3types.CompletedPart, error) {
	partSize := int64(moveDefaultPartSize)
	if partSize > size {
		partSize = size
	}

	numParts := int((size + partSize - 1) / partSize)

	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var completedParts []s3types.CompletedPart
	var errs *multierror.Error

	for i := 0; i < numParts; i++ {
		sem <- struct{}{}
		wg.Add(1)

		go func(partNum int) {
			defer wg.Done()
			defer func() { <-sem }()

			start := int64(partNum) * partSize
			end := start + partSize
			if end > size {
				end = size
			}

			part, err := c.uploadPartCopy(ctx, srcBucket, srcKey, dstBucket, dstKey, uploadID, int32(partNum+1), start, end)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				errs = multierror.Append(errs, err)
			} else if part != nil {
				completedParts = append(completedParts, *part)
			}
		}(i)
	}

	wg.Wait()

	if errs != nil {
		return nil, errs.ErrorOrNil()
	}

	return completedParts, nil
}

func (c *Client) uploadPartCopy(ctx context.Context, srcBucket, srcKey, dstBucket, dstKey, uploadID string, partNumber int32, start, end int64) (*s3types.CompletedPart, error) {
	res, err := c.S3Client.UploadPartCopy(ctx, &s3.UploadPartCopyInput{
		Bucket:          aws.String(dstBucket),
		Key:             aws.String(dstKey),
		UploadId:        aws.String(uploadID),
		PartNumber:      partNumber,
		CopySource:      aws.String(copySource(srcBucket, srcKey)),
		CopySourceRange: aws.String(fmt.Sprintf("bytes=%d-%d", start, end-1)),
	})
	if err != nil {
		return nil, fmt.Errorf("upload part copy: %w", err)
	}

	return &s3types.CompletedPart{
		ETag:       res.CopyPartResult.ETag,
		PartNumber: partNumber,
	}, nil
}
