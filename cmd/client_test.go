package cmd

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/exoapi"
	"github.com/exoscale/cli/pkg/exoapi/mocks"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

func setMockGlobals(t *testing.T) {
	t.Helper()
	oldQuiet, oldAccount := globalstate.Quiet, account.CurrentAccount
	t.Cleanup(func() { globalstate.Quiet, account.CurrentAccount = oldQuiet, oldAccount })
	globalstate.Quiet = true
	account.CurrentAccount = &account.Account{Endpoint: "https://api.invalid/v2"}
}

func TestSwitchClientZoneV3(t *testing.T) {
	setMockGlobals(t)

	const zone = v3.ZoneName("ch-dk-2")
	const endpoint = v3.Endpoint("https://dk2")
	errUnavailable := errors.New("unavailable")
	zoned := mocks.NewMockClient(t)

	tests := []struct {
		name      string
		zone      v3.ZoneName
		client    func(*mocks.MockClient)
		wantZoned bool
		err       error
	}{
		{
			name: "no zone keeps the client",
		},
		{
			name: "zone returns a client for its endpoint",
			zone: zone,
			client: func(c *mocks.MockClient) {
				c.EXPECT().GetZoneAPIEndpoint(mock.Anything, zone).Return(endpoint, nil).Once()
				c.EXPECT().WithEndpoint(endpoint).Return(zoned).Once()
			},
			wantZoned: true,
		},
		{
			name: "lookup failure is returned",
			zone: zone,
			client: func(c *mocks.MockClient) {
				c.EXPECT().GetZoneAPIEndpoint(mock.Anything, zone).Return("", errUnavailable).Once()
			},
			err: errUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := mocks.NewMockClient(t)
			if tt.client != nil {
				tt.client(client)
			}

			got, err := SwitchClientZoneV3(context.Background(), client, tt.zone)
			assert.ErrorIs(t, err, tt.err)
			if tt.wantZoned {
				assert.Same(t, zoned, got)
			} else if tt.err == nil {
				assert.Same(t, client, got)
			}
		})
	}
}

func TestRunAsyncWithMock(t *testing.T) {
	setMockGlobals(t)

	op := &v3.Operation{ID: "op"}
	errRefused := errors.New("refused")
	errFailed := errors.New("operation failed")

	tests := []struct {
		name    string
		client  func(*mocks.MockClient)
		callErr error
		err     error
	}{
		{
			name: "waits for the operation",
			client: func(c *mocks.MockClient) {
				c.EXPECT().Wait(mock.Anything, op, v3.OperationStateSuccess).Return(op, nil).Once()
			},
		},
		{
			name:    "does not wait when the call fails",
			callErr: errRefused,
			err:     errRefused,
		},
		{
			name: "returns the wait failure",
			client: func(c *mocks.MockClient) {
				c.EXPECT().Wait(mock.Anything, op, v3.OperationStateSuccess).Return(nil, errFailed).Once()
			},
			err: errFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := mocks.NewMockClient(t)
			if tt.client != nil {
				tt.client(client)
			}

			err := utils.RunAsync(context.Background(), client, "working", func(context.Context, exoapi.Client) (*v3.Operation, error) {
				return op, tt.callErr
			})
			assert.ErrorIs(t, err, tt.err)
		})
	}
}
