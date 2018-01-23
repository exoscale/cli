package egoscale

import (
	"testing"
)

func TestUsers(t *testing.T) {
	var _ Command = (*RegisterUserKeys)(nil)
}

func TestRegisterUserKeys(t *testing.T) {
	req := &RegisterUserKeys{}
	if req.name() != "registerUserKeys" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*RegisterUserKeysResponse)
}
