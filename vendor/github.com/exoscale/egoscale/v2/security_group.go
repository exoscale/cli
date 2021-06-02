package v2

import (
	"context"
	"errors"
	"net"

	apiv2 "github.com/exoscale/egoscale/v2/api"
	papi "github.com/exoscale/egoscale/v2/internal/public-api"
)

// SecurityGroupRule represents a Security Group rule.
type SecurityGroupRule struct {
	Description     string
	EndPort         uint16
	FlowDirection   string
	ICMPCode        int
	ICMPType        int
	ID              string
	Network         *net.IPNet
	Protocol        string
	SecurityGroupID string
	StartPort       uint16
}

func securityGroupRuleFromAPI(r *papi.SecurityGroupRule) *SecurityGroupRule {
	return &SecurityGroupRule{
		Description: papi.OptionalString(r.Description),
		EndPort: func() (v uint16) {
			if r.EndPort != nil {
				v = uint16(*r.EndPort)
			}
			return
		}(),
		FlowDirection: string(*r.FlowDirection),
		ICMPCode: func() (v int) {
			if r.Icmp != nil {
				v = int(*r.Icmp.Code)
			}
			return
		}(),
		ICMPType: func() (v int) {
			if r.Icmp != nil {
				v = int(*r.Icmp.Type)
			}
			return
		}(),
		ID: *r.Id,
		Network: func() (v *net.IPNet) {
			if r.Network != nil {
				_, v, _ = net.ParseCIDR(*r.Network)
			}
			return
		}(),
		Protocol: string(*r.Protocol),
		SecurityGroupID: func() (v string) {
			if r.SecurityGroup != nil {
				v = r.SecurityGroup.Id
			}
			return
		}(),
		StartPort: func() (v uint16) {
			if r.StartPort != nil {
				v = uint16(*r.StartPort)
			}
			return
		}(),
	}
}

// SecurityGroup represents a Security Group.
type SecurityGroup struct {
	Description string
	ID          string
	Name        string
	Rules       []*SecurityGroupRule

	zone string
	c    *Client
}

func securityGroupFromAPI(client *Client, zone string, s *papi.SecurityGroup) *SecurityGroup {
	return &SecurityGroup{
		Description: papi.OptionalString(s.Description),
		ID:          *s.Id,
		Name:        *s.Name,
		Rules: func() (rules []*SecurityGroupRule) {
			if s.Rules != nil {
				rules = make([]*SecurityGroupRule, 0)
				for _, rule := range *s.Rules {
					rule := rule
					rules = append(rules, securityGroupRuleFromAPI(&rule))
				}
			}
			return rules
		}(),

		c:    client,
		zone: zone,
	}
}

func (s SecurityGroup) get(ctx context.Context, client *Client, zone, id string) (interface{}, error) {
	return client.GetSecurityGroup(ctx, zone, id)
}

// AddRule adds a rule to the Security Group.
func (s *SecurityGroup) AddRule(ctx context.Context, rule *SecurityGroupRule) (*SecurityGroupRule, error) {
	var icmp *struct {
		Code *int64 `json:"code,omitempty"`
		Type *int64 `json:"type,omitempty"`
	}

	if rule.Protocol == "icmp" || rule.Protocol == "icmpv6" {
		icmpCode := int64(rule.ICMPCode)
		icmpType := int64(rule.ICMPType)

		icmp = &struct {
			Code *int64 `json:"code,omitempty"`
			Type *int64 `json:"type,omitempty"`
		}{
			Code: &icmpCode,
			Type: &icmpType,
		}
	}

	// The API doesn't return the Security Group rule created directly, so in order to
	// return a *SecurityGroupRule corresponding to the new rule we have to manually
	// compare the list of rules in the SG before and after the rule creation, and
	// identify the rule that wasn't there before.
	// Note: in case of multiple rules creation in parallel this technique is subject
	// to race condition as we could return an unrelated rule. To prevent this, we
	// also compare the protocol/start port/end port parameters of the new rule to the
	// ones specified in the input rule parameter.
	rules := make(map[string]struct{})
	for _, r := range s.Rules {
		rules[r.ID] = struct{}{}
	}

	startPort := int64(rule.StartPort)
	endPort := int64(rule.EndPort)

	resp, err := s.c.AddRuleToSecurityGroupWithResponse(
		apiv2.WithZone(ctx, s.zone),
		s.ID,
		papi.AddRuleToSecurityGroupJSONRequestBody{
			Description: func() *string {
				if rule.Description != "" {
					return &rule.Description
				}
				return nil
			}(),
			EndPort: func() *int64 {
				if endPort > 0 {
					return &endPort
				}
				return nil
			}(),
			FlowDirection: papi.AddRuleToSecurityGroupJSONBodyFlowDirection(rule.FlowDirection),
			Icmp:          icmp,
			Network: func() (v *string) {
				if rule.Network != nil {
					ip := rule.Network.String()
					v = &ip
				}
				return
			}(),
			Protocol: papi.AddRuleToSecurityGroupJSONBodyProtocol(rule.Protocol),
			SecurityGroup: func() (v *papi.SecurityGroupResource) {
				if rule.SecurityGroupID != "" {
					v = &papi.SecurityGroupResource{Id: rule.SecurityGroupID}
				}
				return
			}(),
			StartPort: func() *int64 {
				if startPort > 0 {
					return &startPort
				}
				return nil
			}(),
		})
	if err != nil {
		return nil, err
	}

	res, err := papi.NewPoller().
		WithTimeout(s.c.timeout).
		WithInterval(s.c.pollInterval).
		Poll(ctx, s.c.OperationPoller(s.zone, *resp.JSON200.Id))
	if err != nil {
		return nil, err
	}

	sgUpdated, err := s.c.GetSecurityGroup(ctx, s.zone, *res.(*papi.Reference).Id)
	if err != nil {
		return nil, err
	}

	// Look for an unknown rule: if we find one we hope it's the one we've just created.
	for _, r := range sgUpdated.Rules {
		if _, ok := rules[r.ID]; !ok && (r.Protocol == rule.Protocol &&
			r.StartPort == rule.StartPort &&
			r.EndPort == rule.EndPort) {
			return r, nil
		}
	}

	return nil, errors.New("unable to identify the rule created")
}

