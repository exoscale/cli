package cmd

import (
	"fmt"
	"net/http"
	"os"

	"github.com/exoscale/egoscale"
	exov2 "github.com/exoscale/egoscale/v2"
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

	for k, v := range headers {
		roundTripper.reqHeaders.Add(k, v)
	}

	return roundTripper
}

func (rt cliRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header = rt.reqHeaders
	return rt.next.RoundTrip(r)
}

func buildClient() {
	if ignoreClientBuild {
		return
	}

	if cs != nil {
		return
	}

	httpClient := &http.Client{Transport: http.DefaultTransport}
	if gCurrentAccount.CustomHeaders != nil {
		httpClient.Transport = newCLIRoundTripper(cs.HTTPClient.Transport, gCurrentAccount.CustomHeaders)
	}

	cs = egoscale.NewClient(
		gCurrentAccount.Endpoint,
		gCurrentAccount.Key,
		gCurrentAccount.APISecret(),
		egoscale.WithHTTPClient(httpClient),
		egoscale.WithoutV2Client())

	// During the Exoscale API V1 -> V2 transition, we need to initialize the
	// V2 client independently from the V1 client because of HTTP middleware
	// (http.Transport) clashes.
	// This can be removed once the only API used is V2.
	clientExoV2, err := exov2.NewClient(
		gCurrentAccount.Key,
		gCurrentAccount.APISecret(),
		exov2.ClientOptWithAPIEndpoint(gCurrentAccount.Endpoint),
		exov2.ClientOptWithHTTPClient(func() *http.Client {
			hc := &http.Client{Transport: http.DefaultTransport}
			if gCurrentAccount.CustomHeaders != nil {
				hc.Transport = newCLIRoundTripper(hc.Transport, gCurrentAccount.CustomHeaders)
			}
			return hc
		}()),
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
	cs.Client = clientExoV2

	csDNS = egoscale.NewClient(gCurrentAccount.DNSEndpoint,
		gCurrentAccount.Key,
		gCurrentAccount.APISecret())

	csRunstatus = egoscale.NewClient(gCurrentAccount.RunstatusEndpoint,
		gCurrentAccount.Key,
		gCurrentAccount.APISecret())
}
