package sks

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	egoscale "github.com/exoscale/egoscale/v3"
	"github.com/exoscale/egoscale/v3/credentials"
)

var (
	fakeClusterVersions = []string{"1.36.1", "1.35.2", "1.34.5"}
)

type fakeSKSVersionsTransport struct {
	versions []string
}

func (t *fakeSKSVersionsTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	body, err := json.Marshal(map[string]any{
		"sks-cluster-versions": fakeClusterVersions,
	})
	if err != nil {
		return nil, err
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(body)),
		Header:     make(http.Header),
	}, nil
}

func TestResolveSKSClusterVersions(t *testing.T) {
	t.Parallel()

	fakeHttpClient := &http.Client{
		Transport: &fakeSKSVersionsTransport{fakeClusterVersions},
	}

	client, err := egoscale.NewClient(
		credentials.NewStaticCredentials("foo", "bar"),
		egoscale.ClientOptWithEndpoint(egoscale.CHGva2),
		egoscale.ClientOptWithHTTPClient(fakeHttpClient),
	)
	if err != nil {
		t.Fatalf("failed to create egoscale dummy client: %v", err)
	}

	ctx := context.Background()

	tests := []struct {
		name            string
		inputVersion    string
		expectedVersion string
		errorMsg        string
	}{
		{
			name:            "major.minor resolves to corresponding major.minor.patch",
			inputVersion:    strings.Join(strings.Split(fakeClusterVersions[1], ".")[:2], "."),
			expectedVersion: fakeClusterVersions[1],
			errorMsg:        "",
		},
		{
			name:            "empty version resolves to latest version",
			inputVersion:    "",
			expectedVersion: fakeClusterVersions[0],
			errorMsg:        "",
		},
		{
			name:            "",
			inputVersion:    fakeClusterVersions[1],
			expectedVersion: fakeClusterVersions[1],
			errorMsg:        "",
		},
		{
			name:            "unsupported major.minor throws error",
			inputVersion:    "1.2", // unsupported
			expectedVersion: "",
			errorMsg:        "the SKS cluster version 1.2 is not supported",
		},
		{
			name:            "version fits no format",
			inputVersion:    "foo", // bad format
			expectedVersion: "",
			errorMsg:        "error resolving the provided SKS cluster version: foo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := resolveSKSClusterVersion(ctx, client, tt.inputVersion)
			assert.Equal(t, tt.expectedVersion, result)
			if tt.errorMsg != "" {
				assert.ErrorContains(t, err, tt.errorMsg)
			}
		})
	}
}
