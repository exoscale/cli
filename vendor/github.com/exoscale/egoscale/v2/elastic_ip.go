package v2

import (
	"context"
	"net"
	"time"

	apiv2 "github.com/exoscale/egoscale/v2/api"
	papi "github.com/exoscale/egoscale/v2/internal/public-api"
)

// ElasticIPHealthcheck represents an Elastic IP healthcheck.
type ElasticIPHealthcheck struct {
	Interval      time.Duration
	Mode          string
	Port          uint16
	StrikesFail   int64
	StrikesOK     int64
	TLSSNI        string
	TLSSkipVerify bool
	Timeout       time.Duration
	URI           string
}

// ElasticIP represents an Elastic IP.
type ElasticIP struct {
	Description string `reset:"description"`
	Healthcheck *ElasticIPHealthcheck
	ID          string
	IPAddress   net.IP

	c    *Client
	zone string
}

func elasticIPFromAPI(client *Client, zone string, e *papi.ElasticIp) *ElasticIP {
	return &ElasticIP{
		Description: papi.OptionalString(e.Description),
		Healthcheck: func() *ElasticIPHealthcheck {
			if hc := e.Healthcheck; hc != nil {
				return &ElasticIPHealthcheck{
					Interval:      time.Duration(papi.OptionalInt64(hc.Interval)) * time.Second,
					Mode:          hc.Mode,
					Port:          uint16(hc.Port),
					StrikesFail:   papi.OptionalInt64(hc.StrikesFail),
					StrikesOK:     papi.OptionalInt64(hc.StrikesOk),
					TLSSNI:        papi.OptionalString(hc.TlsSni),
					TLSSkipVerify: papi.OptionalBool(hc.TlsSkipVerify),
					Timeout:       time.Duration(papi.OptionalInt64(hc.Timeout)) * time.Second,
					URI:           papi.OptionalString(hc.Uri),
				}
			}
			return nil
		}(),
		ID:        papi.OptionalString(e.Id),
		IPAddress: net.ParseIP(papi.OptionalString(e.Ip)),

		c:    client,
		zone: zone,
	}
}

func (e ElasticIP) get(ctx context.Context, client *Client, zone, id string) (interface{}, error) {
	return client.GetElasticIP(ctx, zone, id)
}

// ResetField resets the specified Elastic IP field to its default value.
// The value expected for the field parameter is a pointer to the ElasticIP field to reset.
func (e *ElasticIP) ResetField(ctx context.Context, field interface{}) error {
	resetField, err := resetFieldName(e, field)
	if err != nil {
		return err
	}

	resp, err := e.c.ResetElasticIpFieldWithResponse(apiv2.WithZone(ctx, e.zone), e.ID, resetField)
	if err != nil {
		return err
	}

	_, err = papi.NewPoller().
		WithTimeout(e.c.timeout).
		Poll(ctx, e.c.OperationPoller(e.zone, *resp.JSON200.Id))
	if err != nil {
		return err
	}

	return nil
}

// CreateElasticIP creates an Elastic IP in the specified zone.
func (c *Client) CreateElasticIP(ctx context.Context, zone string, elasticIP *ElasticIP) (*ElasticIP, error) {
	resp, err := c.CreateElasticIpWithResponse(
		apiv2.WithZone(ctx, zone),
		papi.CreateElasticIpJSONRequestBody{
			Description: &elasticIP.Description,
			Healthcheck: func() *papi.ElasticIpHealthcheck {
				if hc := elasticIP.Healthcheck; hc != nil {
					var (
						port     = int64(hc.Port)
						interval = int64(hc.Interval.Seconds())
						timeout  = int64(hc.Timeout.Seconds())
					)

					return &papi.ElasticIpHealthcheck{
						Interval:      &interval,
						Mode:          hc.Mode,
						Port:          port,
						StrikesFail:   &hc.StrikesFail,
						StrikesOk:     &hc.StrikesOK,
						Timeout:       &timeout,
						TlsSkipVerify: &hc.TLSSkipVerify,
						TlsSni:        &hc.TLSSNI,
						Uri:           &hc.URI,
					}
				}
				return nil
			}(),
		})
	if err != nil {
		return nil, err
	}

	res, err := papi.NewPoller().
		WithTimeout(c.timeout).
		Poll(ctx, c.OperationPoller(zone, *resp.JSON200.Id))
	if err != nil {
		return nil, err
	}

	return c.GetElasticIP(ctx, zone, *res.(*papi.Reference).Id)
}

// ListElasticIPs returns the list of existing Elastic IPs in the specified zone.
func (c *Client) ListElasticIPs(ctx context.Context, zone string) ([]*ElasticIP, error) {
	list := make([]*ElasticIP, 0)

	resp, err := c.ListElasticIpsWithResponse(apiv2.WithZone(ctx, zone))
	if err != nil {
		return nil, err
	}

	if resp.JSON200.ElasticIps != nil {
		for i := range *resp.JSON200.ElasticIps {
			list = append(list, elasticIPFromAPI(c, zone, &(*resp.JSON200.ElasticIps)[i]))
		}
	}

	return list, nil
}

// GetElasticIP returns the Elastic IP corresponding to the specified ID in the specified zone.
func (c *Client) GetElasticIP(ctx context.Context, zone, id string) (*ElasticIP, error) {
	resp, err := c.GetElasticIpWithResponse(apiv2.WithZone(ctx, zone), id)
	if err != nil {
		return nil, err
	}

	return elasticIPFromAPI(c, zone, resp.JSON200), nil
}

// UpdateElasticIP updates the specified Elastic IP in the specified zone.
func (c *Client) UpdateElasticIP(ctx context.Context, zone string, elasticIP *ElasticIP) error {
	resp, err := c.UpdateElasticIpWithResponse(
		apiv2.WithZone(ctx, zone),
		elasticIP.ID,
		papi.UpdateElasticIpJSONRequestBody{
			Description: func() *string {
				if elasticIP.Description != "" {
					return &elasticIP.Description
				}
				return nil
			}(),
			Healthcheck: func() *papi.ElasticIpHealthcheck {
				if hc := elasticIP.Healthcheck; hc != nil {
					var (
						port     = int64(hc.Port)
						interval = int64(hc.Interval.Seconds())
						timeout  = int64(hc.Timeout.Seconds())
					)

					return &papi.ElasticIpHealthcheck{
						Interval:      &interval,
						Mode:          hc.Mode,
						Port:          port,
						StrikesFail:   &hc.StrikesFail,
						StrikesOk:     &hc.StrikesOK,
						Timeout:       &timeout,
						TlsSkipVerify: &hc.TLSSkipVerify,
						TlsSni:        &hc.TLSSNI,
						Uri:           &hc.URI,
					}
				}
				return nil
			}(),
		})
	if err != nil {
		return err
	}

	_, err = papi.NewPoller().
		WithTimeout(c.timeout).
		Poll(ctx, c.OperationPoller(zone, *resp.JSON200.Id))
	if err != nil {
		return err
	}

	return nil
}

// DeleteElasticIP deletes the specified Elastic IP in the specified zone.
func (c *Client) DeleteElasticIP(ctx context.Context, zone, id string) error {
	resp, err := c.DeleteElasticIpWithResponse(apiv2.WithZone(ctx, zone), id)
	if err != nil {
		return err
	}

	_, err = papi.NewPoller().
		WithTimeout(c.timeout).
		Poll(ctx, c.OperationPoller(zone, *resp.JSON200.Id))
	if err != nil {
		return err
	}

	return nil
}
