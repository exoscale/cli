package sos

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
)

func (c *Client) deleteBucketCORS(bucket string) error {
	_, err := c.DeleteBucketCors(gContext, &s3.DeleteBucketCorsInput{Bucket: &bucket})
	return err
}

func (c *Client) addBucketCORSRule(bucket string, cors *storageCORSRule) error {
	curCORS, err := c.GetBucketCors(gContext, &s3.GetBucketCorsInput{Bucket: aws.String(bucket)})
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

	_, err = c.PutBucketCors(gContext, &s3.PutBucketCorsInput{
		Bucket: &bucket,
		CORSConfiguration: &s3types.CORSConfiguration{
			CORSRules: append(curCORS.CORSRules, cors.toS3()),
		},
	})

	return err
}
