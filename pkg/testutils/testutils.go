package testutils

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/exoscale/egoscale/v3/credentials"
)

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
func NewV3Client(t *testing.T, endpoint string) *v3.Client {
	t.Helper()
	creds := credentials.NewStaticCredentials("key", "secret")
	client, err := v3.NewClient(creds)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	return client.WithEndpoint(v3.Endpoint(endpoint))
}

// SetupV3Client creates a v3.Client and sets it in the global state.
func SetupV3Client(t *testing.T, endpoint string) {
	t.Helper()
	exocmd.GContext = context.Background()
	globalstate.Quiet = true
	globalstate.EgoscaleV3Client = NewV3Client(t, endpoint)
}
