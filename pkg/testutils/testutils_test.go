package testutils

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/exoapi/mocks"
	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
)

const testZone = v3.ZoneName("ch-dk-2")

func TestNewMockClient(t *testing.T) {
	before := globalstate.EgoscaleV3Client

	t.Run("can switch zone with no setup", func(t *testing.T) {
		client := NewMockClient(t)
		assert.Same(t, client, globalstate.EgoscaleV3Client)

		zoned, err := exocmd.SwitchClientZoneV3(context.Background(), globalstate.EgoscaleV3Client, testZone)
		assert.NoError(t, err)
		assert.Same(t, client, zoned)
	})

	assert.Equal(t, before, globalstate.EgoscaleV3Client, "the previous client must be restored")
}

func TestUseClient(t *testing.T) {
	const endpoint = v3.Endpoint("https://dk2")
	errUnavailable := errors.New("unavailable")
	zoned := mocks.NewMockClient(t)

	tests := []struct {
		name   string
		client func(*mocks.MockClient)
		err    error
	}{
		{
			name: "switches zone through the client",
			client: func(c *mocks.MockClient) {
				c.EXPECT().GetZoneAPIEndpoint(mock.Anything, testZone).Return(endpoint, nil).Once()
				c.EXPECT().WithEndpoint(endpoint).Return(zoned).Once()
			},
		},
		{
			name: "does not retry a failed zone lookup",
			client: func(c *mocks.MockClient) {
				// Once() fails the test if the lookup is retried on another zone.
				c.EXPECT().GetZoneAPIEndpoint(mock.Anything, testZone).Return("", errUnavailable).Once()
			},
			err: errUnavailable,
		},
	}

	for _, tt := range tests {
		before := globalstate.EgoscaleV3Client

		t.Run(tt.name, func(t *testing.T) {
			client := mocks.NewMockClient(t)
			if tt.client != nil {
				tt.client(client)
			}
			UseClient(t, client)
			assert.Equal(t, client, globalstate.EgoscaleV3Client)

			got, err := exocmd.SwitchClientZoneV3(context.Background(), globalstate.EgoscaleV3Client, testZone)
			assert.ErrorIs(t, err, tt.err)
			if tt.err == nil {
				assert.Same(t, zoned, got)
			}
		})

		assert.Equal(t, before, globalstate.EgoscaleV3Client, "the previous client must be restored")
	}
}
