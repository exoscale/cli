package globalstate

import (
	egoscale "github.com/exoscale/egoscale/v2"
)

var (
	OutputFormat   string
	EgoscaleClient *egoscale.Client
	Quiet          bool
	ConfigFolder   string
)
