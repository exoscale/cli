package utils

import (
	"context"

	"github.com/exoscale/cli/pkg/exoapi"
	v3 "github.com/exoscale/egoscale/v3"
)

// RunAsync decorates and waits for an async operation till success.
func RunAsync(ctx context.Context, client exoapi.Client, message string, f func(context.Context, exoapi.Client) (*v3.Operation, error)) (err error) { //nolint:nonamedreturns
	DecorateAsyncOperation(message, func() {
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
