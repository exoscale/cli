package global

import "github.com/exoscale/egoscale/v3/oapi"

// Global provides access to global API resources.
type Global struct {
	oapiClient *oapi.ClientWithResponses
}

// NewGlobal initializes Global with provided oapi Client.
func NewGlobal(c *oapi.ClientWithResponses) *Global {
	return &Global{c}
}

func (a *Global) OrgQuotas() *OrgQuotas {
	return NewOrgQuotas(a.oapiClient)
}

func (a *Global) Operations() *Operation {
	return NewOperation(a.oapiClient)
}
