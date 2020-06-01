package egoscale

import (
	"context"
	"fmt"
	"net/http"
)

// Ping makes the client "ping" the Exoscale API, and returns an error if the API is not reachable, otherwise nil.
// Note: this method doesn't validate client credentials, only network connectivity and API availability.
func (c *Client) Ping(ctx context.Context) error {
	resp, err := c.v2.Ping(ctx)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected API response: %s", resp.Status)
	}

	return nil
}

// optionalString returns the dereferenced string value of v if not nil, otherwise an empty string.
func optionalString(v *string) string {
	if v != nil {
		return *v
	}

	return ""
}

// optionalInt64 returns the dereferenced int64 value of v if not nil, otherwise 0.
func optionalInt64(v *int64) int64 {
	if v != nil {
		return *v
	}

	return 0
}
