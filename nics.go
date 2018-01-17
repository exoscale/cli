package egoscale

import (
	"fmt"
	"net"
)

// Nic represents a Network Interface Controller (NIC)
type Nic struct {
	ID               string            `json:"id,omitempty"`
	BroadcastURI     string            `json:"broadcasturi,omitempty"`
	Gateway          net.IP            `json:"gateway,omitempty"`
	IP6Address       net.IP            `json:"ip6address,omitempty"`
	IP6Cidr          string            `json:"ip6cidr,omitempty"`
	IP6Gateway       net.IP            `json:"ip6gateway,omitempty"`
	IPAddress        net.IP            `json:"ipaddress,omitempty"`
	IsDefault        bool              `json:"isdefault,omitempty"`
	IsolationURI     string            `json:"isolationuri,omitempty"`
	MacAddress       string            `json:"macaddress,omitempty"`
	Netmask          net.IP            `json:"netmask,omitempty"`
	NetworkID        string            `json:"networkid,omitempty"`
	NetworkName      string            `json:"networkname,omitempty"`
	SecondaryIP      []*NicSecondaryIP `json:"secondaryip,omitempty"`
	Traffictype      string            `json:"traffictype,omitempty"`
	Type             string            `json:"type,omitempty"`
	VirtualMachineID string            `json:"virtualmachineid,omitempty"`
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
type ListNics struct {
	VirtualMachineID string `json:"virtualmachineid"`
	ForDisplay       bool   `json:"fordisplay,omitempty"`
	Keyword          string `json:"keyword,omitempty"`
	NetworkID        string `json:"networkid,omitempty"`
	NicID            string `json:"nicid,omitempty"`
	Page             int    `json:"page,omitempty"`
	PageSize         int    `json:"pagesize,omitempty"`
}

func (req *ListNics) name() string {
	return "listNics"
}

func (req *ListNics) response() interface{} {
	return new(ListNicsResponse)
}

// ListNicsResponse represents a list of templates
type ListNicsResponse struct {
	Count int    `json:"count"`
	Nic   []*Nic `json:"nic"`
}

// AddIPToNic represents the assignation of a secondary IP
type AddIPToNic struct {
	NicID     string `json:"nicid"`
	IPAddress net.IP `json:"ipaddress"`
}

func (req *AddIPToNic) name() string {
	return "addIpToNic"
}
func (req *AddIPToNic) asyncResponse() interface{} {
	return new(AddIPToNicResponse)
}

// AddIPToNicResponse represents the addition of an IP to a NIC
type AddIPToNicResponse struct {
	NicSecondaryIP *NicSecondaryIP `json:"nicsecondaryip"`
}

// RemoveIPFromNic represents a deletion request
type RemoveIPFromNic struct {
	ID string `json:"id"`
}

func (req *RemoveIPFromNic) name() string {
	return "removeIpFromNic"
}

func (req *RemoveIPFromNic) asyncResponse() interface{} {
	return new(booleanAsyncResponse)
}

// ListNics lists the NIC of a VM
//
// Deprecated: use the API directly
func (exo *Client) ListNics(req *ListNics) ([]*Nic, error) {
	resp, err := exo.Request(req)
	if err != nil {
		return nil, err
	}

	return resp.(*ListNicsResponse).Nic, nil
}

// AddIPToNic adds an IP to a NIC
//
// Deprecated: use the API directly
func (exo *Client) AddIPToNic(nicID string, ipAddress string, async AsyncInfo) (*NicSecondaryIP, error) {
	ip := net.ParseIP(ipAddress)
	if ip == nil {
		return nil, fmt.Errorf("%s is not a valid IP address", ipAddress)
	}
	req := &AddIPToNic{
		NicID:     nicID,
		IPAddress: ip,
	}
	resp, err := exo.AsyncRequest(req, async)
	if err != nil {
		return nil, err
	}

	return resp.(AddIPToNicResponse).NicSecondaryIP, nil
}

// RemoveIPFromNic removes an IP from a NIC
//
// Deprecated: use the API directly
func (exo *Client) RemoveIPFromNic(secondaryNicID string, async AsyncInfo) error {
	req := &RemoveIPFromNic{
		ID: secondaryNicID,
	}
	return exo.BooleanAsyncRequest(req, async)
}
