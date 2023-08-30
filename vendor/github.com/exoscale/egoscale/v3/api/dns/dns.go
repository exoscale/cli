package dns

import "github.com/exoscale/egoscale/v3/oapi"

// DNS provides access to [Exoscale DNS] API resources.
//
// [Exoscale DNS]: https://community.exoscale.com/documentation/dns/
type DNS struct {
	oapiClient *oapi.ClientWithResponses
}

// NewDNS initializes DNS with provided oapi Client.
func NewDNS(c *oapi.ClientWithResponses) *DNS {
	return &DNS{c}
}

//func (a *DNS) Domains() *Domains {
//return NewDomains(a.oapiClient)
//}
