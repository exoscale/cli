package dedicated_inference

import (
	"context"
	"fmt"

	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

// int64PtrIfNonZero returns a pointer to v if it's non-zero, otherwise nil.
func int64PtrIfNonZero(v int64) *int64 {
	if v == 0 {
		return nil
	}
	return &v
}

// runAsync decorates and waits for an async operation till success.
func runAsync(ctx context.Context, client *v3.Client, message string, f func(context.Context, *v3.Client) (*v3.Operation, error)) (err error) { //nolint:nonamedreturns
	utils.DecorateAsyncOperation(message, func() {
		op, e := f(ctx, client)
		if e != nil {
			err = e
			return
		}
		_, e = client.Wait(ctx, op, v3.OperationStateSuccess)
		if e != nil {
			err = e
		}
	})
	return
}

// resolveDeploymentID resolves a deployment UUID from an ID or a name.
func resolveDeploymentID(ctx context.Context, client *v3.Client, nameOrID string) (v3.UUID, error) {
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
