package v2

import (
	"context"
	"errors"
	"net"
	"strings"

	apiv2 "github.com/exoscale/egoscale/v2/api"
	"github.com/exoscale/egoscale/v2/oapi"
)

// SecurityGroupRule represents a Security Group rule.
type SecurityGroupRule struct {
	Description     *string
	EndPort         *uint16
	FlowDirection   *string `req-for:"create"`
	ICMPCode        *int64
	ICMPType        *int64
	ID              *string `req-for:"delete"`
	Network         *net.IPNet
	Protocol        *string `req-for:"create"`
	SecurityGroupID *string
	StartPort       *uint16
}

func securityGroupRuleFromAPI(r *oapi.SecurityGroupRule) *SecurityGroupRule {
	return &SecurityGroupRule{
		Description: r.Description,
		EndPort: func() (v *uint16) {
			if r.EndPort != nil {
				port := uint16(*r.EndPort)
				v = &port
			}
			return
		}(),
		FlowDirection: (*string)(r.FlowDirection),
		ICMPCode: func() (v *int64) {
			if r.Icmp != nil {
				v = r.Icmp.Code
			}
			return
		}(),
		ICMPType: func() (v *int64) {
			if r.Icmp != nil {
				v = r.Icmp.Type
			}
			return
		}(),
		ID: r.Id,
		Network: func() (v *net.IPNet) {
			if r.Network != nil {
				_, v, _ = net.ParseCIDR(*r.Network)
			}
			return
		}(),
		Protocol: (*string)(r.Protocol),
		SecurityGroupID: func() (v *string) {
			if r.SecurityGroup != nil {
				v = &r.SecurityGroup.Id
			}
			return
		}(),
		StartPort: func() (v *uint16) {
			if r.StartPort != nil {
				port := uint16(*r.StartPort)
				v = &port
			}
			return
		}(),
	}
}

// CreateSecurityGroupRule creates a Security Group rule.
func (c *Client) CreateSecurityGroupRule(
	ctx context.Context,
	zone string,
	securityGroup *SecurityGroup,
	rule *SecurityGroupRule,
) (*SecurityGroupRule, error) {
	if err := validateOperationParams(securityGroup, "update"); err != nil {
		return nil, err
	}
	if err := validateOperationParams(rule, "create"); err != nil {
		return nil, err
	}

	var icmp *struct {
		Code *int64 `json:"code,omitempty"`
		Type *int64 `json:"type,omitempty"`
	}

	if strings.HasPrefix(*rule.Protocol, "icmp") {
		icmp = &struct {
			Code *int64 `json:"code,omitempty"`
			Type *int64 `json:"type,omitempty"`
		}{
			Code: rule.ICMPCode,
			Type: rule.ICMPType,
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
	for _, r := range securityGroup.Rules {
		rules[*r.ID] = struct{}{}
	}

	resp, err := c.AddRuleToSecurityGroupWithResponse(
		apiv2.WithZone(ctx, zone),
		*securityGroup.ID,
		oapi.AddRuleToSecurityGroupJSONRequestBody{
			Description: rule.Description,
			EndPort: func() (v *int64) {
				if rule.EndPort != nil {
					port := int64(*rule.EndPort)
					v = &port
				}
				return
			}(),
			FlowDirection: oapi.AddRuleToSecurityGroupJSONBodyFlowDirection(*rule.FlowDirection),
			Icmp:          icmp,
			Network: func() (v *string) {
				if rule.Network != nil {
					ip := rule.Network.String()
					v = &ip
				}
				return
			}(),
			Protocol: oapi.AddRuleToSecurityGroupJSONBodyProtocol(*rule.Protocol),
			SecurityGroup: func() (v *oapi.SecurityGroupResource) {
				if rule.SecurityGroupID != nil {
					v = &oapi.SecurityGroupResource{Id: *rule.SecurityGroupID}
				}
				return
			}(),
			StartPort: func() (v *int64) {
				if rule.StartPort != nil {
					port := int64(*rule.StartPort)
					v = &port
				}
				return
			}(),
		})
	if err != nil {
		return nil, err
	}

	res, err := oapi.NewPoller().
		WithTimeout(c.timeout).
		WithInterval(c.pollInterval).
		Poll(ctx, oapi.OperationPoller(c, zone, *resp.JSON200.Id))
	if err != nil {
		return nil, err
	}

	sgUpdated, err := c.GetSecurityGroup(ctx, zone, *res.(*oapi.Reference).Id)
	if err != nil {
		return nil, err
	}

	// Look for an unknown rule: if we find one we hope it's the one we've just created.
	for _, r := range sgUpdated.Rules {
		if _, ok := rules[*r.ID]; !ok && (*r.FlowDirection == *rule.FlowDirection && *r.Protocol == *rule.Protocol) {
			return r, nil
		}
	}

	return nil, errors.New("unable to identify the rule created")
}

// DeleteSecurityGroupRule deletes a Security Group rule.
func (c *Client) DeleteSecurityGroupRule(
	ctx context.Context,
	zone string,
	securityGroup *SecurityGroup,
	rule *SecurityGroupRule,
) error {
	if err := validateOperationParams(securityGroup, "update"); err != nil {
		return err
	}
	if err := validateOperationParams(rule, "delete"); err != nil {
		return err
	}

	resp, err := c.DeleteRuleFromSecurityGroupWithResponse(apiv2.WithZone(ctx, zone), *securityGroup.ID, *rule.ID)
	if err != nil {
		return err
	}

	_, err = oapi.NewPoller().
		WithTimeout(c.timeout).
		WithInterval(c.pollInterval).
		Poll(ctx, oapi.OperationPoller(c, zone, *resp.JSON200.Id))
	if err != nil {
		return err
	}

	return nil
}
