package sos

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	"github.com/exoscale/cli/utils"
)

func (c *Client) CreateBucket(name, acl string) error {
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

func (c *Client) ShowBucket(bucket string) (outputter, error) {
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

func (c Client) GetBucketObjectOwnership(ctx context.Context, bucket string) (outputter, error) {
	params := s3.GetBucketOwnershipControlsInput{
		Bucket: aws.String(bucket),
	}

	resp, err := c.GetBucketOwnershipControls(ctx, &params)
	if err != nil {
		// TODO wrap
		return nil, err
	}

	out := storageBucketObjectOwnershipOutput{
		Bucket:          bucket,
		ObjectOwnership: string(resp.OwnershipControls.Rules[0].ObjectOwnership),
	}

	return &out, nil
}

type BucketObjectOwnership string

const (
	ObjectOwnershipObjectWriter         BucketObjectOwnership = BucketObjectOwnership(types.ObjectOwnershipObjectWriter)
	ObjectOwnershipBucketOwnerPreferred BucketObjectOwnership = BucketObjectOwnership(types.ObjectOwnershipBucketOwnerPreferred)
	ObjectOwnershipBucketOwnerEnforced  BucketObjectOwnership = "BucketOwnerEnforced"
)

func (c Client) SetBucketObjectOwnership(ctx context.Context, bucket string, ownership BucketObjectOwnership) error {
	params := s3.PutBucketOwnershipControlsInput{
		Bucket: aws.String(bucket),
		OwnershipControls: &types.OwnershipControls{
			Rules: []types.OwnershipControlsRule{
				{
					ObjectOwnership: types.ObjectOwnership(ownership),
				},
			}},
	}

	resp, err := c.PutBucketOwnershipControls(ctx, &params)
	if err != nil {
		// TODO wrap
		return err
	}

	return nil
}

func (c Client) DeleteBucket(bucket string, recursive bool) error {
	if recursive {
		if _, err := c.DeleteObjects(bucket, "", true); err != nil {
			return fmt.Errorf("error deleting objects: %s", err)
		}
	}

	// Delete dangling multipart uploads preventing bucket deletion.
	res, err := c.ListMultipartUploads(gContext, &s3.ListMultipartUploadsInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		return fmt.Errorf("error listing dangling multipart uploads: %w", err)
	}
	for _, mp := range res.Uploads {
		if _, err = c.AbortMultipartUpload(gContext, &s3.AbortMultipartUploadInput{
			Bucket:   aws.String(bucket),
			Key:      mp.Key,
			UploadId: mp.UploadId,
		}); err != nil {
			return fmt.Errorf("error aborting dangling multipart upload: %w", err)
		}
	}

	if _, err := c.DeleteBucket(gContext, &s3.DeleteBucketInput{Bucket: aws.String(bucket)}); err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) {
			if apiErr.ErrorCode() == "BucketNotEmpty" {
				return errors.New("bucket is not empty, either delete files before or use flag `-r`")
			}
		}

		return fmt.Errorf("unable to retrieve bucket CORS configuration: %w", err)
	}

	return nil
}
