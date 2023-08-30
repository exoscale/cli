package v3

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/hashicorp/go-retryablehttp"

	"github.com/exoscale/egoscale/v3/api/compute"
	"github.com/exoscale/egoscale/v3/api/dbaas"
	"github.com/exoscale/egoscale/v3/api/dns"
	"github.com/exoscale/egoscale/v3/api/global"
	"github.com/exoscale/egoscale/v3/api/iam"
	"github.com/exoscale/egoscale/v3/oapi"
)

const (
	EnvKeyAPIEndpoint = "EXOSCALE_API_ENDPOINT"

	PollingInterval = 3 * time.Second
)

// Client represents Exoscale V3 API Client.
type Client struct {
	creds      *Credentials
	oapiClient *oapi.ClientWithResponses
}

// NewClient returns a new Exoscale API V3 client, or an error if one couldn't be initialized.
// Client is generic (single EP) with no concept of zones/environments.
// For zone-aware client use ZonedClient.
// Default HTTP client is [go-retryablehttp] with static retry configuration.
// To change retry configuration, build new HTTP client and pass it using ClientOptWithHTTPClient.
// API credentials must be passed with ClientOptWithCredentials.
// If EXOSCALE_API_ENDPOINT environment variable is set, it replaces endpoint.
func NewClient(endpoint string, opts ...ClientOpt) (*Client, error) {
	// Env var override
	if h := os.Getenv(EnvKeyAPIEndpoint); h != "" {
		endpoint = h
	}

	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	config := ClientConfig{
		requestEditors: []oapi.RequestEditorFn{},
	}
	for _, opt := range opts {
		if err := opt(&config); err != nil {
			return nil, fmt.Errorf("client configuration error: %w", err)
		}
	}

	client := Client{
		creds: config.creds,
	}

	// Use retryablehttp client by default
	if config.httpClient == nil {
		rc := retryablehttp.NewClient()
		rc.Logger = log.New(io.Discard, "", 0)
		if config.logger != nil {
			rc.Logger = config.logger
		}
		config.httpClient = rc.StandardClient()
	}

	// Mandatory oapi options.
	oapiOpts := []oapi.ClientOption{
		oapi.WithHTTPClient(config.httpClient),
		oapi.WithRequestEditorFn(NewUserAgentProvider(config.uaPrefix).Intercept),
	}

	// We are adding security middleware only if API credentials are specified
	// in order to allow generic usage and local testing.
	// TODO: add log line emphasizing the lack of credentials.
	if client.creds != nil {
		oapiOpts = append(
			oapiOpts,
			oapi.WithRequestEditorFn(NewSecurityProvider(client.creds).Intercept),
		)
	}

	// Attach any custom request editors
	for _, editor := range config.requestEditors {
		oapiOpts = append(
			oapiOpts,
			oapi.WithRequestEditorFn(editor),
		)
	}

	client.oapiClient, err = oapi.NewClientWithResponses(
		u.String(),
		oapiOpts...,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize API client: %w", err)
	}

	return &client, nil
}

// Wait is a helper that waits for async operation to reach the final state.
// Final states are one of: failure, success, timeout.
func (c *Client) Wait(
	ctx context.Context,
	f func(ctx context.Context) (*oapi.Operation, error),
) (*oapi.Operation, error) {
	ticker := time.NewTicker(PollingInterval)
	defer ticker.Stop()

	op, err := f(ctx)
	if err != nil {
		return nil, err
	}
	// Exit right away if operation is already done.
	if *op.State != oapi.OperationStatePending {
		return op, nil
	}

	for {
		select {
		case <-ticker.C:
			op, err := c.Global().Operations().Get(ctx, *op.Id)
			if err != nil {
				return nil, err
			}
			if *op.State != oapi.OperationStatePending {
				continue
			}

			return op, nil
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

// OAPIClient returns configured instance of OpenAPI generated (low-level) API client.
func (c *Client) OAPIClient() *oapi.ClientWithResponses {
	return c.oapiClient
}

// IAM provides access to IAM resources on Exoscale platform.
func (c *Client) IAM() *iam.IAM {
	return iam.NewIAM(c.oapiClient)
}

// DBaaS provides access to DBaaS resources on Exoscale platform.
func (c *Client) DBaaS() *dbaas.DBaaS {
	return dbaas.NewDBaaS(c.oapiClient)
}

// Compute provides access to Compute resources on Exoscale platform.
func (c *Client) Compute() *compute.Compute {
	return compute.NewCompute(c.oapiClient)
}

// DNS provides access to DNS resources on Exoscale platform.
func (c *Client) DNS() *dns.DNS {
	return dns.NewDNS(c.oapiClient)
}

// Global provides access to global resources on Exoscale platform.
func (c *Client) Global() *global.Global {
	return global.NewGlobal(c.oapiClient)
}
