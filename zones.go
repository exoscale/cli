package egoscale

import (
	"context"
	"net"
)

// Zone represents a data center
type Zone struct {
	ID                    string            `json:"id"`
	AllocationState       string            `json:"allocationstate,omitempty"`
	Capacity              string            `json:"capacity,omitempty"`
	Description           string            `json:"description,omitempty"`
	DhcpProvider          string            `json:"dhcpprovider,omitempty"`
	DisplayText           string            `json:"displaytext,omitempty"`
	DNS1                  net.IP            `json:"dns1,omitempty"`
	DNS2                  net.IP            `json:"dns2,omitempty"`
	Domain                string            `json:"domain,omitempty"`
	DomainID              string            `json:"domainid,omitempty"`
	DomainName            string            `json:"domainname,omitempty"`
	GuestCidrAddress      string            `json:"guestcidraddress,omitempty"`
	InternalDNS1          net.IP            `json:"internaldns1,omitempty"`
	InternalDNS2          net.IP            `json:"internaldns2,omitempty"`
	IP6DNS1               net.IP            `json:"ip6dns1,omitempty"`
	IP6DNS2               net.IP            `json:"ip6dns2,omitempty"`
	LocalStorageEnabled   bool              `json:"localstorageenabled,omitempty"`
	Name                  string            `json:"name,omitempty"`
	NetworkType           string            `json:"networktype,omitempty"`
	ResourceDetails       map[string]string `json:"resourcedetails,omitempty"`
	SecurityGroupsEnabled bool              `json:"securitygroupsenabled,omitempty"`
	Vlan                  string            `json:"vlan,omitempty"`
	ZoneToken             string            `json:"zonetoken,omitempty"`
	Tags                  []ResourceTag     `json:"tags,omitempty"`
}

// List fetches all the zones
func (zone *Zone) List(ctx context.Context, client *Client) (<-chan interface{}, <-chan error) {
	pageSize := client.PageSize
	outChan := make(chan interface{}, client.PageSize)
	errChan := make(chan error, 1)

	go func() {
		defer close(outChan)
		defer close(errChan)

		page := 1

		req := &ListZones{
			DomainID: zone.DomainID,
			ID:       zone.ID,
			Name:     zone.Name,
			PageSize: pageSize,
		}

		for {
			req.Page = page

			resp, err := client.RequestWithContext(ctx, req)
			if err != nil {
				errChan <- err
				break
			}

			zones := resp.(*ListZonesResponse)
			for _, zone := range zones.Zone {
				outChan <- zone
			}

			if len(zones.Zone) < pageSize {
				break
			}

			page++
		}
	}()

	return outChan, errChan
}

// ListZones represents a query for zones
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/listZones.html
type ListZones struct {
	Available      *bool         `json:"available,omitempty"`
	DomainID       string        `json:"domainid,omitempty"`
	ID             string        `json:"id,omitempty"`
	Keyword        string        `json:"keyword,omitempty"`
	Name           string        `json:"name,omitempty"`
	Page           int           `json:"page,omitempty"`
	PageSize       int           `json:"pagesize,omitempty"`
	ShowCapacities *bool         `json:"showcapacities,omitempty"`
	Tags           []ResourceTag `json:"tags,omitempty"`
}

// APIName returns the CloudStack API command name
func (*ListZones) APIName() string {
	return "listZones"
}

func (*ListZones) response() interface{} {
	return new(ListZonesResponse)
}

// ListZonesResponse represents a list of zones
type ListZonesResponse struct {
	Count int    `json:"count"`
	Zone  []Zone `json:"zone"`
}
