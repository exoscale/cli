package egoscale

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net"
	"net/url"
)

// ResourceType returns the type of the resource
func (*VirtualMachine) ResourceType() string {
	return "UserVM"
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

func (*DeployVirtualMachine) name() string {
	return "deployVirtualMachine"
}

func (*DeployVirtualMachine) description() string {
	return "Creates and automatically starts a virtual machine based on a service offering, disk offering, and template."
}

func (req *DeployVirtualMachine) onBeforeSend(params *url.Values) error {
	// Either AffinityGroupIDs or AffinityGroupNames must be set
	if len(req.AffinityGroupIDs) > 0 && len(req.AffinityGroupNames) > 0 {
		return fmt.Errorf("either AffinityGroupIDs or AffinityGroupNames must be set")
	}

	// Either SecurityGroupIDs or SecurityGroupNames must be set
	if len(req.SecurityGroupIDs) > 0 && len(req.SecurityGroupNames) > 0 {
		return fmt.Errorf("either SecurityGroupIDs or SecurityGroupNames must be set")
	}

	return nil
}

func (*DeployVirtualMachine) asyncResponse() interface{} {
	return new(VirtualMachine)
}

func (*StartVirtualMachine) name() string {
	return "startVirtualMachine"
}

func (*StartVirtualMachine) description() string {
	return "Starts a virtual machine."
}

func (*StartVirtualMachine) asyncResponse() interface{} {
	return new(VirtualMachine)
}

func (*StopVirtualMachine) name() string {
	return "stopVirtualMachine"
}

func (*StopVirtualMachine) description() string {
	return "Stops a virtual machine."
}

func (*StopVirtualMachine) asyncResponse() interface{} {
	return new(VirtualMachine)
}

func (*RebootVirtualMachine) name() string {
	return "rebootVirtualMachine"
}

func (*RebootVirtualMachine) description() string {
	return "Reboots a virtual machine."
}

func (*RebootVirtualMachine) asyncResponse() interface{} {
	return new(VirtualMachine)
}

func (*RestoreVirtualMachine) name() string {
	return "restoreVirtualMachine"
}

func (*RestoreVirtualMachine) description() string {
	return "Restore a VM to original template/ISO or new template/ISO"
}

func (*RestoreVirtualMachine) asyncResponse() interface{} {
	return new(VirtualMachine)
}

func (*RecoverVirtualMachine) name() string {
	return "recoverVirtualMachine"
}

func (*RecoverVirtualMachine) description() string {
	return "Recovers a virtual machine."
}

func (*RecoverVirtualMachine) response() interface{} {
	return new(VirtualMachine)
}

func (*DestroyVirtualMachine) name() string {
	return "destroyVirtualMachine"
}

func (*DestroyVirtualMachine) description() string {
	return "Destroys a virtual machine."
}

func (*DestroyVirtualMachine) asyncResponse() interface{} {
	return new(VirtualMachine)
}

func (*UpdateVirtualMachine) name() string {
	return "updateVirtualMachine"
}

func (*UpdateVirtualMachine) description() string {
	return "Updates properties of a virtual machine. The VM has to be stopped and restarted for the new properties to take effect. UpdateVirtualMachine does not first check whether the VM is stopped. Therefore, stop the VM manually before issuing this call."
}

func (*UpdateVirtualMachine) response() interface{} {
	return new(VirtualMachine)
}

func (*ExpungeVirtualMachine) name() string {
	return "expungeVirtualMachine"
}

func (*ExpungeVirtualMachine) description() string {
	return "Expunge a virtual machine. Once expunged, it cannot be recoverd."
}

func (*ExpungeVirtualMachine) asyncResponse() interface{} {
	return new(booleanResponse)
}

func (*ScaleVirtualMachine) name() string {
	return "scaleVirtualMachine"
}

func (*ScaleVirtualMachine) description() string {
	return "Scales the virtual machine to a new service offering."
}

func (*ScaleVirtualMachine) asyncResponse() interface{} {
	return new(booleanResponse)
}

func (*ChangeServiceForVirtualMachine) name() string {
	return "changeServiceForVirtualMachine"
}

func (*ChangeServiceForVirtualMachine) description() string {
	return `Changes the service offering for a virtual machine. The virtual machine must be in a "Stopped" state for this command to take effect.`
}

func (*ChangeServiceForVirtualMachine) response() interface{} {
	return new(VirtualMachine)
}

func (*ResetPasswordForVirtualMachine) name() string {
	return "resetPasswordForVirtualMachine"
}

func (*ResetPasswordForVirtualMachine) description() string {
	return `Resets the password for virtual machine. The virtual machine must be in a "Stopped" state and the template must already support this feature for this command to take effect.`
}

func (*ResetPasswordForVirtualMachine) asyncResponse() interface{} {
	return new(VirtualMachine)
}

func (*GetVirtualMachineUserData) name() string {
	return "getVirtualMachineUserData"
}

func (*GetVirtualMachineUserData) description() string {
	return "Returns user data associated with the VM"
}

func (*GetVirtualMachineUserData) response() interface{} {
	return new(VirtualMachineUserData)
}

func (*MigrateVirtualMachine) name() string {
	return "migrateVirtualMachine"
}

func (*MigrateVirtualMachine) description() string {
	return "Attempts Migration of a VM to a different host or Root volume of the vm to a different storage pool"
}

func (*MigrateVirtualMachine) asyncResponse() interface{} {
	return new(VirtualMachine)
}

func (*GetVMPassword) name() string {
	return "getVMPassword"
}

func (*GetVMPassword) description() string {
	return "Returns an encrypted password for the VM"
}

func (*GetVMPassword) response() interface{} {
	return new(Password)
}

func (*ListVirtualMachines) name() string {
	return "listVirtualMachines"
}

func (*ListVirtualMachines) description() string {
	return "List the virtual machines owned by the account."
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

func (*ListVirtualMachines) each(resp interface{}, callback IterateItemFunc) {
	vms, ok := resp.(*ListVirtualMachinesResponse)
	if !ok {
		callback(nil, fmt.Errorf("wrong type. ListVirtualMachinesResponse expected, got %T", resp))
		return
	}

	for i := range vms.VirtualMachine {
		if !callback(&vms.VirtualMachine[i], nil) {
			break
		}
	}
}

func (*AddNicToVirtualMachine) name() string {
	return "addNicToVirtualMachine"
}

func (*AddNicToVirtualMachine) description() string {
	return "Adds VM to specified network by creating a NIC"
}

func (*AddNicToVirtualMachine) asyncResponse() interface{} {
	return new(VirtualMachine)
}

func (*RemoveNicFromVirtualMachine) name() string {
	return "removeNicFromVirtualMachine"
}

func (*RemoveNicFromVirtualMachine) description() string {
	return "Removes VM from specified network by deleting a NIC"
}

func (*RemoveNicFromVirtualMachine) asyncResponse() interface{} {
	return new(VirtualMachine)
}

func (*UpdateDefaultNicForVirtualMachine) name() string {
	return "updateDefaultNicForVirtualMachine"
}

func (*UpdateDefaultNicForVirtualMachine) description() string {
	return "Changes the default NIC on a VM"
}

func (*UpdateDefaultNicForVirtualMachine) asyncResponse() interface{} {
	return new(VirtualMachine)
}

// Decode decodes the base64 / gzipped encoded user data
func (userdata *VirtualMachineUserData) Decode() (string, error) {
	data, err := base64.StdEncoding.DecodeString(userdata.UserData)
	if err != nil {
		return "", err
	}
	// 0x1f8b is the magic number for gzip
	if len(data) < 2 || data[0] != 0x1f || data[1] != 0x8b {
		return string(data), nil
	}
	gr, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return "", err
	}
	defer gr.Close() // nolint: errcheck

	str, err := ioutil.ReadAll(gr)
	if err != nil {
		return "", err
	}
	return string(str), nil
}
