package iam

import "github.com/exoscale/egoscale/v3/oapi"

// IAM provides access to [Exoscale IAM] API resources.
//
// [Exoscale IAM]: https://community.exoscale.com/documentation/iam/
type IAM struct {
	oapiClient *oapi.ClientWithResponses
}

// NewIAM initializes IAM with provided oapi Client.
func NewIAM(c *oapi.ClientWithResponses) *IAM {
	return &IAM{c}
}

func (a *IAM) Roles() *Roles {
	return NewRoles(a.oapiClient)
}

func (a *IAM) OrgPolicy() *OrgPolicy {
	return NewOrgPolicy(a.oapiClient)
}

func (a *IAM) AccessKey() *AccessKey {
	return NewAccessKey(a.oapiClient)
}
