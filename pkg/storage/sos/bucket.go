package sos

import (
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
	"github.com/exoscale/cli/utils"
)

func (c *Client) createBucket(name, acl string) error {
	s3Bucket := s3.CreateBucketInput{Bucket: aws.String(name)}

	if acl != "" {
		if !utils.IsInList(s3BucketCannedACLToStrings(), acl) {
			return fmt.Errorf("invalid canned ACL %q, supported values are: %s",
				acl,
				strings.Join(s3BucketCannedACLToStrings(), ", "))
		}

		s3Bucket.ACL = s3types.BucketCannedACL(acl)
	}

	_, err := c.CreateBucket(gContext, &s3Bucket)
	return err
}

func (c *Client) showBucket(bucket string) (outputter, error) {
	acl, err := c.GetBucketAcl(gContext, &s3.GetBucketAclInput{Bucket: aws.String(bucket)})
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve bucket ACL: %w", err)
	}

	cors, err := c.GetBucketCors(gContext, &s3.GetBucketCorsInput{Bucket: aws.String(bucket)})
	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) {
			if apiErr.ErrorCode() == "NoSuchCORSConfiguration" {
				cors = &s3.GetBucketCorsOutput{}
			}
		}

		if cors == nil {
			return nil, fmt.Errorf("unable to retrieve bucket CORS configuration: %w", err)
		}
	}

	out := storageShowBucketOutput{
		Name: bucket,
		Zone: c.zone,
		ACL:  storageACLFromS3(acl.Grants),
		CORS: storageCORSRulesFromS3(cors),
	}

	return &out, nil
}
