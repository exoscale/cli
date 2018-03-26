package egoscale

import (
	"context"
	"net"
)

// Nic represents a Network Interface Controller (NIC)
type Nic struct {
	ID               string           `json:"id,omitempty"`
	BroadcastURI     string           `json:"broadcasturi,omitempty"`
	Gateway          net.IP           `json:"gateway,omitempty"`
	IP6Address       net.IP           `json:"ip6address,omitempty"`
	IP6Cidr          string           `json:"ip6cidr,omitempty"`
	IP6Gateway       net.IP           `json:"ip6gateway,omitempty"`
	IPAddress        net.IP           `json:"ipaddress,omitempty"`
	IsDefault        bool             `json:"isdefault,omitempty"`
	IsolationURI     string           `json:"isolationuri,omitempty"`
	MacAddress       string           `json:"macaddress,omitempty"`
	Netmask          net.IP           `json:"netmask,omitempty"`
	NetworkID        string           `json:"networkid,omitempty"`
	NetworkName      string           `json:"networkname,omitempty"`
	SecondaryIP      []NicSecondaryIP `json:"secondaryip,omitempty"`
	TrafficType      string           `json:"traffictype,omitempty"`
	Type             string           `json:"type,omitempty"`
	VirtualMachineID string           `json:"virtualmachineid,omitempty"`
}

// List fetches all the nics
func (nic *Nic) List(ctx context.Context, client *Client) (<-chan interface{}, <-chan error) {
	pageSize := client.PageSize
	outChan := make(chan interface{}, client.PageSize)
	errChan := make(chan error, 1)

	go func() {
		defer close(outChan)
		defer close(errChan)

		page := 1

		req := &ListNics{
			VirtualMachineID: nic.VirtualMachineID,
			NicID:            nic.ID,
			NetworkID:        nic.NetworkID,
			PageSize:         pageSize,
		}

		for {
			req.Page = page

			resp, err := client.RequestWithContext(ctx, req)
			if err != nil {
				errChan <- err
				break
			}

			nics := resp.(*ListNicsResponse)
			for _, zone := range nics.Nic {
				outChan <- zone
			}

			if len(nics.Nic) < pageSize {
				break
			}

			page++
		}
	}()

	return outChan, errChan
}

// NicSecondaryIP represents a link between NicID and IPAddress
type NicSecondaryIP struct {
	ID               string `json:"id"`
	IPAddress        net.IP `json:"ipaddress"`
	NetworkID        string `json:"networkid"`
	NicID            string `json:"nicid"`
	VirtualMachineID string `json:"virtualmachineid,omitempty"`
}

// ListNics represents the NIC search
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/listNics.html
type ListNics struct {
	VirtualMachineID string `json:"virtualmachineid"`
	ForDisplay       bool   `json:"fordisplay,omitempty"`
	Keyword          string `json:"keyword,omitempty"`
	NetworkID        string `json:"networkid,omitempty"`
	NicID            string `json:"nicid,omitempty"`
	Page             int    `json:"page,omitempty"`
	PageSize         int    `json:"pagesize,omitempty"`
}

// APIName returns the CloudStack API command name
func (*ListNics) APIName() string {
	return "listNics"
}

func (*ListNics) response() interface{} {
	return new(ListNicsResponse)
}

// ListNicsResponse represents a list of templates
type ListNicsResponse struct {
	Count int   `json:"count"`
	Nic   []Nic `json:"nic"`
}

// AddIPToNic (Async) represents the assignation of a secondary IP
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/addIpToNic.html
type AddIPToNic struct {
	NicID     string `json:"nicid"`
	IPAddress net.IP `json:"ipaddress"`
}

// APIName returns the CloudStack API command name: addIpToNic
func (*AddIPToNic) APIName() string {
	return "addIpToNic"
}
func (*AddIPToNic) asyncResponse() interface{} {
	return new(AddIPToNicResponse)
}

// AddIPToNicResponse represents the addition of an IP to a NIC
type AddIPToNicResponse struct {
	NicSecondaryIP NicSecondaryIP `json:"nicsecondaryip"`
}

// RemoveIPFromNic (Async) represents a deletion request
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/removeIpFromNic.html
type RemoveIPFromNic struct {
	ID string `json:"id"`
}

// APIName returns the CloudStack API command name: removeIpFromNic
func (*RemoveIPFromNic) APIName() string {
	return "removeIpFromNic"
}

func (*RemoveIPFromNic) asyncResponse() interface{} {
	return new(booleanAsyncResponse)
}

// ActivateIP6 (Async) activates the IP6 on the given NIC
//
// Exoscale specific API: https://community.exoscale.ch/api/compute/#activateip6_GET
type ActivateIP6 struct {
	NicID string `json:"nicid"`
}

// APIName returns the CloudStack API command name: activateIp6
func (*ActivateIP6) APIName() string {
	return "activateIp6"
}

func (*ActivateIP6) asyncResponse() interface{} {
	return new(ActivateIP6Response)
}

// ActivateIP6Response represents the modified NIC
type ActivateIP6Response struct {
	Nic Nic `json:"nic"`
}
