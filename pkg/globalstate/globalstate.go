package globalstate

import (
	v3 "github.com/exoscale/egoscale/v3"
)

var (
	OutputFormat     string
	EgoscaleV3Client *v3.Client
	Quiet            bool
	ConfigFolder     string
)
