package sos

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

const (
	ObjectHeaderCacheControl       = "Cache-Control"
	ObjectHeaderContentDisposition = "Content-Disposition"
	ObjectHeaderContentEncoding    = "Content-Encoding"
	ObjectHeaderContentLanguage    = "Content-Language"
	ObjectHeaderContentType        = "Content-Type"
	ObjectHeaderExpires            = "Expires"
)

func (c *Client) UpdateObjectHeaders(ctx context.Context, bucket, key string, headers map[string]*string) error {
	object, err := c.CopyObject(ctx, bucket, key)
	if err != nil {
		return err
	}

	lookupHeader := func(key string, fallback *string) *string {
		if v, ok := headers[key]; ok {
			return v
		}
		return fallback
	}

	object.CacheControl = lookupHeader(ObjectHeaderCacheControl, object.CacheControl)
	object.ContentDisposition = lookupHeader(ObjectHeaderContentDisposition, object.ContentDisposition)
	object.ContentEncoding = lookupHeader(ObjectHeaderContentEncoding, object.ContentEncoding)
	object.ContentLanguage = lookupHeader(ObjectHeaderContentLanguage, object.ContentLanguage)
	object.ContentType = lookupHeader(ObjectHeaderContentType, object.ContentType)

	// For some reason, the AWS SDK doesn't use the same type for the "Expires"
	// header in GetObject (*string) and CopyObject (*time.Time)...
	if v, ok := headers[ObjectHeaderExpires]; ok {
		t, err := time.Parse(time.RFC822, aws.ToString(v))
		if err != nil {
			return fmt.Errorf(`invalid "Expires" header value %q, expecting RFC822 format`, aws.ToString(v))
		}
		object.Expires = &t
	}

	_, err = c.CopyObject(ctx, object)
	return err
}

func (c *Client) UpdateObjectsHeaders(ctx context.Context, bucket, prefix string, headers map[string]*string, recursive bool) error {
	return c.ForEachObject(ctx, bucket, prefix, recursive, func(o *s3types.Object) error {
		return c.UpdateObjectHeaders(ctx, bucket, aws.ToString(o.Key), headers)
	})
}

func (c *Client) DeleteObjectHeaders(ctx context.Context, bucket, key string, headers []string) error {
	object, err := c.CopyObject(ctx, bucket, key)
	if err != nil {
		return err
	}

	for _, header := range headers {
		switch header {
		case ObjectHeaderCacheControl:
			object.CacheControl = nil

		case ObjectHeaderContentDisposition:
			object.ContentDisposition = nil

		case ObjectHeaderContentEncoding:
			object.ContentEncoding = nil

		case ObjectHeaderContentLanguage:
			object.ContentLanguage = nil

		case ObjectHeaderContentType:
			object.ContentType = aws.String("application/binary")

		case ObjectHeaderExpires:
			object.Expires = nil
		}
	}

	_, err = c.CopyObject(ctx, object)
	return err
}

func (c *Client) DeleteObjectsHeaders(bucket, prefix string, headers []string, recursive bool) error {
	return c.forEachObject(bucket, prefix, recursive, func(o *s3types.Object) error {
		return c.deleteObjectHeaders(bucket, aws.ToString(o.Key), headers)
	})
}
