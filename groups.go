package egoscale

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// CreateEgressRule is an alias of AuthorizeSecurityGroupEgress
func (exo *Client) CreateEgressRule(rule SecurityGroupRule, async AsyncInfo) (*SecurityGroupRule, error) {
	return exo.AuthorizeSecurityGroupEgress(rule, async)
}

// AuthorizeSecurityGroupEgress authorizes a particular egress rule for this security group
func (exo *Client) AuthorizeSecurityGroupEgress(rule SecurityGroupRule, async AsyncInfo) (*SecurityGroupRule, error) {
	return exo.doSecurityGroupRule("authorize", "Egress", rule, async)
}

// CreateIngressRule is an alias of AuthorizeSecurityGroupIngress
func (exo *Client) CreateIngressRule(rule SecurityGroupRule, async AsyncInfo) (*SecurityGroupRule, error) {
	return exo.AuthorizeSecurityGroupIngress(rule, async)
}

// AuthorizeSecurityGroupIngress authorizes a particular ingress rule for this security group
func (exo *Client) AuthorizeSecurityGroupIngress(rule SecurityGroupRule, async AsyncInfo) (*SecurityGroupRule, error) {
	return exo.doSecurityGroupRule("authorize", "Ingress", rule, async)
}

func (exo *Client) doSecurityGroupRule(action string, kind string, rule SecurityGroupRule, async AsyncInfo) (*SecurityGroupRule, error) {
	params := url.Values{}
	params.Set("securitygroupid", rule.SecurityGroupId)

	if rule.Cidr != "" {
		params.Set("cidrlist", rule.Cidr)
	} else if len(rule.UserSecurityGroupList) > 0 {
		for i, usg := range rule.UserSecurityGroupList {
			key := fmt.Sprintf("usersecuritygrouplist[%d]", i)
			params.Set(key+".account", usg.Account)
			params.Set(key+".group", usg.Group)
		}
	} else {
		return nil, fmt.Errorf("No CIDR or Security Group List provided")
	}

	params.Set("protocol", rule.Protocol)

	if rule.Protocol == "ICMP" {
		params.Set("icmpcode", fmt.Sprintf("%d", rule.IcmpCode))
		params.Set("icmptype", fmt.Sprintf("%d", rule.IcmpType))
	} else if rule.Protocol == "TCP" || rule.Protocol == "UDP" {
		params.Set("startport", fmt.Sprintf("%d", rule.StartPort))
		params.Set("endport", fmt.Sprintf("%d", rule.EndPort))
	} else {
		return nil, fmt.Errorf("Invalid rule Protocol: %s", rule.Protocol)
	}

	resp, err := exo.AsyncRequest(action+"SecurityGroup"+kind, params, async)
	if err != nil {
		return nil, err
	}

	var r SecurityGroupRuleResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return r.SecurityGroupRule, nil
}

func (exo *Client) CreateSecurityGroupWithRules(name string, ingress []SecurityGroupRule, egress []SecurityGroupRule, async AsyncInfo) (*CreateSecurityGroupResponse, error) {

	params := url.Values{}
	params.Set("name", name)

	resp, err := exo.Request("createSecurityGroup", params)

	var r CreateSecurityGroupResponseWrapper
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	sgid := r.Wrapped.Id

	for _, erule := range egress {
		erule.SecurityGroupId = sgid
		_, err = exo.CreateEgressRule(erule, async)
		if err != nil {
			return nil, err
		}
	}

	for _, inrule := range ingress {
		inrule.SecurityGroupId = sgid
		_, err = exo.CreateIngressRule(inrule, async)
		if err != nil {
			return nil, err
		}
	}

	return &r.Wrapped, nil
}

func (exo *Client) DeleteSecurityGroup(name string) error {
	params := url.Values{}
	params.Set("name", name)

	resp, err := exo.Request("deleteSecurityGroup", params)
	if err != nil {
		return err
	}

	fmt.Printf("## response: %+v\n", resp)
	return nil
}
