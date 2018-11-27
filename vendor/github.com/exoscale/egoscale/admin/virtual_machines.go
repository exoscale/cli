package admin

import "github.com/exoscale/egoscale"

type VirtualMachine struct {
	egoscale.VirtualMachine
	Account   string         `json:"account,omitempty" doc:"list resources by account"`
	AccountID *egoscale.UUID `json:"accountid,omitempty"`
	HostID    *egoscale.UUID `json:"hostid,omitempty" doc:"the host ID"`
	PodID     *egoscale.UUID `json:"storageid,omitempty" doc:"the pod ID"`
	StorageID *egoscale.UUID `json:"storageid,omitempty" doc:"the storage ID where vm's volumes belong to"`
	HostName  string         `json:"string,omitempty"`
}

type ListVirtualMachines struct {
	egoscale.ListVirtualMachines
	ListAll *bool `json:"listall,omitempty"`
}

type LisVirtualMachinesResponse struct {
	Count          int              `json:"count"`
	VirtualMachine []VirtualMachine `json:"virtualmachine"`
}
