package cmd

import (
	"net/http"

	"github.com/exoscale/egoscale"
)

// roundTripper implements the http.RoundTripper interface and allows client
// request customization, such as HTTP headers injection. If provided with a
// non-nil rt parameter, it will wrap around it when performing requests.
type roundTripper struct {
	reqHeaders http.Header
	rt         http.RoundTripper
}

func newRoundTripper(rt http.RoundTripper, headers map[string]string) roundTripper {
	var roundTripper = roundTripper{
		rt:         http.DefaultTransport,
		reqHeaders: http.Header{},
	}

	if rt != nil {
		roundTripper.rt = rt
	}

	for k, v := range headers {
		roundTripper.reqHeaders.Add(k, v)
	}

	return roundTripper
}

func (rt roundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header = rt.reqHeaders
	return rt.rt.RoundTrip(r)
}

func buildClient() {
	if ignoreClientBuild {
		return
	}

	if cs != nil {
		return
	}

	cs = egoscale.NewClient(gCurrentAccount.Endpoint,
		gCurrentAccount.Key,
		gCurrentAccount.APISecret())

	if gCurrentAccount.CustomHeaders != nil {
		cs.HTTPClient.Transport = newRoundTripper(cs.HTTPClient.Transport, gCurrentAccount.CustomHeaders)
	}

	csDNS = egoscale.NewClient(gCurrentAccount.DNSEndpoint,
		gCurrentAccount.Key,
		gCurrentAccount.APISecret())

	csRunstatus = egoscale.NewClient(gCurrentAccount.RunstatusEndpoint,
		gCurrentAccount.Key,
		gCurrentAccount.APISecret())
}
