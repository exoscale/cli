package sos

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
)

func (c *storageClient) updateObjectHeaders(bucket, key string, headers map[string]*string) error {
	object, err := c.copyObject(bucket, key)
	if err != nil {
		return err
	}

	lookupHeader := func(key string, fallback *string) *string {
		if v, ok := headers[key]; ok {
			return v
		}
		return fallback
	}

	object.CacheControl = lookupHeader(storageObjectHeaderCacheControl, object.CacheControl)
	object.ContentDisposition = lookupHeader(storageObjectHeaderContentDisposition, object.ContentDisposition)
	object.ContentEncoding = lookupHeader(storageObjectHeaderContentEncoding, object.ContentEncoding)
	object.ContentLanguage = lookupHeader(storageObjectHeaderContentLanguage, object.ContentLanguage)
	object.ContentType = lookupHeader(storageObjectHeaderContentType, object.ContentType)

	// For some reason, the AWS SDK doesn't use the same type for the "Expires"
	// header in GetObject (*string) and CopyObject (*time.Time)...
	if v, ok := headers[storageObjectHeaderExpires]; ok {
		t, err := time.Parse(time.RFC822, aws.ToString(v))
		if err != nil {
			return fmt.Errorf(`invalid "Expires" header value %q, expecting RFC822 format`, aws.ToString(v))
		}
		object.Expires = &t
	}

	_, err = c.CopyObject(gContext, object)
	return err
}

func (c *storageClient) updateObjectsHeaders(bucket, prefix string, headers map[string]*string, recursive bool) error {
	return c.forEachObject(bucket, prefix, recursive, func(o *s3types.Object) error {
		return c.updateObjectHeaders(bucket, aws.ToString(o.Key), headers)
	})
}

func (c *storageClient) deleteObjectHeaders(bucket, key string, headers []string) error {
	object, err := c.copyObject(bucket, key)
	if err != nil {
		return err
	}

	for _, header := range headers {
		switch header {
		case storageObjectHeaderCacheControl:
			object.CacheControl = nil

		case storageObjectHeaderContentDisposition:
			object.ContentDisposition = nil

		case storageObjectHeaderContentEncoding:
			object.ContentEncoding = nil

		case storageObjectHeaderContentLanguage:
			object.ContentLanguage = nil

		case storageObjectHeaderContentType:
			object.ContentType = aws.String("application/binary")

		case storageObjectHeaderExpires:
			object.Expires = nil
		}
	}

	_, err = c.CopyObject(gContext, object)
	return err
}

func (c *storageClient) deleteObjectsHeaders(bucket, prefix string, headers []string, recursive bool) error {
	return c.forEachObject(bucket, prefix, recursive, func(o *s3types.Object) error {
		return c.deleteObjectHeaders(bucket, aws.ToString(o.Key), headers)
	})
}
