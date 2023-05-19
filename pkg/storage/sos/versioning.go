package sos

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/table"
)

type storageBucketObjectVersioningOutput struct {
	Bucket           string `json:"bucket"`
	ObjectVersioning string `json:"objectVersioning"`
}

func (o *storageBucketObjectVersioningOutput) ToJSON() { output.JSON(o) }
func (o *storageBucketObjectVersioningOutput) ToText() { output.Text(o) }
func (o *storageBucketObjectVersioningOutput) ToTable() {
	t := table.NewTable(os.Stdout)
	defer t.Render()
	t.SetHeader([]string{"Bucket Object Versioning"})

	t.Append([]string{"Bucket", o.Bucket})
	t.Append([]string{"Object Versioning", o.ObjectVersioning})
}

func (c *Client) GetBucketVersioning(ctx context.Context, bucket string) (types.BucketVersioningStatus, error) {
	result, err := c.S3Client.GetBucketVersioning(ctx, &s3.GetBucketVersioningInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		return "", err
	}

	if string(result.Status) == "" {
		return types.BucketVersioningStatus("Disabled"), nil
	}

	return result.Status, nil
}

func (c *Client) BucketVersioningStatus(ctx context.Context, bucket string) (output.Outputter, error) {
	status, err := c.GetBucketVersioning(ctx, bucket)
	if err != nil {
		return nil, err
	}

	return &storageBucketObjectVersioningOutput{
		Bucket:           bucket,
		ObjectVersioning: string(status),
	}, nil
}

func (c *Client) EnableBucketVersioning(ctx context.Context, bucket string) error {
	_, err := c.S3Client.PutBucketVersioning(ctx, &s3.PutBucketVersioningInput{
		Bucket: aws.String(bucket),
		VersioningConfiguration: &types.VersioningConfiguration{
			Status: types.BucketVersioningStatusEnabled,
		},
	})

	return err
}

func (c *Client) SuspendBucketVersioning(ctx context.Context, bucket string) error {
	_, err := c.S3Client.PutBucketVersioning(ctx, &s3.PutBucketVersioningInput{
		Bucket: aws.String(bucket),
		VersioningConfiguration: &types.VersioningConfiguration{
			Status: types.BucketVersioningStatusSuspended,
		},
	})

	return err
}
