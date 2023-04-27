package cmd

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
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

	roundTripper.reqHeaders.Add("User-Agent", fmt.Sprintf("Exoscale-CLI/%s (%s) %s",
		gVersion, gCommit, egoscale.UserAgent))

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

	globalstate.EgoscaleClient = egoscale.NewClient(
		account.CurrentAccount.Endpoint,
		account.CurrentAccount.Key,
		account.CurrentAccount.APISecret(),
		egoscale.WithHTTPClient(httpClient),
		egoscale.WithoutV2Client())

	// During the Exoscale API V1 -> V2 transition, we need to initialize the
	// V2 client independently of the V1 client because of HTTP middleware
	// (http.Transport) clashes.
	// This can be removed once the only API used is V2.
	clientExoV2, err := exov2.NewClient(
		account.CurrentAccount.Key,
		account.CurrentAccount.APISecret(),
		exov2.ClientOptWithAPIEndpoint(account.CurrentAccount.Endpoint),
		exov2.ClientOptWithTimeout(time.Minute*time.Duration(account.CurrentAccount.ClientTimeout)),
		exov2.ClientOptWithHTTPClient(func() *http.Client {
			return &http.Client{
				Transport: newCLIRoundTripper(http.DefaultTransport, account.CurrentAccount.CustomHeaders),
			}
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
	globalstate.EgoscaleClient.Client = clientExoV2

	csRunstatus = egoscale.NewClient(account.CurrentAccount.RunstatusEndpoint,
		account.CurrentAccount.Key,
		account.CurrentAccount.APISecret())
}
