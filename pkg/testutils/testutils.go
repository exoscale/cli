package testutils

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/exoapi"
	"github.com/exoscale/cli/pkg/exoapi/mocks"
	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/exoscale/egoscale/v3/credentials"
	"github.com/stretchr/testify/mock"
)

const mockEndpoint = v3.Endpoint("https://api.invalid/v2")

// WriteJSON encodes v to w as JSON, and sets the response Content-Type to
// application/json with the given status code.
func WriteJSON(t *testing.T, w http.ResponseWriter, code int, v any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		t.Fatalf("encode json: %v", err)
	}
}

// NewV3Client returns a v3.Client with test credentials and the given endpoint.
func NewV3Client(t *testing.T, endpoint string) exoapi.Client {
	t.Helper()
	creds := credentials.NewStaticCredentials("key", "secret")
	client, err := v3.NewClient(creds)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	return exoapi.New(client).WithEndpoint(v3.Endpoint(endpoint))
}

// SetupV3Client creates a v3.Client and sets it in the global state.
func SetupV3Client(t *testing.T, endpoint string) {
	t.Helper()
	exocmd.GContext = context.Background()
	globalstate.Quiet = true
	globalstate.EgoscaleV3Client = NewV3Client(t, endpoint)
}

// NewMockClient returns a mock client and sets it in the global state with
// UseClient. WithEndpoint returns the mock itself and GetZoneAPIEndpoint returns
// a dummy endpoint, so that commands switching zone need no specific setup.
//
// These two default expectations match any argument and take precedence over
// the ones added later: to assert on them, for instance to return a different
// client per zone, pass your own mocks.NewMockClient to UseClient instead.
func NewMockClient(t *testing.T) *mocks.MockClient {
	t.Helper()
	client := mocks.NewMockClient(t)
	client.EXPECT().WithEndpoint(mock.Anything).Return(client).Maybe()
	client.EXPECT().GetZoneAPIEndpoint(mock.Anything, mock.Anything).Return(mockEndpoint, nil).Maybe()
	UseClient(t, client)
	return client
}

// UseClient sets client in the global state for the duration of the test, along
// with the other globals commands rely on.
func UseClient(t *testing.T, client exoapi.Client) {
	t.Helper()
	oldClient, oldContext, oldQuiet, oldAccount := globalstate.EgoscaleV3Client, exocmd.GContext, globalstate.Quiet, account.CurrentAccount
	t.Cleanup(func() {
		globalstate.EgoscaleV3Client, exocmd.GContext, globalstate.Quiet, account.CurrentAccount = oldClient, oldContext, oldQuiet, oldAccount
	})

	globalstate.EgoscaleV3Client = client
	exocmd.GContext = context.Background()
	globalstate.Quiet = true
	// A configured endpoint disables the retry of failed zone lookups on another
	// zone, which would otherwise call the mock a second time.
	account.CurrentAccount = &account.Account{Endpoint: string(mockEndpoint)}
}
