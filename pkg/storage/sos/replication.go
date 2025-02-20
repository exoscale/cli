package sos

import (
	"bytes"
	"context"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/table"
)

// Decoding json into the default types.ReplicationRuleFilter fails so we
// partially recreate the original struct with conversion methods
type bucketReplicationConfRuleFilter struct {
	Prefix string
}

type bucketReplicationRule struct {
	Destination *types.Destination

	Status types.ReplicationRuleStatus

	DeleteMarkerReplication *types.DeleteMarkerReplication

	ExistingObjectReplication *types.ExistingObjectReplication

	Filter bucketReplicationConfRuleFilter

	ID *string

	Priority int32

	SourceSelectionCriteria *types.SourceSelectionCriteria
}

type BucketReplicationConf struct {
	Role  *string
	Rules []bucketReplicationRule
}

type bucketReplication struct {
	BucketReplicationConf
	Bucket string
}

func (o *BucketReplicationConf) ToS3() *types.ReplicationConfiguration {
	c := &types.ReplicationConfiguration{
		Role:  o.Role,
		Rules: []types.ReplicationRule{},
	}
	for _, r := range o.Rules {

		c.Rules = append(c.Rules, types.ReplicationRule{
			Destination:               r.Destination,
			Status:                    r.Status,
			DeleteMarkerReplication:   r.DeleteMarkerReplication,
			ExistingObjectReplication: r.ExistingObjectReplication,
			Filter:                    &types.ReplicationRuleFilterMemberPrefix{Value: r.Filter.Prefix},
			ID:                        r.ID,
			Priority:                  r.Priority,
			SourceSelectionCriteria:   r.SourceSelectionCriteria,
		})
	}
	return c
}

func (o *BucketReplicationConf) FromS3(c *types.ReplicationConfiguration) {
	o.Role = c.Role
	o.Rules = make([]bucketReplicationRule, len(c.Rules))

	for i, r := range c.Rules {

		p := r.Filter.(*types.ReplicationRuleFilterMemberPrefix)

		o.Rules[i] = bucketReplicationRule{
			Destination:               r.Destination,
			Status:                    r.Status,
			DeleteMarkerReplication:   r.DeleteMarkerReplication,
			ExistingObjectReplication: r.ExistingObjectReplication,
			Filter:                    bucketReplicationConfRuleFilter{Prefix: p.Value},
			ID:                        r.ID,
			Priority:                  r.Priority,
			SourceSelectionCriteria:   r.SourceSelectionCriteria,
		}
	}
}

func (o *bucketReplication) ToJSON() { output.JSON(o) }
func (o *bucketReplication) ToText() { output.Text(o) }
func (o *bucketReplication) ToTable() {
	t := table.NewTable(os.Stdout)
	defer t.Render()
	t.SetHeader([]string{"Bucket Replication"})

	t.Append([]string{"Bucket", o.Bucket})
	t.Append([]string{"Role", func() string {
		if o.Role != nil {
			return *o.Role
		} else {
			return "-"
		}
	}()})

	t.Append([]string{"Rules", func() string {
		buf := bytes.NewBuffer(nil)
		ct := table.NewEmbeddedTable(buf)
		ct.SetHeader([]string{" "})
		for _, r := range o.Rules {

			if r.ID != nil {
				ct.Append([]string{
					"ID", *r.ID,
				})
			}
			ct.Append([]string{"Status", string(r.Status)})
			ct.Append([]string{"Priority", strconv.Itoa(int(r.Priority))})
			if r.ExistingObjectReplication != nil {
				ct.Append([]string{"ExistingObjectReplication", string(r.ExistingObjectReplication.Status)})
			}
			ct.Append([]string{"DeleteMarkerReplication", string(r.DeleteMarkerReplication.Status)})

			ct.Append([]string{"Filter prefix", r.Filter.Prefix})

			if r.Destination != nil {
				if r.Destination.Bucket != nil {
					ct.Append([]string{"Destination Bucket", *r.Destination.Bucket})
				}
				if r.Destination.Account != nil {
					ct.Append([]string{"Destination Account", *r.Destination.Account})
				}

			}
		}
		ct.Render()
		return buf.String()
	}()})

}

func (c *Client) GetBucketReplication(ctx context.Context, bucket string) (*bucketReplication, error) {
	result, err := c.S3Client.GetBucketReplication(ctx, &s3.GetBucketReplicationInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		return nil, err
	}

	replConf := BucketReplicationConf{}
	replConf.FromS3(result.ReplicationConfiguration)
	return &bucketReplication{
		BucketReplicationConf: replConf,
		Bucket:                bucket,
	}, nil
}

func (c *Client) PutBucketReplication(ctx context.Context, bucket string, conf *types.ReplicationConfiguration) error {
	_, err := c.S3Client.PutBucketReplication(ctx, &s3.PutBucketReplicationInput{
		Bucket:                   aws.String(bucket),
		ReplicationConfiguration: conf,
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) DeleteBucketReplication(ctx context.Context, bucket string) error {
	_, err := c.S3Client.DeleteBucketReplication(ctx, &s3.DeleteBucketReplicationInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		return err
	}
	return nil
}
