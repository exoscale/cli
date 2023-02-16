package utils

import (
	"context"
	"crypto/rand"
	"encoding/base64"

	"github.com/exoscale/egoscale"
	v2 "github.com/exoscale/egoscale/v2"
)

// RandStringBytes Generate random string of n bytes
func RandStringBytes(n int) (string, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

func GetInstancesInSecurityGroup(ctx context.Context, client *egoscale.Client, securityGroupID, zone string) ([]*v2.Instance, error) {
	allInstances, err := client.ListInstances(ctx, zone)
	if err != nil {
		return nil, err
	}

	var instancesInSG []*v2.Instance
	for _, instance := range allInstances {
		for _, sgID := range *instance.SecurityGroupIDs {
			if sgID == securityGroupID {
				instancesInSG = append(instancesInSG, instance)
			}
		}
	}

	return instancesInSG, nil
}

func GetDefaultStringOrZero(v *string, defaultStr string) string {
	if v == nil {
		return defaultStr
	}

	return *v
}
