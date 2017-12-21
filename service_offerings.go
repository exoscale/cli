/*
Service Offerings

A service offering correspond to some hardware features (CPU, RAM).

See: http://docs.cloudstack.apache.org/projects/cloudstack-administration/en/latest/service_offerings.html
*/
package egoscale

// ServiceOffering corresponds to the Compute Offerings
type ServiceOffering struct {
	CpuNumber              int               `json:"cpunumber,omitempty"`
	CpuSpeed               int               `json:"cpuspeed,omitempty"`
	DisplayText            string            `json:"displaytext,omitempty"`
	Domain                 string            `json:"domain,omitempty"`
	DomainId               string            `json:"domainid,omitempty"`
	HostTags               string            `json:"hosttags,omitempty"`
	Id                     string            `json:"id,omitempty"`
	IsCustomized           bool              `json:"iscustomized,omitempty"`
	IsSystem               bool              `json:"issystem,omitempty"`
	IsVolatile             bool              `json:"isvolatile,omitempty"`
	Memory                 int               `json:"memory,omitempty"`
	Name                   string            `json:"name,omitempty"`
	NetworkRate            int               `json:"networkrate,omitempty"`
	ServiceOfferingDetails map[string]string `json:"serviceofferingdetails,omitempty"`
}

// ListServiceOfferingRequest represents a query for service offerings
type ListServiceOfferingsRequest struct {
	DomainId         string `json:"domainid,omitempty"`
	Id               string `json:"id,omitempty"`
	IsSystem         bool   `json:"issystem,omitempty"`
	Keyword          string `json:"keyword,omitempty"`
	Name             string `json:"name,omitempty"`
	Page             int    `json:"page,omitempty"`
	PageSize         int    `json:"pagesize,omitempty"`
	SystemVmType     string `json:"systemvmtype"`
	VirtualMachineId string `json:"virtualmachineid"`
}

// Command returns the CloudStack API command
func (req *ListServiceOfferingsRequest) Command() string {
	return "listServiceOfferings"
}

// ListServiceOfferingsResponse represents a list of service offerings
type ListServiceOfferingsResponse struct {
	Count           int                `json:"count"`
	ServiceOffering []*ServiceOffering `json:"serviceoffering"`
}

func (exo *Client) ListServiceOfferings(req *ListServiceOfferingsRequest) ([]*ServiceOffering, error) {
	var r ListServiceOfferingsResponse
	err := exo.Request(req, &r)
	if err != nil {
		return nil, err
	}

	return r.ServiceOffering, nil
}
