//go:build integration_api
// +build integration_api

package integration_with_api_test

import (
	"fmt"
	"math/rand"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/exoscale/cli/internal/integ"
	"github.com/exoscale/cli/pkg/storage/sos"
	"github.com/stretchr/testify/require"
)

func TestLifecycle(t *testing.T) {
	// Write a temporary lifecycle configuration file consumed by the set command.
	lcJSON := `{
    "Rules": [
        {
            "Status": "Enabled",
            "Expiration": { "Days": 30 },
            "Filter": { "Prefix": "logs/" },
            "ID": "expire-logs-after-30-days"
        },
        {
            "Status": "Enabled",
            "Expiration": { "Days": 90 },
            "Filter": { "Prefix": "backups/" },
            "ID": "expire-backups-after-90-days"
        },
        {
            "Status": "Enabled",
            "Expiration": { "Days": 7 },
            "Filter": { "Prefix": "temp/" },
            "ID": "expire-temp-after-7-days"
        }
    ]
}`

	tmpFile, err := os.CreateTemp("", "lifecycle-*.json")
	require.NoError(t, err)
	_, err = tmpFile.WriteString(lcJSON)
	require.NoError(t, err)
	require.NoError(t, tmpFile.Close())
	defer os.Remove(tmpFile.Name())

	bucketName := fmt.Sprintf("test-lifecycle-bucket-%d", rand.Int())

	expectedShow := sos.BucketLifecycle{
		Bucket: bucketName,
		BucketLifecycleConf: sos.BucketLifecycleConf{
			Rules: []sos.BucketLifecycleRule{
				{
					Status: types.ExpirationStatus("Enabled"),
					ID:     aws.String("expire-logs-after-30-days"),
					Filter: &sos.BucketLifecycleConfRuleFilter{
						Prefix: aws.String("logs/"),
					},
					Expiration: &types.LifecycleExpiration{
						Days: aws.Int32(30),
					},
				},
				{
					Status: types.ExpirationStatus("Enabled"),
					ID:     aws.String("expire-backups-after-90-days"),
					Filter: &sos.BucketLifecycleConfRuleFilter{
						Prefix: aws.String("backups/"),
					},
					Expiration: &types.LifecycleExpiration{
						Days: aws.Int32(90),
					},
				},
				{
					Status: types.ExpirationStatus("Enabled"),
					ID:     aws.String("expire-temp-after-7-days"),
					Filter: &sos.BucketLifecycleConfRuleFilter{
						Prefix: aws.String("temp/"),
					},
					Expiration: &types.LifecycleExpiration{
						Days: aws.Int32(7),
					},
				},
			},
		},
	}

	params := struct {
		BucketName string
		LCFilePath string
	}{
		BucketName: bucketName,
		LCFilePath: tmpFile.Name(),
	}

	s := integ.Suite{
		Zone:       "ch-dk-2",
		Parameters: params,
		Steps: []integ.Step{
			{
				Description: "create bucket",
				Command:     "exo storage mb sos://{{.BucketName}}",
			},
			{
				Description: "put lifecycle configuration",
				Command:     "exo storage bucket lifecycle set sos://{{.BucketName}} {{.LCFilePath}}",
			},
			{
				Description: "get lifecycle configuration",
				Command:     "exo storage bucket lifecycle show sos://{{.BucketName}}",
				Expected:    expectedShow,
			},
			{
				Description: "delete lifecycle configuration",
				Command:     "exo storage bucket lifecycle delete sos://{{.BucketName}}",
			},
			{
				Description: "delete bucket",
				Command:     "exo storage rb sos://{{.BucketName}} -f",
				NoZone:      true,
			},
		},
		T: t,
	}

	s.Run()
}
