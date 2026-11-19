// Package v3 provides primitives to interact with the openapi HTTP API.
package v3

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"runtime"
	"slices"
	"strings"
	"time"

	"github.com/exoscale/egoscale/v3/credentials"
	"github.com/exoscale/egoscale/v3/utils"
	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/go-retryablehttp"
)

// Client represents an Exoscale API client.
type Client struct {
	// credentials are read for every request, so that new ones (see credentials.ReloadableProvider)
	// are used without creating a new client. Clones share them.
	credentials       *credentials.Credentials
	userAgent         string
	serverEndpoint    string
	defaultHTTPClient *retryablehttp.Client
	httpClient        *http.Client
	waitTimeout       time.Duration
	validate          *validator.Validate

	logger *slog.Logger

	// A list of callbacks for modifying requests which are generated before sending over
	// the network.
	requestInterceptors []RequestInterceptorFn
}

// NewClient returns a new Exoscale API client.
func NewClient(credentials *credentials.Credentials, opts ...ClientOpt) (*Client, error) {
	if _, err := credentials.Get(); err != nil {
		return nil, err
	}

	client := &Client{
		credentials:    credentials,
		serverEndpoint: string(CHGva2),
		validate:       validator.New(),
		userAgent:      getDefaultUserAgent(),
		logger:         slog.New(slog.DiscardHandler), // no logs by default
	}

	client.defaultHTTPClient = retryablehttp.NewClient()
	client.defaultHTTPClient.Logger = slog.New(slog.DiscardHandler)
	client.httpClient = client.defaultHTTPClient.StandardClient()

	for _, opt := range opts {
		if err := opt(client); err != nil {
			return nil, fmt.Errorf("client configuration error: %s", err)
		}
	}

	return client, nil
}

func (c Client) GetZoneName(ctx context.Context, endpoint Endpoint) (ZoneName, error) {
	zone, err := c.findZone(ctx, string(endpoint))
	if err != nil {
		return "", fmt.Errorf("get zone name: %w", err)
	}

	return zone.Name, nil
}

func (c Client) GetZoneAPIEndpoint(ctx context.Context, zoneName ZoneName) (Endpoint, error) {
	zone, err := c.findZone(ctx, string(zoneName))
	if err != nil {
		return "", fmt.Errorf("get zone api endpoint: %w", err)
	}

	return zone.APIEndpoint, nil
}

// findZone looks a zone up by name or API endpoint.
// ListZones is sent unauthenticated (signing it makes IAM reject restricted keys),
// so zones only enabled for some organizations are missing from its response:
// on a miss, retry once with a signed request.
func (c Client) findZone(ctx context.Context, nameOrAPIEndpoint string) (Zone, error) {
	resp, err := c.ListZones(ctx)
	if err != nil {
		return Zone{}, fmt.Errorf("list zones: %w", err)
	}

	zone, err := resp.FindZone(nameOrAPIEndpoint)
	if err == nil {
		return zone, nil
	}
	if !errors.Is(err, ErrNotFound) {
		return Zone{}, fmt.Errorf("find zone: %w", err)
	}

	signed := cloneClient(&c)

	signed.requestInterceptors = append(slices.Clip(c.requestInterceptors), func(_ context.Context, req *http.Request) error {
		return c.signRequest(req)
	})
	signedResp, signedErr := signed.ListZones(ctx)
	if signedErr != nil {
		return Zone{}, fmt.Errorf("find zone: %w", err)
	}

	zone, err = signedResp.FindZone(nameOrAPIEndpoint)
	if err != nil {
		return Zone{}, fmt.Errorf("find zone: %w", err)
	}

	return zone, nil
}

// RequestInterceptorFn is the function signature for the RequestInterceptor callback function
type RequestInterceptorFn func(ctx context.Context, req *http.Request) error

// Deprecated: use ClientOptWithUserAgent instead.
var UserAgent = getDefaultUserAgent()

// ClientOpt represents a function setting Exoscale API client option.
type ClientOpt func(*Client) error

// Deprecated: use ClientOptWithLogHandler instead.
func ClientOptWithTrace() ClientOpt {
	return func(c *Client) error {
		c.logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
		if c.defaultHTTPClient != nil {
			c.defaultHTTPClient.Logger = c.logger.With("subsystem", "retryablehttp")
		}

		return nil
	}
}

// ClientOptWithLogHandler sets provided log handler.
func ClientOptWithLogHandler(h slog.Handler) ClientOpt {
	return func(c *Client) error {
		c.logger = slog.New(h)
		if c.defaultHTTPClient != nil {
			c.defaultHTTPClient.Logger = c.logger.With("subsystem", "retryablehttp")
		}
		return nil
	}
}

// ClientOptWithUserAgent returns a ClientOpt setting the user agent header.
func ClientOptWithUserAgent(ua string) ClientOpt {
	return func(c *Client) error {
		c.userAgent = ua + " " + getDefaultUserAgent()
		return nil
	}
}

