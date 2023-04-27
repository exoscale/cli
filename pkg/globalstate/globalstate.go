package globalstate

import (
	"github.com/exoscale/egoscale"
)

var (
	OutputFormat   string
	EgoscaleClient *egoscale.Client
	Quiet          bool
)
