package cmd

import (
	"github.com/exoscale/egoscale"
)

var defaultRules = map[string]*egoscale.IngressRule{
	"ssh": {
		Protocol:    "tcp",
		Cidr:        "0.0.0.0/0",
		StartPort:   22,
		EndPort:     22,
		Description: "",
	},
	"rdp": {
		Protocol:    "tcp",
		Cidr:        "0.0.0.0/0",
		StartPort:   3389,
		EndPort:     3389,
		Description: "",
	},
	"ping": {
		Protocol:    "icmp",
		Cidr:        "0.0.0.0/0",
		IcmpType:    8,
		IcmpCode:    0,
		Description: "",
	},
}
