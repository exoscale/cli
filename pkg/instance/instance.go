package instance

import (
	"context"
	"fmt"

	egov3 "github.com/exoscale/egoscale/v3"
	"github.com/exoscale/egoscale/v3/oapi"
)

func FindInstanceByName(ctx context.Context, client *egov3.Client, name string) (*oapi.InstancesListElement, error) {
	instanceList, err := client.Compute().Instance().List(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list instances: %w", err)
	}

	for _, instance := range instanceList {
		if *instance.Name == name {
			return &instance, nil
		}
	}

	return nil, nil
}
