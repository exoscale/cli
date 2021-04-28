package v2

import (
	"context"

	apiv2 "github.com/exoscale/egoscale/v2/api"
	papi "github.com/exoscale/egoscale/v2/internal/public-api"
)

// DeployTarget represents a Deploy Target.
type DeployTarget struct {
	Description string
	ID          string
	Name        string
	Type        string
}

func deployTargetFromAPI(d *papi.DeployTarget) *DeployTarget {
	return &DeployTarget{
		Description: papi.OptionalString(d.Description),
		ID:          *d.Id,
		Name:        papi.OptionalString(d.Name),
		Type:        *d.Type,
	}
}

// ListDeployTargets returns the list of existing Deploy Targets in the specified zone.
func (c *Client) ListDeployTargets(ctx context.Context, zone string) ([]*DeployTarget, error) {
	list := make([]*DeployTarget, 0)

	resp, err := c.ListDeployTargetsWithResponse(apiv2.WithZone(ctx, zone))
	if err != nil {
		return nil, err
	}

	if resp.JSON200.DeployTargets != nil {
		for i := range *resp.JSON200.DeployTargets {
			list = append(list, deployTargetFromAPI(&(*resp.JSON200.DeployTargets)[i]))
		}
	}

	return list, nil
}

// GetDeployTarget returns the Deploy Target corresponding to the specified ID in the specified zone.
func (c *Client) GetDeployTarget(ctx context.Context, zone, id string) (*DeployTarget, error) {
	resp, err := c.GetDeployTargetWithResponse(apiv2.WithZone(ctx, zone), id)
	if err != nil {
		return nil, err
	}

	return deployTargetFromAPI(resp.JSON200), nil
}
