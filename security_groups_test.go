package egoscale

import (
	"testing"
)

func TestGroupsRequests(t *testing.T) {
	var _ Taggable = (*SecurityGroup)(nil)
	var _ asyncCommand = (*AuthorizeSecurityGroupEgress)(nil)
	var _ onBeforeHook = (*AuthorizeSecurityGroupEgress)(nil)
	var _ asyncCommand = (*AuthorizeSecurityGroupIngress)(nil)
	var _ onBeforeHook = (*AuthorizeSecurityGroupIngress)(nil)
	var _ syncCommand = (*CreateSecurityGroup)(nil)
	var _ syncCommand = (*DeleteSecurityGroup)(nil)
	var _ syncCommand = (*ListSecurityGroups)(nil)
	var _ asyncCommand = (*RevokeSecurityGroupEgress)(nil)
	var _ asyncCommand = (*RevokeSecurityGroupIngress)(nil)
}

func TestSecurityGroup(t *testing.T) {
	instance := &SecurityGroup{}
	if instance.ResourceType() != "SecurityGroup" {
		t.Errorf("ResourceType doesn't match")
	}
}

func TestAuthorizeSecurityGroupEgress(t *testing.T) {
	req := &AuthorizeSecurityGroupEgress{}
	if req.APIName() != "authorizeSecurityGroupEgress" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*AuthorizeSecurityGroupEgressResponse)
}

func TestAuthorizeSecurityGroupIngress(t *testing.T) {
	req := &AuthorizeSecurityGroupIngress{}
	if req.APIName() != "authorizeSecurityGroupIngress" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*AuthorizeSecurityGroupIngressResponse)
}

func TestCreateSecurityGroup(t *testing.T) {
	req := &CreateSecurityGroup{}
	if req.APIName() != "createSecurityGroup" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*CreateSecurityGroupResponse)
}

func TestDeleteSecurityGroup(t *testing.T) {
	req := &DeleteSecurityGroup{}
	if req.APIName() != "deleteSecurityGroup" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*booleanSyncResponse)
}

func TestListSecurityGroups(t *testing.T) {
	req := &ListSecurityGroups{}
	if req.APIName() != "listSecurityGroups" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*ListSecurityGroupsResponse)
}

func TestRevokeSecurityGroupEgress(t *testing.T) {
	req := &RevokeSecurityGroupEgress{}
	if req.APIName() != "revokeSecurityGroupEgress" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*booleanAsyncResponse)
}

func TestRevokeSecurityGroupIngress(t *testing.T) {
	req := &RevokeSecurityGroupIngress{}
	if req.APIName() != "revokeSecurityGroupIngress" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*booleanAsyncResponse)
}