// ClientOptWithValidator returns a ClientOpt with a given validator.
func ClientOptWithValidator(validate *validator.Validate) ClientOpt {
	return func(c *Client) error {
		c.validate = validate
		return nil
	}
}

// ClientOptWithEndpoint returns a ClientOpt With a given zone Endpoint.
func ClientOptWithEndpoint(endpoint Endpoint) ClientOpt {
	return func(c *Client) error {
		c.serverEndpoint = string(endpoint)
		return nil
	}
}

// ClientOptWithWaitTimeout returns a ClientOpt With a given wait timeout.
func ClientOptWithWaitTimeout(t time.Duration) ClientOpt {
	return func(c *Client) error {
		c.waitTimeout = t
		return nil
	}
}

// ClientOptWithRequestInterceptors returns a ClientOpt With given RequestInterceptors.
func ClientOptWithRequestInterceptors(f ...RequestInterceptorFn) ClientOpt {
	return func(c *Client) error {
		c.requestInterceptors = append(c.requestInterceptors, f...)
		return nil
	}
}

// ClientOptWithHTTPClient returns a ClientOpt overriding the default http.Client.
// Note: the Exoscale API client will chain additional middleware
// (http.RoundTripper) on the HTTP client internally, which can alter the HTTP
// requests and responses. If you don't want any other middleware than the ones
// currently set to your HTTP client, you should duplicate it and pass a copy
// instead.
func ClientOptWithHTTPClient(v *http.Client) ClientOpt {
	return func(c *Client) error {
		c.defaultHTTPClient = nil
		c.httpClient = v

		return nil
	}
}

// getDefaultUserAgent returns the "User-Agent" HTTP request header added to outgoing HTTP requests.
func getDefaultUserAgent() string {
	return fmt.Sprintf("egoscale/%s (%s; %s/%s)",
		Version,
		runtime.Version(),
		runtime.GOOS,
		runtime.GOARCH)
}

// getUserAgent only for compatibility with UserAgent.
func (c *Client) getUserAgent() string {
	defaultUA := getDefaultUserAgent()

	if c.userAgent != defaultUA {
		return c.userAgent
	}

	if UserAgent != defaultUA {
		return UserAgent
	}

	return c.userAgent
}

// WithEndpoint returns a copy of Client with new zone Endpoint.
func (c *Client) WithEndpoint(endpoint Endpoint) *Client {
	clone := cloneClient(c)

	clone.serverEndpoint = string(endpoint)

	return clone
}

// WithWaitTimeout returns a copy of Client with new wait timeout.
func (c *Client) WithWaitTimeout(t time.Duration) *Client {
	clone := cloneClient(c)

	clone.waitTimeout = t

	return clone
}

// WithUserAgent returns a copy of Client with new User-Agent.
func (c *Client) WithUserAgent(ua string) *Client {
	clone := cloneClient(c)

	clone.userAgent = ua + " " + getDefaultUserAgent()

	return clone
}

// Deprecated: use WithLogHandler instead.
func (c *Client) WithTrace() *Client {
	clone := cloneClient(c)

	clone.logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
	if clone.defaultHTTPClient != nil {
		clone.defaultHTTPClient.Logger = clone.logger.With("subsystem", "retryablehttp")
	}

	return clone
}

// WithLogHandler returns a copy of Client with new log handler set.
func (c *Client) WithLogHandler(h slog.Handler) *Client {
	clone := cloneClient(c)

	clone.logger = slog.New(h)
	if clone.defaultHTTPClient != nil {
		clone.defaultHTTPClient.Logger = clone.logger.With("subsystem", "retryablehttp")
	}

	return clone
}

// WithHttpClient returns a copy of Client with new http.Client.
// Deprecated: use WithHTTPClient instead.
func (c *Client) WithHttpClient(client *http.Client) *Client {
	clone := cloneClient(c)

	clone.httpClient = client

	return clone
}

// WithHTTPClient returns a copy of Client with new http.Client.
func (c *Client) WithHTTPClient(client *http.Client) *Client {
	clone := cloneClient(c)

	clone.httpClient = client

	return clone
}

// WithRequestInterceptor returns a copy of Client with new RequestInterceptors.
func (c *Client) WithRequestInterceptor(f ...RequestInterceptorFn) *Client {
	clone := cloneClient(c)

	clone.requestInterceptors = append(clone.requestInterceptors, f...)

	return clone
}

func (c *Client) executeRequestInterceptors(ctx context.Context, req *http.Request) error {
	for _, fn := range c.requestInterceptors {
		if err := fn(ctx, req); err != nil {
			return err
		}
	}

	return nil
}

