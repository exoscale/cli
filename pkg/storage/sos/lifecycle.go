package sos

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/table"
)

// Decoding json into the default types.LifecycleRuleFilter fails so we
// partially recreate the original struct with conversion methods
type BucketLifecycleConfRuleFilter struct {
	Prefix                *string
	And                   *BucketLifecycleAndOperator
	ObjectSizeGreaterThan *int64
	ObjectSizeLessThan    *int64
}

type BucketLifecycleAndOperator struct {
	Prefix                *string
	ObjectSizeGreaterThan *int64
	ObjectSizeLessThan    *int64
}

type BucketLifecycleRule struct {
	Status types.ExpirationStatus

	AbortIncompleteMultipartUpload *types.AbortIncompleteMultipartUpload

	Expiration *types.LifecycleExpiration

	Filter *BucketLifecycleConfRuleFilter

	ID *string

	NoncurrentVersionExpiration *types.NoncurrentVersionExpiration
}

type BucketLifecycleConf struct {
	Rules []BucketLifecycleRule
}

type BucketLifecycle struct {
	BucketLifecycleConf
	Bucket string
}

func (o *BucketLifecycleConf) ToS3() *types.BucketLifecycleConfiguration {
	c := &types.BucketLifecycleConfiguration{
		Rules: []types.LifecycleRule{},
	}
	for _, r := range o.Rules {

		rule := types.LifecycleRule{
			Status:                         r.Status,
			AbortIncompleteMultipartUpload: r.AbortIncompleteMultipartUpload,
			Expiration:                     r.Expiration,
			Filter:                         r.filterToS3(),
			ID:                             r.ID,
			NoncurrentVersionExpiration:    r.NoncurrentVersionExpiration,
		}

		c.Rules = append(c.Rules, rule)
	}
	return c
}

func (r BucketLifecycleRule) filterToS3() *types.LifecycleRuleFilter {
	if r.Filter == nil {
		return nil
	}
	filter := &types.LifecycleRuleFilter{
		Prefix:                r.Filter.Prefix,
		ObjectSizeGreaterThan: r.Filter.ObjectSizeGreaterThan,
		ObjectSizeLessThan:    r.Filter.ObjectSizeLessThan,
	}
	if r.Filter.And != nil {
		filter.And = &types.LifecycleRuleAndOperator{
			ObjectSizeGreaterThan: r.Filter.And.ObjectSizeGreaterThan,
			ObjectSizeLessThan:    r.Filter.And.ObjectSizeLessThan,
			Prefix:                r.Filter.And.Prefix,
		}
	}
	return filter
}

func (o *BucketLifecycleConf) FromS3(c *types.BucketLifecycleConfiguration) {
	o.Rules = make([]BucketLifecycleRule, len(c.Rules))

	for i, r := range c.Rules {

		var filter *BucketLifecycleConfRuleFilter
		if f := r.Filter; f != nil {
			filter = &BucketLifecycleConfRuleFilter{
				Prefix:                f.Prefix,
				ObjectSizeGreaterThan: f.ObjectSizeGreaterThan,
				ObjectSizeLessThan:    f.ObjectSizeLessThan,
			}
			if f.And != nil {
				filter.And = &BucketLifecycleAndOperator{
					Prefix:                f.And.Prefix,
					ObjectSizeGreaterThan: f.And.ObjectSizeGreaterThan,
					ObjectSizeLessThan:    f.And.ObjectSizeLessThan,
				}
			}
		}

		o.Rules[i] = BucketLifecycleRule{
			Status:                         r.Status,
			AbortIncompleteMultipartUpload: r.AbortIncompleteMultipartUpload,
			Expiration:                     r.Expiration,
			Filter:                         filter,
			ID:                             r.ID,
			NoncurrentVersionExpiration:    r.NoncurrentVersionExpiration,
		}
	}
}

