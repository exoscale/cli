package v2

import (
	"context"

	apiv2 "github.com/exoscale/egoscale/v2/api"
	papi "github.com/exoscale/egoscale/v2/internal/public-api"
)

// InstanceType represents a Compute instance type.
type InstanceType struct {
	Authorized bool
	CPUs       int64
	Family     string
	GPUs       int64
	ID         string
	Memory     int64
	Size       string
}

func instanceTypeFromAPI(t *papi.InstanceType) *InstanceType {
	return &InstanceType{
		Authorized: *t.Authorized,
		CPUs:       *t.Cpus,
		Family:     *t.Family,
		GPUs:       papi.OptionalInt64(t.Gpus),
		ID:         *t.Id,
		Memory:     *t.Memory,
		Size:       *t.Size,
	}
}

// ListInstanceTypes returns the list of existing Instance types in the specified zone.
func (c *Client) ListInstanceTypes(ctx context.Context, zone string) ([]*InstanceType, error) {
	list := make([]*InstanceType, 0)

	resp, err := c.ListInstanceTypesWithResponse(apiv2.WithZone(ctx, zone))
	if err != nil {
		return nil, err
	}

	if resp.JSON200.InstanceTypes != nil {
		for i := range *resp.JSON200.InstanceTypes {
			list = append(list, instanceTypeFromAPI(&(*resp.JSON200.InstanceTypes)[i]))
		}
	}

	return list, nil
}

// GetInstanceType returns the Instance type corresponding to the specified ID in the specified zone.
func (c *Client) GetInstanceType(ctx context.Context, zone, id string) (*InstanceType, error) {
	resp, err := c.GetInstanceTypeWithResponse(apiv2.WithZone(ctx, zone), id)
	if err != nil {
		return nil, err
	}

	return instanceTypeFromAPI(resp.JSON200), nil
}
