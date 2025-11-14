package utils

import (
	"context"

	v3 "github.com/exoscale/egoscale/v3"
)

// RunAsync decorates and waits for an async operation till success.
func RunAsync(ctx context.Context, client *v3.Client, message string, f func(context.Context, *v3.Client) (*v3.Operation, error)) (err error) { //nolint:nonamedreturns
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
