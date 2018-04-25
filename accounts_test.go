package egoscale

import (
	"testing"
)

func TestAccounts(t *testing.T) {
	var _ Command = (*ListAccounts)(nil)
	var _ Command = (*EnableAccount)(nil)
	var _ Command = (*DisableAccount)(nil)
}

func TestListAccounts(t *testing.T) {
	req := &ListAccounts{}
	if req.name() != "listAccounts" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*ListAccountsResponse)
}

func TestEnableAccount(t *testing.T) {
	req := &EnableAccount{}
	if req.name() != "enableAccount" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*EnableAccountResponse)
}

func TestDisableAccount(t *testing.T) {
	req := &DisableAccount{}
	if req.name() != "disableAccount" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*DisableAccountResponse)
}
