/*
Affinity and Anti-Affinity groups

Affinity and Anti-Affinity groups provide a way to influence where VMs should run. See: http://docs.cloudstack.apache.org/projects/cloudstack-administration/en/stable/virtual_machines.html#affinity-groups
*/

package egoscale

// AffinityGroup represents an (anti-)affinity group
type AffinityGroup struct {
	Id                string   `json:"id,omitempty"`
	Account           string   `json:"account,omitempty"`
	Description       string   `json:"description,omitempty"`
	Domain            string   `json:"domain,omitempty"`
	DomainId          string   `json:"domainid,omitempty"`
	Name              string   `json:"name,omitempty"`
	Type              string   `json:"type,omitempty"`
	VirtualMachineIds []string `json:"virtualmachineIds,omitempty"` // *I*ds is not a typo
}

// AffinityGroupType represent an affinity group type
type AffinityGroupType struct {
	Type string `json:"type"`
}

// CreateAffinityGroupRequest represents a new (anti-)affinity group
type CreateAffinityGroupRequest struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Account     string `json:"account,omitempty"`
	Description string `json:"description,omitempty"`
	DomainId    string `json:"domainid,omitempty"`
}

// Command return the CloudStack API
func (req *CreateAffinityGroupRequest) Command() string {
	return "createAffinityGroupRequest"
}

// DeleteAffinityGroupRequest represents an (anti-)affinity group to be deleted
type DeleteAffinityGroupRequest struct {
	Id          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Type        string `json:"type,omitempty"`
	Account     string `json:"account,omitempty"`
	Description string `json:"description,omitempty"`
	DomainId    string `json:"domainid,omitempty"`
}

// Command returns the CloudStack API command
func (req *DeleteAffinityGroupRequest) Command() string {
	return "deleteAffinityGroup"
}

// ListAffinityGroupsRequest represents an (anti-)affinity groups search
type ListAffinityGroupsRequest struct {
	Account          string `json:"account,omitempty"`
	DomainId         string `json:"domainid,omitempty"`
	Id               string `json:"id,omitempty"`
	IsRecursive      bool   `json:"isrecursive,omitempty"`
	Keyword          string `json:"keyword,omitempty"`
	ListAll          bool   `json:"listall,omitempty"`
	Name             string `json:"name,omitempty"`
	Page             string `json:"page,omitempty"`
	PageSize         string `json:"pagesize,omitempty"`
	Type             string `json:"type,omitempty"`
	VirtualMachineId string `json:"virtualmachineid,omitempty"`
}

// Command return the CloudStack API command
func (req *ListAffinityGroupsRequest) Command() string {
	return "listAffinityGroups"
}

// CreateAffinityGroupResponse represents the response of the creation of an (anti-)affinity group
type CreateAffinityGroupResponse struct {
	AffinityGroup *AffinityGroup `json:"affinitygroup"`
}

// ListAffinityGroupTypesRequest represents an (anti-)affinity groups search
type ListAffinityGroupTypesRequest struct {
	Keyword  string `json:"keyword,omitempty"`
	Page     string `json:"page,omitempty"`
	PageSize string `json:"pagesize,omitempty"`
}

// Command return the CloudStack API command
func (req *ListAffinityGroupTypesRequest) Command() string {
	return "listAffinityGroupTypes"
}

// ListAffinityGroupsResponse represents a list of (anti-)affinity groups
type ListAffinityGroupsResponse struct {
	Count         int              `json:"count"`
	AffinityGroup []*AffinityGroup `json:"affinitygroup"`
}

// ListAffinityGroupTypesResponse represents a list of (anti-)affinity group types
type ListAffinityGroupTypesResponse struct {
	Count             int                  `json:"count"`
	AffinityGroupType []*AffinityGroupType `json:"affinitygrouptype"`
}

// XXX UpdateVmAffinityGroup
//
// http://cloudstack.apache.org/api/apidocs-4.10/apis/updateVMAffinityGroup.html

// Legacy methods

// Deprecated: CreateAffinityGroup creates a group
func (exo *Client) CreateAffinityGroup(name string, async AsyncInfo) (*AffinityGroup, error) {
	req := &CreateAffinityGroupRequest{
		Name: name,
	}
	resp := new(CreateAffinityGroupResponse)
	err := exo.AsyncRequest(req, resp, async)
	if err != nil {
		return nil, err
	}

	return resp.AffinityGroup, nil
}

// Deprecated: DeleteAffinityGroup deletes a group
func (exo *Client) DeleteAffinityGroup(name string, async AsyncInfo) error {
	req := &DeleteAffinityGroupRequest{
		Name: name,
	}
	return exo.BooleanAsyncRequest(req, async)
}
