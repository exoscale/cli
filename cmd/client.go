package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	exov2 "github.com/exoscale/egoscale/v2"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/exoscale/egoscale/v3/credentials"
)

// cliRoundTripper implements the http.RoundTripper interface and allows client
// request customization, such as HTTP headers injection. If provided with a
// non-nil next parameter, it will wrap around it when performing requests.
type cliRoundTripper struct {
	next http.RoundTripper

	reqHeaders http.Header
}

func newCLIRoundTripper(next http.RoundTripper, headers map[string]string) cliRoundTripper {
	roundTripper := cliRoundTripper{
		next:       http.DefaultTransport,
		reqHeaders: http.Header{},
	}

	if next != nil {
		roundTripper.next = next
	}

	roundTripper.reqHeaders.Add("User-Agent", fmt.Sprintf("Exoscale-CLI/%s (%s) %s",
		gVersion, gCommit, exov2.UserAgent))

	for k, v := range headers {
		roundTripper.reqHeaders.Add(k, v)
	}

	return roundTripper
}

func (rt cliRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	for h := range rt.reqHeaders {
		r.Header.Add(h, rt.reqHeaders.Get(h))
	}

	return rt.next.RoundTrip(r)
}

func buildClient() {
	if ignoreClientBuild {
		return
	}

	if globalstate.EgoscaleClient != nil {
		return
	}

	httpClient := &http.Client{Transport: newCLIRoundTripper(http.DefaultTransport, account.CurrentAccount.CustomHeaders)}

	clientTimeout := account.CurrentAccount.ClientTimeout
	if clientTimeout == 0 {
		clientTimeout = defaultClientTimeout
	}
	clientExoV2, err := exov2.NewClient(
		account.CurrentAccount.Key,
		account.CurrentAccount.APISecret(),
		exov2.ClientOptWithTimeout(time.Minute*time.Duration(clientTimeout)),
		exov2.ClientOptWithHTTPClient(httpClient),
		exov2.ClientOptCond(func() bool {
			if v := os.Getenv("EXOSCALE_TRACE"); v != "" {
				return true
			}
			return false
		}, exov2.ClientOptWithTrace()),
	)
	if err != nil {
		panic(fmt.Sprintf("unable to initialize Exoscale API V2 client: %v", err))
	}
	globalstate.EgoscaleClient = clientExoV2

	creds := credentials.NewStaticCredentials(
		account.CurrentAccount.Key,
		account.CurrentAccount.APISecret(),
	)

	clientV3, err := v3.NewClient(
		creds,
		v3.ClientOptWithHTTPClient(httpClient),
	)
	if err != nil {
		panic(fmt.Sprintf("unable to initialize Exoscale API V3 client: %v", err))
	}

	if account.CurrentAccount.Endpoint != "" {
		clientV3 = clientV3.WithEndpoint(v3.Endpoint(account.CurrentAccount.Endpoint))
	}

	if v := os.Getenv("EXOSCALE_TRACE"); v != "" {
		clientV3 = clientV3.WithTrace()
	}
	globalstate.EgoscaleV3Client = clientV3
}

func switchClientZoneV3(ctx context.Context, client *v3.Client, zone v3.ZoneName) (*v3.Client, error) {
	if zone == "" {
		return client, nil
	}
	endpoint, err := client.GetZoneAPIEndpoint(ctx, zone)
	if err != nil {
		return nil, fmt.Errorf("switch client zone v3: %w", err)
	}

	return client.WithEndpoint(endpoint), nil
}
