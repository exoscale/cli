package egoscale

import (
	"context"
	"fmt"
	"net"
	"net/url"

	"github.com/jinzhu/copier"
)

// ResourceType returns the type of the resource
func (*VirtualMachine) ResourceType() string {
	return "UserVM"
}

// Get fills the VM
func (vm *VirtualMachine) Get(ctx context.Context, client *Client) error {
	if vm.ID == "" && vm.Name == "" && vm.DefaultNic() == nil {
		return fmt.Errorf("A VirtualMachine may only be searched using ID, Name or IPAddress")
	}

	vms, err := client.ListWithContext(ctx, vm)
	if err != nil {
		return err
	}

	count := len(vms)
	if count == 0 {
		return &ErrorResponse{
			ErrorCode: ParamError,
			ErrorText: fmt.Sprintf("VirtualMachine not found. id: %s, name: %s", vm.ID, vm.Name),
		}
	} else if count > 1 {
		return fmt.Errorf("More than one VirtualMachine was found. Query: id: %s, name: %s", vm.ID, vm.Name)
	}

	return copier.Copy(vm, vms[0])
}

// Delete destroys the VM
func (vm *VirtualMachine) Delete(ctx context.Context, client *Client) error {
	_, err := client.RequestWithContext(ctx, &DestroyVirtualMachine{
		ID: vm.ID,
	})

	return err
}

// ListRequest builds the ListVirtualMachines request
func (vm *VirtualMachine) ListRequest() (ListCommand, error) {
	// XXX: AffinityGroupID, SecurityGroupID, Tags

	req := &ListVirtualMachines{
		Account:    vm.Account,
		DomainID:   vm.DomainID,
		GroupID:    vm.GroupID,
		ID:         vm.ID,
		Name:       vm.Name,
		ProjectID:  vm.ProjectID,
		State:      vm.State,
		TemplateID: vm.TemplateID,
		ZoneID:     vm.ZoneID,
	}

	nic := vm.DefaultNic()
	if nic != nil {
		req.IPAddress = nic.IPAddress
	}

	return req, nil
}

// DefaultNic returns the default nic
func (vm *VirtualMachine) DefaultNic() *Nic {
	for _, nic := range vm.Nic {
		if nic.IsDefault {
			return &nic
		}
	}

	return nil
}

// IP returns the default nic IP address
func (vm *VirtualMachine) IP() *net.IP {
	nic := vm.DefaultNic()
	if nic != nil {
		ip := nic.IPAddress
		return &ip
	}

	return nil
}

// NicsByType returns the corresponding interfaces base on the given type
func (vm *VirtualMachine) NicsByType(nicType string) []Nic {
	nics := make([]Nic, 0)
	for _, nic := range vm.Nic {
		if nic.Type == nicType {
			// XXX The CloudStack API forgets to specify it
			nic.VirtualMachineID = vm.ID
			nics = append(nics, nic)
		}
	}
	return nics
}

// NicByNetworkID returns the corresponding interface based on the given NetworkID
//
// A VM cannot be connected twice to a same network.
func (vm *VirtualMachine) NicByNetworkID(networkID string) *Nic {
	for _, nic := range vm.Nic {
		if nic.NetworkID == networkID {
			nic.VirtualMachineID = vm.ID
			return &nic
		}
	}
	return nil
}

// NicByID returns the corresponding interface base on its ID
func (vm *VirtualMachine) NicByID(nicID string) *Nic {
	for _, nic := range vm.Nic {
		if nic.ID == nicID {
			nic.VirtualMachineID = vm.ID
			return &nic
		}
	}

	return nil
}

// APIName returns the CloudStack API command name
func (*DeployVirtualMachine) APIName() string {
	return "deployVirtualMachine"
}

func (req *DeployVirtualMachine) onBeforeSend(params *url.Values) error {
	// Either AffinityGroupIDs or AffinityGroupNames must be set
	if len(req.AffinityGroupIDs) > 0 && len(req.AffinityGroupNames) > 0 {
		return fmt.Errorf("Either AffinityGroupIDs or AffinityGroupNames must be set")
	}

	// Either SecurityGroupIDs or SecurityGroupNames must be set
	if len(req.SecurityGroupIDs) > 0 && len(req.SecurityGroupNames) > 0 {
		return fmt.Errorf("Either SecurityGroupIDs or SecurityGroupNames must be set")
	}

	return nil
}

func (*DeployVirtualMachine) asyncResponse() interface{} {
	return new(DeployVirtualMachineResponse)
}

// APIName returns the CloudStack API command name
func (*StartVirtualMachine) APIName() string {
	return "startVirtualMachine"
}
func (*StartVirtualMachine) asyncResponse() interface{} {
	return new(StartVirtualMachineResponse)
}

// APIName returns the CloudStack API command name
func (*StopVirtualMachine) APIName() string {
	return "stopVirtualMachine"
}

