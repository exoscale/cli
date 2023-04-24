package sos

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
)

func (c *Client) addObjectMetadata(bucket, key string, metadata map[string]string) error {
	object, err := c.copyObject(bucket, key)
	if err != nil {
		return err
	}

	if len(object.Metadata) == 0 {
		object.Metadata = make(map[string]string)
	}

	for k, v := range metadata {
		if strings.ContainsAny(k, storageMetadataForbiddenCharset) {
			return fmt.Errorf("%s: invalid value", k)
		}

		object.Metadata[k] = v
	}

	_, err = c.CopyObject(gContext, object)
	return err
}

func (c *Client) addObjectsMetadata(bucket, prefix string, metadata map[string]string, recursive bool) error {
	return c.forEachObject(bucket, prefix, recursive, func(o *s3types.Object) error {
		return c.addObjectMetadata(bucket, aws.ToString(o.Key), metadata)
	})
}

func (c *Client) deleteObjectMetadata(bucket, key string, mdKeys []string) error {
	object, err := c.copyObject(bucket, key)
	if err != nil {
		return err
	}

	for _, k := range mdKeys {
		if _, ok := object.Metadata[k]; !ok {
			return fmt.Errorf("key %q not found in current metadata", k)
		}
		delete(object.Metadata, k)
	}

	_, err = c.CopyObject(gContext, object)
	return err
}

func (c *Client) DeleteObjectsMetadata(bucket, prefix string, mdKeys []string, recursive bool) error {
	return c.forEachObject(bucket, prefix, recursive, func(o *s3types.Object) error {
		return c.deleteObjectMetadata(bucket, aws.ToString(o.Key), mdKeys)
	})
}
