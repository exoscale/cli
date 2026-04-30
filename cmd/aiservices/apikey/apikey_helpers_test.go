package apikey

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	v3 "github.com/exoscale/egoscale/v3"
	"github.com/exoscale/egoscale/v3/credentials"
)

type apikeyHelperServer struct {
	server *httptest.Server
	keys   []v3.AIAPIKey
}

func newAPIKeyHelperServer(t *testing.T) *apikeyHelperServer {
	ts := &apikeyHelperServer{}
	mux := http.NewServeMux()
	mux.HandleFunc("/ai/api-key", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			writeJSON(t, w, http.StatusOK, v3.ListAIAPIKeysResponse{AIAPIKeys: ts.keys})
		case http.MethodPost:
			writeJSON(t, w, http.StatusOK, v3.AIAPIKeyWithValue{
				ID:        v3.UUID("new-key-id"),
				Name:      "new-key",
				Scope:     "public",
				CreatedAT: time.Now(),
				UpdatedAT: time.Now(),
				Value:     "exo_ai_test_value",
			})
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/ai/api-key/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path[len("/ai/api-key/"):]
		if r.Method == http.MethodPost && strings.HasSuffix(path, "/rotate") {
			id := strings.TrimSuffix(path, "/rotate")
			writeJSON(t, w, http.StatusOK, v3.AIAPIKeyWithValue{
				ID:        v3.UUID(id),
				Name:      "key",
				Scope:     "public",
				CreatedAT: time.Now(),
				UpdatedAT: time.Now(),
				Value:     "exo_ai_rotated_value",
			})
			return
		}
		if r.Method == http.MethodPost && strings.HasSuffix(path, "/reveal") {
			id := strings.TrimSuffix(path, "/reveal")
			writeJSON(t, w, http.StatusOK, v3.AIAPIKeyWithValue{
				ID:        v3.UUID(id),
				Name:      "key",
				Scope:     "public",
				CreatedAT: time.Now(),
				UpdatedAT: time.Now(),
				Value:     "exo_ai_revealed_value",
			})
			return
		}
		id := path
		switch r.Method {
		case http.MethodGet:
			for _, k := range ts.keys {
				if string(k.ID) == id {
					writeJSON(t, w, http.StatusOK, k)
					return
				}
			}
			w.WriteHeader(http.StatusNotFound)
		case http.MethodPatch:
			writeJSON(t, w, http.StatusOK, v3.AIAPIKey{
				ID:        v3.UUID(id),
				Name:      "updated-key",
				Scope:     "public",
				CreatedAT: time.Now(),
				UpdatedAT: time.Now(),
			})
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
	ts.server = httptest.NewServer(mux)
	return ts
}

func writeJSON(t *testing.T, w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		t.Fatalf("encode json: %v", err)
	}
}

func TestFindAIAPIKeyByIDAndName(t *testing.T) {
	ts := newAPIKeyHelperServer(t)
	defer ts.server.Close()
	now := time.Now()
	ts.keys = []v3.AIAPIKey{
		{ID: v3.UUID("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"), Name: "alpha", Scope: "public", CreatedAT: now, UpdatedAT: now},
	}
	creds := credentials.NewStaticCredentials("key", "secret")
	client, err := v3.NewClient(creds)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	client = client.WithEndpoint(v3.Endpoint(ts.server.URL))
	ctx := context.Background()

	// by ID
	list, err := client.ListAIAPIKeys(ctx)
	if err != nil {
		t.Fatalf("list api keys: %v", err)
	}
	entry, err := list.FindAIAPIKey("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	if err != nil || string(entry.ID) != "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa" {
		t.Fatalf("resolve by id failed: %v %v", entry.ID, err)
	}
	// by name
	entry, err = list.FindAIAPIKey("alpha")
	if err != nil || string(entry.ID) != "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa" {
		t.Fatalf("resolve by name failed: %v %v", entry.ID, err)
	}
	// not found
	_, err = list.FindAIAPIKey("missing")
	if err == nil {
		t.Fatalf("expected not found error, got %v", err)
	}
}