func (*StopVirtualMachine) asyncResponse() interface{} {
	return new(StopVirtualMachineResponse)
}

// APIName returns the CloudStack API command name
func (*RebootVirtualMachine) APIName() string {
	return "rebootVirtualMachine"
}

func (*RebootVirtualMachine) asyncResponse() interface{} {
	return new(RebootVirtualMachineResponse)
}

// RebootVirtualMachineResponse represents a rebooted VM instance
type RebootVirtualMachineResponse VirtualMachineResponse

// APIName returns the CloudStack API command name
func (*RestoreVirtualMachine) APIName() string {
	return "restoreVirtualMachine"
}

func (*RestoreVirtualMachine) asyncResponse() interface{} {
	return new(RestoreVirtualMachineResponse)
}

// APIName returns the CloudStack API command name
func (*RecoverVirtualMachine) APIName() string {
	return "recoverVirtualMachine"
}

func (*RecoverVirtualMachine) response() interface{} {
	return new(RecoverVirtualMachineResponse)
}

// APIName returns the CloudStack API command name
func (*DestroyVirtualMachine) APIName() string {
	return "destroyVirtualMachine"
}

func (*DestroyVirtualMachine) asyncResponse() interface{} {
	return new(DestroyVirtualMachineResponse)
}

// APIName returns the CloudStack API command name
func (*UpdateVirtualMachine) APIName() string {
	return "updateVirtualMachine"
}

func (*UpdateVirtualMachine) response() interface{} {
	return new(UpdateVirtualMachineResponse)
}

// UpdateVirtualMachineResponse represents an updated VM instance
type UpdateVirtualMachineResponse VirtualMachineResponse

// ExpungeVirtualMachine represents the annihilation of a VM
type ExpungeVirtualMachine struct {
	ID string `json:"id"`
}

// APIName returns the CloudStack API command name
func (*ExpungeVirtualMachine) APIName() string {
	return "expungeVirtualMachine"
}

func (*ExpungeVirtualMachine) asyncResponse() interface{} {
	return new(booleanAsyncResponse)
}

// APIName returns the CloudStack API command name
func (*ScaleVirtualMachine) APIName() string {
	return "scaleVirtualMachine"
}

func (*ScaleVirtualMachine) asyncResponse() interface{} {
	return new(booleanAsyncResponse)
}

// APIName returns the CloudStack API command name
func (*ChangeServiceForVirtualMachine) APIName() string {
	return "changeServiceForVirtualMachine"
}

func (*ChangeServiceForVirtualMachine) response() interface{} {
	return new(ChangeServiceForVirtualMachineResponse)
}

// APIName returns the CloudStack API command name
func (*ResetPasswordForVirtualMachine) APIName() string {
	return "resetPasswordForVirtualMachine"
}

func (*ResetPasswordForVirtualMachine) asyncResponse() interface{} {
	return new(ResetPasswordForVirtualMachineResponse)
}

// APIName returns the CloudStack API command name
func (*GetVMPassword) APIName() string {
	return "getVMPassword"
}

func (*GetVMPassword) response() interface{} {
	return new(GetVMPasswordResponse)
}

// APIName returns the CloudStack API command name
func (*ListVirtualMachines) APIName() string {
	return "listVirtualMachines"
}

func (*ListVirtualMachines) response() interface{} {
	return new(ListVirtualMachinesResponse)
}

// SetPage sets the current page
func (ls *ListVirtualMachines) SetPage(page int) {
	ls.Page = page
}

// SetPageSize sets the page size
func (ls *ListVirtualMachines) SetPageSize(pageSize int) {
	ls.PageSize = pageSize
}

func (*ListVirtualMachines) each(resp interface{}, callback ListCommandFunc) {
	vms := resp.(*ListVirtualMachinesResponse)
	for _, vm := range vms.VirtualMachine {
		callback(vm, nil)
	}
}

// APIName returns the CloudStack API command name
func (*AddNicToVirtualMachine) APIName() string {
	return "addNicToVirtualMachine"
}

func (*AddNicToVirtualMachine) asyncResponse() interface{} {
	return new(AddNicToVirtualMachineResponse)
}

// APIName returns the CloudStack API command name
func (*RemoveNicFromVirtualMachine) APIName() string {
	return "removeNicFromVirtualMachine"
}

func (*RemoveNicFromVirtualMachine) asyncResponse() interface{} {
	return new(RemoveNicFromVirtualMachineResponse)
}

// APIName returns the CloudStack API command name
func (*UpdateDefaultNicForVirtualMachine) APIName() string {
	return "updateDefaultNicForVirtualMachine"
}

func (*UpdateDefaultNicForVirtualMachine) asyncResponse() interface{} {
	return new(UpdateDefaultNicForVirtualMachineResponse)
}
