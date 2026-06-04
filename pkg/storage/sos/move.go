package sos

import (
	"context"
	"fmt"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

const (
	moveLargeObjectThreshold = 5 * 1024 * 1024 * 1024 // 5 GiB
	moveDefaultPartSize      = 100 * 1024 * 1024      // 100 MiB
	moveMaxConcurrency       = 10
)

func (c *Client) MoveObject(ctx context.Context, srcBucket, srcKey, dstBucket, dstKey string, multipartConcurrency int, verbose bool) error {
	if multipartConcurrency <= 0 {
		multipartConcurrency = 1
	}
	if multipartConcurrency > moveMaxConcurrency {
		multipartConcurrency = moveMaxConcurrency
	}

	headRes, err := c.S3Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(srcBucket),
		Key:    aws.String(srcKey),
	})
	if err != nil {
		return fmt.Errorf("unable to retrieve source object info: %w", err)
	}

	size := headRes.ContentLength

	if size > moveLargeObjectThreshold {
		return c.moveLargeObject(ctx, srcBucket, srcKey, dstBucket, dstKey, headRes, multipartConcurrency, verbose)
	}

	return c.moveObject(ctx, srcBucket, srcKey, dstBucket, dstKey, headRes, verbose)
}

func (c *Client) moveObject(ctx context.Context, srcBucket, srcKey, dstBucket, dstKey string, headRes *s3.HeadObjectOutput, verbose bool) error {
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

	copyInput := &s3.CopyObjectInput{
		Bucket:            aws.String(dstBucket),
		Key:               aws.String(dstKey),
		CopySource:        aws.String(copySource(srcBucket, srcKey)),
		Metadata:          headRes.Metadata,
		MetadataDirective: s3types.MetadataDirectiveReplace,
		ACL:               getACLFromGrants(acl.Grants),
	}

	if headRes.CacheControl != nil {
		copyInput.CacheControl = headRes.CacheControl
	}
	if headRes.ContentDisposition != nil {
		copyInput.ContentDisposition = headRes.ContentDisposition
	}
	if headRes.ContentEncoding != nil {
		copyInput.ContentEncoding = headRes.ContentEncoding
	}
	if headRes.ContentLanguage != nil {
		copyInput.ContentLanguage = headRes.ContentLanguage
	}
	if headRes.ContentType != nil {
		copyInput.ContentType = headRes.ContentType
	}
	if headRes.Expires != nil {
		copyInput.Expires = headRes.Expires
	}

	if _, err := c.S3Client.CopyObject(ctx, copyInput); err != nil {
		return fmt.Errorf("copy: %w", err)
	}

	if verbose {
		fmt.Printf("deleting: %s\n", srcURL)
	}

	if err := c.DeleteObject(ctx, srcBucket, srcKey); err != nil {
		return fmt.Errorf("delete source: %w", err)
	}

	return nil
}

func (c *Client) DeleteObject(ctx context.Context, bucket, key string) error {
	_, err := c.S3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	return err
}

func copySource(bucket, key string) string {
	return bucket + "/" + url.PathEscape(key)
}

// getACLFromGrants maps S3 object grants to a canned ACL. Note: complex
// grant sets (e.g. per-user CanonicalUser grants) are not preserved; they
// fall back to private. Only the common public/authenticated-read group
// grants are mapped.
func getACLFromGrants(grants []s3types.Grant) s3types.ObjectCannedACL {
	for _, grant := range grants {
		if grant.Grantee.Type != s3types.TypeGroup {
			continue
		}
		uri := aws.ToString(grant.Grantee.URI)
		if uri == "http://acs.amazonaws.com/groups/global/AllUsers" {
			return s3types.ObjectCannedACLPublicRead
		}
		if uri == "http://acs.amazonaws.com/groups/global/AuthenticatedUsers" {
			return s3types.ObjectCannedACLAuthenticatedRead
		}
	}
	return s3types.ObjectCannedACLPrivate
}
