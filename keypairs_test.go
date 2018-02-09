package egoscale

import (
	"testing"
)

func TestSSHKeyPairs(t *testing.T) {
	var _ asyncCommand = (*ResetSSHKeyForVirtualMachine)(nil)
	var _ syncCommand = (*RegisterSSHKeyPair)(nil)
	var _ syncCommand = (*CreateSSHKeyPair)(nil)
	var _ syncCommand = (*DeleteSSHKeyPair)(nil)
	var _ syncCommand = (*ListSSHKeyPairs)(nil)
}

func TestResetSSHKeyForVirtualMachine(t *testing.T) {
	req := &ResetSSHKeyForVirtualMachine{}
	if req.APIName() != "resetSSHKeyForVirtualMachine" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*ResetSSHKeyForVirtualMachineResponse)
}

func TestRegisterSSHKeyPair(t *testing.T) {
	req := &RegisterSSHKeyPair{}
	if req.APIName() != "registerSSHKeyPair" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*RegisterSSHKeyPairResponse)
}

func TestCreateSSHKeyPair(t *testing.T) {
	req := &CreateSSHKeyPair{}
	if req.APIName() != "createSSHKeyPair" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*CreateSSHKeyPairResponse)
}

func TestDeleteSSHKeyPair(t *testing.T) {
	req := &DeleteSSHKeyPair{}
	if req.APIName() != "deleteSSHKeyPair" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*booleanSyncResponse)
}

func TestListSSHKeyPairs(t *testing.T) {
	req := &ListSSHKeyPairs{}
	if req.APIName() != "listSSHKeyPairs" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*ListSSHKeyPairsResponse)
}