func (o *BucketLifecycle) ToJSON() { output.JSON(o) }
func (o *BucketLifecycle) ToText() { output.Text(o) }
func (o *BucketLifecycle) ToTable() {
	t := table.NewTable(os.Stdout)
	defer t.Render()
	t.SetHeader([]string{"Bucket Lifecycle"})

	t.Append([]string{"Bucket", o.Bucket})

	t.Append([]string{"Rules", func() string {
		buf := bytes.NewBuffer(nil)
		ct := table.NewEmbeddedTable(buf)
		ct.SetHeader([]string{" "})
		for _, r := range o.Rules {

			if r.ID != nil {
				ct.Append([]string{
					"ID", *r.ID,
				})
			} else {
				ct.Append([]string{
					"ID", "",
				})
			}
			ct.Append([]string{"Status", string(r.Status)})

			if r.Filter != nil {
				if r.Filter.And != nil {
					if r.Filter.And.Prefix != nil {
						ct.Append([]string{"Filter (And) prefix", *r.Filter.And.Prefix})
					}
					if r.Filter.And.ObjectSizeGreaterThan != nil {
						ct.Append([]string{"Filter (And) object-size-greater-than", fmt.Sprintf("%d", *r.Filter.And.ObjectSizeGreaterThan)})
					}
					if r.Filter.And.ObjectSizeLessThan != nil {
						ct.Append([]string{"Filter (And) object-size-less-than", fmt.Sprintf("%d", *r.Filter.And.ObjectSizeLessThan)})
					}
				} else if r.Filter.Prefix != nil {
					ct.Append([]string{"Filter prefix", *r.Filter.Prefix})
				} else if r.Filter.ObjectSizeGreaterThan != nil {
					ct.Append([]string{"Filter object-size-greater-than", fmt.Sprintf("%d", *r.Filter.ObjectSizeGreaterThan)})
				} else if r.Filter.ObjectSizeLessThan != nil {
					ct.Append([]string{"Filter object-size-less-than", fmt.Sprintf("%d", *r.Filter.ObjectSizeLessThan)})
				}
			}

			if r.Expiration != nil {
				if r.Expiration.Days != nil && *r.Expiration.Days > 0 {
					ct.Append([]string{"Expiration days", fmt.Sprintf("%d", *r.Expiration.Days)})
				}
				if r.Expiration.Date != nil {
					ct.Append([]string{"Expiration date", r.Expiration.Date.String()})
				}
				if r.Expiration.ExpiredObjectDeleteMarker != nil {
					ct.Append([]string{"Expire delete marker", fmt.Sprintf("%v", *r.Expiration.ExpiredObjectDeleteMarker)})
				}
			}

			if r.NoncurrentVersionExpiration != nil &&
				r.NoncurrentVersionExpiration.NoncurrentDays != nil {
				ct.Append([]string{"Noncurrent expiration days", fmt.Sprintf("%d", *r.NoncurrentVersionExpiration.NoncurrentDays)})
			}

			if r.AbortIncompleteMultipartUpload != nil &&
				r.AbortIncompleteMultipartUpload.DaysAfterInitiation != nil {
				ct.Append([]string{"Abort incomplete multipart", fmt.Sprintf("%d days", *r.AbortIncompleteMultipartUpload.DaysAfterInitiation)})
			}
		}
		ct.Render()
		return buf.String()
	}()})

}

func (c *Client) GetBucketLifecycle(ctx context.Context, bucket string) (*BucketLifecycle, error) {
	result, err := c.S3Client.GetBucketLifecycleConfiguration(ctx, &s3.GetBucketLifecycleConfigurationInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		return nil, err
	}

	lcConf := BucketLifecycleConf{}
	lcConf.FromS3(&types.BucketLifecycleConfiguration{Rules: result.Rules})
	return &BucketLifecycle{
		BucketLifecycleConf: lcConf,
		Bucket:              bucket,
	}, nil
}

func (c *Client) PutBucketLifecycle(ctx context.Context, bucket string, conf *types.BucketLifecycleConfiguration) error {
	_, err := c.S3Client.PutBucketLifecycleConfiguration(ctx, &s3.PutBucketLifecycleConfigurationInput{
		Bucket:                 aws.String(bucket),
		LifecycleConfiguration: conf,
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) DeleteBucketLifecycle(ctx context.Context, bucket string) error {
	_, err := c.S3Client.DeleteBucketLifecycle(ctx, &s3.DeleteBucketLifecycleInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		return err
	}
	return nil
}
