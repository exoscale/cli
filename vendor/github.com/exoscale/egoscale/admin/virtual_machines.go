package admin

import "github.com/exoscale/egoscale"

// VirtualMachine represents the enriched VirtualMachine in the admin world
type VirtualMachine struct {
	egoscale.VirtualMachine
	Account   string         `json:"account,omitempty" doc:"list resources by account"`
	AccountID *egoscale.UUID `json:"accountid,omitempty"`
	HostID    *egoscale.UUID `json:"hostid,omitempty" doc:"the host ID"`
	PodID     *egoscale.UUID `json:"podid,omitempty" doc:"the pod ID"`
	StorageID *egoscale.UUID `json:"storageid,omitempty" doc:"the storage ID where vm's volumes belong to"`
	HostName  string         `json:"string,omitempty"`
}

// ListVirtualMachines represents the enriched ListVirtualMachines command
type ListVirtualMachines struct {
	egoscale.ListVirtualMachines
	ListAll *bool `json:"listall,omitempty"`
}

// ListVirtualMachinesResponse represents the list of VirtualMachine in the admin world
type ListVirtualMachinesResponse struct {
	Count          int              `json:"count"`
	VirtualMachine []VirtualMachine `json:"virtualmachine"`
}

// Response returns the struct to unmarshal
func (ListVirtualMachines) Response() interface{} {
	return new(ListVirtualMachinesResponse)
}
