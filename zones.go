/*
Zones

A Zone corresponds to a Data Center.
*/
package egoscale

// Zone represents a data center
type Zone struct {
	Id                    string            `json:"id"`
	AllocationState       string            `json:"allocationstate,omitempty"`
	Capacity              string            `json:"capacity,omitempty"`
	Description           string            `json:"description,omitempty"`
	DhcpProvider          string            `json:"dhcpprovider,omitempty"`
	DisplayText           string            `json:"displaytext,omitempty"`
	Dns1                  string            `json:"dns1,omitempty"`
	Dns2                  string            `json:"dns2,omitempty"`
	Domain                string            `json:"domain,omitempty"`
	DomainId              string            `json:"domainid,omitempty"`
	DomainName            string            `json:"domainname,omitempty"`
	GuestCidrAddress      string            `json:"guestcidraddress,omitempty"`
	InternalDns1          string            `json:"internaldns1,omitempty"`
	InternalDns2          string            `json:"internaldns2,omitempty"`
	Ip6Dns1               string            `json:"ip6dns1,omitempty"`
	Ip6Dns2               string            `json:"ip6dns2,omitempty"`
	LocalStorageEnabled   bool              `json:"localstorageenabled,omitempty"`
	Name                  string            `json:"name,omitempty"`
	NetworkType           string            `json:"networktype,omitempty"`
	ResourceDetails       map[string]string `json:"resourcedetails,omitempty"`
	SecurityGroupsEnabled bool              `json:"securitygroupsenabled,omitempty"`
	Vlan                  string            `json:"vlan,omitempty"`
	ZoneToken             string            `json:"zonetoken,omitempty"`
	Tags                  []*ResourceTag    `json:"tags,omitempty"`
}

// ListZonesRequest represents a query for zones
type ListZonesRequest struct {
	Available      bool           `json:"available,omitempty"`
	DomainId       string         `json:"domainid,omitempty"`
	Id             string         `json:"id,omitempty"`
	Keyword        string         `json:"keyword,omitempty"`
	Name           string         `json:"name,omitempty"`
	Page           int            `json:"page,omitempty"`
	PageSize       int            `json:"pagesize,omitempty"`
	ShowCapacities bool           `json:"showcapacities,omitempty"`
	Tags           []*ResourceTag `json:"tags,omitempty"`
}

// Command returns the CloudStack API command
func (req *ListZonesRequest) Command() string {
	return "listZones"
}

// ListZonesResponse represents a list of zones
type ListZonesResponse struct {
	Count int     `json:"count"`
	Zone  []*Zone `json:"zone"`
}
