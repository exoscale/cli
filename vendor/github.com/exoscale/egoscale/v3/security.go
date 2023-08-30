package v3

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"
)

// ReqExpire represents the request expiration duration.
const RequestExpire = 10 * time.Minute

// SecurityProvider is an request interceptor that implements [Exoscale API Request Signature] authentication.
// It is used by default in the Client.
//
// [API Request Signature]: https://openapi-v2.exoscale.com/#topic-api-request-signature
type SecurityProvider struct {
	creds *Credentials
}

// NewSecurityProvider creates a SecurityProvider interceptor using provided credentials.
// Credentials struct is passed as a pointer as it is safe for concurrent use.
func NewSecurityProvider(creds *Credentials) *SecurityProvider {
	return &SecurityProvider{
		creds: creds,
	}
}

// Intercept is oapi.RequestEditorFn that will attach authentication header to API call.
func (p *SecurityProvider) Intercept(ctx context.Context, req *http.Request) error {
	var (
		sigParts    []string
		headerParts []string
	)

	// Request method/URL path
	sigParts = append(sigParts, fmt.Sprintf("%s %s", req.Method, req.URL.EscapedPath()))
	headerParts = append(headerParts, "EXO2-HMAC-SHA256 credential="+p.creds.APIKey())

	// Request body if present
	body := ""
	if req.Body != nil {
		data, err := io.ReadAll(req.Body)
		if err != nil {
			return err
		}
		err = req.Body.Close()
		if err != nil {
			return err
		}
		body = string(data)
		req.Body = io.NopCloser(bytes.NewReader(data))
	}
	sigParts = append(sigParts, body)

	// Request query string parameters
	// Important: this is order-sensitive, we have to have to sort parameters alphabetically to ensure signed
	// values match the names listed in the "signed-query-args=" signature pragma.
	signedParams, paramsValues := extractRequestParameters(req)
	sigParts = append(sigParts, paramsValues)
	if len(signedParams) > 0 {
		headerParts = append(headerParts, "signed-query-args="+strings.Join(signedParams, ";"))
	}

	// Request headers -- none at the moment
	// Note: the same order-sensitive caution for query string parameters applies to headers.
	sigParts = append(sigParts, "")

	// Request expiration date (UNIX timestamp, no line return)
	sigParts = append(sigParts, fmt.Sprint(time.Now().Add(RequestExpire).Unix()))
	headerParts = append(headerParts, "expires="+fmt.Sprint(time.Now().Add(RequestExpire).Unix()))

	h := hmac.New(sha256.New, []byte(p.creds.APISecret()))
	if _, err := h.Write([]byte(strings.Join(sigParts, "\n"))); err != nil {
		return err
	}
	headerParts = append(headerParts, "signature="+base64.StdEncoding.EncodeToString(h.Sum(nil)))

	req.Header.Set("Authorization", strings.Join(headerParts, ","))

	return nil

}

// extractRequestParameters returns the list of request URL parameters names
// and a strings concatenating the values of the parameters.
func extractRequestParameters(req *http.Request) ([]string, string) {
	var (
		names  []string
		values string
	)

	for param, values := range req.URL.Query() {
		// Keep only parameters that hold exactly 1 value (i.e. no empty or multi-valued parameters)
		if len(values) == 1 {
			names = append(names, param)
		}
	}
	sort.Strings(names)

	for _, param := range names {
		values += req.URL.Query().Get(param)
	}

	return names, values
}
