package egoscale

import (
	"net"
	"net/url"
	"testing"
)

func TestVirtualMachine(t *testing.T) {
	instance := &VirtualMachine{}
	if instance.ResourceType() != "UserVM" {
		t.Errorf("ResourceType doesn't match")
	}
}

func TestDeployVirtualMachine(t *testing.T) {
	req := &DeployVirtualMachine{}
	if req.name() != "deployVirtualMachine" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*DeployVirtualMachineResponse)
}

func TestDestroyVirtualMachine(t *testing.T) {
	req := &DestroyVirtualMachine{}
	if req.name() != "destroyVirtualMachine" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*DestroyVirtualMachineResponse)
}

func TestRebootVirtualMachine(t *testing.T) {
	req := &RebootVirtualMachine{}
	if req.name() != "rebootVirtualMachine" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*RebootVirtualMachineResponse)
}

func TestStartVirtualMachine(t *testing.T) {
	req := &StartVirtualMachine{}
	if req.name() != "startVirtualMachine" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*StartVirtualMachineResponse)
}

func TestStopVirtualMachine(t *testing.T) {
	req := &StopVirtualMachine{}
	if req.name() != "stopVirtualMachine" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*StopVirtualMachineResponse)
}

func TestResetPasswordForVirtualMachine(t *testing.T) {
	req := &ResetPasswordForVirtualMachine{}
	if req.name() != "resetPasswordForVirtualMachine" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*ResetPasswordForVirtualMachineResponse)
}

func TestUpdateVirtualMachine(t *testing.T) {
	req := &UpdateVirtualMachine{}
	if req.name() != "updateVirtualMachine" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*UpdateVirtualMachineResponse)
}

func TestListVirtualMachines(t *testing.T) {
	req := &ListVirtualMachines{}
	if req.name() != "listVirtualMachines" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*ListVirtualMachinesResponse)
}

func TestGetVMPassword(t *testing.T) {
	req := &GetVMPassword{}
	if req.name() != "getVMPassword" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*GetVMPasswordResponse)
}

func TestRestoreVirtualMachine(t *testing.T) {
	req := &RestoreVirtualMachine{}
	if req.name() != "restoreVirtualMachine" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*RestoreVirtualMachineResponse)
}

func TestChangeServiceForVirtualMachine(t *testing.T) {
	req := &ChangeServiceForVirtualMachine{}
	if req.name() != "changeServiceForVirtualMachine" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*ChangeServiceForVirtualMachineResponse)
}

func TestScaleVirtualMachine(t *testing.T) {
	req := &ScaleVirtualMachine{}
	if req.name() != "scaleVirtualMachine" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*booleanAsyncResponse)
}

func TestRecoverVirtualMachine(t *testing.T) {
	req := &RecoverVirtualMachine{}
	if req.name() != "recoverVirtualMachine" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*RecoverVirtualMachineResponse)
}

func TestExpungeVirtualMachine(t *testing.T) {
	req := &ExpungeVirtualMachine{}
	if req.name() != "expungeVirtualMachine" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*booleanAsyncResponse)
}

func TestGetVirtualMachineUserData(t *testing.T) {
	req := &GetVirtualMachineUserData{}
	if req.name() != "getVirtualMachineUserData" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*GetVirtualMachineUserDataResponse)
}

func TestAddNicToVirtualMachine(t *testing.T) {
	req := &AddNicToVirtualMachine{}
	if req.name() != "addNicToVirtualMachine" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*AddNicToVirtualMachineResponse)
}

func TestRemoveNicFromVirtualMachine(t *testing.T) {
	req := &RemoveNicFromVirtualMachine{}
	if req.name() != "removeNicFromVirtualMachine" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*RemoveNicFromVirtualMachineResponse)
}

