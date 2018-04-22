package egoscale

import (
	"testing"
)

func TestUsers(t *testing.T) {
	var _ Command = (*CreateUser)(nil)
	var _ Command = (*RegisterUserKeys)(nil)
	var _ Command = (*UpdateUser)(nil)
	var _ Command = (*ListUsers)(nil)
}

func TestCreateUser(t *testing.T) {
	req := &CreateUser{}
	if req.name() != "createUser" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*CreateUserResponse)
}

func TestRegisterUserKeys(t *testing.T) {
	req := &RegisterUserKeys{}
	if req.name() != "registerUserKeys" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*RegisterUserKeysResponse)
}

func TestUpdateUser(t *testing.T) {
	req := &UpdateUser{}
	if req.name() != "updateUser" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*UpdateUserResponse)
}

func TestListUsers(t *testing.T) {
	req := &ListUsers{}
	if req.name() != "listUsers" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*ListUsersResponse)
}
