package globalstate

import (
	egoscale "github.com/exoscale/egoscale/v2"
	v3 "github.com/exoscale/egoscale/v3"
)

var (
	OutputFormat     string
	EgoscaleClient   *egoscale.Client
	EgoscaleV3Client *v3.Client
	Quiet            bool
	ConfigFolder     string
)
