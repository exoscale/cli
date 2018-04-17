package egoscale

import (
	"net/url"
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

func TestListSecurityGroupsApiName(t *testing.T) {
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

func TestAuthorizeSecurityGroupEgressOnBeforeSendICMP(t *testing.T) {
	req := &AuthorizeSecurityGroupEgress{
		Protocol: "ICMP",
	}
	params := url.Values{}

	if err := req.onBeforeSend(&params); err != nil {
		t.Error(err)
	}

	if _, ok := params["icmpcode"]; !ok {
		t.Errorf("icmpcode should have been set")
	}
	if _, ok := params["icmptype"]; !ok {
		t.Errorf("icmptype should have been set")
	}
}

func TestAuthorizeSecurityGroupEgressOnBeforeSendTCP(t *testing.T) {
	req := &AuthorizeSecurityGroupEgress{
		Protocol:  "TCP",
		StartPort: 0,
		EndPort:   1024,
	}
	params := url.Values{}

	if err := req.onBeforeSend(&params); err != nil {
		t.Error(err)
	}

	if _, ok := params["startport"]; !ok {
		t.Errorf("startport should have been set")
	}
}

func TestGetSecurityGroup(t *testing.T) {
	ts := newServer(response{200, `
{"listsecuritygroupsresponse": {
	"count": 1,
	"securitygroup": [
		{
			"account": "yoan.blanc@exoscale.ch",
			"description": "dummy (for test)",
			"domain": "yoan.blanc@exoscale.ch",
			"domainid": "2da0d0d3-e7b2-42ef-805d-eb2ea90ae7ef",
			"egressrule": [],
			"id": "4bfe1073-a6d4-48bd-8f24-2ab586674092",
			"ingressrule": [
				{
					"cidr": "0.0.0.0/0",
					"description": "SSH",
					"endport": 22,
					"protocol": "tcp",
					"ruleid": "fc03b5b1-1d15-4933-99c3-afa0b8f2ab25",
					"startport": 22,
					"tags": []
				}
			],
			"name": "ssh",
			"tags": []
		}
	]
}}`})
	defer ts.Close()

	cs := NewClient(ts.URL, "KEY", "SECRET")
	sg := &SecurityGroup{
		ID: "4bfe1073-a6d4-48bd-8f24-2ab586674092",
	}
	if err := cs.Get(sg); err != nil {
		t.Error(err)
	}

	if sg.Account != "yoan.blanc@exoscale.ch" {
		t.Errorf("Account doesn't match, got %v", sg.Account)
	}
}

func TestGetSecurityGroupMissing(t *testing.T) {
	ts := newServer(response{200, `
{"listsecuritygroupsresponse": {
	"count": 0,
	"securitygroup": []
}}`})
	defer ts.Close()

	cs := NewClient(ts.URL, "KEY", "SECRET")
	sg := &SecurityGroup{
		ID: "4bfe1073-a6d4-48bd-8f24-2ab586674092",
	}
	if err := cs.Get(sg); err == nil {
		t.Errorf("Missing Security Group should have failed")
	}
}

func TestGetSecurityGroupError(t *testing.T) {
	ts := newServer(response{200, `
{"listsecuritygroupsresponse": {
	"cserrorcode": 9999,
	"errorcode": 431,
	"errortext": "Unable to execute API command listsecuritygroups due to invalid value. Invalid parameter id value=4bfe1073-a6d4-48bd-8f24-2ab5866740 due to incorrect long value format, or entity does not exist or due to incorrect parameter annotation for the field in api cmd class.",
	"uuidList": []
}}`})
	defer ts.Close()

	cs := NewClient(ts.URL, "KEY", "SECRET")
	sg := &SecurityGroup{
		ID: "4bfe1073-a6d4-48bd-8f24-2ab5866740",
	}
	if err := cs.Get(sg); err == nil {
		t.Errorf("Missing Security Group should have failed")
	}
}

func TestListSecurityGroups(t *testing.T) {
	ts := newServer(response{200, `
		{"listsecuritygroupsresponse":{
			"count": 2,
			"securitygroup": [
			  {
				"account": "exoscale-1",
				"description": "test",
				"domain": "exoscale-1",
				"domainid": "5b2f621e-3eb6-4a14-a315-d4d7d62f28ff",
				"egressrule": [],
				"id": "55c3b385-0a9b-4970-a5d9-ad1e7f13157d",
				"ingressrule": [
				  {
					"cidr": "0.0.0.0/0",
					"icmpcode": 0,
					"icmptype": 8,
					"protocol": "icmp",
					"ruleid": "1d64f828-9267-4aeb-9cf9-703c3ed99627",
					"tags": []
				  },
				  {
					"cidr": "0.0.0.0/0",
					"endport": 22,
					"protocol": "tcp",
					"ruleid": "1a84c747-1ad5-48ea-a75d-521638c403ea",
					"startport": 22,
					"tags": []
				  },
				  {
					"cidr": "0.0.0.0/0",
					"endport": 3389,
					"protocol": "tcp",
					"ruleid": "f2ab2e27-65a1-40b8-b8c2-9252dc75b5b3",
					"startport": 3389,
					"tags": []
				  }
				],
				"name": "hello",
				"tags": []
			  },
			  {
				"account": "exoscale-1",
				"description": "Default Security Group",
				"domain": "exoscale-1",
				"domainid": "5b2f621e-3eb6-4a14-a315-d4d7d62f28ff",
				"egressrule": [],
				"id": "b1b05d21-11de-4c38-804e-c9bdacdaaa70",
				"ingressrule": [
				  {
					"cidr": "0.0.0.0/0",
					"endport": 22,
					"protocol": "tcp",
					"ruleid": "4aa47b2c-9d1f-4856-9893-286fb8befa76",
					"startport": 22,
					"tags": []
				  },
				  {
					"cidr": "0.0.0.0/0",
					"description": "znc",
					"endport": 1025,
					"protocol": "tcp",
					"ruleid": "45d0feca-f726-4b15-8e02-4de912c6dea7",
					"startport": 1025,
					"tags": []
				  }
				],
				"name": "default",
				"tags": []
			  }
			]
		  }}`})

	defer ts.Close()

	cs := NewClient(ts.URL, "KEY", "SECRET")
	sgs, err := cs.List(&SecurityGroup{})
	if err != nil {
		t.Errorf("%v", err)
	}

	if len(sgs) != 2 {
		t.Errorf("Expected two sg, got %d", len(sgs))
	}

	sg := sgs[0].(*SecurityGroup)

	if sg.ID != "55c3b385-0a9b-4970-a5d9-ad1e7f13157d" {
		t.Errorf("Wrong security group")
	}

}
