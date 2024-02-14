package cmd

import (
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

	clientExoV2, err := exov2.NewClient(
		account.CurrentAccount.Key,
		account.CurrentAccount.APISecret(),
		exov2.ClientOptWithTimeout(time.Minute*time.Duration(account.CurrentAccount.ClientTimeout)),
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

	if v := os.Getenv("EXOSCALE_TRACE"); v != "" {
		clientV3 = clientV3.WithTrace()
	}
	globalstate.EgoscaleV3Client = clientV3
}
