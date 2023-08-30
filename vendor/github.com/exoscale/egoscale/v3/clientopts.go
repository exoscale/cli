package v3

import (
	"fmt"
	"net/http"
	"os"

	"github.com/exoscale/egoscale/v3/oapi"
)

const (
	EnvKeyAPIKey    = "EXOSCALE_API_KEY"
	EnvKeyAPISecret = "EXOSCALE_API_SECRET"
)

// ClientConfig hold Client configuration options.
type ClientConfig struct {
	creds      *Credentials
	httpClient *http.Client
	logger     Logger

	requestEditors []oapi.RequestEditorFn
	// TODO: implement response editors (not available in oapi, should be embeded in consumer API.

	// User-Agent prefix
	uaPrefix string
}

// ClientOpt represents a function setting Exoscale API client option.
type ClientOpt func(*ClientConfig) error

// ClientOptWithCredentials returns a ClientOpt that sets credentials.
// If credentials are empty, error will be returned.
func ClientOptWithCredentials(key, secret string) ClientOpt {
	return func(c *ClientConfig) error {
		if key == "" || secret == "" {
			return fmt.Errorf("missing API credentials")
		}
		c.creds = NewCredentials(key, secret)

		return nil
	}
}

// ClientOptWithCredentialsFromEnv returns a ClientOpt that reads credentials from environment.
// Returns error of any value is missing in environment.
func ClientOptWithCredentialsFromEnv() ClientOpt {
	return func(c *ClientConfig) error {
		key := os.Getenv(EnvKeyAPIKey)
		secret := os.Getenv(EnvKeyAPISecret)
		if key == "" || secret == "" {
			return fmt.Errorf("API credentials not found in environment: %s %s", EnvKeyAPIKey, EnvKeyAPISecret)
		}

		c.creds = NewCredentials(key, secret)

		return nil
	}
}

// ClientOptWithHTTPClient returns a ClientOpt overriding the default http.Client.
// Default HTTP client is [go-retryablehttp] with static retry configuration.
// If you want to keep it your custom client should extend it.
//
// [go-retryablehttp]: https://github.com/hashicorp/go-retryablehttp
func ClientOptWithHTTPClient(v *http.Client) ClientOpt {
	return func(c *ClientConfig) error {
		c.httpClient = v

		return nil
	}
}

// ClientOptWithRequestEditor returns a ClientOpt that adds oapi.RequestEditorFn to oapi client.
// Editors run sequentialy and this function appends provided editor funtion to the end of the list.
func ClientOptWithRequestEditor(e oapi.RequestEditorFn) ClientOpt {
	return func(c *ClientConfig) error {
		c.requestEditors = append(c.requestEditors, e)

		return nil
	}
}

// ClientOptWithUserAgent returns a ClientOpt that sets User-Agent string to the provided value.
// Value provided is always suffixed with library version and host info.
func ClientOptWithUserAgent(ua string) ClientOpt {
	return func(c *ClientConfig) error {
		c.uaPrefix = ua

		return nil
	}
}

// ClientOptWithLogger returns ClientOpt that configures logging with provided Logger.
func ClientOptWithLogger(logger Logger) ClientOpt {
	return func(c *ClientConfig) error {
		c.logger = logger

		return nil
	}
}
