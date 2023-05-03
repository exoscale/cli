package sos

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
)

func (c *Client) DeleteBucketCORS(ctx context.Context, bucket string) error {
	_, err := c.S3Client.DeleteBucketCors(ctx, &s3.DeleteBucketCorsInput{Bucket: &bucket})
	return err
}

type CORSRule struct {
	AllowedOrigins []string `json:"allowed_origins,omitempty"`
	AllowedMethods []string `json:"allowed_methods,omitempty"`
	AllowedHeaders []string `json:"allowed_headers,omitempty"`
}

func (c *Client) AddBucketCORSRule(ctx context.Context, bucket string, cors *CORSRule) error {
	curCORS, err := c.S3Client.GetBucketCors(ctx, &s3.GetBucketCorsInput{Bucket: aws.String(bucket)})
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

	_, err = c.S3Client.PutBucketCors(ctx, &s3.PutBucketCorsInput{
		Bucket: &bucket,
		CORSConfiguration: &s3types.CORSConfiguration{
			CORSRules: append(curCORS.CORSRules, cors.toS3()),
		},
	})

	return err
}

// toS3 converts a sos.CORSRule object to the S3 CORS rule format.
func (r *CORSRule) toS3() s3types.CORSRule {
	return s3types.CORSRule{
		AllowedOrigins: r.AllowedOrigins,
		AllowedMethods: r.AllowedMethods,
		AllowedHeaders: r.AllowedHeaders,
	}
}
