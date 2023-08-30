package dbaas

import "github.com/exoscale/egoscale/v3/oapi"

// DBaaS provides access to [Exoscale DBaaS] API resources.
//
// [Exoscale DBaaS]: https://community.exoscale.com/documentation/dbaas/
type DBaaS struct {
	oapiClient *oapi.ClientWithResponses
}

// NewDBaaS initializes DBaas with provided oapi Client.
func NewDBaaS(c *oapi.ClientWithResponses) *DBaaS {
	return &DBaaS{c}
}

func (a *DBaaS) Integrations() *Integrations {
	return NewIntegrations(a.oapiClient)
}
