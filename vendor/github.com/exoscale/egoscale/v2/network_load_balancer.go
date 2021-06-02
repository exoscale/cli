package v2

import (
	"context"
	"errors"
	"net"
	"strings"
	"time"

	apiv2 "github.com/exoscale/egoscale/v2/api"
	papi "github.com/exoscale/egoscale/v2/internal/public-api"
)

// NetworkLoadBalancerServerStatus represents a Network Load Balancer service target server status.
type NetworkLoadBalancerServerStatus struct {
	InstanceIP net.IP
	Status     string
}

func nlbServerStatusFromAPI(st *papi.LoadBalancerServerStatus) *NetworkLoadBalancerServerStatus {
	return &NetworkLoadBalancerServerStatus{
		InstanceIP: net.ParseIP(papi.OptionalString(st.PublicIp)),
		Status:     string(*st.Status),
	}
}

// NetworkLoadBalancerServiceHealthcheck represents a Network Load Balancer service healthcheck.
type NetworkLoadBalancerServiceHealthcheck struct {
	Interval time.Duration
	Mode     string
	Port     uint16
	Retries  int64
	TLSSNI   string
	Timeout  time.Duration
	URI      string
}

// NetworkLoadBalancerService represents a Network Load Balancer service.
type NetworkLoadBalancerService struct {
	Description       string
	Healthcheck       NetworkLoadBalancerServiceHealthcheck
	HealthcheckStatus []*NetworkLoadBalancerServerStatus
	ID                string
	InstancePoolID    string
	Name              string
	Port              uint16
	Protocol          string
	State             string
	Strategy          string
	TargetPort        uint16
}

func nlbServiceFromAPI(svc *papi.LoadBalancerService) *NetworkLoadBalancerService {
	return &NetworkLoadBalancerService{
		Description: papi.OptionalString(svc.Description),
		Healthcheck: NetworkLoadBalancerServiceHealthcheck{
			Interval: time.Duration(papi.OptionalInt64(svc.Healthcheck.Interval)) * time.Second,
			Mode:     string(svc.Healthcheck.Mode),
			Port:     uint16(svc.Healthcheck.Port),
			Retries:  papi.OptionalInt64(svc.Healthcheck.Retries),
			TLSSNI:   papi.OptionalString(svc.Healthcheck.TlsSni),
			Timeout:  time.Duration(papi.OptionalInt64(svc.Healthcheck.Timeout)) * time.Second,
			URI:      papi.OptionalString(svc.Healthcheck.Uri),
		},
		HealthcheckStatus: func() []*NetworkLoadBalancerServerStatus {
			statuses := make([]*NetworkLoadBalancerServerStatus, 0)
			if svc.HealthcheckStatus != nil {
				for _, st := range *svc.HealthcheckStatus {
					st := st
					statuses = append(statuses, nlbServerStatusFromAPI(&st))
				}
			}
			return statuses
		}(),
		ID:             *svc.Id,
		InstancePoolID: *svc.InstancePool.Id,
		Name:           *svc.Name,
		Port:           uint16(*svc.Port),
		Protocol:       string(*svc.Protocol),
		Strategy:       string(*svc.Strategy),
		TargetPort:     uint16(*svc.TargetPort),
		State:          string(*svc.State),
	}
}

// NetworkLoadBalancer represents a Network Load Balancer instance.
type NetworkLoadBalancer struct {
	CreatedAt   time.Time
	Description string
	ID          string
	IPAddress   net.IP
	Labels      map[string]string `reset:"labels"`
	Name        string
	Services    []*NetworkLoadBalancerService
	State       string

	c    *Client
	zone string
}