func TestUpdateDefaultNicForVirtualMachine(t *testing.T) {
	req := &UpdateDefaultNicForVirtualMachine{}
	if req.name() != "updateDefaultNicForVirtualMachine" {
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

func TestGetVirtualMachine(t *testing.T) {
	ts := newServer(response{200, `
{"listvirtualmachinesresponse": {
	"count": 1,
	"virtualmachine": [
		{
			"account": "yoan.blanc@exoscale.ch",
			"affinitygroup": [],
			"cpunumber": 1,
			"cpuspeed": 2198,
			"cpuused": "0%",
			"created": "2018-01-19T14:37:08+0100",
			"diskioread": 0,
			"diskiowrite": 13734,
			"diskkbsread": 0,
			"diskkbswrite": 94342,
			"displayname": "test",
			"displayvm": true,
			"domain": "ROOT",
			"domainid": "1874276d-4cac-448b-aa5e-de00fd4157e8",
			"haenable": false,
			"hostid": "70c12af4-b1cb-4133-97dd-3579bb88a8ce",
			"hostname": "virt-hv-pp005.dk2.p.exoscale.net",
			"hypervisor": "KVM",
			"id": "69069d5e-1591-4214-937e-4c8cba63fcfb",
			"instancename": "i-2-188150-VM",
			"isdynamicallyscalable": false,
			"keypair": "test-yoan",
			"memory": 1024,
			"name": "test",
			"networkkbsread": 5542,
			"networkkbswrite": 8813,
			"nic": [
				{
					"broadcasturi": "vlan://untagged",
					"gateway": "159.100.248.1",
					"id": "75d1367c-319a-4658-b31d-1a26496061ff",
					"ipaddress": "159.100.251.247",
					"isdefault": true,
					"macaddress": "06:6d:cc:00:00:3c",
					"netmask": "255.255.252.0",
					"networkid": "d48bfccc-c11f-438f-8177-9cf6a40dc4f8",
					"networkname": "defaultGuestNetwork",
					"traffictype": "Guest",
					"type": "Shared"
				}
			],
			"oscategoryid": "ca158095-a6d2-4b0c-95e3-9a2e5123cbfc",
			"passwordenabled": true,
			"securitygroup": [
				{
					"account": "default",
					"description": "",
					"id": "41471067-d3c2-41ef-ae45-f57e00078843",
					"name": "default security group",
					"tags": []
				}
			],
			"serviceofferingid": "84925525-7825-418b-845b-1aed179bbc40",
			"serviceofferingname": "Tiny",
			"state": "Running",
			"tags": [],
			"templatedisplaytext": "Linux CentOS 7.4 64-bit 10G Disk (2018-01-08-d617dd)",
			"templateid": "934b4d48-d82e-42f5-8f14-b34de3af9854",
			"templatename": "Linux CentOS 7.4 64-bit",
			"zoneid": "381d0a95-ed3a-4ad9-b41c-b97073c1a433",
			"zonename": "ch-dk-2"
		}
	]
}}`})
	defer ts.Close()

	cs := NewClient(ts.URL, "KEY", "SECRET")
	vm := &VirtualMachine{
		ID: "69069d5e-1591-4214-937e-4c8cba63fcfb",
	}
	if err := cs.Get(vm); err != nil {
		t.Error(err)
	}

	if vm.Account != "yoan.blanc@exoscale.ch" {
		t.Errorf("Account doesn't match, got %v", vm.Account)
	}
}

func TestGetVirtualMachinePassword(t *testing.T) {
	ts := newServer(response{200, `
{"getvmresponse": {
	"password": {
		"encryptedpassword": "test"
	}
}}`})
	defer ts.Close()

	cs := NewClient(ts.URL, "KEY", "SECRET")
	req := &GetVMPassword{
		ID: "test",
	}
	resp, err := cs.Request(req)
	if err != nil {
		t.Error(err)
	}
	if resp.(*GetVMPasswordResponse).Password.EncryptedPassword != "test" {
		t.Errorf("Encrypted password missing")
	}
}

func TestListMachines(t *testing.T) {
	ts := newServer(response{200, `
{"listvirtualmachinesresponse": {
	"count": 3,
	"virtualmachine": [
		{
			"id": "84752707-a1d6-4e93-8207-bafeda83fe15"
		},
		{
			"id": "f93238e1-cc6e-484b-9650-fe8921631b7b"
		},
		{
			"id": "487eda20-eea1-43f7-9456-e870a359b173"
		}
	]
}}`})
	defer ts.Close()

	cs := NewClient(ts.URL, "KEY", "SECRET")
	req := &VirtualMachine{}
	vms, err := cs.List(req)
	if err != nil {
		t.Error(err)
	}

	if len(vms) != 3 {
		t.Errorf("Expected three vms, got %d", len(vms))
	}
}

func TestListMachinesFailure(t *testing.T) {
	ts := newServer(response{200, `
{"listvirtualmachinesresponse": {
	"count": 3,
	"virtualmachine": {}
}}`})
	defer ts.Close()

	cs := NewClient(ts.URL, "KEY", "SECRET")
	req := &VirtualMachine{}
	vms, err := cs.List(req)
	if err == nil {
		t.Errorf("Expected an error got %v", err)
	}

	if len(vms) != 0 {
		t.Errorf("Expected 0 vms, got %d", len(vms))
	}
}

func TestListMachinesPaginate(t *testing.T) {
	ts := newServer(response{200, `
{"listvirtualmachinesresponse": {
	"count": 3,
	"virtualmachine": [
		{
			"id": "84752707-a1d6-4e93-8207-bafeda83fe15"
		},
		{
			"id": "f93238e1-cc6e-484b-9650-fe8921631b7b"
		},
		{
			"id": "487eda20-eea1-43f7-9456-e870a359b173"
		}
	]
}}`})
	defer ts.Close()

	cs := NewClient(ts.URL, "KEY", "SECRET")
	vm := &VirtualMachine{}
	req, err := vm.ListRequest()
	if err != nil {
		t.Error(err)
	}
	cs.Paginate(req, func(i interface{}, err error) bool {
		if i.(*VirtualMachine).ID != "84752707-a1d6-4e93-8207-bafeda83fe15" {
			t.Errorf("Expected id '84752707-a1d6-4e93-8207-bafeda83fe15', got %v", i.(*VirtualMachine).ID)
		}
		return false
	})

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

	ip := vm.IP()
	if ip.String() != "192.168.0.10" {
		t.Errorf("IP Address doesn't match")
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
