package egoscale

import (
	"testing"
)

func TestCreateUser(t *testing.T) {
	req := &CreateUser{}
	if req.name() != "createUser" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*User)
}

func TestRegisterUserKeys(t *testing.T) {
	req := &RegisterUserKeys{}
	if req.name() != "registerUserKeys" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*User)
}

func TestUpdateUser(t *testing.T) {
	req := &UpdateUser{}
	if req.name() != "updateUser" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*User)
}

func TestListUsers(t *testing.T) {
	req := &ListUsers{}
	if req.name() != "listUsers" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*ListUsersResponse)
}

func TestDeleteUser(t *testing.T) {
	req := &DeleteUser{}
	if req.name() != "deleteUser" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*booleanResponse)
}