func nlbFromAPI(client *Client, zone string, nlb *papi.LoadBalancer) *NetworkLoadBalancer {
	return &NetworkLoadBalancer{
		CreatedAt:   *nlb.CreatedAt,
		Description: papi.OptionalString(nlb.Description),
		ID:          *nlb.Id,
		IPAddress:   net.ParseIP(papi.OptionalString(nlb.Ip)),
		Labels: func() map[string]string {
			if nlb.Labels != nil {
				return nlb.Labels.AdditionalProperties
			}
			return nil
		}(),
		Name: *nlb.Name,
		Services: func() []*NetworkLoadBalancerService {
			services := make([]*NetworkLoadBalancerService, 0)
			if nlb.Services != nil {
				for _, svc := range *nlb.Services {
					svc := svc
					services = append(services, nlbServiceFromAPI(&svc))
				}
			}
			return services
		}(),
		State: string(*nlb.State),

		c:    client,
		zone: zone,
	}
}

// AddService adds a service to the Network Load Balancer instance.
func (nlb *NetworkLoadBalancer) AddService(
	ctx context.Context,
	svc *NetworkLoadBalancerService,
) (*NetworkLoadBalancerService, error) {
	var (
		port                = int64(svc.Port)
		targetPort          = int64(svc.TargetPort)
		healthcheckPort     = int64(svc.Healthcheck.Port)
		healthcheckInterval = int64(svc.Healthcheck.Interval.Seconds())
		healthcheckTimeout  = int64(svc.Healthcheck.Timeout.Seconds())
	)

	// The API doesn't return the NLB service created directly, so in order to return a
	// *NetworkLoadBalancerService corresponding to the new service we have to manually
	// compare the list of services on the NLB instance before and after the service
	// creation, and identify the service that wasn't there before.
	// Note: in case of multiple services creation in parallel this technique is subject
	// to race condition as we could return an unrelated service. To prevent this, we
	// also compare the name of the new service to the name specified in the svc
	// parameter.
	services := make(map[string]struct{})
	for _, svc := range nlb.Services {
		services[svc.ID] = struct{}{}
	}

	resp, err := nlb.c.AddServiceToLoadBalancerWithResponse(
		apiv2.WithZone(ctx, nlb.zone),
		nlb.ID,
		papi.AddServiceToLoadBalancerJSONRequestBody{
			Description: func() *string {
				if svc.Description != "" {
					return &svc.Description
				}
				return nil
			}(),
			Healthcheck: papi.LoadBalancerServiceHealthcheck{
				Interval: &healthcheckInterval,
				Mode:     papi.LoadBalancerServiceHealthcheckMode(svc.Healthcheck.Mode),
				Port:     healthcheckPort,
				Retries:  &svc.Healthcheck.Retries,
				Timeout:  &healthcheckTimeout,
				TlsSni: func() *string {
					if svc.Healthcheck.Mode == "https" && svc.Healthcheck.TLSSNI != "" {
						return &svc.Healthcheck.TLSSNI
					}
					return nil
				}(),
				Uri: func() *string {
					if strings.HasPrefix(svc.Healthcheck.Mode, "http") {
						return &svc.Healthcheck.URI
					}
					return nil
				}(),
			},
			InstancePool: papi.InstancePool{Id: &svc.InstancePoolID},
			Name:         svc.Name,
			Port:         port,
			Protocol:     papi.AddServiceToLoadBalancerJSONBodyProtocol(svc.Protocol),
			Strategy:     papi.AddServiceToLoadBalancerJSONBodyStrategy(svc.Strategy),
			TargetPort:   targetPort,
		})
	if err != nil {
		return nil, err
	}

	res, err := papi.NewPoller().
		WithTimeout(nlb.c.timeout).
		WithInterval(nlb.c.pollInterval).
		Poll(ctx, nlb.c.OperationPoller(nlb.zone, *resp.JSON200.Id))
	if err != nil {
		return nil, err
	}

	nlbUpdated, err := nlb.c.GetNetworkLoadBalancer(ctx, nlb.zone, *res.(*papi.Reference).Id)
	if err != nil {
		return nil, err
	}

	// Look for an unknown service: if we find one we hope it's the one we've just created.
	for _, s := range nlbUpdated.Services {
		if _, ok := services[svc.ID]; !ok && s.Name == svc.Name {
			return s, nil
		}
	}

	return nil, errors.New("unable to identify the service created")
}

