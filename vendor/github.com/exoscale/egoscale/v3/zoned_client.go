package v3

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/exoscale/egoscale/v3/oapi"
)

const (
	DefaultHostPattern = "https://api-%s.exoscale.com/v2"

	EnvKeyAPIEndpointPattern = "EXOSCALE_API_ENDPOINT_PATTERN"
	EnvKeyAPIEndpointZones   = "EXOSCALE_API_ENDPOINT_ZONES"
)

var (
	// DefaultZones list (available in oapi code).
	// When new zone is added or existing removed this slice needs to be updated.
	// First zone in the slice is used as default in DefaultZonedClient.
	DefaultZones = []oapi.ZoneName{
		oapi.ChGva2,
		oapi.AtVie1,
		oapi.AtVie2,
		oapi.BgSof1,
		oapi.ChDk2,
		oapi.DeFra1,
		oapi.DeMuc1,
	}
)

// ZonedClient is an Exoscale API Client that can communicate with API servers in different zones.
// It has the same interface as Client and uses currently selected zone to run API calls.
// Consumer is expected to select zone before invoking API calls.
type ZonedClient struct {
	zones       map[oapi.ZoneName]*oapi.ClientWithResponses
	currentZone oapi.ZoneName
	mx          sync.RWMutex

	Client
}

// NewZonedClient creates a new ZonedClient using URL pattern, list of zones and Client options.
// URL pattern must be a valid URL with exactly one substitution verb '%s', for example:
//
//	https://api-%s.exoscale.com/v2
//
// ClientOpt options will be passed down to Client as provided.
// If EXOSCALE_API_ENDPOINT_PATTERN environment variable is set, it replaces urlPattern.
// If EXOSCALE_API_ENDPOINT_ZONES environment variable is set (CSV format), it replaces zones.
func NewZonedClient(urlPattern string, zones []oapi.ZoneName, opts ...ClientOpt) (*ZonedClient, error) {
	if len(zones) == 0 {
		return nil, errors.New("list of zones cannot be empty")
	}

	// Env overrides
	if h := os.Getenv(EnvKeyAPIEndpointPattern); h != "" {
		urlPattern = h
	}
	if z := os.Getenv(EnvKeyAPIEndpointZones); z != "" {
		zones = []oapi.ZoneName{}
		parts := strings.Split(z, ",")
		for _, part := range parts {
			zones = append(zones, oapi.ZoneName(part))
		}
	}

	zonedClient := ZonedClient{
		zones: map[oapi.ZoneName]*oapi.ClientWithResponses{},
	}

	for _, zone := range zones {
		client, err := NewClient(fmt.Sprintf(urlPattern, zone), opts...)
		if err != nil {
			return nil, err
		}

		if zonedClient.creds == nil {
			zonedClient.creds = client.creds
		}

		zonedClient.zones[zone] = client.oapiClient
	}

	// Set default zone to first zone in the provided slice
	zonedClient.currentZone = zones[0]
	zonedClient.oapiClient = zonedClient.zones[zones[0]]

	return &zonedClient, nil
}

// DefaultClient creates a ZonedClient with preset API URL pattern and zone and provided options.
// This is what should be used by default.
func DefaultClient(opts ...ClientOpt) (*ZonedClient, error) {
	return NewZonedClient(DefaultHostPattern, DefaultZones, opts...)
}

// Zone returns the current zone identifier.
func (c *ZonedClient) Zone() oapi.ZoneName {
	c.mx.RLock()
	defer c.mx.RUnlock()

	return c.currentZone
}

// SetZone selects the current zone.
func (c *ZonedClient) SetZone(z oapi.ZoneName) {
	c.mx.Lock()
	c.oapiClient = c.zones[z]
	c.mx.Unlock()
}

// InZone selects returns the instance of the Client in selected zone so the methods may be chained:
//
//	zonedClient.InZone(oapi.ChGva2).OAPIClient()...
func (c *ZonedClient) InZone(z oapi.ZoneName) *Client {
	return &Client{
		creds:      c.creds,
		oapiClient: c.zones[z],
	}
}

// OAPIClient returns configured instance of OpenAPI generated (low-level) API client in the selected zone.
func (c *ZonedClient) OAPIClient() *oapi.ClientWithResponses {
	c.mx.RLock()
	defer c.mx.RUnlock()

	return c.Client.OAPIClient()
}

// ForEachZone runs function f in each configured zone.
// Argument of function f is configured Client for the zone.
func (c *ZonedClient) ForEachZone(f func(c *Client, zone oapi.ZoneName)) {
	for zone, oapiClient := range c.zones {
		f(&Client{creds: c.creds, oapiClient: oapiClient}, zone)
	}
}
