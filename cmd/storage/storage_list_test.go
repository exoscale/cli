package storage

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/storage/sos"
)

func TestStorageListUsesExplicitZoneForBucket(t *testing.T) {
	const (
		bucket      = "escape-test"
		defaultZone = "ch-gva-2"
		zone        = "ch-dk-2"
	)

	var requests []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, fmt.Sprintf("%s %s", r.Method, r.URL.Path))
		if r.Method == http.MethodHead {
			w.Header().Set("X-Amz-Bucket-Region", zone)
			return
		}

		w.Header().Set("Content-Type", "application/xml")
		_, _ = fmt.Fprint(w, `<ListBucketResult xmlns="http://s3.amazonaws.com/doc/2006-03-01/"><IsTruncated>false</IsTruncated></ListBucketResult>`)
	}))
	t.Cleanup(server.Close)

	oldAccount := account.CurrentAccount
	oldContext := exocmd.GContext
	oldOpts := sos.CommonConfigOptFns
	t.Cleanup(func() {
		account.CurrentAccount = oldAccount
		exocmd.GContext = oldContext
		sos.CommonConfigOptFns = oldOpts
	})

	account.CurrentAccount = &account.Account{
		Key:         "key",
		Secret:      "secret",
		DefaultZone: defaultZone,
		SosEndpoint: server.URL + "/{zone}",
	}
	exocmd.GContext = context.Background()
	sos.CommonConfigOptFns = nil

	require.NoError(t, storageListCmd.Flags().Set("zone", zone))
	t.Cleanup(func() { require.NoError(t, storageListCmd.Flags().Set("zone", "")) })

	require.NoError(t, storageListCmd.RunE(storageListCmd, []string{bucket}))
	require.Equal(t, []string{"GET /ch-dk-2/escape-test"}, requests)
}