// UpdateService updates the specified Network Load Balancer service.
func (nlb *NetworkLoadBalancer) UpdateService(ctx context.Context, svc *NetworkLoadBalancerService) error {
	var (
		healthcheckPort     = int64(svc.Healthcheck.Port)
		healthcheckInterval = int64(svc.Healthcheck.Interval.Seconds())
		healthcheckTimeout  = int64(svc.Healthcheck.Timeout.Seconds())
	)

	resp, err := nlb.c.UpdateLoadBalancerServiceWithResponse(
		apiv2.WithZone(ctx, nlb.zone),
		nlb.ID,
		svc.ID,
		papi.UpdateLoadBalancerServiceJSONRequestBody{
			Description: func() *string {
				if svc.Description != "" {
					return &svc.Description
				}
				return nil
			}(),
			Healthcheck: &papi.LoadBalancerServiceHealthcheck{
				Interval: &healthcheckInterval,
				Mode:     papi.LoadBalancerServiceHealthcheckMode(svc.Healthcheck.Mode),
				Port:     healthcheckPort,
				Retries:  &svc.Healthcheck.Retries,
				Timeout:  &healthcheckTimeout,
				TlsSni: func() *string {
					if svc.Healthcheck.Mode == "https" && svc.Healthcheck.TLSSNI != "" {
						return &svc.Healthcheck.TLSSNI
					}
					return nil
				}(),
				Uri: func() *string {
					if strings.HasPrefix(svc.Healthcheck.Mode, "http") {
						return &svc.Healthcheck.URI
					}
					return nil
				}(),
			},
			Name: func() *string {
				if svc.Name != "" {
					return &svc.Name
				}
				return nil
			}(),
			Port: func() *int64 {
				if v := svc.Port; v > 0 {
					port := int64(v)
					return &port
				}
				return nil
			}(),
			Protocol: func() *papi.UpdateLoadBalancerServiceJSONBodyProtocol {
				if svc.Protocol != "" {
					return (*papi.UpdateLoadBalancerServiceJSONBodyProtocol)(&svc.Protocol)
				}
				return nil
			}(),
			Strategy: func() *papi.UpdateLoadBalancerServiceJSONBodyStrategy {
				if svc.Strategy != "" {
					return (*papi.UpdateLoadBalancerServiceJSONBodyStrategy)(&svc.Strategy)
				}
				return nil
			}(),
			TargetPort: func() *int64 {
				if v := svc.TargetPort; v > 0 {
					targetPort := int64(v)
					return &targetPort
				}
				return nil
			}(),
		})
	if err != nil {
		return err
	}

	_, err = papi.NewPoller().
		WithTimeout(nlb.c.timeout).
		WithInterval(nlb.c.pollInterval).
		Poll(ctx, nlb.c.OperationPoller(nlb.zone, *resp.JSON200.Id))
	if err != nil {
		return err
	}

	return nil
}

// DeleteService deletes the specified service from the Network Load Balancer instance.
func (nlb *NetworkLoadBalancer) DeleteService(ctx context.Context, svc *NetworkLoadBalancerService) error {
	resp, err := nlb.c.DeleteLoadBalancerServiceWithResponse(
		apiv2.WithZone(ctx, nlb.zone),
		nlb.ID,
		svc.ID,
	)
	if err != nil {
		return err
	}

	_, err = papi.NewPoller().
		WithTimeout(nlb.c.timeout).
		WithInterval(nlb.c.pollInterval).
		Poll(ctx, nlb.c.OperationPoller(nlb.zone, *resp.JSON200.Id))
	if err != nil {
		return err
	}

	return nil
}

// ResetField resets the specified Network Load Balancer field to its default value.
// The value expected for the field parameter is a pointer to the NetworkLoadBalancer field to reset.
func (nlb *NetworkLoadBalancer) ResetField(ctx context.Context, field interface{}) error {
	resetField, err := resetFieldName(nlb, field)
	if err != nil {
		return err
	}

	resp, err := nlb.c.ResetLoadBalancerFieldWithResponse(
		apiv2.WithZone(ctx, nlb.zone),
		nlb.ID,
		papi.ResetLoadBalancerFieldParamsField(resetField),
	)
	if err != nil {
		return err
	}

	_, err = papi.NewPoller().
		WithTimeout(nlb.c.timeout).
		WithInterval(nlb.c.pollInterval).
		Poll(ctx, nlb.c.OperationPoller(nlb.zone, *resp.JSON200.Id))
	if err != nil {
		return err
	}

	return nil
}

