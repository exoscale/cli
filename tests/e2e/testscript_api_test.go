//go:build api

package e2e_test

import (
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
)

// TestScriptsAPI runs testscript scenarios that require API access.
// These tests only run when built with the 'api' build tag:
//
//	go test -tags=api
//
// API tests require EXOSCALE_API_KEY and EXOSCALE_API_SECRET
// environment variables to be set with valid API credentials.
func TestScriptsAPI(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir: "scenarios/api",
		Setup: func(e *testscript.Env) error {
			return setupTestEnv(e, true)
		},
	})
}
