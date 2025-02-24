package sos_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/exoscale/cli/pkg/storage/sos"
	"github.com/stretchr/testify/assert"
)

func TestClientBucketReplication(t *testing.T) {

	ctx := context.Background()
	bucket := "bucket"
	role := "some-role"
	id := "foo"
	bucketReplica := "bucket-repl"
	c1 := sos.BucketReplicationConf{
		Role: &role,
		Rules: []sos.BucketReplicationRule{
			{
				ID:                      &id,
				Status:                  types.ReplicationRuleStatusEnabled,
				Destination:             &types.Destination{Bucket: &bucketReplica},
				DeleteMarkerReplication: &types.DeleteMarkerReplication{Status: types.DeleteMarkerReplicationStatusDisabled},
				Filter: sos.BucketReplicationConfRuleFilter{
					Prefix: "*",
				},
				ExistingObjectReplication: &types.ExistingObjectReplication{
					Status: types.ExistingObjectReplicationStatusEnabled,
				},
				Priority:                13,
				SourceSelectionCriteria: &types.SourceSelectionCriteria{},
			},
		},
	}

	t.Run("Data type conversion", func(t *testing.T) {

		s3C := c1.ToS3()

		c2 := sos.BucketReplicationConf{}
		c2.FromS3(s3C)

		assert.Equal(t, c1.Role, c2.Role)

		assert.Equal(t, len(c1.Rules), len(c2.Rules))
		assert.Equal(t, len(c1.Rules), 1)

		assert.Equal(t, c1, c2)

	})

	t.Run("Get", func(t *testing.T) {

		client := &sos.Client{
			S3Client: &MockS3API{
				mockGetBucketReplication: func(ctx context.Context, params *s3.GetBucketReplicationInput, optFns ...func(*s3.Options)) (*s3.GetBucketReplicationOutput, error) {
					return &s3.GetBucketReplicationOutput{
						ReplicationConfiguration: &types.ReplicationConfiguration{
							Role: &role,
							Rules: []types.ReplicationRule{
								{
									ID:                      &id,
									Status:                  types.ReplicationRuleStatusEnabled,
									Destination:             &types.Destination{Bucket: &bucketReplica},
									DeleteMarkerReplication: &types.DeleteMarkerReplication{Status: types.DeleteMarkerReplicationStatusDisabled},
									Filter: &types.ReplicationRuleFilterMemberPrefix{
										Value: "*",
									},
									ExistingObjectReplication: &types.ExistingObjectReplication{
										Status: types.ExistingObjectReplicationStatusEnabled,
									},
									Priority:                13,
									SourceSelectionCriteria: &types.SourceSelectionCriteria{},
								},
							},
						},
					}, nil
				},
			},
		}

		g, err := client.GetBucketReplication(ctx, bucket)
		assert.NoError(t, err)

		assert.Equal(t, g.BucketReplicationConf, c1)
	})

}
