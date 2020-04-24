package egoscale

import (
	"context"
	"fmt"
)

// Ping makes the client "ping" the Exoscale API, and returns an error if the API is not reachable, otherwise nil.
// Note: this method doesn't validate client credentials, only network connectivity and API availability.
func (c *Client) Ping(ctx context.Context) error {
	resp, err := c.v2.Ping(ctx)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("unexpected API response: %s", resp.Status)
	}

	return nil
}
