package v2

import (
	"context"
	"net"

	apiv2 "github.com/exoscale/egoscale/v2/api"
	papi "github.com/exoscale/egoscale/v2/internal/public-api"
)

// PrivateNetwork represents a Private Network.
type PrivateNetwork struct {
	Description string
	EndIP       net.IP
	ID          string
	Name        string
	Netmask     net.IP
	StartIP     net.IP
}

func privateNetworkFromAPI(p *papi.PrivateNetwork) *PrivateNetwork {
	return &PrivateNetwork{
		Description: papi.OptionalString(p.Description),
		EndIP:       net.ParseIP(papi.OptionalString(p.EndIp)),
		ID:          *p.Id,
		Name:        *p.Name,
		Netmask:     net.ParseIP(papi.OptionalString(p.Netmask)),
		StartIP:     net.ParseIP(papi.OptionalString(p.StartIp)),
	}
}

func (p PrivateNetwork) get(ctx context.Context, client *Client, zone, id string) (interface{}, error) {
	return client.GetPrivateNetwork(ctx, zone, id)
}

// CreatePrivateNetwork creates a Private Network in the specified zone.
func (c *Client) CreatePrivateNetwork(
	ctx context.Context,
	zone string,
	privateNetwork *PrivateNetwork,
) (*PrivateNetwork, error) {
	resp, err := c.CreatePrivateNetworkWithResponse(
		apiv2.WithZone(ctx, zone),
		papi.CreatePrivateNetworkJSONRequestBody{
			Description: func() *string {
				if privateNetwork.Description != "" {
					return &privateNetwork.Description
				}
				return nil
			}(),
			EndIp: func() (ip *string) {
				if privateNetwork.EndIP != nil {
					v := privateNetwork.EndIP.String()
					return &v
				}
				return
			}(),
			Name: privateNetwork.Name,
			Netmask: func() (ip *string) {
				if privateNetwork.Netmask != nil {
					v := privateNetwork.Netmask.String()
					return &v
				}
				return
			}(),
			StartIp: func() (ip *string) {
				if privateNetwork.StartIP != nil {
					v := privateNetwork.StartIP.String()
					return &v
				}
				return
			}(),
		})
	if err != nil {
		return nil, err
	}

	res, err := papi.NewPoller().
		WithTimeout(c.timeout).
		WithInterval(c.pollInterval).
		Poll(ctx, c.OperationPoller(zone, *resp.JSON200.Id))
	if err != nil {
		return nil, err
	}

	return c.GetPrivateNetwork(ctx, zone, *res.(*papi.Reference).Id)
}

// ListPrivateNetworks returns the list of existing Private Networks in the specified zone.
func (c *Client) ListPrivateNetworks(ctx context.Context, zone string) ([]*PrivateNetwork, error) {
	list := make([]*PrivateNetwork, 0)

	resp, err := c.ListPrivateNetworksWithResponse(apiv2.WithZone(ctx, zone))
	if err != nil {
		return nil, err
	}

	if resp.JSON200.PrivateNetworks != nil {
		for i := range *resp.JSON200.PrivateNetworks {
			list = append(list, privateNetworkFromAPI(&(*resp.JSON200.PrivateNetworks)[i]))
		}
	}

	return list, nil
}

// GetPrivateNetwork returns the Private Network corresponding to the specified ID in the specified zone.
func (c *Client) GetPrivateNetwork(ctx context.Context, zone, id string) (*PrivateNetwork, error) {
	resp, err := c.GetPrivateNetworkWithResponse(apiv2.WithZone(ctx, zone), id)
	if err != nil {
		return nil, err
	}

	return privateNetworkFromAPI(resp.JSON200), nil
}

// FindPrivateNetwork attempts to find a Private Network by name or ID in the specified zone.
// In case the identifier is a name and multiple resources match, an ErrTooManyFound error is returned.
func (c *Client) FindPrivateNetwork(ctx context.Context, zone, v string) (*PrivateNetwork, error) {
	res, err := c.ListPrivateNetworks(ctx, zone)
	if err != nil {
		return nil, err
	}

	var found *PrivateNetwork
	for _, r := range res {
		if r.ID == v {
			return c.GetPrivateNetwork(ctx, zone, r.ID)
		}

		// Historically, the Exoscale API allowed users to create multiple Private Networks sharing a common name.
		// This function being expected to return one resource at most, in case the specified identifier is a name
		// we have to check that there aren't more that one matching result before returning it.
		if r.Name == v {
			if found != nil {
				return nil, apiv2.ErrTooManyFound
			}
			found = r
		}
	}

	if found != nil {
		return found, nil
	}

	return nil, apiv2.ErrNotFound
}

// UpdatePrivateNetwork updates the specified Private Network in the specified zone.
func (c *Client) UpdatePrivateNetwork(ctx context.Context, zone string, privateNetwork *PrivateNetwork) error {
	resp, err := c.UpdatePrivateNetworkWithResponse(
		apiv2.WithZone(ctx, zone),
		privateNetwork.ID,
		papi.UpdatePrivateNetworkJSONRequestBody{
			Description: func() *string {
				if privateNetwork.Description != "" {
					return &privateNetwork.Description
				}
				return nil
			}(),
			EndIp: func() (ip *string) {
				if privateNetwork.EndIP != nil {
					v := privateNetwork.EndIP.String()
					return &v
				}
				return
			}(),
			Name: &privateNetwork.Name,
			Netmask: func() (ip *string) {
				if privateNetwork.Netmask != nil {
					v := privateNetwork.Netmask.String()
					return &v
				}
				return
			}(),
			StartIp: func() (ip *string) {
				if privateNetwork.StartIP != nil {
					v := privateNetwork.StartIP.String()
					return &v
				}
				return
			}(),
		})
	if err != nil {
		return err
	}

	_, err = papi.NewPoller().
		WithTimeout(c.timeout).
		WithInterval(c.pollInterval).
		Poll(ctx, c.OperationPoller(zone, *resp.JSON200.Id))
	if err != nil {
		return err
	}

	return nil
}

// DeletePrivateNetwork deletes the specified Private Network in the specified zone.
func (c *Client) DeletePrivateNetwork(ctx context.Context, zone, id string) error {
	resp, err := c.DeletePrivateNetworkWithResponse(apiv2.WithZone(ctx, zone), id)
	if err != nil {
		return err
	}

	_, err = papi.NewPoller().
		WithTimeout(c.timeout).
		WithInterval(c.pollInterval).
		Poll(ctx, c.OperationPoller(zone, *resp.JSON200.Id))
	if err != nil {
		return err
	}

	return nil
}
