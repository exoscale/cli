package sos

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
)

func (c *Client) DeleteBucketCORS(ctx context.Context, bucket string) error {
	_, err := c.DeleteBucketCors(ctx, &s3.DeleteBucketCorsInput{Bucket: &bucket})
	return err
}

type CORSRule struct {
	AllowedOrigins []string `json:"allowed_origins,omitempty"`
	AllowedMethods []string `json:"allowed_methods,omitempty"`
	AllowedHeaders []string `json:"allowed_headers,omitempty"`
}

func (c *Client) AddBucketCORSRule(ctx context.Context, bucket string, cors *CORSRule) error {
	curCORS, err := c.GetBucketCors(ctx, &s3.GetBucketCorsInput{Bucket: aws.String(bucket)})
	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) {
			if apiErr.ErrorCode() == "NoSuchCORSConfiguration" {
				curCORS = &s3.GetBucketCorsOutput{}
			}
		}

		if cors == nil {
			return fmt.Errorf("unable to retrieve bucket CORS configuration: %w", err)
		}
	}

	_, err = c.PutBucketCors(ctx, &s3.PutBucketCorsInput{
		Bucket: &bucket,
		CORSConfiguration: &s3types.CORSConfiguration{
			CORSRules: append(curCORS.CORSRules, cors.toS3()),
		},
	})

	return err
}
