package deployment

import (
	"context"
	"fmt"

	v3 "github.com/exoscale/egoscale/v3"
)

// ResolveDeploymentID resolves a deployment UUID from an ID or a name.
func ResolveDeploymentID(ctx context.Context, client *v3.Client, nameOrID string) (v3.UUID, error) {
	if id, err := v3.ParseUUID(nameOrID); err == nil {
		return id, nil
	}
	resp, err := client.ListDeployments(ctx)
	if err != nil {
		var zero v3.UUID
		return zero, err
	}
	for _, d := range resp.Deployments {
		if d.Name == nameOrID {
			return d.ID, nil
		}
	}
	var zero v3.UUID
	return zero, fmt.Errorf("deployment %q not found", nameOrID)
}
