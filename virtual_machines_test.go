package egoscale

import (
	"net"
	"net/url"
	"testing"
)

func TestVirtualMachines(t *testing.T) {
	var _ Taggable = (*VirtualMachine)(nil)
	var _ AsyncCommand = (*DeployVirtualMachine)(nil)
	var _ AsyncCommand = (*DestroyVirtualMachine)(nil)
	var _ AsyncCommand = (*RebootVirtualMachine)(nil)
	var _ AsyncCommand = (*StartVirtualMachine)(nil)
	var _ AsyncCommand = (*StopVirtualMachine)(nil)
	var _ AsyncCommand = (*ResetPasswordForVirtualMachine)(nil)
	var _ Command = (*UpdateVirtualMachine)(nil)
	var _ Command = (*ListVirtualMachines)(nil)
	var _ Command = (*GetVMPassword)(nil)
	var _ AsyncCommand = (*RestoreVirtualMachine)(nil)
	var _ Command = (*ChangeServiceForVirtualMachine)(nil)
	var _ AsyncCommand = (*ScaleVirtualMachine)(nil)
	var _ Command = (*RecoverVirtualMachine)(nil)
	var _ AsyncCommand = (*ExpungeVirtualMachine)(nil)
	var _ AsyncCommand = (*AddNicToVirtualMachine)(nil)
	var _ AsyncCommand = (*RemoveNicFromVirtualMachine)(nil)
	var _ AsyncCommand = (*UpdateDefaultNicForVirtualMachine)(nil)
}

func TestVirtualMachine(t *testing.T) {
	instance := &VirtualMachine{}
	if instance.ResourceType() != "UserVM" {
		t.Errorf("ResourceType doesn't match")
	}
}

func TestDeployVirtualMachine(t *testing.T) {
	req := &DeployVirtualMachine{}
	if req.APIName() != "deployVirtualMachine" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*DeployVirtualMachineResponse)
}

func TestDestroyVirtualMachine(t *testing.T) {
	req := &DestroyVirtualMachine{}
	if req.APIName() != "destroyVirtualMachine" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*DestroyVirtualMachineResponse)
}

func TestRebootVirtualMachine(t *testing.T) {
	req := &RebootVirtualMachine{}
	if req.APIName() != "rebootVirtualMachine" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*RebootVirtualMachineResponse)
}

func TestStartVirtualMachine(t *testing.T) {
	req := &StartVirtualMachine{}
	if req.APIName() != "startVirtualMachine" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*StartVirtualMachineResponse)
}

func TestStopVirtualMachine(t *testing.T) {
	req := &StopVirtualMachine{}
	if req.APIName() != "stopVirtualMachine" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*StopVirtualMachineResponse)
}

func TestResetPasswordForVirtualMachine(t *testing.T) {
	req := &ResetPasswordForVirtualMachine{}
	if req.APIName() != "resetPasswordForVirtualMachine" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*ResetPasswordForVirtualMachineResponse)
}

func TestUpdateVirtualMachine(t *testing.T) {
	req := &UpdateVirtualMachine{}
	if req.APIName() != "updateVirtualMachine" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*UpdateVirtualMachineResponse)
}

func TestListVirtualMachines(t *testing.T) {
	req := &ListVirtualMachines{}
	if req.APIName() != "listVirtualMachines" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*ListVirtualMachinesResponse)
}

func TestGetVMPassword(t *testing.T) {
	req := &GetVMPassword{}
	if req.APIName() != "getVMPassword" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*GetVMPasswordResponse)
}

func TestRestoreVirtualMachine(t *testing.T) {
	req := &RestoreVirtualMachine{}
	if req.APIName() != "restoreVirtualMachine" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*RestoreVirtualMachineResponse)
}

func TestChangeServiceForVirtualMachine(t *testing.T) {
	req := &ChangeServiceForVirtualMachine{}
	if req.APIName() != "changeServiceForVirtualMachine" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*ChangeServiceForVirtualMachineResponse)
}

func TestScaleVirtualMachine(t *testing.T) {
	req := &ScaleVirtualMachine{}
	if req.APIName() != "scaleVirtualMachine" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*booleanAsyncResponse)
}

func TestRecoverVirtualMachine(t *testing.T) {
	req := &RecoverVirtualMachine{}
	if req.APIName() != "recoverVirtualMachine" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*RecoverVirtualMachineResponse)
}

func TestExpungeVirtualMachine(t *testing.T) {
	req := &ExpungeVirtualMachine{}
	if req.APIName() != "expungeVirtualMachine" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*booleanAsyncResponse)
}

func TestAddNicToVirtualMachine(t *testing.T) {
	req := &AddNicToVirtualMachine{}
	if req.APIName() != "addNicToVirtualMachine" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*AddNicToVirtualMachineResponse)
}