// CreateNetworkLoadBalancer creates a Network Load Balancer instance in the specified zone.
func (c *Client) CreateNetworkLoadBalancer(
	ctx context.Context,
	zone string,
	nlb *NetworkLoadBalancer,
) (*NetworkLoadBalancer, error) {
	resp, err := c.CreateLoadBalancerWithResponse(
		apiv2.WithZone(ctx, zone),
		papi.CreateLoadBalancerJSONRequestBody{
			Description: func() *string {
				if nlb.Description != "" {
					return &nlb.Description
				}
				return nil
			}(),
			Labels: func() *papi.Labels {
				if len(nlb.Labels) > 0 {
					return &papi.Labels{AdditionalProperties: nlb.Labels}
				}
				return nil
			}(),
			Name: nlb.Name,
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

	return c.GetNetworkLoadBalancer(ctx, zone, *res.(*papi.Reference).Id)
}

// ListNetworkLoadBalancers returns the list of existing Network Load Balancers in the
// specified zone.
func (c *Client) ListNetworkLoadBalancers(ctx context.Context, zone string) ([]*NetworkLoadBalancer, error) {
	list := make([]*NetworkLoadBalancer, 0)

	resp, err := c.ListLoadBalancersWithResponse(apiv2.WithZone(ctx, zone))
	if err != nil {
		return nil, err
	}

	if resp.JSON200.LoadBalancers != nil {
		for i := range *resp.JSON200.LoadBalancers {
			list = append(list, nlbFromAPI(c, zone, &(*resp.JSON200.LoadBalancers)[i]))
		}
	}

	return list, nil
}

// GetNetworkLoadBalancer returns the Network Load Balancer instance corresponding to the
// specified ID in the specified zone.
func (c *Client) GetNetworkLoadBalancer(ctx context.Context, zone, id string) (*NetworkLoadBalancer, error) {
	resp, err := c.GetLoadBalancerWithResponse(apiv2.WithZone(ctx, zone), id)
	if err != nil {
		return nil, err
	}

	return nlbFromAPI(c, zone, resp.JSON200), nil
}

// FindNetworkLoadBalancer attempts to find a Network Load Balancer by name or ID in the specified zone.
func (c *Client) FindNetworkLoadBalancer(ctx context.Context, zone, v string) (*NetworkLoadBalancer, error) {
	res, err := c.ListNetworkLoadBalancers(ctx, zone)
	if err != nil {
		return nil, err
	}

	for _, r := range res {
		if r.ID == v || r.Name == v {
			return c.GetNetworkLoadBalancer(ctx, zone, r.ID)
		}
	}

	return nil, apiv2.ErrNotFound
}

// UpdateNetworkLoadBalancer updates the specified Network Load Balancer instance in the specified zone.
func (c *Client) UpdateNetworkLoadBalancer(ctx context.Context, zone string, nlb *NetworkLoadBalancer) error {
	resp, err := c.UpdateLoadBalancerWithResponse(
		apiv2.WithZone(ctx, zone),
		nlb.ID,
		papi.UpdateLoadBalancerJSONRequestBody{
			Description: func() *string {
				if nlb.Description != "" {
					return &nlb.Description
				}
				return nil
			}(),
			Labels: func() *papi.Labels {
				if len(nlb.Labels) > 0 {
					return &papi.Labels{AdditionalProperties: nlb.Labels}
				}
				return nil
			}(),
			Name: func() *string {
				if nlb.Name != "" {
					return &nlb.Name
				}
				return nil
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

// DeleteNetworkLoadBalancer deletes the specified Network Load Balancer instance in the specified zone.
func (c *Client) DeleteNetworkLoadBalancer(ctx context.Context, zone, id string) error {
	resp, err := c.DeleteLoadBalancerWithResponse(apiv2.WithZone(ctx, zone), id)
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
