package egoscale

import (
	"fmt"
)

// Volume represents a volume linked to a VM
type Volume struct {
	Id                         string        `json:"id"`
	Account                    string        `json:"account,omitempty"`
	AttachedAt                 string        `json:"attached,omitempty"`
	ChainInfo                  string        `json:"chaininfo,omitempty"`
	CreatedAt                  string        `json:"created,omitempty"`
	Destroyed                  bool          `json:"destroyed,omitempty"`
	DisplayVolume              bool          `json:"displayvolume,omitempty"`
	Domain                     string        `json:"domain,omitempty"`
	DomainId                   string        `json:"domainid,omitempty"`
	Name                       string        `json:"name,omitempty"`
	QuiesceVm                  bool          `json:"quiescevm,omitempty"`
	ServiceOfferingDisplayText string        `json:"serviceofferingdisplaytext,omitempty"`
	ServiceOfferingId          string        `json:"serviceofferingid,omitempty"`
	ServiceOfferingName        string        `json:"serviceofferingname,omitempty"`
	Size                       uint64        `json:"size,omitempty"`
	State                      string        `json:"state,omitempty"`
	Type                       string        `json:"type,omitempty"`
	VirtualMachineId           string        `json:"virtualmachineid,omitempty"`
	VmName                     string        `json:"vmname,omitempty"`
	VmState                    string        `json:"vmstate,omitempty"`
	ZoneId                     string        `json:"zoneid,omitempty"`
	ZoneName                   string        `json:"zonename,omitempty"`
	Tags                       []*VolumeTag  `json:"tags,omitempty"`
	JobId                      string        `json:"jobid,omitempty"`
	JobStatus                  JobStatusType `json:"jobstatus,omitempty"`
}

// VolumeTag represents a tag associated with a Volume
type VolumeTag struct {
	Account    string `json:"account,omitempty"`
	Customer   string `json:"customer,omitempty"`
	Domain     string `json:"domain,omitempty"`
	DomainId   string `json:"domainid,omitempty"`
	Key        string `json:"key,omitempty"`
	Project    string `json:"project,omitempty"`
	ProjectId  string `json:"projectid,omitempty"`
	Resource   string `json:"resource,omitempty"`
	ResourceId string `json:"resourceid,omitempty"`
	Value      string `json:"value,omitempty"`
}

// ListVolumeRequest represents a query listing volumes
type ListVolumesRequest struct {
	Account          string         `json:"account,omitempty"`
	DiskOfferingId   string         `json:"diskoffering,omitempty"`
	DisplayVolume    string         `json:"displayvolume,omitempty"` // root only
	DomainId         string         `json:"domainid,omitempty"`
	HostId           string         `json:"hostid,omitempty"`
	Id               string         `json:"id,omitempty"`
	IsRecursive      bool           `json:"isrecursive,omitempty"`
	Keyword          string         `json:"keyword,omitempty"`
	ListAll          bool           `json:"listall,omitempty"`
	Name             string         `json:"name,omitempty"`
	Page             int            `json:"page,omitempty"`
	PageSize         int            `json:"pagesize,omitempty"`
	PodId            string         `json:"podid,omitempty"`
	ProjectId        string         `json:"projectid,omitempty"`
	StorageId        string         `json:"storageid,omitempty"`
	Tags             []*ResourceTag `json:"tags,omitempty"`
	Type             string         `json:"type,omitempty"`
	VirtualMachineId string         `json:"virtualmachineid,omitempty"`
	ZoneId           string         `json:"zoneid,omitempty"`
}

// Command returns the CloudStack API command
func (req *ListVolumesRequest) Command() string {
	return "listVolumes"
}

// ListVolumesResponse represents a list of volumes
type ListVolumesResponse struct {
	Count  int       `json:"count"`
	Volume []*Volume `json:"volume"`
}

// ListVolumes
func (exo *Client) ListVolumes(req *ListVolumesRequest) ([]*Volume, error) {
	var r ListVolumesResponse
	err := exo.Request(req, &r)
	if err != nil {
		return nil, err
	}

	return r.Volume, nil
}

// GetRootVolumeForVirtualMachine(d.Id())
func (exo *Client) GetRootVolumeForVirtualMachine(virtualMachineId string) (*Volume, error) {
	volumes, err := exo.ListVolumes(&ListVolumesRequest{
		VirtualMachineId: virtualMachineId,
		Type:             "ROOT",
	})
	if err != nil {
		return nil, err
	}

	if len(volumes) != 1 {
		return nil, fmt.Errorf("Expected exactly one volume for %v, got %d", virtualMachineId, len(volumes))
	}

	return volumes[0], nil
}