func TestRemoveNicFromVirtualMachine(t *testing.T) {
	req := &RemoveNicFromVirtualMachine{}
	if req.APIName() != "removeNicFromVirtualMachine" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*RemoveNicFromVirtualMachineResponse)
}

func TestUpdateDefaultNicForVirtualMachine(t *testing.T) {
	req := &UpdateDefaultNicForVirtualMachine{}
	if req.APIName() != "updateDefaultNicForVirtualMachine" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*UpdateDefaultNicForVirtualMachineResponse)
}

func TestDeployOnBeforeSend(t *testing.T) {
	req := &DeployVirtualMachine{
		SecurityGroupNames: []string{"default"},
	}
	params := new(url.Values)

	if err := req.onBeforeSend(params); err != nil {
		t.Error(err)
	}
}

func TestDeployOnBeforeSendNoSG(t *testing.T) {
	req := &DeployVirtualMachine{}
	params := new(url.Values)

	// CS will pick the default oiine
	if err := req.onBeforeSend(params); err != nil {
		t.Error(err)
	}
}

func TestDeployOnBeforeSendBothSG(t *testing.T) {
	req := &DeployVirtualMachine{
		SecurityGroupIDs:   []string{"1"},
		SecurityGroupNames: []string{"foo"},
	}
	params := new(url.Values)

	if err := req.onBeforeSend(params); err == nil {
		t.Errorf("DeployVM should only accept SG ids or names")
	}
}

func TestDeployOnBeforeSendBothAG(t *testing.T) {
	req := &DeployVirtualMachine{
		AffinityGroupIDs:   []string{"2"},
		AffinityGroupNames: []string{"foo"},
	}
	params := new(url.Values)

	if err := req.onBeforeSend(params); err == nil {
		t.Errorf("DeployVM should only accept SG ids or names")
	}
}

func TestNicHelpers(t *testing.T) {
	vm := &VirtualMachine{
		Nic: []Nic{
			{
				ID:           "2b50e232-b6d3-491c-92ce-12b24c6123e5",
				IsDefault:    true,
				MacAddress:   "06:aa:14:00:00:18",
				IPAddress:    net.ParseIP("192.168.0.10"),
				Gateway:      net.ParseIP("192.168.0.1"),
				Netmask:      net.ParseIP("255.255.255.0"),
				NetworkID:    "d48bfccc-c11f-438f-8177-9cf6a40dc4d8",
				NetworkName:  "defaultGuestNetwork",
				BroadcastURI: "vlan://untagged",
				TrafficType:  "Guest",
				Type:         "Shared",
			}, {
				BroadcastURI: "vxlan://001",
				ID:           "10b8ffc8-62b3-4b87-82d0-fb7f31bc99b6",
				IsDefault:    false,
				MacAddress:   "0a:7b:5e:00:25:fa",
				NetworkID:    "5f1033fe-2abd-4dda-80b6-c946e21a78ec",
				NetworkName:  "privNetForBasicZone1",
				TrafficType:  "Guest",
				Type:         "Isolated",
			}, {
				BroadcastURI: "vxlan://002",
				ID:           "10b8ffc8-62b3-4b87-82d0-fb7f31bc99b7",
				IsDefault:    false,
				MacAddress:   "0a:7b:5e:00:25:ff",
				NetworkID:    "5f1033fe-2abd-4dda-80b6-c946e21a72ec",
				NetworkName:  "privNetForBasicZone2",
				TrafficType:  "Guest",
				Type:         "Isolated",
			},
		},
	}

	nic := vm.DefaultNic()
	if nic.IPAddress.String() != "192.168.0.10" {
		t.Errorf("Default NIC doesn't match")
	}

	nic1 := vm.NicByID("2b50e232-b6d3-491c-92ce-12b24c6123e5")
	if nic.ID != nic1.ID {
		t.Errorf("NicByID does not match %#v %#v", nic, nic1)
	}

	if len(vm.NicsByType("Isolated")) != 2 {
		t.Errorf("Isolated nics count does not match")
	}

	if len(vm.NicsByType("Shared")) != 1 {
		t.Errorf("Shared nics count does not match")
	}

	if len(vm.NicsByType("Dummy")) != 0 {
		t.Errorf("Dummy nics count does not match")
	}

	if vm.NicByNetworkID("5f1033fe-2abd-4dda-80b6-c946e21a78ec") == nil {
		t.Errorf("NetworkID nic wasn't found")
	}

	if vm.NicByNetworkID("5f1033fe-2abd-4dda-80b6-c946e21a78ed") != nil {
		t.Errorf("NetworkID nic was found??")
	}
}

func TestNicNoDefault(t *testing.T) {
	vm := &VirtualMachine{
		Nic: []Nic{},
	}

	// code coverage...
	nic := vm.DefaultNic()
	if nic != nil {
		t.Errorf("Default NIC wasn't nil?")
	}
}
