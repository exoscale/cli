package globalstate

import (
	"time"

	"github.com/exoscale/cli/pkg/exoapi"
)

var (
	OutputFormat          string
	EgoscaleV3Client      exoapi.Client
	Quiet                 bool
	ConfigFolder          string
	GitVersion, GitCommit string
	RequestTimeout        time.Duration
)
