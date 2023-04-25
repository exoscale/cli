package sos

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/table"
	"github.com/exoscale/cli/utils"
)

// CORSRulesFromS3 converts a list of S3 CORS rules to a list of
// CORSRule.
func CORSRulesFromS3(v *s3.GetBucketCorsOutput) []CORSRule {
	rules := make([]CORSRule, 0)

	for _, rule := range v.CORSRules {
		rules = append(rules, CORSRule{
			AllowedOrigins: rule.AllowedOrigins,
			AllowedMethods: rule.AllowedMethods,
			AllowedHeaders: rule.AllowedHeaders,
		})
	}

	return rules
}

func (c *Client) CreateNewBucket(ctx context.Context, name, acl string) error {
	s3Bucket := s3.CreateBucketInput{Bucket: aws.String(name)}

	if acl != "" {
		if !utils.IsInList(S3BucketCannedACLToStrings(), acl) {
			return fmt.Errorf("invalid canned ACL %q, supported values are: %s",
				acl,
				strings.Join(S3BucketCannedACLToStrings(), ", "))
		}

		s3Bucket.ACL = s3types.BucketCannedACL(acl)
	}

	_, err := c.CreateBucket(ctx, &s3Bucket)
	return err
}

func (c *Client) ShowBucket(ctx context.Context, bucket string) (output.Outputter, error) {
	acl, err := c.GetBucketAcl(ctx, &s3.GetBucketAclInput{Bucket: aws.String(bucket)})
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve bucket ACL: %w", err)
	}

	cors, err := c.GetBucketCors(ctx, &s3.GetBucketCorsInput{Bucket: aws.String(bucket)})
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
		ACL:  ACLFromS3(acl.Grants),
		CORS: CORSRulesFromS3(cors),
	}

	return &out, nil
}

type storageBucketObjectOwnershipOutput struct {
	Bucket          string `json:"bucket"`
	ObjectOwnership string `json:"objectOwnership"`
}

func (o *storageBucketObjectOwnershipOutput) toJSON() { output.JSON(o) }
func (o *storageBucketObjectOwnershipOutput) toText() { output.Text(o) }
func (o *storageBucketObjectOwnershipOutput) toTable() {
	t := table.NewTable(os.Stdout)
	defer t.Render()
	t.SetHeader([]string{"Bucket Object Ownership"})

	t.Append([]string{"Bucket", o.Bucket})
	// TODO naming?
	t.Append([]string{"Object Ownership", o.ObjectOwnership})
}

func (c Client) GetBucketObjectOwnership(ctx context.Context, bucket string) (output.Outputter, error) {
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
	ObjectOwnershipObjectWriter         BucketObjectOwnership = BucketObjectOwnership(s3types.ObjectOwnershipObjectWriter)
	ObjectOwnershipBucketOwnerPreferred BucketObjectOwnership = BucketObjectOwnership(s3types.ObjectOwnershipBucketOwnerPreferred)
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

	// TODO
	_, err := c.PutBucketOwnershipControls(ctx, &params)
	if err != nil {
		// TODO wrap
		return err
	}

	return nil
}

func (c Client) DeleteBucket(ctx context.Context, bucket string, recursive bool) error {
	if recursive {
		if _, err := c.DeleteObjects(bucket, "", true); err != nil {
			return fmt.Errorf("error deleting objects: %s", err)
		}
	}

	// Delete dangling multipart uploads preventing bucket deletion.
	res, err := c.ListMultipartUploads(ctx, &s3.ListMultipartUploadsInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		return fmt.Errorf("error listing dangling multipart uploads: %w", err)
	}
	for _, mp := range res.Uploads {
		if _, err = c.AbortMultipartUpload(ctx, &s3.AbortMultipartUploadInput{
			Bucket:   aws.String(bucket),
			Key:      mp.Key,
			UploadId: mp.UploadId,
		}); err != nil {
			return fmt.Errorf("error aborting dangling multipart upload: %w", err)
		}
	}

	if _, err := c.DeleteBucket(ctx, &s3.DeleteBucketInput{Bucket: aws.String(bucket)}); err != nil {
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
