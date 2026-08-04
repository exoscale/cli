package compute

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	v3 "github.com/exoscale/egoscale/v3"
	"github.com/exoscale/egoscale/v3/credentials"
)

func TestResolveTemplateIDPassesUUIDThrough(t *testing.T) {
	const id = "5b473841-fd33-4b88-bcfe-a776abf15034"

	got, err := ResolveTemplateID(context.Background(), nil, id, "public", "ch-gva-2")
	if err != nil {
		t.Fatalf("ResolveTemplateID() error = %v", err)
	}
	if got != v3.UUID(id) {
		t.Errorf("ResolveTemplateID() = %q, want %q", got, id)
	}
}

func TestResolveTemplateGetsOptionalUUIDMetadata(t *testing.T) {
	const id = "5b473841-fd33-4b88-bcfe-a776abf15034"

	tests := []struct {
		name     string
		status   int
		response string
		wantSize int64
	}{
		{
			name:     "metadata available",
			status:   http.StatusOK,
			response: `{"id":"5b473841-fd33-4b88-bcfe-a776abf15034","size":85899345920}`,
			wantSize: 80 << 30,
		},
		{
			name:     "metadata unavailable",
			status:   http.StatusNotFound,
			response: `{"message":"not found"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet || r.URL.Path != "/template/"+id {
					t.Errorf("request = %s %s, want GET /template/%s", r.Method, r.URL.Path, id)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.response))
			}))
			defer server.Close()

			client, err := v3.NewClient(
				credentials.NewStaticCredentials("key", "secret"),
				v3.ClientOptWithEndpoint(v3.Endpoint(server.URL)),
			)
			if err != nil {
				t.Fatalf("NewClient() error = %v", err)
			}

			got, err := ResolveTemplate(context.Background(), client, id, "public", "ch-gva-2")
			if err != nil {
				t.Fatalf("ResolveTemplate() error = %v", err)
			}
			if got.ID != v3.UUID(id) || got.Size != tt.wantSize {
				t.Errorf("ResolveTemplate() = {ID: %q, Size: %d}, want {ID: %q, Size: %d}", got.ID, got.Size, id, tt.wantSize)
			}
		})
	}
}
