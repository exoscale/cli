package sos

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

const MetadataForbiddenCharset = `()<>@,;!:\\'&"/[]?_={} `

func (c *Client) AddObjectMetadata(ctx context.Context, bucket, key string, metadata map[string]string) error {
	object, err := c.CopyObject(ctx, bucket, key)
	if err != nil {
		return err
	}

	if len(object.Metadata) == 0 {
		object.Metadata = make(map[string]string)
	}

	for k, v := range metadata {
		if strings.ContainsAny(k, MetadataForbiddenCharset) {
			return fmt.Errorf("%s: invalid value", k)
		}

		object.Metadata[k] = v
	}

	_, err = c.s3Client.CopyObject(ctx, object)
	return err
}

func (c *Client) AddObjectsMetadata(ctx context.Context, bucket, prefix string, metadata map[string]string, recursive bool) error {
	return c.ForEachObject(ctx, bucket, prefix, recursive, func(o *s3types.Object) error {
		return c.AddObjectMetadata(ctx, bucket, aws.ToString(o.Key), metadata)
	})
}

func (c *Client) DeleteObjectMetadata(ctx context.Context, bucket, key string, mdKeys []string) error {
	object, err := c.CopyObject(ctx, bucket, key)
	if err != nil {
		return err
	}

	for _, k := range mdKeys {
		if _, ok := object.Metadata[k]; !ok {
			return fmt.Errorf("key %q not found in current metadata", k)
		}
		delete(object.Metadata, k)
	}

	_, err = c.s3Client.CopyObject(ctx, object)
	return err
}

func (c *Client) DeleteObjectsMetadata(ctx context.Context, bucket, prefix string, mdKeys []string, recursive bool) error {
	return c.ForEachObject(ctx, bucket, prefix, recursive, func(o *s3types.Object) error {
		return c.DeleteObjectMetadata(ctx, bucket, aws.ToString(o.Key), mdKeys)
	})
}
