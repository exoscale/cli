package acctests

import (
	"os"

	"github.com/stretchr/testify/suite"
)

const (
	AccTestGuardEnvVar = "EXO_CLI_ACC"
)

type AcceptanceTestSuite struct {
	suite.Suite

	ExoCLIExecutable string
}

func (s *AcceptanceTestSuite) SetupSuite() {
	if os.Getenv(AccTestGuardEnvVar) == "" {
		s.T().Logf("Skipping acceptance test; Set the environment variable %s to run the acceptance tests.", AccTestGuardEnvVar)
		s.T().SkipNow()
	}
}