// DeleteRule deletes the specified rule from the Security Group.
func (s *SecurityGroup) DeleteRule(ctx context.Context, rule *SecurityGroupRule) error {
	resp, err := s.c.DeleteRuleFromSecurityGroupWithResponse(
		apiv2.WithZone(ctx, s.zone),
		s.ID,
		rule.ID,
	)
	if err != nil {
		return err
	}

	_, err = papi.NewPoller().
		WithTimeout(s.c.timeout).
		WithInterval(s.c.pollInterval).
		Poll(ctx, s.c.OperationPoller(s.zone, *resp.JSON200.Id))
	if err != nil {
		return err
	}

	return nil
}

// CreateSecurityGroup creates a Security Group.
func (c *Client) CreateSecurityGroup(
	ctx context.Context,
	zone string,
	securityGroup *SecurityGroup,
) (*SecurityGroup, error) {
	resp, err := c.CreateSecurityGroupWithResponse(ctx, papi.CreateSecurityGroupJSONRequestBody{
		Description: func() *string {
			if securityGroup.Description != "" {
				return &securityGroup.Description
			}
			return nil
		}(),
		Name: securityGroup.Name,
	})
	if err != nil {
		return nil, err
	}

	res, err := papi.NewPoller().
		WithTimeout(c.timeout).
		WithInterval(c.pollInterval).
		Poll(ctx, c.OperationPoller(zone, *resp.JSON200.Id))
	if err != nil {
		return nil, err
	}

	return c.GetSecurityGroup(ctx, zone, *res.(*papi.Reference).Id)
}

// ListSecurityGroups returns the list of existing Security Groups.
func (c *Client) ListSecurityGroups(ctx context.Context, zone string) ([]*SecurityGroup, error) {
	list := make([]*SecurityGroup, 0)

	resp, err := c.ListSecurityGroupsWithResponse(ctx)
	if err != nil {
		return nil, err
	}

	if resp.JSON200.SecurityGroups != nil {
		for i := range *resp.JSON200.SecurityGroups {
			list = append(list, securityGroupFromAPI(c, zone, &(*resp.JSON200.SecurityGroups)[i]))
		}
	}

	return list, nil
}

// GetSecurityGroup returns the Security Group corresponding to the specified ID in the specified zone.
func (c *Client) GetSecurityGroup(ctx context.Context, zone, id string) (*SecurityGroup, error) {
	resp, err := c.GetSecurityGroupWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}

	return securityGroupFromAPI(c, zone, resp.JSON200), nil
}

// FindSecurityGroup attempts to find a Security Group by name or ID in the specified zone.
func (c *Client) FindSecurityGroup(ctx context.Context, zone, v string) (*SecurityGroup, error) {
	res, err := c.ListSecurityGroups(ctx, zone)
	if err != nil {
		return nil, err
	}

	for _, r := range res {
		if r.ID == v || r.Name == v {
			return c.GetSecurityGroup(ctx, zone, r.ID)
		}
	}

	return nil, apiv2.ErrNotFound
}

// DeleteSecurityGroup deletes the specified Security Group in the specified zone.
func (c *Client) DeleteSecurityGroup(ctx context.Context, zone, id string) error {
	resp, err := c.DeleteSecurityGroupWithResponse(apiv2.WithZone(ctx, zone), id)
	if err != nil {
		return err
	}

	_, err = papi.NewPoller().
		WithTimeout(c.timeout).
		WithInterval(c.pollInterval).
		Poll(ctx, c.OperationPoller(zone, *resp.JSON200.Id))
	if err != nil {
		return err
	}

	return nil
}
