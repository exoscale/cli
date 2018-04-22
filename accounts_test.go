package egoscale

import (
	"testing"
)

func TestAccounts(t *testing.T) {
	var _ Command = (*ListAccounts)(nil)
}

func TestListAccounts(t *testing.T) {
	req := &ListAccounts{}
	if req.name() != "listAccounts" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*ListAccountsResponse)
}