// do sends request. If the API rejects the API key the request was signed with, it reloads the
// credentials and, if they changed, signs the request again and retries it once.
func (c *Client) do(request *http.Request) (*http.Response, error) {
	response, err := c.send(request)
	if err != nil || !isInvalidKeyResponse(response) {
		return response, err
	}

	retry, ok := c.retryWithNewCredentials(request)
	if !ok {
		return response, nil
	}

	_ = response.Body.Close()

	c.logger.DebugContext(request.Context(), "API key rejected, retrying with the reloaded credentials")

	return c.send(retry)
}

// retryWithNewCredentials reloads the credentials of a request signed by c and rejected by the
// API, and returns a copy of the request signed with the new ones. It returns false if the
// credentials didn't change, or the request can't be sent again.
func (c *Client) retryWithNewCredentials(request *http.Request) (*http.Request, bool) {
	signedKey, ok := signedAPIKey(request)
	if !ok || c.credentials == nil || (request.Body != nil && request.GetBody == nil) {
		return nil, false
	}

	ctx := request.Context()
	if err := c.credentials.Reload(ctx); err != nil {
		c.logger.DebugContext(ctx, "unable to reload the credentials", "error", err)
	}

	values, err := c.credentials.Get()
	if err != nil || values.APIKey == signedKey {
		return nil, false
	}

	retry := request.Clone(ctx)
	if request.GetBody != nil {
		body, err := request.GetBody()
		if err != nil {
			return nil, false
		}
		retry.Body = body
	}

	if err := SignHTTPRequest(retry, values.APIKey, values.APISecret, signatureExpiration()); err != nil {
		return nil, false
	}

	return retry, true
}

// isInvalidKeyResponse reports whether the API rejected the API key of the request
// (unknown or deleted key, or wrong secret).
// It keeps the response body readable.
func isInvalidKeyResponse(response *http.Response) bool {
	switch response.StatusCode {
	case http.StatusUnauthorized:
		return true
	case http.StatusForbidden:
	default:
		return false
	}

	body, err := io.ReadAll(response.Body)
	_ = response.Body.Close()
	response.Body = io.NopCloser(bytes.NewReader(body))
	if err != nil {
		return false
	}

	var apiErr struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &apiErr); err != nil {
		return false
	}

	return strings.Contains(strings.ToLower(apiErr.Message), "invalid key")
}

// signedAPIKey returns the API key of a request signed by SignHTTPRequest.
func signedAPIKey(request *http.Request) (string, bool) {
	auth, ok := strings.CutPrefix(request.Header.Get("Authorization"), signatureScheme+" credential=")
	if !ok {
		return "", false
	}
	key, _, _ := strings.Cut(auth, ",")

	return key, key != ""
}

// send is a wrapper for c.httpClient.Do to apply middleware.
func (c *Client) send(request *http.Request) (*http.Response, error) {
	ctx := request.Context()
	if request == nil || !c.logger.Enabled(ctx, slog.LevelDebug) {
		return c.httpClient.Do(request)
	}

	var (
		meta []any
	)

	requestMeta := []any{
		"method", request.Method,
		"host", request.URL.Host,
		"path", request.URL.EscapedPath(),
		"headers", utils.PrintHeaders(request.Header),
	}

	if request.ContentLength != 0 &&
		request.Body != nil &&
		utils.IsLoggableContentType(request.Header) {
		requestBody, err := io.ReadAll(request.Body)
		if err != nil {
			return nil, err
		}
		if err = request.Body.Close(); err != nil {
			return nil, err
		}
		request.Body = io.NopCloser(bytes.NewReader(requestBody))

		requestMeta = append(requestMeta, "json_body", string(requestBody))
	}

	meta = append(meta, slog.Group("request", requestMeta...))

	response, err := c.httpClient.Do(request)
	if err != nil {
		return response, err
	}

	responseMeta := []any{
		"status", response.Status,
		"headers", utils.PrintHeaders(response.Header),
	}

	if response.ContentLength != 0 &&
		response.Body != nil &&
		request.URL.Query().Get("streaming") != "true" &&
		utils.IsLoggableContentType(response.Header) {
		responseBody, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}
		if err = response.Body.Close(); err != nil {
			return nil, err
		}
		response.Body = io.NopCloser(bytes.NewReader(responseBody))

		responseMeta = append(responseMeta, "json_body", string(responseBody))
	}

	meta = append(meta, slog.Group("response", responseMeta...))

	c.logger.DebugContext(
		ctx,
		"api call trace",
		meta...,
	)

	return response, err
}

func cloneClient(c *Client) *Client {
	newc := Client{
		credentials:         c.credentials,
		userAgent:           c.userAgent,
		serverEndpoint:      c.serverEndpoint,
		defaultHTTPClient:   c.defaultHTTPClient,
		httpClient:          c.httpClient,
		requestInterceptors: c.requestInterceptors,
		waitTimeout:         c.waitTimeout,
		validate:            c.validate,
		logger:              c.logger,
	}

	if c.defaultHTTPClient != nil {
		rc := retryablehttp.NewClient()
		rc.Logger = c.defaultHTTPClient.Logger
		newc.defaultHTTPClient = rc
	}

	return &newc
}
