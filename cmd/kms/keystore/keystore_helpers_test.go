package keystore

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/exoscale/cli/pkg/testutils"
	v3 "github.com/exoscale/egoscale/v3"
)

const testKeyStoreID = v3.UUID("2b1e4e2a-6b0c-4f5c-9d6e-6f1f0f3a7c11")

// unknownKeyStoreID is a well-formed ID that the fake API does not know.
const unknownKeyStoreID = "8f3e6a1c-2d4b-4c5e-9f7a-1b2c3d4e5f60"

// fakeKeyStoreAPI is an in-memory key store API recording the calls it receives.
type fakeKeyStoreAPI struct {
	stores    map[v3.UUID]*v3.GetKeyStoreResponse
	calls     []string
	createReq *v3.CreateKeyStoreRequest
	updateReq *v3.UpdateKeyStoreRequest
}

func newFakeKeyStoreAPI(t *testing.T) *fakeKeyStoreAPI {
	t.Helper()

	api := &fakeKeyStoreAPI{stores: map[v3.UUID]*v3.GetKeyStoreResponse{}}

	mux := http.NewServeMux()
	mux.HandleFunc("/key-store", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			api.calls = append(api.calls, "list")
			var list v3.ListKeyStoresResponse
			for _, ks := range api.stores {
				list.KeyStores = append(list.KeyStores, v3.ListKeyStoresResponseEntry{
					ID:     ks.ID,
					Name:   ks.Name,
					Type:   v3.ListKeyStoresResponseEntryType(ks.Type),
					Status: v3.ListKeyStoresResponseEntryStatus(ks.Status),
					Proxy:  ks.Proxy,
				})
			}
			testutils.WriteJSON(t, w, http.StatusOK, list)
		case http.MethodPost:
			api.calls = append(api.calls, "create")
			var req v3.CreateKeyStoreRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatalf("decode create request: %v", err)
			}
			api.createReq = &req
			ks := &v3.GetKeyStoreResponse{
				ID:     testKeyStoreID,
				Name:   req.Name,
				Type:   v3.GetKeyStoreResponseTypeExternalKeyStore,
				Status: v3.GetKeyStoreResponseStatusDisconnected,
				Proxy: &v3.KeyStoreProxyResponse{
					Endpoint: req.Proxy.Endpoint,
					Auth:     &v3.KeyStoreProxyAuthResponse{Key: req.Proxy.Auth.Key},
				},
			}
			api.stores[ks.ID] = ks
			testutils.WriteJSON(t, w, http.StatusOK, v3.ListKeyStoresResponseEntry{ID: ks.ID, Name: ks.Name})
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/key-store/", func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/key-store/"), "/")
		ks, ok := api.stores[v3.UUID(parts[0])]
		if !ok {
			testutils.WriteJSON(t, w, http.StatusNotFound, v3.ErrorResponse{Status: http.StatusNotFound, Title: "Not Found", Detail: "key store not found"})
			return
		}

		action := ""
		if len(parts) > 1 {
			action = parts[1]
		}

		switch {
		case r.Method == http.MethodGet && action == "":
			api.calls = append(api.calls, "get")
			testutils.WriteJSON(t, w, http.StatusOK, ks)
		case r.Method == http.MethodDelete && action == "":
			api.calls = append(api.calls, "delete")
			delete(api.stores, ks.ID)
			testutils.WriteJSON(t, w, http.StatusOK, v3.SuccessResponse{Status: "success"})
		case r.Method == http.MethodPost && action == "connect":
			api.calls = append(api.calls, "connect")
			ks.Status = v3.GetKeyStoreResponseStatusConnected
			testutils.WriteJSON(t, w, http.StatusOK, v3.SuccessResponse{Status: "success"})
		case r.Method == http.MethodPost && action == "disconnect":
			api.calls = append(api.calls, "disconnect")
			ks.Status = v3.GetKeyStoreResponseStatusDisconnected
			testutils.WriteJSON(t, w, http.StatusOK, v3.SuccessResponse{Status: "success"})
		case r.Method == http.MethodPost && action == "update":
			api.calls = append(api.calls, "update")
			var req v3.UpdateKeyStoreRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatalf("decode update request: %v", err)
			}
			api.updateReq = &req
			if req.Description != "" {
				ks.Description = req.Description
			}
			if req.Proxy != nil && req.Proxy.Endpoint != "" {
				ks.Proxy.Endpoint = req.Proxy.Endpoint
			}
			testutils.WriteJSON(t, w, http.StatusOK, ks)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	testutils.SetupV3Client(t, srv.URL)

	return api
}

// addKeyStore seeds the fake API with an existing key store.
func (api *fakeKeyStoreAPI) addKeyStore(name string, status v3.GetKeyStoreResponseStatus) *v3.GetKeyStoreResponse {
	ks := &v3.GetKeyStoreResponse{
		ID:     testKeyStoreID,
		Name:   name,
		Type:   v3.GetKeyStoreResponseTypeExternalKeyStore,
		Status: status,
		Proxy: &v3.KeyStoreProxyResponse{
			Endpoint: "https://xks.example.com",
			Auth:     &v3.KeyStoreProxyAuthResponse{Key: "AKIDEXAMPLE"},
		},
		Health: &v3.KeyStoreHealth{Status: v3.KeyStoreHealthStatusHealthy},
	}
	api.stores[ks.ID] = ks
	return ks
}

func (api *fakeKeyStoreAPI) called(call string) bool {
	for _, c := range api.calls {
		if c == call {
			return true
		}
	}
	return false
}
