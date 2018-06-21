package egoscale

import (
	"testing"
)

func TestListAccounts(t *testing.T) {
	req := &ListAccounts{}
	_ = req.response().(*ListAccountsResponse)
}

func TestEnableAccount(t *testing.T) {
	req := &EnableAccount{}
	_ = req.response().(*Account)
}

func TestDisableAccount(t *testing.T) {
	req := &DisableAccount{}
	_ = req.response().(*AsyncJobResult)
	_ = req.asyncResponse().(*Account)
}
